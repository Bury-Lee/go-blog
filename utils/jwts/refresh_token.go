package jwts

import (
	"StarDreamerCyberNook/global"
	"StarDreamerCyberNook/models"
	"StarDreamerCyberNook/models/enum"
	"errors"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/sirupsen/logrus"
)

// RefreshClaims 刷新令牌声明
// 说明:仅携带用户ID和标准声明,减少泄露面
type RefreshClaims struct {
	ID uint `json:"id"`
	jwt.RegisteredClaims
}

// GetRefreshToken 生成刷新令牌
// 参数:userID - 用户ID
// 返回:string - refresh token
// 返回:error - 生成错误
// 说明:使用refresh密钥签名,过期时间使用RefreshExpire
func GetRefreshToken(userID uint) (string, error) {
	cla := jwt.NewWithClaims(jwtSigningMethod, RefreshClaims{
		ID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(global.Config.Jwt.RefreshExpire) * time.Hour)),
			Issuer:    global.Config.Jwt.Issuer,
		},
	})
	return cla.SignedString([]byte(global.Config.Jwt.RefreshTokenSecret))
}

// ParseRefreshToken 解析刷新令牌
// 参数:tokenString - 刷新token字符串
// 返回:*RefreshClaims - 解析后的声明
// 返回:error - 解析错误
// 说明:校验算法与密钥,拒绝算法混淆
func ParseRefreshToken(tokenString string) (*RefreshClaims, error) {
	if tokenString == "" {
		return nil, errors.New("请登录")
	}

	token, err := jwt.ParseWithClaims(tokenString, &RefreshClaims{}, func(token *jwt.Token) (interface{}, error) {
		if token.Method != jwtSigningMethod {
			logrus.Errorf("非法的算法: %v", token.Header["alg"])
			return nil, fmt.Errorf("非法的算法: %v", token.Header["alg"])
		}
		return []byte(global.Config.Jwt.RefreshTokenSecret), nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*RefreshClaims)
	if !ok || !token.Valid {
		return nil, errors.New("无效的token")
	}
	return claims, nil
}

// GetTokenPair 生成访问+刷新令牌
// 参数:claims - 访问令牌业务声明
// 返回:string - access token
// 返回:string - refresh token
// 返回:error - 生成错误
// 说明:登录和注册都统一调用,避免分散逻辑
func GetToken(claims Claims) (string, string, error) {
	accessToken, err := GetAccessToken(claims)
	if err != nil {
		return "", "", err
	}
	refreshToken, err := GetRefreshToken(claims.UserID)
	if err != nil {
		return "", "", err
	}
	return accessToken, refreshToken, nil
}

// RefreshTokenPair 刷新通行令牌
// 参数:refreshToken - 旧refresh token
// 返回:string - 新access token
// 返回:error - 刷新错误
// 说明:校验refresh,查用户状态,重新签发AccessToken
func RefreshAccessToken(refreshToken string) (string, error) {
	rc, err := ParseRefreshToken(refreshToken)
	if err != nil {
		return "", err
	}

	var user models.UserModel
	if err = global.DB.Take(&user, rc.ID).Error; err != nil {
		return "", errors.New("用户不存在或已失效")
	}
	if user.Role == enum.BlackRole {
		return "", errors.New("用户已被封禁")
	}

	//通过验证,重新签发accessToken
	accessToken, err := GetAccessToken(Claims{
		UserID:   user.ID,
		Username: user.UserName,
		Role:     user.Role,
	})
	if err != nil {
		return "", err
	}

	return accessToken, nil
}

func ParseRefreshTokenByGin(c *gin.Context) (*RefreshClaims, error) {
	// 只允许从HTTP请求体中获取token
	token := c.PostForm("token")
	if token == "" {
		return nil, errors.New("请登录")
	}

	return ParseRefreshToken(token)
}

func GetRefreshClaims(c *gin.Context) *RefreshClaims {
	// 从Gin上下文获取名为"claims"的值
	_claims, ok := c.Get("claims")
	if !ok {
		// 如果没有找到claims，返回nil
		return nil
	}
	// 类型转换为*RefreshClaims
	claims, ok := _claims.(*RefreshClaims)
	if !ok {
		// 如果类型转换失败，返回nil
		return nil
	}
	// 返回转换成功的声明对象
	return claims
}
