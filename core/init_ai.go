package core

import (
	"StarDreamerCyberNook/global"

	"github.com/sashabaranov/go-openai"
	"github.com/sirupsen/logrus"
)

func InitAI() (*openai.Client, error) {
	if !global.Config.AI.Enable {
		logrus.Info("AI模型已禁用")
		return nil, nil
	}

	if global.Config.AI.Model == "local" {
		conf := openai.DefaultConfig("")
		conf.BaseURL = global.Config.AI.Host
		client := openai.NewClientWithConfig(conf)
		logrus.Info("模型已加载")
		return client, nil
	}

	//TODO:加入可用性检测

	//TODO:根据不同厂家返回不同的AI模型客户端
	return nil, nil
}
