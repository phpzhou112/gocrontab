package tools

import (
	log "github.com/alecthomas/log4go"
	"github.com/gin-gonic/gin"
	"os"
	"reflect"
	"runtime"
	"runtime/debug"
	"time"
)

//异常捕获
func Catch(c *gin.Context) {

	var data = map[string]string{}

	if err := recover(); err != nil {
		debug.PrintStack()
		exeName := os.Args[0] //获取程序名称
		now := time.Now()     //获取当前时间
		pid := os.Getpid()    //获取进程ID
		log.Error("Exception:", exeName, now, pid)
		log.Error("Exception:", err, reflect.TypeOf(err))

		code := 1
		data = nil
		msg := "exception"

		switch err.(type) {
		case *runtime.TypeAssertionError:
			c.AbortWithStatusJSON(200, gin.H{
				"code":    code,
				"message": msg,
				"data":   data,
			})
		case *runtime.Error:
			c.AbortWithStatusJSON(200, gin.H{
				"code":    code,
				"message": msg,
				"data":   data,
			})
		case string:
			c.AbortWithStatusJSON(200, gin.H{
				"code":    code,
				"message": msg,
				"data":   data,
			})
		case int:
			c.AbortWithStatusJSON(200, gin.H{
				"code":    code,
				"message": msg,
				"data":   data,
			})
		default:
			c.AbortWithStatusJSON(200, gin.H{
				"code":    code,
				"message": msg,
				"data":   nil,
			})
		}
	}

	return
}
