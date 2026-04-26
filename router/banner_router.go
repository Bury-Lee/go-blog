package router

import (
	"StarDreamerCyberNook/api"
	"StarDreamerCyberNook/middleware"

	"github.com/gin-gonic/gin"
)

func BannerRouter(r *gin.RouterGroup) {
	api := api.App.BannerApi
	r.GET("/banner", api.BannerListView)
	r.DELETE("/banner", middleware.AdminMiddleware, api.BannerRemoveView)
	r.POST("/banner", middleware.AdminMiddleware, api.BannerCreateView)
	r.PUT("/banner/:id", middleware.AdminMiddleware, api.BannerUpdateView)
}
