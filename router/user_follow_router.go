package router

import (
	"StarDreamerCyberNook/api"
	"StarDreamerCyberNook/middleware"

	"github.com/gin-gonic/gin"
)

func UserFollowRouter(r *gin.RouterGroup) {
	app := api.App.FollowApi
	r.GET("/user/follow/list", app.FollowUserListView) //关注列表
	r.GET("/user/follower/list", app.FollowerListView) //粉丝列表

	r.GET("/user/friend/check")                                                      //检查是否关注用户
	r.POST("/user/follow", middleware.AuthMiddleware, app.FollowUserView)            //关注用户
	r.POST("/user/follow/unfollow", middleware.AuthMiddleware, app.UnFollowUserView) //取消关注用户
}
