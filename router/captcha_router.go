package router

import (
	"StarDreamerCyberNook/api"

	"github.com/gin-gonic/gin"
)

func CaptcharRouter(r *gin.RouterGroup) {
	app := api.App.CaptchaApi
	//验证码相关
	r.GET("/captcha", app.CaptchaView)
}
