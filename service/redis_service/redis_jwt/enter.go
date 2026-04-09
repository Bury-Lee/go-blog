// service/redis_service/redis_jwt/enter.go
package redis_jwt

import (
	"StarDreamerCyberNook/global"
	jwts "StarDreamerCyberNook/utils/jwts"
	"context"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

type BlackType int8

const ( //分为手动下线,管理员拉黑,多设备登录被挤下线
	UserBlackType   BlackType = 1 //用户手动下线
	AdminBlackType  BlackType = 2 //管理员拉黑
	DeviceBlackType BlackType = 3 //多设备登录被挤下线
)

func (b BlackType) String() string {
	switch b {
	case UserBlackType:
		return "已下线"
	case AdminBlackType:
		return "已被管理员拉黑"
	case DeviceBlackType:
		return "因多设备登录,已下线"
	default:
		return "未知"
	}
}
func FromString(blackType string) BlackType {
	switch blackType {
	case "1":
		return UserBlackType
	case "2":
		return AdminBlackType
	case "3":
		return DeviceBlackType
	default:
		return 0
	}
}

// 记录jwt的黑名单
func TokenBlack(accessToken string, refreshToken string, blackType BlackType) {
	//将token加入黑名单,包括access token和refresh token
	ctx := context.Background()
	//可能会有的激进做法:把access token也加入黑名单,对于性能消耗可能有点高
	// AccessTokenkey := fmt.Sprintf("token_black_%s", accessToken)
	RefreshTokenkey := fmt.Sprintf("token_black_%s", refreshToken)

	// //解析通行token
	// claim, err := jwts.ParseAccessToken(accessToken)
	// if err != nil || claim == nil {
	// 	logrus.Errorf("Token解析失败: %v", err)
	// 	return
	// }
	//解析刷新token
	refreshClaim, err := jwts.ParseRefreshToken(refreshToken)
	if err != nil || refreshClaim == nil {
		logrus.Errorf("Token解析失败: %v", err)
		return
	}
	//设置黑名单时间
	refreshSecond := refreshClaim.ExpiresAt.Time.Unix() - time.Now().Unix()
	// accessSecond := claim.ExpiresAt.Time.Unix() - time.Now().Unix()

	// res, err := global.RedisTimeCache.Set(ctx, AccessTokenkey, blackType, time.Duration(accessSecond)*time.Second).Result()
	// if err != nil {
	// 	logrus.Errorf("Redis Set失败: %s,%v", res, err)
	// 	return
	// }
	res, err := global.RedisTimeCache.Set(ctx, RefreshTokenkey, blackType, time.Duration(refreshSecond)*time.Second).Result()
	if err != nil {
		logrus.Errorf("Redis Set失败: %s,%v", res, err)
		return
	}
}

func HasTokenBlack(Token string) (BlackType, bool) {
	//判断通行token是否在黑名单中
	// AccessTonkenkey := fmt.Sprintf("token_black_%s", accessToken)
	RefreshTokenkey := fmt.Sprintf("token_black_%s", Token)

	ctx := context.Background() //凑数的空白ctx
	//可能会有激进的做法:判断刷新token是否在黑名单中
	// 判断通行token是否在黑名单中
	// res, err := global.RedisTimeCache.Get(ctx, AccessTonkenkey).Result()
	// if err != nil {
	// 	logrus.Errorf("Redis Get失败:%s, %v", AccessTonkenkey, err) //也许没必要记录日志
	// 	return 0, false
	// }

	//判断刷新token是否在黑名单中
	res, err := global.RedisTimeCache.Get(ctx, RefreshTokenkey).Result()
	if err != nil && err != redis.Nil { //redis.Nil表示key不存在
		logrus.Errorf("Redis Get失败:%s, %v", RefreshTokenkey, err) //也许没必要记录日志
		return 0, false
	}
	blk := FromString(res) //理论上都是同一种黑名单的类型
	return blk, blk != 0
}

func HasTokenByGin(c *gin.Context) (BlackType, bool) {
	token := c.GetHeader("token")
	if token == "" {
		token = c.Query("token")
	}
	return HasTokenBlack(token)
}
