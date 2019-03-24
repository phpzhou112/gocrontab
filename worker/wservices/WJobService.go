package wservices

import (
	"context"
	"go.etcd.io/etcd/clientv3"
	"go.etcd.io/etcd/mvcc/mvccpb"
	"gocrontab/common/constants"
	"gocrontab/common/etcdclient"
	"gocrontab/common/tools"

)

// 任务管理器
type JobMgr struct {
	Client *clientv3.Client
	Kv clientv3.KV
	Lease clientv3.Lease
	Watcher clientv3.Watcher
}

// 监听任务变化
func (jobMgr *JobMgr) WatchJobs() (err error) {
	var (
		getResp *clientv3.GetResponse
		kvpair *mvccpb.KeyValue
		job *tools.Job
		watchStartRevision int64
		watchChan clientv3.WatchChan
		watchResp clientv3.WatchResponse
		watchEvent *clientv3.Event
		jobName string
		jobEvent *tools.JobEvent
	)

	// 1, get一下/cron/jobs/目录下的所有任务，并且获知当前集群的revision
	if getResp, err = jobMgr.Kv.Get(context.TODO(), constants.JOB_SAVE_DIR, clientv3.WithPrefix()); err != nil {
		return
	}

	// 当前有哪些任务
	for _, kvpair = range getResp.Kvs {
		// 反序列化json得到Job
		if job, err = tools.UnpackJob(kvpair.Value); err == nil {
			jobEvent = tools.BuildJobEvent(constants.JOB_EVENT_SAVE, job)
			// 同步给scheduler(调度协程)
			etcdclient.GScheduler.PushJobEvent(jobEvent)
		}
	}

	// 2, 从该revision向后监听变化事件
	go func() { // 监听协程
		// 从GET时刻的后续版本开始监听变化
		watchStartRevision = getResp.Header.Revision + 1
		// 监听/cron/jobs/目录的后续变化
		watchChan = jobMgr.Watcher.Watch(context.TODO(), constants.JOB_SAVE_DIR, clientv3.WithRev(watchStartRevision), clientv3.WithPrefix())
		// 处理监听事件
		for watchResp = range watchChan {
			for _, watchEvent = range watchResp.Events {
				switch watchEvent.Type {
				case mvccpb.PUT: // 任务保存事件
					if job, err = tools.UnpackJob(watchEvent.Kv.Value); err != nil {
						continue
					}
					// 构建一个更新Event
					jobEvent = tools.BuildJobEvent(constants.JOB_EVENT_SAVE, job)
				case mvccpb.DELETE: // 任务被删除了
					// Delete /cron/jobs/job10
					jobName = tools.ExtractJobName(string(watchEvent.Kv.Key))

					job = &tools.Job{Name: jobName}

					// 构建一个删除Event
					jobEvent = tools.BuildJobEvent(constants.JOB_EVENT_DELETE, job)
				}
				// 变化推给scheduler
				etcdclient.GScheduler.PushJobEvent(jobEvent)
			}
		}
	}()
	return
}

// 监听强杀任务通知
func (jobMgr *JobMgr) WatchKiller() {
	var (
		watchChan clientv3.WatchChan
		watchResp clientv3.WatchResponse
		watchEvent *clientv3.Event
		jobEvent *tools.JobEvent
		jobName string
		job *tools.Job
	)
	// 监听/cron/killer目录
	go func() { // 监听协程
		// 监听/cron/killer/目录的变化
		watchChan = jobMgr.Watcher.Watch(context.TODO(), constants.JOB_KILLER_DIR, clientv3.WithPrefix())
		// 处理监听事件
		for watchResp = range watchChan {
			for _, watchEvent = range watchResp.Events {
				switch watchEvent.Type {
				case mvccpb.PUT: // 杀死任务事件
					jobName = tools.ExtractKillerName(string(watchEvent.Kv.Key))
					job = &tools.Job{Name: jobName}
					jobEvent = tools.BuildJobEvent(constants.JOB_EVENT_KILL, job)
					// 事件推给scheduler
					etcdclient.GScheduler.PushJobEvent(jobEvent)
				case mvccpb.DELETE: // killer标记过期, 被自动删除
				}
			}
		}
	}()
}



// 创建任务执行锁
func (jobMgr *JobMgr) CreateJobLock(jobName string) (jobLock *tools.JobLock){
	jobLock = tools.InitJobLock(jobName, jobMgr.Kv, jobMgr.Lease)
	return
}
