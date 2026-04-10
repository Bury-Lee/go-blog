package comment_api

import (
	"StarDreamerCyberNook/common/response"
	"StarDreamerCyberNook/global"
	"StarDreamerCyberNook/models"
	"StarDreamerCyberNook/service/message_service"
	jwts "StarDreamerCyberNook/utils/jwts"

	"github.com/gin-gonic/gin"
)

func (CommentApi) AtOther(c *gin.Context) { //@别人并且发送时前端自动调用的接口
	var req models.IDRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithMsg("参数错误", c)
		return
	}
	claim := jwts.GetClaims(c)
	var actor models.UserModel
	err := global.DB.Take(&actor, "id = ?", claim.UserID).Error
	if err != nil {
		response.FailWithMsg("用户不存在", c)
		return
	}
	message_service.InsertAtMessage(actor, req.ID)
}
