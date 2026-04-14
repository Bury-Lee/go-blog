package article_api

import (
	"StarDreamerCyberNook/common/response"
	"StarDreamerCyberNook/global"
	"StarDreamerCyberNook/models"
	"StarDreamerCyberNook/models/enum"
	"StarDreamerCyberNook/utils/jwts"
	"fmt"

	"github.com/gin-gonic/gin"
)

type ArticleTopRequest struct {
	TopType   string `json:"topType"`                      //,置顶类型，用户置顶还是管理员置顶,不填写时默认为用户置顶
	ArticleID uint   `json:"articleID" binding:"required"` //文章ID
}

func (ArticleApi) ArticleTopView(c *gin.Context) {
	//TODO:用户最多置顶n篇文章，管理员置顶不受限制，置顶的文章会在文章列表接口中优先展示
	var req ArticleTopRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMsg("参数错误", c)
		return
	}
	claim, err := jwts.ParseTokenByGin(c)
	if err != nil {
		response.FailWithMsg("token解析失败", c)
		return
	}
	//查询是否已经有这篇文章的置顶
	if global.DB.Where("user_id = ? AND article_id = ?", claim.UserID, req.ArticleID).First(&models.UserTopArticleModel{}).Error == nil {
		response.FailWithMsg("这篇文章已经被你置顶了", c)
		return
	}
	if claim.Role != enum.AdminRole {
		var count int64
		global.DB.Model(&models.UserTopArticleModel{}).Where("user_id = ?", claim.UserID).Count(&count)
		if count > maxTopLimit {
			response.FailWithMsg(fmt.Sprintf("你已经有%d篇文章被置顶了,不能再多了", maxTopLimit), c)
			return
		}
	}
	topModel := models.UserTopArticleModel{
		UserID:    claim.UserID,
		ArticleID: req.ArticleID,
	}

	if global.DB.Create(&topModel).Error != nil {
		response.FailWithMsg("置顶失败", c)
		return
	}
	response.OkWithMsg("置顶成功", c)
}

// 写个管理员可以强制删除的接口
func (ArticleApi) ArticleCancleTopView(c *gin.Context) {
	var req ArticleTopRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMsg("参数错误", c)
		return
	}
	claim, err := jwts.ParseTokenByGin(c)
	if err != nil {
		response.FailWithMsg("token解析失败", c)
		return
	}
	//查询是否已经有这篇文章的置顶
	if global.DB.Where("user_id = ? AND article_id = ?", claim.UserID, req.ArticleID).First(&models.UserTopArticleModel{}).Error != nil {
		response.FailWithMsg("你还没置顶过文章", c)
		return
	}
	//如果有就删除
	topModel := models.UserTopArticleModel{
		UserID:    claim.UserID,
		ArticleID: req.ArticleID,
	}

	if global.DB.Delete(&topModel).Error != nil {
		response.FailWithMsg("取消置顶失败", c)
		return
	}
	response.OkWithMsg("取消置顶成功", c)
}

// 写个管理员可以强制删除的接口
type AdminArticleTopRequest struct {
	UserID    uint `json:"userID"`
	ArticleID uint `json:"articleID"`
}

func (ArticleApi) AdminArticleDeleteView(c *gin.Context) {
	var req AdminArticleTopRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMsg("参数错误", c)
		return
	}
	claim, err := jwts.ParseTokenByGin(c)
	if err != nil || claim.Role != enum.AdminRole {
		response.FailWithMsg("权限不足", c)
		return
	}

	//查询是否已经有这篇文章的置顶
	if global.DB.Where("user_id = ? AND article_id = ?", req.UserID, req.ArticleID).First(&models.UserTopArticleModel{}).Error != nil {
		response.FailWithMsg("没有相关记录", c)
		return
	}
	//如果有就删除
	topModel := models.UserTopArticleModel{
		UserID:    req.UserID,
		ArticleID: req.ArticleID,
	}

	if global.DB.Delete(&topModel).Error != nil {
		response.FailWithMsg("取消置顶失败", c)
		return
	}
	response.OkWithMsg("取消置顶成功", c)
}
