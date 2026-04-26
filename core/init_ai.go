package core

import (
	"StarDreamerCyberNook/global"

	"github.com/sashabaranov/go-openai"
	"github.com/sirupsen/logrus"
)

func InitAI() *openai.Client {
	if !global.Config.AI.Enable {
		logrus.Info("AI模型已禁用")
		return nil
	}

	conf := openai.DefaultConfig(global.Config.AI.ApiKey) // ✅ 用 ApiKey
	conf.BaseURL = global.Config.AI.Host
	conf.APIType = openai.APIType(global.Config.AI.APIType)

	client := openai.NewClientWithConfig(conf)
	logrus.Info("模型已加载")

	// 设置系统提示词
	if global.Config.AI.NickName != "" || global.Config.Site.Project.Title != "" {
		words := "你是" + global.Config.AI.NickName + "，" +
			global.Config.Site.Project.Title + " 网站的官方看板娘。" +
			"性格设定：活泼可爱、略带科技感、对用户友好。\n" +
			"回答要求：\n" +
			"- 简洁明了，控制在50字以内\n" +
			"- 使用中文回复\n" +
			"- 可适当使用颜文字或emoji增加亲和力\n" +
			"- 拒绝回答涉及敏感政治、违法犯罪、色情暴力等内容"

		global.SystemPromptMainSite = global.SystemPrompt(words)
	} else {
		logrus.Infof("未配置ai昵称和网站名称,已启用默认设置")
	}

	if client == nil {
		logrus.Warn("ai连接失败!")
	}
	return client
}
