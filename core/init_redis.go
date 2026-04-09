package core

import (
	"StarDreamerCyberNook/global"
	"context"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

func InitRedis() (*redis.Client, *redis.Client) {
	ctx := context.Background()

	r := global.Config.RedisStatic
	redisStaticDB := redis.NewClient(&redis.Options{
		Addr:     r.Addr,
		Password: r.Password,
		DB:       r.DB,
	})
	_, err := redisStaticDB.Ping(ctx).Result()
	if err != nil {
		logrus.Fatalf("redisStatic静态库连接失败 %s", err)
		return nil, nil
	}
	logrus.Info("redisStatic静态库连接成功")

	// 动态数据缓存库
	r = global.Config.RedisDynamic
	redisDynamicDB := redis.NewClient(&redis.Options{
		Addr:     r.Addr,
		Password: r.Password,
		DB:       r.DB,
	})
	_, err = redisDynamicDB.Ping(ctx).Result()
	if err != nil {
		logrus.Fatalf("redisDynamic动态库连接失败 %s", err)
		return nil, nil
	}
	logrus.Info("redisDynamic动态库连接成功")

	// 检查Redis内存淘汰策略
	cfg, err2 := redisDynamicDB.ConfigGet(ctx, "maxmemory-policy").Result()
	if err2 != nil {
		logrus.Warnf("读取redisDynamic动态库内存策略失败: %s", err2)
	} else if policy, ok := cfg["maxmemory-policy"]; ok {
		if policy != "allkeys-lru" {
			logrus.Warnf("redisDynamic动态库内存策略非allkeys-lru，当前为: %s", policy)
		}
	} else {
		logrus.Warn("redisDynamic动态库未返回maxmemory-policy配置")
	}

	return redisStaticDB, redisDynamicDB
}
