package router

import (
	"StarDreamerCyberNook/api"
	"StarDreamerCyberNook/middleware"

	"github.com/gin-gonic/gin"
)

func MessageRouter(r *gin.RouterGroup) { //消息系统路由注册函数
	app := api.App.SiteMessageApi
	r.GET("/msg/conf", middleware.AuthMiddleware, app.SiteMessageConfView) //有问题
	r.POST("/msg/conf/update", middleware.AuthMiddleware, app.SiteMessageConfUpdateView)
	r.GET("/msg/check", middleware.AuthMiddleware, app.SiteMessageCheckView)
	r.POST("/msg/clear", middleware.AuthMiddleware, app.OneKeyClearView)
	r.GET("/msg", middleware.AuthMiddleware, app.SiteMessageListView)
	r.DELETE("/msg", middleware.AuthMiddleware, app.MessageRemoveView)
}
