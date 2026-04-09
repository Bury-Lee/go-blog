package article_api

import (
	"StarDreamerCyberNook/common/response"
	"StarDreamerCyberNook/global"
	"StarDreamerCyberNook/models"
	"StarDreamerCyberNook/service/message_service"
	"StarDreamerCyberNook/service/redis_service/redis_count"
	jwts "StarDreamerCyberNook/utils/jwts"

	"github.com/gin-gonic/gin"
)

func (ArticleApi) ArticleDiggView(c *gin.Context) { //TODO:这里的逻辑需要优化,Redis的点赞存储应该是用户+文章ID,然后定时任务写入Mysql,因为这是最高频的操作之一了,每次都写数据库太逆天了
	var IDRequest models.IDRequest
	if err := c.ShouldBindUri(&IDRequest); err != nil {
		response.FailWithMsg("参数错误", c)
		return
	}

	var article models.ArticleModel
	err := global.DB.Take(&article, "status = ? and id = ?", models.StatusPublished, IDRequest.ID).Error
	if err != nil {
		response.FailWithMsg("文章不存在", c)
		return
	}

	claims := jwts.GetClaims(c)

	// 查一下之前有没有点过
	var userDiggArticle models.ArticleDiggModel
	err = global.DB.Take(&userDiggArticle, "user_id = ? and article_id = ?", claims.UserID, article.ID).Error
	if err != nil {
		// 点赞
		err = global.DB.Create(&models.ArticleDiggModel{
			UserID:    claims.UserID,
			ArticleID: IDRequest.ID,
		}).Error
		if err != nil {
			response.FailWithMsg("点赞失败", c)
			return
		}
		// TODO: 更新点赞数到缓存里面
		redis_count.SetCacheDigg(IDRequest.ID, true)
		response.OkWithMsg("点赞成功", c)
		// 发送点赞消息
		err = message_service.InsertArticleDiggMessage(userDiggArticle)
		if err != nil {
			response.FailWithMsg("发送点赞消息失败", c)
		}
		return
	}
	// 取消点赞
	redis_count.SetCacheDigg(IDRequest.ID, false)
	global.DB.Delete(&models.ArticleDiggModel{}, "user_id = ? and article_id = ?", claims.UserID, article.ID) //没必要进行错误处理
	response.OkWithMsg("取消点赞成功", c)
}
