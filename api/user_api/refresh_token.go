package user_api

import (
	"StarDreamerCyberNook/common/response"
	jwts "StarDreamerCyberNook/utils/jwts"

	"github.com/gin-gonic/gin"
)

func (UserApi) RefreshAccessToken(c *gin.Context) {
	// 优先兼容标准 Authorization: Bearer <token>，其次兼容历史 refreshToken 头
	token := ""
	authHeader := c.GetHeader("Authorization")
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		token = authHeader[7:]
	}
	if token == "" {
		token = c.GetHeader("refreshToken")
	}
	if token == "" {
		response.FailWithMsg("缺少 refreshToken", c)
		return
	}

	accessToken, err := jwts.RefreshAccessToken(token)
	// 处理错误
	if err != nil {
		response.FailWithMsg("刷新失败: "+err.Error(), c)
		return
	}
	response.OkWithData(accessToken, c)
}
