package user_api

import (
	"StarDreamerCyberNook/common/response"
	"StarDreamerCyberNook/service/redis_service/redis_jwt"

	"github.com/gin-gonic/gin"
)

func (UserApi) LogoutView(c *gin.Context) {
	AccessToken := c.GetHeader("token")
	RefreshToken := c.GetHeader("refreshToken")
	redis_jwt.TokenBlack(AccessToken, RefreshToken, redis_jwt.UserBlackType)

	response.OkWithMsg("退出登录成功", c)
}
