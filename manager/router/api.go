package routes

import (
	"github.com/gin-gonic/gin"
	"gocrontab/manager/controllers"
)

func Bind(port string) {
	var (
		router *gin.Engine
	)
	router = gin.New()

	projectRouter(router)
	router.NoRoute(func(c *gin.Context) {
		c.JSON(404, gin.H{"ret": 404})
	})
	router.Run(":" + port)
}

//项目访问路径
func projectRouter(r *gin.Engine) {

	var (
		ed     controllers.EtcdController
		route  *gin.RouterGroup
		wRoute *gin.RouterGroup
	)
	//加载html文件
	r.LoadHTMLGlob("./manager/views/*")
	r.GET("/", ed.Index)

	route = r.Group("job")
	{
		route.POST("/save", ed.JobAdd)
		route.POST("/delete", ed.JobDelete)
		route.POST("/kill", ed.JobKill)
		route.POST("/list", ed.JobList)
	}

	wRoute = r.Group("worker")
	{
		wRoute.POST("/list", ed.WorkerList)
	}

}
