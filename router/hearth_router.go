package router

import (
	"StarDreamerCyberNook/api"

	"github.com/gin-gonic/gin"
)

func HearthRouter(r *gin.RouterGroup) { //心跳路由注册函数
	api := api.App.HearthApi
	r.GET("/heartbeat", api.Hearth)

}
