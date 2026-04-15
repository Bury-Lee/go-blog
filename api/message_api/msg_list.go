package message_api

import (
	"StarDreamerCyberNook/common"
	"StarDreamerCyberNook/common/response"
	"StarDreamerCyberNook/global"
	"StarDreamerCyberNook/models"
	jwts "StarDreamerCyberNook/utils/jwts"

	"github.com/gin-gonic/gin"
)

type SiteMessageListViewRequest struct {
	common.PageInfo
	Type models.MessageType `query:"type" binding:"required"` //查询的消息类型
}

func (MessageApi) SiteMessageListView(c *gin.Context) { //可以在读取之后,返回响应之后吧已读字段一起写回数据库
	var req SiteMessageListViewRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithMsg("参数错误", c)
		return
	}
	claim := jwts.GetClaims(c)

	query := global.DB.Where("rev_user_id = ?", claim.UserID)
	switch req.Type {
	case models.MessageTypeAt, models.MessageTypeComment, models.MessageTypeReply: //不好有bug,自己回复自己也会有消息,需要过滤
		query.Where("type in ?", []models.MessageType{models.MessageTypeAt, models.MessageTypeComment, models.MessageTypeReply})
		query.Where("action_user_id <> ?", claim.UserID) //过滤掉自己回复自己的消息
	case models.MessageTypeCollect, models.MessageTypeDigg:
		query.Where("type in ?", []models.MessageType{models.MessageTypeCollect, models.MessageTypeDigg})
	case models.MessageTypePrivate:
		query.Where("type = ?", models.MessageTypePrivate)
	case models.MessageTypeSystem:
		query.Where("type = ?", models.MessageTypeSystem)
	default:
		response.FailWithMsg("消息类型错误", c)
		return
	}

	var Options common.Options
	Options.PageInfo = req.PageInfo
	Options.Where = query
	list, count, err := common.ListQuery[models.MessageModel](
		models.MessageModel{RevUserID: claim.UserID},
		Options,
	)
	if err != nil {
		response.FailWithMsg("查询消息失败", c)
		return
	}
	response.OkWithList(list, count, c)

	//把读取的消息设置为已读
	var updateList []uint
	for _, item := range list {
		if item.IsRead == false {
			updateList = append(updateList, item.ID)
		}
	}
	if len(updateList) > 0 {
		global.DB.Model(&models.MessageModel{}).Where("id IN ?", updateList).Update("is_read", true)
	}
}
