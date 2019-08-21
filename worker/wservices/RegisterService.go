package wservices

import (
	"context"
	"go.etcd.io/etcd/clientv3"
	"gocrontab/common/constants"
	"gocrontab/common/etcdclient"
	"gocrontab/common/tools"
	"time"
)

var GRegister *Register

// 注册节点到etcd： /cron/workers/IP地址
type Register struct {
	Client  *clientv3.Client
	Kv      clientv3.KV
	Lease   clientv3.Lease
	LocalIP string // 本机IP
}

//初始化注册服务
func InitRegister() (err error) {
	var (
		client  *clientv3.Client
		kv      clientv3.KV
		lease   clientv3.Lease
		localIp string
	)

	// 初始化配置
	if etcdclient.GClient == nil {
		if client, err = etcdclient.InitEtcd(); err != nil {
			return
		}
		etcdclient.GClient = client
	}

	// 本机IP
	if localIp, err = tools.GetLocalIP(); err != nil {
		return
	}

	// 得到KV和Lease的API子集
	kv = clientv3.NewKV(client)
	lease = clientv3.NewLease(client)

	GRegister = &Register{
		Client:  client,
		Kv:      kv,
		Lease:   lease,
		LocalIP: localIp,
	}

	// 服务注册
	go GRegister.KeepOnline()

	return
}

// 注册到/cron/workers/IP, 并自动续租
func (register *Register) KeepOnline() {
	var (
		regKey         string
		leaseGrantResp *clientv3.LeaseGrantResponse
		err            error
		keepAliveChan  <-chan *clientv3.LeaseKeepAliveResponse
		keepAliveResp  *clientv3.LeaseKeepAliveResponse
		cancelCtx      context.Context
		cancelFunc     context.CancelFunc
	)

	for {
		// 注册路径
		regKey = constants.JOB_WORKER_DIR + register.LocalIP

		cancelFunc = nil

		// 创建租约
		if leaseGrantResp, err = register.Lease.Grant(context.TODO(), 10); err != nil {
			goto RETRY
		}

		// 自动续租
		if keepAliveChan, err = register.Lease.KeepAlive(context.TODO(), leaseGrantResp.ID); err != nil {
			goto RETRY
		}

		cancelCtx, cancelFunc = context.WithCancel(context.TODO())

		// 注册到etcd
		if _, err = register.Kv.Put(cancelCtx, regKey, "", clientv3.WithLease(leaseGrantResp.ID)); err != nil {
			goto RETRY
		}

		// 处理续租应答
		for {
			select {
			case keepAliveResp = <-keepAliveChan:
				if keepAliveResp == nil { // 续租失败
					goto RETRY
				}
			}
		}

	RETRY:
		time.Sleep(1 * time.Second)
		if cancelFunc != nil {
			cancelFunc()
		}
	}
}
