package cron_service

import (
	"time"

	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
)

func CronArticle() {

	var crontab *cron.Cron
	timezone, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		logrus.Warnf("无法设置时区,已使用UTC时区:%v", err)
		crontab = cron.New(cron.WithSeconds(), cron.WithLocation(time.UTC))
	} else {
		crontab = cron.New(cron.WithSeconds(), cron.WithLocation(timezone))
	}

	//debug使用
	// crontab.AddFunc("*/10 * * * * *", SyncArticle)//debug:10秒一次
	// crontab.AddFunc("*/4 * * * * *", SyncArticle)
	// crontab.AddFunc("*/4 * * * * *", SyncComment)
	//以上的记得删除

	// 注册一个每10分钟执行一次数据同步的任务
	crontab.AddFunc("0 */10 * * * *", SyncArticle)
	time.Sleep(time.Minute * 5) //交叉执行任务
	crontab.AddFunc("0 */10 * * * *", SyncComment)

	crontab.Start()
}
