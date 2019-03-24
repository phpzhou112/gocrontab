package services

import (
	"context"
	log "github.com/alecthomas/log4go"
	"go.etcd.io/etcd/clientv3"
	"go.etcd.io/etcd/mvcc/mvccpb"
	"gocrontab/common/constants"
	"gocrontab/common/tools"
)

// /cron/workers/
type WorkerMgr struct {
	Client *clientv3.Client
	Kv     clientv3.KV
	Lease  clientv3.Lease
}

// 获取在线worker列表
func (workerMgr *WorkerMgr) ListWorkers() (workerArr []string, err error) {
	var (
		getResp  *clientv3.GetResponse
		kv       *mvccpb.KeyValue
		workerIP string
	)

	// 初始化数组
	workerArr = make([]string, 0)

	// 获取目录下所有Kv
	if getResp, err = workerMgr.Kv.Get(context.TODO(), constants.JOB_WORKER_DIR, clientv3.WithPrefix()); err != nil {
		return
	}

	// 解析每个节点的IP
	for _, kv = range getResp.Kvs {
		// kv.Key : /cron/workers/192.168.2.1
		workerIP = tools.ExtractWorkerIP(string(kv.Key))
		workerArr = append(workerArr, workerIP)
	}
	//打印worker
	log.Info("worker节点:", workerArr)
	return
}
