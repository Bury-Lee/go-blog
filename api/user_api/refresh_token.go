package user_api

import (
	"StarDreamerCyberNook/common/response"
	jwts "StarDreamerCyberNook/utils/jwts"

	"github.com/gin-gonic/gin"
)

func (UserApi) RefreshAccessToken(c *gin.Context) {
	// 从请求头的 Authorization 字段获取 token
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		response.FailWithMsg("缺少 Authorization 头", c)
		return
	}

	// 验证格式为 "Bearer {token}"
	var token string
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		token = authHeader[7:]
	} else {
		response.FailWithMsg("Authorization 头格式错误", c)
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
