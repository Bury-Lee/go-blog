package message_api

import (
	"StarDreamerCyberNook/common/response"
	"StarDreamerCyberNook/global"
	"StarDreamerCyberNook/models"
	jwts "StarDreamerCyberNook/utils/jwts"
	utils_other "StarDreamerCyberNook/utils/other"

	"github.com/gin-gonic/gin"
)

type MessageApi struct {
}

func (MessageApi) SiteMessageConfView(c *gin.Context) { //查询站点消息列表配置
	//查询站点消息列表,然后查询"已读"的数据结构,如果没有查到,说明是未读消息,否则是已读消息?
	claim := jwts.GetClaims(c)
	if claim == nil {
		response.FailWithMsg("未登录", c)
		return
	}
	var userMessageConf models.UserMessageConfModel
	if err := global.DB.Where("user_id = ?", claim.UserID).Take(&userMessageConf).Error; err != nil {
		response.FailWithMsg("查询用户消息配置失败", c)
		return
	}
	userMessageConf.UserModel = models.UserModel{} //过滤用户信息
	response.OkWithData(userMessageConf, c)
}

type SiteMessageConfUpdateRequest struct {
	ID                 uint             `json:"id"`
	UserID             uint             `json:"userID"`                                      //TODO:迟点检查一下
	User               models.UserModel `gorm:"foreignKey:UserID" json:"user"`               // 关联的用户表
	OpenCommentMessage *bool            `json:"openCommentMessage" u:"open_comment_message"` // 是否开启评论通知
	OpenReplyMessage   *bool            `json:"openReplyMessage" u:"open_reply_message"`     // 是否开启回复通知
	OpenDiggMessage    *bool            `json:"openDiggMessage" u:"open_digg_message"`       // 是否开启点赞通知
	OpenCollectMessage *bool            `json:"openCollectMessage" u:"open_collect_message"` // 是否开启收藏通知
	OpenPrivateMessage *bool            `json:"openPrivateMessage" u:"open_private_message"` // 是否开启私信通知
}

func (MessageApi) SiteMessageConfUpdateView(c *gin.Context) {
	claim := jwts.GetClaims(c)
	if claim == nil {
		response.FailWithMsg("未登录", c)
		return
	}
	var userMessageConf models.UserMessageConfModel
	if err := c.ShouldBindJSON(&userMessageConf); err != nil {
		response.FailWithMsg("参数绑定失败", c)
		return
	}

	if err := global.DB.Where("user_id = ?", claim.UserID).Take(&userMessageConf).Error; err != nil {
		response.FailWithMsg("查询用户消息配置失败", c)
		return
	}
	userConfMap := utils_other.StructToMap(userMessageConf, "u")
	global.DB.Model(&userMessageConf).Updates(userConfMap)
	response.OkWithMsg("更新用户消息配置成功", c)
}
