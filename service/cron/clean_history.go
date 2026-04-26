package cron_service

import (
	"StarDreamerCyberNook/global"
	"StarDreamerCyberNook/models"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// 清理超过一个月的浏览记录
var cleanIng sync.Mutex

func SyncCleanHistory() {
	// 尝试加锁，避免 cron 并发执行
	if !cleanIng.TryLock() {
		return
	}
	defer cleanIng.Unlock()

	// 计算过期时间（30天前）
	expireTime := time.Now().AddDate(0, 0, -30)

	for {
		logrus.Infof("开始清理超过一个月的浏览记录")

		// 执行删除（每次最多删除50条）
		tx := global.DB.
			Where("created_at < ?", expireTime).
			Limit(50).
			Delete(&models.UserArticleHistoryModel{})

		if tx.Error != nil {
			logrus.Errorf("清理失败: %v", tx.Error)
			return
		}

		affected := tx.RowsAffected

		// 如果本次一条都没删，说明已经清理完了
		if affected == 0 {
			logrus.Infof("浏览记录清理完成")
			return
		}

		// logrus.Infof("本次清理 %d 条记录", affected)

		// 每批之间休眠，避免数据库压力过大
		time.Sleep(5 * time.Second)
	}
}
