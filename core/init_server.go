package core

import (
	"StarDreamerCyberNook/global"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func InitServer(router *gin.Engine) *http.Server {
	return &http.Server{
		Addr:           global.Config.System.Addr(), // 设置服务器监听的地址
		Handler:        router,                      // 设置请求处理器（路由）
		ReadTimeout:    10 * time.Second,            // 设置请求的读取超时时间为 10 秒
		WriteTimeout:   10 * time.Second,            // 设置响应的写入超时时间为 10 秒
		MaxHeaderBytes: 1 << 20,                     // 设置最大请求头的大小（1MB）
	}
}
