package ai_api

import (
	"StarDreamerCyberNook/common/response"
	"StarDreamerCyberNook/global"
	"context"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sashabaranov/go-openai"
)

type UserAIRequest struct { //TODO:想想怎么优化这个请求体,让用户不能传入被修改过的system_prompt
	Messages  []openai.ChatCompletionMessage `json:"messages"`   // 对话历史消息
	UserInput string                         `json:"user_input"` // 用户输入
	ImageID   uint                           `json:"image_ID"`   // 图片ID（可选）//TODO:加入图片的支持功能
	Model     string                         `json:"model"`      // 模型名称（可选，默认使用全局配置）
}

// AIResponse 表示AI响应的结果
type AIResponse struct {
	Success bool   `json:"success"`
	Content string `json:"content"`
	Error   string `json:"error"`
}

func (AIApi) Chat(c *gin.Context) {
	var req UserAIRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMsg("请求参数错误", c)
		return
	}

	//扫描用户输入,只允许包含AI和用户角色的输入
	for _, v := range req.Messages {
		if v.Role != openai.ChatMessageRoleUser && v.Role != openai.ChatMessageRoleAssistant {
			response.FailWithMsg("对话历史消息中包含非用户或助手角色", c)
			return
		}
	}
	if strings.Contains(req.UserInput, global.SystemPromptMainSite.String()) {
		response.FailWithMsg("用户输入中包含system_prompt", c)
		return
	}

	ctx := context.Background()

	//由于这个函数的特殊性,需要保留原本的逻辑,不使用ai的service

	// 构建完整的消息列表
	var messages []openai.ChatCompletionMessage

	// 添加系统提示词作为第一条消息

	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleSystem,
		Content: global.SystemPromptMainSite.String(), //使用看板娘的提示词
	})

	// 添加对话历史消息
	messages = append(messages, req.Messages...)

	// 添加用户输入作为最后一条消息
	userMessage := openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: req.UserInput,
	}
	messages = append(messages, userMessage)

	// 创建非流式请求
	resp, err := global.AIClient.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model:       global.Config.AI.Model,
			Temperature: global.Config.AI.Temperature,
			MaxTokens:   global.Config.AI.MaxTokens,
			Messages:    messages,
		},
	)

	if err != nil {
		response.FailWithMsg(fmt.Sprintf("创建聊天完成时出错: %v", err), c)
		return
	}

	if len(resp.Choices) == 0 {
		response.FailWithMsg("AI未返回有效结果", c)
		return
	}

	content := resp.Choices[0].Message.Content

	// 返回一次性结果
	result := AIResponse{
		Success: true,
		Content: content,
		Error:   "",
	}

	response.OkWithData(result, c)
}
