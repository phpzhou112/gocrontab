package controllers

import (
	"errors"
	log "github.com/alecthomas/log4go"
	"github.com/gin-gonic/gin"
)

type BaseController struct{}

func (b *BaseController) Analysis(c *gin.Context) (data []byte, err error) {
	var reqData string
	reqData = c.Request.FormValue("data")
	log.Info("[BaseController]接受参数data结束，入参：", reqData)
	if reqData == "" {
		err = errors.New("请求参数为空")
		return
	}
	data = []byte(reqData)
	return
}
