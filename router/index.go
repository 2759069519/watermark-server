package router

import (
	"github.com/gin-gonic/gin"
	"watermarkServer/controllers"
)

var ctr = controllers.IndexController{}

func IndexRouter(r *gin.Engine) {
	r.GET("/parse", ctr.Parse)
	r.POST("/parse", ctr.Parse)
	r.GET("/proxy/*encoded", controllers.Proxy)
	r.POST("/proxy", controllers.Proxy)
}
