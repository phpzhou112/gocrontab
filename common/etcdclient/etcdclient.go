package etcdclient

import (
	"errors"
	log "github.com/alecthomas/log4go"
	"go.etcd.io/etcd/clientv3"
	"gocrontab/common/tools"
	"gocrontab/manager/services"
	services2 "gocrontab/worker/wservices"
	"time"
)

var (
	GWorkerMgr *services.WorkerMgr
	GJobMgr    *services.JobMgr
	GClient    *clientv3.Client //全局连接
	GRegister *services2.Register
	GWJobMgr   *services2.JobMgr
	GExecutor *services2.Executor
	GScheduler *services2.Scheduler
)

type EtcdClient struct {
}

//初始化连接etcd
func InitEtcd() (client *clientv3.Client, err error) {
	var (
		config    clientv3.Config
		endPoints []string
	)
	if tools.Get("manager", "etcdEndpoints", true).IsArray() {
		for _, v := range tools.Get("manager", "etcdEndpoints", true).Array() {
			endPoints = append(endPoints, v.Str)
		}
	}

	// 初始化配置
	config = clientv3.Config{
		Endpoints:   endPoints,                                                                             // 集群地址
		DialTimeout: time.Duration(tools.Get("manager", "etcdDialTimeout", true).Int()) * time.Millisecond, // 连接超时
	}

	// 建立连接
	if client, err = clientv3.New(config); err != nil {
		log.Error("连接etcd失败", err)
		return
	}
	GClient = client
	log.Info("连接etcd成功", endPoints)
	return
}

//初始化连接
func InitWorkerMgr() (err error) {
	var (
		client *clientv3.Client
		kv     clientv3.KV
		lease  clientv3.Lease
	)
	if GClient == nil {
		if client, err = InitEtcd(); err != nil {
			return
		}
		GClient = client
	}

	// 得到KV和Lease的API子集
	kv = clientv3.NewKV(client)
	lease = clientv3.NewLease(client)

	GWorkerMgr = &services.WorkerMgr{
		Client: client,
		Kv:     kv,
		Lease:  lease,
	}
	return
}

// 初始化管理器
func InitJobMgr() (err error) {
	var (
		client *clientv3.Client
		kv     clientv3.KV
		lease  clientv3.Lease
	)

	if GClient == nil {
		return errors.New("连接etcd失败,初始化管理器")
	}
	// 得到KV和Lease的API子集
	kv = clientv3.NewKV(GClient)
	lease = clientv3.NewLease(GClient)

	// 赋值单例
	GJobMgr = &services.JobMgr{
		Client: client,
		Kv:     kv,
		Lease:  lease,
	}
	return
}
//初始化注册服务
func InitRegister() (err error) {
	var (
		client *clientv3.Client
		kv clientv3.KV
		lease clientv3.Lease
		localIp string
	)

	// 初始化配置
	if GClient == nil {
		if client, err = InitEtcd(); err != nil {
			return
		}
		GClient = client
	}

	// 本机IP
	if localIp, err = tools.GetLocalIP(); err != nil {
		return
	}

	// 得到KV和Lease的API子集
	kv = clientv3.NewKV(client)
	lease = clientv3.NewLease(client)

	GRegister = &services2.Register{
		Client: client,
		Kv: kv,
		Lease: lease,
		LocalIP: localIp,
	}

	// 服务注册
	go GRegister.KeepOnline()

	return
}


// 初始化worker管理器
func InitWJobMgr() (err error) {
	var (
		client *clientv3.Client
		kv clientv3.KV
		lease clientv3.Lease
		watcher clientv3.Watcher
	)

	// 初始化配置
	if GClient == nil {
		if client, err = InitEtcd(); err != nil {
			return
		}
		GClient = client
	}

	// 得到KV和Lease的API子集
	kv = clientv3.NewKV(client)
	lease = clientv3.NewLease(client)
	watcher = clientv3.NewWatcher(client)

	// 赋值单例
	GWJobMgr = &services2.JobMgr{
		Client: client,
		Kv: kv,
		Lease: lease,
		Watcher: watcher,
	}

	// 启动任务监听
	GWJobMgr.WatchJobs()

	// 启动监听killer
	GWJobMgr.WatchKiller()

	return
}

//  初始化执行器
func InitExecutor() (err error) {
	GExecutor = &services2.Executor{}
	return
}

