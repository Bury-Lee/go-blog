package ai_service

import (
	"StarDreamerCyberNook/global"
	"context"
	"errors"

	"github.com/sashabaranov/go-openai"
	"github.com/sirupsen/logrus"
)

// 输入内容和提示词,返回模型回复和错误信息
func CreateSingleReply(content string, prompt string) (string, error) {
	ctx := context.Background()

	// 构建完整的消息列表
	var messages []openai.ChatCompletionMessage

	// 添加系统提示词作为第一条消息

	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleSystem,
		Content: prompt,
	})

	// 添加对话历史消息
	messages = append(messages,
		openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleUser,
			Content: content,
		})

	// 创建非流式请求
	res, err := global.AIClient.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model:       global.Config.AI.Model,
			Temperature: global.Config.AI.Temperature,
			MaxTokens:   global.Config.AI.MaxTokens,
			Messages:    messages,
		},
	)
	if err != nil {
		if global.Config.System.RunMode == "debug" {
			logrus.Errorf("创建单条回复失败:%v", err)
		}
		return "", err
	}
	if global.Config.System.RunMode == "debug" {
		logrus.Debugf("输入内容:%s,提示词:%s", content, prompt)
		if len(res.Choices) == 0 {
			logrus.Debugf("模型空回复")
		}
		logrus.Debugf("模型回复:%s", res.Choices[0].Message.Content)
	}
	if len(res.Choices) == 0 {
		// logrus.Error("模型空回复")
		return "", errors.New("模型空回复")
	}
	return res.Choices[0].Message.Content, nil
}
