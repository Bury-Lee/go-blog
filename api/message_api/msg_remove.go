package message_api

import (
	"StarDreamerCyberNook/common/response"
	"StarDreamerCyberNook/global"
	"StarDreamerCyberNook/models"
	jwts "StarDreamerCyberNook/utils/jwts"

	"github.com/gin-gonic/gin"
)

type MessageRemoveRequest struct {
	MessageID []uint `json:"messageID"` // 消息ID//其实使用model.RemoveRequest也是可以的
}

func (MessageApi) MessageRemoveView(c *gin.Context) {
	var req MessageRemoveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMsg("参数错误", c)
		return
	}
	if len(req.MessageID) == 0 || len(req.MessageID) > 100 {
		response.FailWithMsg("删除的消息数量必须在1-100之间", c)
		return
	}
	claim := jwts.GetClaims(c)
	if err := global.DB.Model(&models.MessageModel{}).Where("rev_user_id = ?", claim.UserID).Where("id in ?", req.MessageID).Delete(&models.MessageModel{}).Error; err != nil { //删除接收者是自己的ID为req的消息
		response.FailWithMsg("删除消息失败", c)
		return
	}
	response.OkWithMsg("删除消息成功", c)
}
