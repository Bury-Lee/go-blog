// router/image_router.go
package router

import (
	"StarDreamerCyberNook/api"
	"StarDreamerCyberNook/middleware"

	"github.com/gin-gonic/gin"
)

func LocalImageRouter(r *gin.RouterGroup) { //图片路由注册函数
	api := api.App.ImageApi
	r.GET("/image", api.GetImage)
	r.POST("/images", middleware.ImgPostLimitMiddleware, middleware.AuthMiddleware, api.ImageUploadView)
	r.GET("/images", middleware.AdminMiddleware, api.ImageList)
	r.DELETE("/image", middleware.AdminMiddleware, api.ImageRemoveView) //考虑进行分开,一个真删除一个假删除
}

func OSSImageRouter(r *gin.RouterGroup) { //图片路由注册函数
	api := api.App.OSSImgApi
	r.GET("/image", api.GetImage)
	r.POST("/image", middleware.ImgPostLimitMiddleware, middleware.AuthMiddleware, api.ImageUploadView)
	r.GET("/images", middleware.AdminMiddleware, api.ImageList)
	r.DELETE("/image", middleware.AdminMiddleware, api.ImageRemoveView)
}

//TODO:允许用户删除自己的图片,从jwts中获取参数,然后允许批量删除
