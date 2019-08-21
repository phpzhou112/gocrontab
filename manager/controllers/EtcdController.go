package controllers

import (
	"encoding/json"
	log "github.com/alecthomas/log4go"
	"github.com/gin-gonic/gin"
	"gocrontab/common/constants"
	"gocrontab/common/tools"
	"gocrontab/manager/services"
	"net/http"
)

type EtcdController struct {
	BaseController
}

//列出所有任务
func (this *EtcdController) JobList(c *gin.Context) {
	defer tools.Catch(c)

	var (
		jobList []*tools.Job
		err     error
	)

	// 获取任务列表

	if jobList, err = services.GJobMgr.ListJobs(); err != nil {
		tools.BuildResponse(c, constants.CODE_FAILED, "获取列表失败", "")
		return
	}

	//成功返回
	tools.BuildResponse(c, constants.CODE_SUCCESS, constants.MESSAGE_SUCCESS, jobList)
	return
}

//添加
func (this *EtcdController) JobAdd(c *gin.Context) {
	//加入异常捕捉
	defer tools.Catch(c)
	var (
		err    error
		data   []byte
		job    tools.Job
		oldJob *tools.Job
	)

	if data, err = this.Analysis(c); err != nil {
		log.Error("解析请求json失败", err)
		tools.BuildResponse(c, 1, "请求参数不正确", "")
		return
	}
	if err = json.Unmarshal(data, &job); err != nil {
		log.Error("解析json失败", err)
		return
	}
	// 4, 保存到etcd
	if oldJob, err = services.GJobMgr.SaveJob(&job); err != nil {
		log.Error("保存到etcd失败", err)
		return
	}
	//返回数据
	tools.BuildResponse(c, constants.CODE_SUCCESS, constants.MESSAGE_SUCCESS, oldJob)
	log.Info("保存成功", oldJob)
	return
}

//更新
func (this *EtcdController) update(c *gin.Context) {

}

//删除
func (this *EtcdController) JobDelete(c *gin.Context) {
	defer tools.Catch(c)
	var (
		err  error
		job  tools.Job
		data []byte
	)

	if data, err = this.Analysis(c); err != nil {
		log.Error("解析请求失败", err)
		tools.BuildResponse(c, constants.CODE_FAILED, "请求参数不正确", "")
		return
	}

	if err = json.Unmarshal(data, &job); err != nil {
		log.Error("解析json失败", err)
		return
	}
	//去删除任务
	if _, err = services.GJobMgr.DeleteJob(job.Name); err != nil {
		tools.BuildResponse(c, constants.CODE_FAILED, "删除任务失败", "")
		return
	}
	//返回数据
	tools.BuildResponse(c, constants.CODE_SUCCESS, constants.MESSAGE_SUCCESS, "")
	return
}

//杀死某个任务
func (this *EtcdController) JobKill(c *gin.Context) {
	defer tools.Catch(c)
	var (
		err  error
		data []byte
		job  tools.Job
	)
	if data, err = this.Analysis(c); err != nil {
		log.Error("解析请求json失败", err)
		tools.BuildResponse(c, 1, "请求参数不正确", "")
		return
	}

	// 要杀死的任务名
	if err = json.Unmarshal(data, &job); err != nil {
		log.Error("解析json失败", err)
		return
	}

	// 杀死任务
	if err = services.GJobMgr.KillJob(job.Name); err != nil {
		tools.BuildResponse(c, constants.CODE_FAILED, "杀死任务失败", "")
		return
	}

	//返回数据
	tools.BuildResponse(c, constants.CODE_SUCCESS, constants.MESSAGE_SUCCESS, "")
	return

}

func (this *EtcdController) WorkerList(c *gin.Context) {
	defer tools.Catch(c)
	var (
		workerArr []string
		err       error
	)

	if workerArr, err = services.GWorkerMgr.ListWorkers(); err != nil {
		tools.BuildResponse(c, constants.CODE_FAILED, "获取worker节点失败", "")
		return
	}

	// 正常应答
	tools.BuildResponse(c, constants.CODE_SUCCESS, constants.MESSAGE_SUCCESS, workerArr)
	return

}

//展示页面
func (this *EtcdController) Index(c *gin.Context) {
	defer tools.Catch(c)
	c.HTML(http.StatusOK, "index.html", nil)
}
