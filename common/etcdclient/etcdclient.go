package etcdclient

import (
	log "github.com/alecthomas/log4go"
	"go.etcd.io/etcd/clientv3"
	"gocrontab/common/tools"
	"time"
)

var (
	GClient *clientv3.Client //全局连接
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
