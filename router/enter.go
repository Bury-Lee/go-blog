// router/enter.go
package router

import (
	_ "embed"

	"StarDreamerCyberNook/global"
	"StarDreamerCyberNook/middleware"

	"github.com/gin-gonic/gin"
)

//go:embed 404page.html
var UnFoundPage string //这是404页面的HTML内容,用于自定义404页面,由于前端不想做,所以这里直接写在这里了

func InitRouter() *gin.Engine {
	gin.SetMode(global.Config.System.RunMode) //设置gin模式
	r := gin.Default()

	r.Static("/web", "static") //静态文件目录,注册个路由到时候写网页?而且还可以用于获取图片//TODO:后期把这个路径移到配置里,不要写死在代码里

	// nr := r.Group("/api")

	nr := r.Group("/api") //TODO:测试使用无前缀api,开发完成了要给app组加上/api的前缀,nr := r.Group("/api")这样

	if global.Config.System.RunMode == "debug" {
		nr.Use(middleware.RequestLogMiddleware)
		nr.Use(middleware.CORS)
		TestRouter(nr)
	}
	nr.Use(middleware.LogMiddleware)
	nr.Use(middleware.ActLimitMiddleware) //防攻击中间件
	HearthRouter(nr)                      //心跳路由注册函数

	SiteRouter(nr) //已测试完毕
	LogRouter(nr)  //由于条件问题,待测

	if global.Config.ObjectStorage.Enable {
		OSSImageRouter(nr)
	} else {
		//如果对象存储未启用,则使用本地存储
		LocalImageRouter(nr) //已测试完毕
	}
	BannerRouter(nr)   //已测试完毕
	UserRouter(nr)     //已测试完毕
	CaptcharRouter(nr) //已测试完毕
	ArticleRouter(nr)  //已测试完毕
	CommentRouter(nr)  //已测试完毕
	MessageRouter(nr)
	ChatRouter(nr)
	AIRouter(nr) //已测试完毕
	FriendRouter(nr)

	//硬编码的HTML内容,我信不过前端
	r.NoRoute(func(ctx *gin.Context) { //自定义的404页面
		ctx.Data(200, "text/html; charset=utf-8", []byte(UnFoundPage))
	})

	// add := global.Config.System.Addr()
	// r.Run(add) //监听地址为配置文件中的IP和端口
	return r
}
