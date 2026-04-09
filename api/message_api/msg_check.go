package message_api

import (
	"StarDreamerCyberNook/common/response"
	"StarDreamerCyberNook/global"
	"StarDreamerCyberNook/models"
	jwts "StarDreamerCyberNook/utils/jwts"

	"github.com/gin-gonic/gin"
)

func (MessageApi) SiteMessageCheckView(c *gin.Context) {
	claim := jwts.GetClaims(c)

	// 定义一个临时结构体来接收查询结果
	type MessageCount struct {
		MessageType models.MessageType `json:"message_type"` // 查询的字段名
		Count       uint8              `json:"count"`        // COUNT(*) 的结果
	}

	var counts []MessageCount
	// 使用 DB.Raw 或者更推荐的 Model 方式进行聚合查询
	if err := global.DB.Model(&models.MessageModel{}).
		Select("type, count(*) as count"). // 选择类型和计数
		Where("rev_user_id = ? AND is_read = ?", claim.UserID, false).
		Group("type"). // 按类型分组
		Find(&counts).Error; err != nil {
		response.FailWithMsg("未读消息查询失败", c)
		return
	}

	// 组装结果
	IsRead := make(map[models.MessageType]uint8)
	for _, item := range counts {
		IsRead[item.MessageType] = item.Count // 将每种类型的未读数量存入map
	}

	response.OkWithData(IsRead, c)
}
