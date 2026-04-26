package cron_service

import (
	"StarDreamerCyberNook/global"
	"StarDreamerCyberNook/models"
	"log"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// 清理超过一个月的浏览记录
var cleanIng sync.Mutex

func SyncCleanHistory() {
	// 避免重复执行
	if !cleanIng.TryLock() {
		return
	}
	cleanIng.Unlock()

	// 每次删除50条数据，避免负载过高
	for {
		logrus.Infof("开始清理超过一个月的浏览记录")
		var count int64
		// 查询符合条件的数据数量
		global.DB.Model(&models.UserArticleHistoryModel{}).Where("created_at < ?", time.Now().AddDate(0, 0, -30)).Count(&count)

		// 如果没有符合条件的数据，退出循环
		if count == 0 {
			cleanIng.Unlock()
			logrus.Infof("浏览记录清理完成")
			return
		}

		// 删除50条数据
		if err := global.DB.Where("created_at < ?", time.Now().AddDate(0, 0, -30)).
			Limit(50).Delete(&models.UserArticleHistoryModel{}).Error; err != nil {
			log.Printf("Failed to delete history records: %v", err)
			cleanIng.Unlock()
			return
		}

		// 删除完一批后休眠5秒钟
		time.Sleep(5 * time.Second)
	}
}

//一会处理分页查询的事,包括Likelist匹配
