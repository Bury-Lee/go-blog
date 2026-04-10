package core

import (
	"StarDreamerCyberNook/global"
	"log"
	"os"

	"github.com/olivere/elastic/v7"
	"github.com/sirupsen/logrus"
)

func InitElasticSearch() *elastic.Client { //TODO:升级到v8
	es := global.Config.ES
	if es.Url == "" {
		logrus.Errorf("你es忘记填写地址了")
		return nil
	}
	var client *elastic.Client
	var err error
	if global.Config.System.RunMode == "debug" {
		// TODO:输出好长,考虑把这里的调试注释掉
		// 创建一个 Logger 实例（使用标准库的 log.Logger，输出到控制台）
		ESLog := log.New(os.Stdout, "[ES-Debug]", log.LstdFlags)
		// 启用跟踪日志
		// 创建 Elasticsearch 客户端并设置跟踪日志
		client, err = elastic.NewClient(
			elastic.SetURL(es.Url),
			elastic.SetSniff(false),
			elastic.SetBasicAuth(es.UserName, es.Password),
			elastic.SetTraceLog(ESLog),
		)
		logrus.Debug("es调试模式已开启")
	} else {
		client, err = elastic.NewClient(
			elastic.SetURL(es.Url),
			elastic.SetSniff(false),
			elastic.SetBasicAuth(es.UserName, es.Password),
		)
	}
	if err != nil {
		logrus.Errorf("es初始化失败%s", err.Error())
		panic("ES初始化失败") //服务应该一定要上es
	}
	logrus.Info("es已连接")
	return client
}
