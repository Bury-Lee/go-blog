package router

import (
	"StarDreamerCyberNook/api"
	"StarDreamerCyberNook/middleware"

	"github.com/gin-gonic/gin"
)

func UserRouter(r *gin.RouterGroup) {
	app := api.App.UserApi
	r.POST("/user/send_email", middleware.CaptchaMiddleware, middleware.ActLimitMiddleware, app.SendEmailView)
	r.POST("/user/email", middleware.EmailVerifyMiddleware, app.RegisterEmailView)                                                                    //邮箱注册
	r.POST("/user/login", middleware.CaptchaMiddleware, app.Login)                                                                                    //登录
	r.DELETE("/user/logout", middleware.AuthMiddleware, app.LogoutView)                                                                               //发送邮箱验证码
	r.GET("/user/detail", middleware.AuthMiddleware, app.UserDetailView)                                                                              //获取用户详情
	r.GET("/user/info/:id", app.CheckUserBaseInfoView)                                                                                                //检查用户基础信息
	r.GET("/user/loginlog", middleware.AuthMiddleware, app.UserLoginListView)                                                                         //获取用户登录日志
	r.PUT("/user/resetEmail", middleware.AuthMiddleware, middleware.ResetEmailVerifyMiddleware, middleware.EmailVerifyMiddleware, app.ResetEmailView) //重置邮箱,要经过两次邮箱验证码验证,一次验证原邮箱,一次验证新邮箱
	r.PUT("/user/update", middleware.AuthMiddleware, app.UserInfoUpdateView)                                                                          //更新用户信息
	r.PUT("/user/admin/update", middleware.AdminMiddleware, app.AdminUserInfoUpdateView)                                                              //管理员更新用户信息
	r.POST("/user/token", app.RefreshAccessToken)                                                                                                     //刷新access token
	//TODO:加入邮箱验证码登录
}
