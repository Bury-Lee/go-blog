package router

import (
	"StarDreamerCyberNook/api"
	"StarDreamerCyberNook/middleware"

	"github.com/gin-gonic/gin"
)

func ChatRouter(r *gin.RouterGroup) {
	app := api.App.ChatApi
	r.POST("/chat/send", middleware.AuthMiddleware, app.ChatSendView)      // 发送消息
	r.GET("/chat/get", middleware.AuthMiddleware, app.ChatListView)        // 查我发送的消息
	r.GET("/chat/session", middleware.AuthMiddleware, app.SessionListView) // 查我的会话列表
}
