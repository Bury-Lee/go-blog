package cron_service

import (
	"time"

	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
)

func Cron() {
	//抢锁执行

	var crontab *cron.Cron
	timezone, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		logrus.Warnf("定时任务无法设置时区,已使用UTC时区:%v", err)
		crontab = cron.New(cron.WithSeconds(), cron.WithLocation(time.UTC))
	} else {
		crontab = cron.New(cron.WithSeconds(), cron.WithLocation(timezone))
	}

	//debug使用
	// crontab.AddFunc("*/10 * * * * *", SyncArticle)//debug:10秒一次
	// crontab.AddFunc("*/4 * * * * *", SyncArticle)
	// crontab.AddFunc("*/4 * * * * *", SyncComment)

	// 每10分钟，0秒触发
	crontab.AddFunc("0 */10 * * * *", GetLock) //30分钟的锁,每30分钟抢一次

	// 每10分钟
	crontab.AddFunc("0 1-59/10 * * * *", SyncArticle) //文章数据同步

	// 每10分钟
	crontab.AddFunc("0 6-59/10 * * * *", SyncComment) //评论数据同步

	crontab.Start()
}
