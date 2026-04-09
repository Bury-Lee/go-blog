package core

import (
	"StarDreamerCyberNook/global"

	"github.com/olivere/elastic/v7"
	"github.com/sirupsen/logrus"
)

func InitElasticSearch() *elastic.Client { //TODO:升级到v8
	es := global.Config.ES
	if es.Url == "" {
		logrus.Errorf("你es忘记填写地址了")
		return nil
	}
	client, err := elastic.NewClient(
		elastic.SetURL(es.Url),
		elastic.SetSniff(false),
		elastic.SetBasicAuth(es.UserName, es.Password),
	)
	if err != nil {
		logrus.Errorf("es初始化失败%s", err.Error())
		panic("ES初始化失败") //服务应该一定要上es
	}
	logrus.Info("es已连接")
	return client
}
