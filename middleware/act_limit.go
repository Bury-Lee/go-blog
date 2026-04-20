package middleware

import (
	"StarDreamerCyberNook/common/response"
	"StarDreamerCyberNook/global"
	"context"
	"encoding/json"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

// TODO:升级一下,改为对于不同的路由使用不同的限制,比如登录接口可以限制为每分钟5次,其他接口可以限制为每分钟100次,还可以根据用户ID进行限制,比如每个用户每天只能修改密码3次之类的
type ActTimes struct {
	IP        string `json:"ip"`         //ip地址//其实可以不用记啊,不过为了方便管理员看看是谁在攻击网站,就记着
	Times     int    `json:"times"`      //请求次数
	BlockTime int64  `json:"block_time"` //封禁时间
}

func ActLimitMiddleware(c *gin.Context) {
	IP := c.ClientIP()
	now := time.Now().Unix()
	ctx := context.Background()
	count, err := global.RedisTimeCache.Get(ctx, IP).Result()
	if err != nil {
		if err == redis.Nil {
			first := ActTimes{IP: IP, Times: 1, BlockTime: 0}
			if data, marshalErr := json.Marshal(first); marshalErr == nil {
				_ = global.RedisTimeCache.Set(ctx, IP, data, time.Minute).Err()
			}
		}
		c.Next()
		return
	}

	var actTimes ActTimes
	if err = json.Unmarshal([]byte(count), &actTimes); err != nil {
		reset := ActTimes{IP: IP, Times: 1, BlockTime: 0}
		if data, marshalErr := json.Marshal(reset); marshalErr == nil {
			_ = global.RedisTimeCache.Set(ctx, IP, data, time.Minute).Err()
		}
		c.Next()
		return
	}

	if actTimes.BlockTime > now {
		response.FailWithMsg("请求过于频繁", c)
		c.Abort()
		return
	}

	if actTimes.Times >= 64 {
		blocked := ActTimes{IP: IP, Times: 0, BlockTime: now + 60}
		ctx := context.Background()
		if data, marshalErr := json.Marshal(blocked); marshalErr == nil {
			_ = global.RedisTimeCache.Set(ctx, IP, data, time.Minute*2).Err()
		}
		response.FailWithMsg("请求过于频繁", c)
		c.Abort()
		return
	}
	ttl, err := global.RedisTimeCache.TTL(ctx, IP).Result()
	if err != nil || ttl <= 0 {
		ttl = time.Second * 30
	}

	next := ActTimes{IP: IP, Times: actTimes.Times + 1, BlockTime: actTimes.BlockTime}
	if data, marshalErr := json.Marshal(next); marshalErr == nil {
		_ = global.RedisTimeCache.Set(ctx, IP, data, ttl).Err()
	}

	c.Next()
}

func EmailSendLimitMiddleware(c *gin.Context) { //限制一分钟只能发一次邮箱
	IP := c.ClientIP()
	ctx := context.Background()
	_, err := global.RedisTimeCache.Get(ctx, IP).Result()
	if err != redis.Nil {
		if err == redis.Nil {
			first := ActTimes{IP: IP, Times: 1, BlockTime: 0}
			if data, marshalErr := json.Marshal(first); marshalErr == nil {
				_ = global.RedisTimeCache.Set(ctx, IP, data, time.Minute).Err()
			}
		}
		c.Next() //如果没有记录,说明是第一次发送邮件,直接放行
		return
	} else if err == nil {
		c.Abort() //查到了,说明已经发送过邮件了,不能重复发送
		response.FailWithMsg("请求过于频繁", c)
		return
	} else { //未知错误
		c.Abort()
		logrus.Errorf("邮箱发送发生错误:%v", err)
		response.FailWithMsg("未知错误,请联系管理员", c)
		return
	}

}

func ImgPostLimitMiddleware(c *gin.Context) {
	IP := c.ClientIP()
	now := time.Now().Unix()
	ctx := context.Background()
	count, err := global.RedisTimeCache.Get(ctx, IP).Result()
	if err != nil {
		if err == redis.Nil {
			first := ActTimes{IP: IP, Times: 1, BlockTime: 0}
			if data, marshalErr := json.Marshal(first); marshalErr == nil {
				_ = global.RedisTimeCache.Set(ctx, IP, data, time.Minute).Err()
			}
		}
		c.Next()
		return
	}

	var actTimes ActTimes
	if err = json.Unmarshal([]byte(count), &actTimes); err != nil {
		reset := ActTimes{IP: IP, Times: 1, BlockTime: 0}
		if data, marshalErr := json.Marshal(reset); marshalErr == nil {
			_ = global.RedisTimeCache.Set(ctx, IP, data, time.Minute).Err()
		}
		c.Next()
		return
	}

	if actTimes.BlockTime > now {
		response.FailWithMsg("请求过于频繁", c)
		c.Abort()
		return
	}

	if actTimes.Times >= 20 { //限制一分钟只能发20次图片
		blocked := ActTimes{IP: IP, Times: 0, BlockTime: now + 60}
		ctx := context.Background()
		if data, marshalErr := json.Marshal(blocked); marshalErr == nil {
			_ = global.RedisTimeCache.Set(ctx, IP, data, time.Minute*2).Err()
		}
		response.FailWithMsg("请求过于频繁", c)
		c.Abort()
		return
	}
	ttl, err := global.RedisTimeCache.TTL(ctx, IP).Result()
	if err != nil || ttl <= 0 {
		ttl = time.Second * 30
	}

	next := ActTimes{IP: IP, Times: actTimes.Times + 1, BlockTime: actTimes.BlockTime}
	if data, marshalErr := json.Marshal(next); marshalErr == nil {
		_ = global.RedisTimeCache.Set(ctx, IP, data, ttl).Err()
	}

	c.Next()
}
