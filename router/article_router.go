package router

import (
	"StarDreamerCyberNook/api"
	"StarDreamerCyberNook/middleware"

	"github.com/gin-gonic/gin"
)

func ArticleRouter(r *gin.RouterGroup) {
	app := api.App.ArticleApi
	r.POST("/article", middleware.AuthMiddleware, app.ArticleCreateView)     //创建文章//待测
	r.PUT("/article", middleware.AuthMiddleware, app.ArticleUpdateView)      //更新文章//待测
	r.PUT("/article/inc", middleware.AuthMiddleware, app.ArticleUpdateView2) //增量更新文章//待测
	r.GET("/article", app.ArticleListView)                                   //获取文章列表 //待测
	r.GET("/article/search", app.ArticleSearchView)                          //搜索文章//待测
	r.GET("/article/:id", app.ArticleDetailView)                             //获取文章详情//待测

	r.POST("/article/top/:id", middleware.AuthMiddleware, app.ArticleTopView)              //置顶文章
	r.DELETE("/article/top", middleware.AuthMiddleware, app.ArticleCancleTopView)          //取消置顶
	r.DELETE("/article/admingTop", middleware.AdminMiddleware, app.AdminArticleDeleteView) //管理员取消置顶

	r.GET("/article/review", middleware.AdminMiddleware, app.ArticleReviewListView)  //获取审核文章列表
	r.POST("/article/review/:id", middleware.AdminMiddleware, app.ArticleReviewView) //审核文章

	r.POST("/article/look", middleware.AuthMiddleware, app.ArticleLookView)       //创建浏览记录,这样单独加个接口还能开无痕模式设置
	r.POST("/article/digg/:id", middleware.AuthMiddleware, app.ArticleDiggView)   //点赞文章
	r.DELETE("/article", middleware.AuthMiddleware, app.ArticleRemoveUserView)    //删除文章
	r.DELETE("/article/admin", middleware.AdminMiddleware, app.ArticleRemoveView) //删除文章(管理员)

	r.GET("/article/history", app.ArticleLookListView)                                 //获取文章浏览记录
	r.DELETE("/article/history", middleware.AuthMiddleware, app.ArticleLookRemoveView) //删除文章浏览记录

	r.POST("/article/category", middleware.AuthMiddleware, app.CategoryCreateView)   //创建文章分类
	r.GET("/article/category", app.CategoryListView)                                 //获取文章分类列表
	r.DELETE("/article/category", middleware.AuthMiddleware, app.CategoryRemoveView) //删除文章分类

	r.POST("/article/collect", middleware.AuthMiddleware, app.ArticleCollectView)         //收藏文章
	r.GET("/article/collect/folder", app.CollectListView)                                 //查看收藏夹列表
	r.GET("/article/collect/list", app.CollectArticleListView)                            //查看收藏夹内文章列表
	r.POST("/article/collect/folder", middleware.AuthMiddleware, app.CollectCreateView)   //创建收藏夹
	r.PUT("/article/collect/folder", middleware.AuthMiddleware, app.CollectUpdateView)    //更新收藏夹
	r.DELETE("/article/collect/folder", middleware.AuthMiddleware, app.CollectRemoveView) //删除收藏夹
}
