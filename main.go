package main

import (
	"StarDreamerCyberNook/core"
	"StarDreamerCyberNook/flags"
	"StarDreamerCyberNook/global"
	router "StarDreamerCyberNook/router"
	cron_service "StarDreamerCyberNook/service/cron"
	"encoding/json"

	"github.com/sirupsen/logrus"
)

func main() {
	flags.Parse()                                                 //解析命令行参数
	global.Config = core.ReadConf()                               //读取文件
	core.InitLogrus()                                             //初始化日志
	global.IPsearcher = core.InitIPDB()                           //初始化ip地址库
	global.DB = core.InitDB()                                     //初始化数据库
	global.RedisTimeCache, global.RedisHotPool = core.InitRedis() //初始化redis
	global.ES = core.InitElasticSearch()                          //初始化elasticsearch
	global.LocalAIClient, _ = core.InitAI()                       //初始化AI模型

	flags.Run() //运行命令行参数
	//debug模式下打印配置
	if global.Config.System.RunMode == "debug" {
		configDebug, err := json.MarshalIndent(global.Config, "", "  ")
		if err != nil {
			logrus.Error("Failed to marshal config:", err)
			return
		}
		logrus.Debug(string(configDebug))
	}

	// 启动定时任务
	go cron_service.CronArticle()

	router.Run() //运行路由
}
