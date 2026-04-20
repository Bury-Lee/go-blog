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
	global.LocalAIClient = core.InitAI()                          //初始化AI模型
	if global.Config.ObjectStorage.Enable {
		logrus.Infof("对象存储已启用,存储桶:%s", global.Config.ObjectStorage.Bucket)
		global.StorageClient = core.InitClient()
	} else {
		logrus.Info("对象存储未启用,使用本地存储")
	}

	flags.Run() //运行命令行参数
	//debug模式下打印配置
	if global.Config.System.RunMode == "debug" {
		configDebug, err := json.MarshalIndent(global.Config, "", "  ")
		if err != nil {
			logrus.Error("反序列化失败:", err)
			return
		}
		logrus.Debug(string(configDebug))
	}

	if global.Config.System.Cron { //当在分布式环境下时建议只启用一个实例的定时任务,避免导致并发下的多写问题
		// 启动定时任务
		go cron_service.Cron()
	}

	router := router.InitRouter() //注册路由
	// router.Run(global.Config.System.Addr())
	server := core.InitServer(router)
	err := server.ListenAndServe()
	if err != nil {
		logrus.Error("服务器启动失败:", err)
		return
	}
}
