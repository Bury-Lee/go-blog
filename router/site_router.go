// router/site_router.go
package router

import (
	"StarDreamerCyberNook/api"

	"github.com/gin-gonic/gin"
)

func SiteRouter(r *gin.RouterGroup) { //站点路由注册函数 TODO:这里很多方法都要重写一下了
	app := api.App.SiteApi

	//主站配置修改相关
	r.GET("/site/qq_login", app.SiteInfoQQView) //查询QQ登录配置
	r.GET("/site/:name", app.SiteInfoView)
	// r.PUT("/site/:name", middleware.AdminMiddleware, app.SiteUpdateView)//更新站点配置,分布式环境下就算了
}
