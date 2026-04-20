package cron_service

import (
	"StarDreamerCyberNook/global"
	"context"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

func GetLock() {
	ctx := context.Background()
	lockKey := "cron_lock"
	myAddr := global.Config.System.Addr()

	// 尝试获取锁
	// 注意：go-redis 的 SetArgs 在 NX 模式下，如果成功，err 为 nil
	// 如果因为 NX (key存在) 导致失败，err 也是 nil，但 val 会是 "" (空字符串)
	val, err := global.RedisTimeCache.SetArgs(ctx, lockKey, myAddr, redis.SetArgs{
		Mode: "NX",
		TTL:  30 * time.Minute,
	}).Result()

	if err != nil && err != redis.Nil {
		logrus.Errorf("Redis 连接异常，获取锁失败: %v", err)
		return
	}

	// 2. 判断是否抢到了锁
	// 如果 val 是 "OK" (或者非空，取决于具体驱动版本，但在 NX 成功时通常有值)
	// 最稳妥的方式是看 val 是否为 "OK" 或者 err 是否为 nil 且 val 不为空
	if val == "OK" {
		logrus.Info("成功获取定时任务锁，开始执行...")
		return
	}

	// 3. 没抢到锁 (val 为空)
	logrus.Info("定时任务锁已被占用，跳过本次启动")
}
