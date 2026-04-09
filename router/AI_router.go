package router

import (
	"StarDreamerCyberNook/api"
	"StarDreamerCyberNook/global"

	"github.com/gin-gonic/gin"
)

func AIRouter(r *gin.RouterGroup) {
	if !global.Config.AI.Enable { //如果AI功能未开启,则不注册路由,停用ai功能
		return
	}
	app := api.App.AIApi
	r.POST("/chat", app.Chat)
}
