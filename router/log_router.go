// router/log_router.go
package router

import (
	"StarDreamerCyberNook/api"
	"StarDreamerCyberNook/middleware"

	"github.com/gin-gonic/gin"
)

func LogRouter(r *gin.RouterGroup) { //日志系统路由注册函数

	/*
	   反面教材:管理员中间件不能以r.use(认证中间件)的形式放中间,不然就会绑定到所有的路由上,包括那些不需要管理员权限的路由
	   现在是正确示例
	*/

	api := api.App.LogApi
	r.GET("/logs", middleware.AdminMiddleware, api.LogListView)
	r.GET("/logs/:id", middleware.AdminMiddleware, api.LogReadView)
	r.DELETE("/logs", middleware.AdminMiddleware, api.LogRemoveView)

}
