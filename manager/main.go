package main

import (
	log "github.com/alecthomas/log4go"
	"gocrontab/common/etcdclient"
	"gocrontab/common/tools"
	"gocrontab/manager/router"
	"math"
	"runtime"
)

//管理者主函数
func main() {

	var NumCPU = runtime.NumCPU()
	runtime.GOMAXPROCS(int(math.Max(float64(NumCPU-1), 1)))

	//初始化连接
	_init()
	//启动http server
	var port = tools.Get("manager", "webPort", true).String()
	if port == "" {
		port = "12828"
	}
	routes.Bind(port)

}

//初始化函数
func _init() {
	_checkConfig()
	var (
		err error
	)
	//启动etcd连接
	if err = etcdclient.InitWorkerMgr(); err != nil {
		log.Error("初始化服务模块失败 ", err)
		return
	}
	//任务管理器
	if err = etcdclient.InitJobMgr(); err != nil {
		log.Error("初始化任务管理器失败 ", err)
		return
	}
	//加载log4j配置文件
	log.LoadConfiguration("common/log.xml")
	log.Info("init  success")
}

/*检查配置文件格式*/
func _checkConfig() {
	filename, _ := tools.WalkDir("manager/config", "")
	for _, s := range filename {
		tools.Valid(s)
	}
	log.Info("init check config success")
}

//关闭一些连接
func _deferer() {

}
