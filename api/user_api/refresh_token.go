package user_api

import (
	"StarDreamerCyberNook/common/response"
	jwts "StarDreamerCyberNook/utils/jwts"

	"github.com/gin-gonic/gin"
)

func (UserApi) RefreshAccessToken(c *gin.Context) {
	// 从请求头的 refreshToken 字段获取 token
	authHeader := c.GetHeader("refreshToken") // 或者使用 "refresh-token"
	if authHeader == "" {
		response.FailWithMsg("缺少 refreshToken 头", c)
		return
	}

	// 直接使用 token，不需要 Bearer 前缀
	token := authHeader

	accessToken, err := jwts.RefreshAccessToken(token)
	// 处理错误
	if err != nil {
		response.FailWithMsg("刷新失败: "+err.Error(), c)
		return
	}
	response.OkWithData(accessToken, c)
}
