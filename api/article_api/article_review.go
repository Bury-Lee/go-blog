package article_api

import (
	"StarDreamerCyberNook/common/response"
	"StarDreamerCyberNook/global"
	"StarDreamerCyberNook/models"
	"StarDreamerCyberNook/service/message_service"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type ArticleExamineRequest struct {
	ArticleID uint          `json:"articleID" binding:"required"`
	Status    models.Status `json:"status" binding:"required"` //审核状态,2为通过,1,3为不通过
	Msg       string        `json:"msg"`                       // 为4的时候，传递进来
}

func (ArticleApi) ArticleReviewView(c *gin.Context) {
	// TODO:审核改为使用AI,在Canal哪里用go连接一个Python脚本,Python脚本配置几个ai,让ail审核文章,审核通过就更新文章的状态为3,审核不通过就更新文章的状态为1
	var req ArticleExamineRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMsg("参数错误", c)
		return
	}

	var article models.ArticleModel
	err := global.DB.Take(&article, req.ArticleID).Error
	if err != nil {
		response.FailWithMsg("文章不存在", c)
		return
	}

	global.DB.Model(&article).Update("status", req.Status)

	content := ""
	switch req.Status {
	case models.StatusPublished:
		content = fmt.Sprintf("文章通过审核:%s\n备注:%s", article.Title, req.Msg)
	case models.StatusDraft:
		content = fmt.Sprintf("文章不通过审核:%s\n备注:%s", article.Title, req.Msg)
	default:
		content = ""
	}

	message := models.MessageModel{
		RevUserID:          article.UserID,
		ActionUserID:       0,
		ActionUserNickname: "系统",
		ActionUserAvatar:   "", //这里可以改用站内默认头像
		Title:              "文章审核通知",
		ArticleID:          req.ArticleID,
		ArticleTitle:       article.Title,
		Content:            content,
	}
	err = message_service.InsertSystemMessage(message)
	if err != nil {
		logrus.Error("发送系统消息失败", err)
	}
	response.OkWithMsg("审核成功", c)
}
