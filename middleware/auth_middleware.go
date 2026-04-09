package middleware

import (
	"StarDreamerCyberNook/common/response"
	"StarDreamerCyberNook/models/enum"
	"StarDreamerCyberNook/service/redis_service/redis_jwt"
	jwts "StarDreamerCyberNook/utils/jwts"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware(c *gin.Context) { //鉴权中间件
	claim, err := jwts.ParseTokenByGin(c)
	if err != nil {
		response.FailWithError(err, c)
		c.Abort()
		return
	}
	blackType, ok := redis_jwt.HasTokenByGin(c)
	if ok {
		response.FailWithMsg("由于"+blackType.String()+",服务已不可用", c)
		c.Abort()
		return
	}
	c.Set("claims", claim)
}

func AdminMiddleware(c *gin.Context) { //管理员中间件
	claim, err := jwts.ParseTokenByGin(c)
	if err != nil {
		response.FailWithMsg("无记录", c)
		c.Abort()
		return
	}
	blackType, ok := redis_jwt.HasTokenByGin(c)
	if ok {
		response.FailWithMsg("由于"+blackType.String()+",服务已不可用", c)
		c.Abort()
		return
	}
	if claim.Role != enum.AdminRole {
		response.FailWithMsg("权限不足", c)
		c.Abort()
		return
	}
}

func VipMiddleware(c *gin.Context) { //会员中间件
	claim, err := jwts.ParseTokenByGin(c)
	if err != nil {
		response.FailWithError(err, c)
		c.Abort()
		return
	}
	blackType, ok := redis_jwt.HasTokenByGin(c)
	if ok {
		response.FailWithMsg("由于"+blackType.String()+",服务已不可用", c)
		c.Abort()
		return
	}
	if claim.Role != enum.VipRole && claim.Role != enum.AdminRole {
		response.FailWithMsg("权限不足", c)
		c.Abort()
		return
	}
}
