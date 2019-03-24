package main

import (
	log "github.com/alecthomas/log4go"
	"gocrontab/common/etcdclient"
	"gocrontab/common/tools"
	"gocrontab/worker/wservices"
	"math"
	"runtime"
	"time"
)

//worker执行模块
func main() {
	var NumCPU = runtime.NumCPU()
	runtime.GOMAXPROCS(int(math.Max(float64(NumCPU-1), 1)))
	_init()
	for {
		time.Sleep(1 * time.Second)
	}
}

func _init() {
	_checkConfig()
	var (
		err error
	)
	//启动etcd连接
	if err = etcdclient.InitRegister(); err != nil {
		log.Error("初始化注册失败 ", err)
		return
	}

	if err = etcdclient.InitExecutor(); err != nil {
		log.Error("初始化执行器失败 ", err)
		return
	}

	//// 启动调度器
	if err = wservices.InitScheduler(); err != nil {
		log.Error("初始化调度器器失败 ", err)
		return
	}
	// 初始化任务管理器
	if err = etcdclient.InitWJobMgr(); err != nil {
		log.Error("初始化调度器器失败 ", err)
		return
	}
	//加载log4j配置文件
	log.LoadConfiguration("common/log.xml")
	log.Info("init  success")
}

/*检查配置文件格式*/
func _checkConfig() {
	filename, _ := tools.WalkDir("worker/config", "")
	for _, s := range filename {
		tools.Valid(s)
	}
	log.Info("init check config success")
}
