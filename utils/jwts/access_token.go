// utils/jwts/enter.go
package jwts

//TODO:目前使用的是单token模式,以后改为双token
import (
	"StarDreamerCyberNook/global"
	"StarDreamerCyberNook/models"
	"StarDreamerCyberNook/models/enum"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/sirupsen/logrus"

	"github.com/gin-gonic/gin"
)

// Claims 自定义JWT声明结构体
// 包含用户的基本身份信息：用户ID、用户名和角色类型
type Claims struct {
	UserID   uint          `json:"ID"`   // 用户唯一标识符
	Username string        `json:"name"` // 用户名
	Role     enum.RoleType `json:"role"` // 用户角色类型（如管理员、普通用户等）
}

// MyClaims 继承自定义声明并嵌入标准声明
// 将自定义声明与JWT标准声明合并，形成完整的JWT声明结构
type MyClaims struct {
	Claims
	jwt.RegisteredClaims // 包含标准JWT声明（如exp过期时间、iss发行人等）
}

// GetUser 根据JWT中的UserID从数据库中获取用户完整信息
// 这个方法用于验证用户是否仍然有效（比如用户未被删除或禁用）
func (this *MyClaims) GetUser() (models.UserModel, error) {
	// 使用全局数据库连接，根据UserID查询用户信息
	var user models.UserModel
	err := global.DB.Take(&user, this.UserID).Error
	return user, err
}

var jwtSigningMethod = jwt.SigningMethodHS256

// GetAccessToken 生成JWT令牌
// 输入：用户身份信息（Claims）
// 输出：JWT字符串和错误信息
func GetAccessToken(claims Claims) (string, error) {
	// 创建JWT对象，使用HS256算法签名，包含自定义声明和标准声明
	cla := jwt.NewWithClaims(jwtSigningMethod, MyClaims{
		Claims: claims,
		RegisteredClaims: jwt.RegisteredClaims{
			// 设置过期时间：当前时间 + 配置文件中的AccessExpire小时数
			// 这里设置为24小时后过期（具体时长由配置决定）
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(global.Config.Jwt.AccessExpire) * time.Minute)),
			// IssuedAt:  jwt.NewNumericDate(time.Now()),                     // 当前签发时间设置（已注释）
			// 设置JWT发行人，用于验证JWT的来源
			Issuer: global.Config.Jwt.Issuer,
		},
	})

	// 使用配置文件中的密钥对JWT进行签名，生成最终的token字符串
	return cla.SignedString([]byte(global.Config.Jwt.AccessTokenSecret))
}

// ParseAccessToken 解析和验证JWT令牌
// 输入：JWT字符串
// 输出：解析后的声明信息和错误信息
func ParseAccessToken(tokenString string) (*MyClaims, error) {
	// 如果token为空，直接返回错误，提示需要登录
	if tokenString == "" {
		return nil, errors.New("请登录")
	}

	// 解析JWT字符串，验证签名并提取声明
	token, err := jwt.ParseWithClaims(tokenString, &MyClaims{}, func(token *jwt.Token) (interface{}, error) {
		// 验证签名算法是否为预期的算法，防止算法混淆攻击
		if token.Method != jwtSigningMethod {
			logrus.Errorf("算法混淆攻击: %v!", token.Header["alg"])
			return nil, fmt.Errorf("非法的算法: %v", token.Header["alg"])
		}

		// 返回用于验证签名的密钥
		return []byte(global.Config.Jwt.AccessTokenSecret), nil
	})

	// 如果解析过程中出现错误，返回该错误
	if err != nil {
		logrus.Errorf("token解析失败: %v", err) // 日志记录
		return nil, err
	}

	// 检查解析结果是否有效：声明类型正确且token未过期/被篡改
	if claims, ok := token.Claims.(*MyClaims); ok && token.Valid {
		return claims, nil
	}
	// 如果验证失败，返回无效token错误
	return nil, errors.New("invalid token")
}

// ParseTokenByGin 从Gin上下文中解析JWT通行令牌
// 优先从请求头的"token"字段获取，如果不存在则从URL查询参数中获取
func ParseTokenByGin(c *gin.Context) (*MyClaims, error) {
	// 从HTTP请求头中获取token
	token := c.GetHeader("token")
	if token == "" {
		// 如果请求头中没有token，则尝试从URL查询参数中获取
		token = c.Query("token")
	}

	return ParseAccessToken(token)
}

// GetClaims 从Gin上下文获取已解析的JWT声明
// 这通常是在中间件已经解析并验证了accessToken之后使用
func GetClaims(c *gin.Context) *MyClaims {
	// 从Gin上下文获取名为"claims"的值
	_claims, ok := c.Get("claims")
	if !ok {
		// 如果没有找到claims，返回nil
		return nil
	}
	// 类型转换为*MyClaims
	claims, ok := _claims.(*MyClaims)
	if !ok {
		// 如果类型转换失败，返回nil
		return nil
	}
	// 返回转换成功的声明对象
	return claims
}
