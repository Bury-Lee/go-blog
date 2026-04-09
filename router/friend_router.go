package router

import (
	"StarDreamerCyberNook/api"
	"StarDreamerCyberNook/middleware"

	"github.com/gin-gonic/gin"
)

func FriendRouter(r *gin.RouterGroup) { //图片路由注册函数
	api := api.App.FriendApi
	r.GET("/friendLink", api.FriendLinkListView)
	r.DELETE("/friendLink", middleware.AdminMiddleware, api.FriendLinkRemoveView)
	r.POST("/friendLink", middleware.AdminMiddleware, api.FriendLinkCreateView)
	r.PUT("/friendLink/:id", middleware.AdminMiddleware, api.FriendLinkUpdateView)
	r.GET("/friendPromotion", api.FriendPromotionListView)
	r.DELETE("/friendPromotion", middleware.AdminMiddleware, api.FriendPromotionRemoveView)
	r.POST("/friendPromotion", middleware.AdminMiddleware, api.FriendPromotionCreateView)
	r.PUT("/friendPromotion/:id", middleware.AdminMiddleware, api.FriendPromotionUpdateView)

}
