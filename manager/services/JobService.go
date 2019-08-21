package services

import (
	"context"
	"encoding/json"
	"errors"
	"go.etcd.io/etcd/clientv3"
	"go.etcd.io/etcd/mvcc/mvccpb"
	"gocrontab/common/constants"
	"gocrontab/common/etcdclient"
	"gocrontab/common/tools"
)

// 任务管理器
type JobMgr struct {
	Client *clientv3.Client
	Kv     clientv3.KV
	Lease  clientv3.Lease
}

var GJobMgr *JobMgr

func InitJobMgr() (err error) {
	var (
		client *clientv3.Client
		kv     clientv3.KV
		lease  clientv3.Lease
	)

	if etcdclient.GClient == nil {
		return errors.New("连接etcd失败,初始化管理器")
	}
	// 得到KV和Lease的API子集
	kv = clientv3.NewKV(etcdclient.GClient)
	lease = clientv3.NewLease(etcdclient.GClient)

	// 赋值单例
	GJobMgr = &JobMgr{
		Client: client,
		Kv:     kv,
		Lease:  lease,
	}
	return
}

// 保存任务
func (jobMgr *JobMgr) SaveJob(job *tools.Job) (oldJob *tools.Job, err error) {
	// 把任务保存到/cron/jobs/任务名 -> json
	var (
		jobKey    string
		jobValue  []byte
		putResp   *clientv3.PutResponse
		oldJobObj tools.Job
	)

	// etcd的保存key
	jobKey = constants.JOB_SAVE_DIR + job.Name
	// 任务信息json
	if jobValue, err = json.Marshal(job); err != nil {
		return
	}
	// 保存到etcd
	if putResp, err = jobMgr.Kv.Put(context.TODO(), jobKey, string(jobValue), clientv3.WithPrevKV()); err != nil {
		return
	}
	// 如果是更新, 那么返回旧值
	if putResp.PrevKv != nil {
		// 对旧值做一个反序列化
		if err = json.Unmarshal(putResp.PrevKv.Value, &oldJobObj); err != nil {
			err = nil
			return
		}
		oldJob = &oldJobObj
	}
	return
}

// 删除任务
func (jobMgr *JobMgr) DeleteJob(name string) (oldJob *tools.Job, err error) {
	var (
		jobKey    string
		delResp   *clientv3.DeleteResponse
		oldJobObj tools.Job
	)

	// etcd中保存任务的key
	jobKey = constants.JOB_SAVE_DIR + name

	// 从etcd中删除它
	if delResp, err = jobMgr.Kv.Delete(context.TODO(), jobKey, clientv3.WithPrevKV()); err != nil {
		return
	}

	// 返回被删除的任务信息
	if len(delResp.PrevKvs) != 0 {
		// 解析一下旧值, 返回它
		if err = json.Unmarshal(delResp.PrevKvs[0].Value, &oldJobObj); err != nil {
			err = nil
			return
		}
		oldJob = &oldJobObj
	}
	return
}

// 列举任务
func (jobMgr *JobMgr) ListJobs() (jobList []*tools.Job, err error) {
	var (
		dirKey  string
		getResp *clientv3.GetResponse
		kvPair  *mvccpb.KeyValue
		job     *tools.Job
	)

	// 任务保存的目录
	dirKey = constants.JOB_SAVE_DIR

	// 获取目录下所有任务信息
	if getResp, err = jobMgr.Kv.Get(context.TODO(), dirKey, clientv3.WithPrefix()); err != nil {
		return
	}

	// 初始化数组空间
	jobList = make([]*tools.Job, 0)

	// 遍历所有任务, 进行反序列化
	for _, kvPair = range getResp.Kvs {
		job = &tools.Job{}
		if err = json.Unmarshal(kvPair.Value, job); err != nil {
			err = nil
			continue
		}
		jobList = append(jobList, job)
	}

	return
}

// 杀死任务
func (jobMgr *JobMgr) KillJob(name string) (err error) {
	// 更新一下key=/cron/killer/任务名
	var (
		killerKey      string
		leaseGrantResp *clientv3.LeaseGrantResponse
		leaseId        clientv3.LeaseID
	)

	// 通知worker杀死对应任务
	killerKey = constants.JOB_KILLER_DIR + name

	// 让worker监听到一次put操作, 创建一个租约让其稍后自动过期即可
	if leaseGrantResp, err = jobMgr.Lease.Grant(context.TODO(), 1); err != nil {
		return
	}

	// 租约ID
	leaseId = leaseGrantResp.ID

	// 设置killer标记
	if _, err = jobMgr.Kv.Put(context.TODO(), killerKey, "", clientv3.WithLease(leaseId)); err != nil {
		return
	}
	return
}
