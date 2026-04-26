package utils

import (
	"StarDreamerCyberNook/global"
	"StarDreamerCyberNook/models"
	"StarDreamerCyberNook/utils"
	"context"
	"fmt"
	"strings"

	openai "github.com/sashabaranov/go-openai"
)

// AIResponse 表示AI响应的结果
type AIResponse struct {
	Success bool   `json:"success"`
	Content string `json:"content"`
	Error   string `json:"error"`
}

// AIRequest 表示AI请求的参数
type AIRequest struct {
	Messages     []openai.ChatCompletionMessage `json:"messages"`      // 对话历史消息
	UserInput    string                         `json:"user_input"`    // 用户输入
	ImagePath    uint                           `json:"image_ID"`      // 图片路径（可选）
	SystemPrompt string                         `json:"system_prompt"` // 系统提示词（可选）
	Model        string                         `json:"model"`         // 模型名称（可选，默认使用全局配置）
}

// ChatWithAI 接收AIRequest结构体，返回AIResponse
func ChatWithAI(request AIRequest) AIResponse { //保留函数
	// 配置客户端
	client := global.AIClient
	ctx := context.Background()

	// 构建消息列表
	messages := make([]openai.ChatCompletionMessage, len(request.Messages))

	// 复制现有消息
	copy(messages, request.Messages)

	// 添加系统提示词（如果存在）
	if request.SystemPrompt != "" {
		systemMessage := openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleSystem,
			Content: request.SystemPrompt,
		}
		messages = append([]openai.ChatCompletionMessage{systemMessage}, messages...)
	}

	// 处理用户输入和图片
	var userMessage openai.ChatCompletionMessage

	if request.ImagePath != 0 {
		// 包含图片的消息
		var imageModel models.ImageModel
		global.DB.First(&imageModel, "id = ?", request.ImagePath)
		base64Image, err := utils.EncodeImageToBase64(imageModel.Path)
		if err != nil {
			return AIResponse{
				Success: false,
				Content: "",
				Error:   fmt.Sprintf("读取图片失败: %v", err),
			}
		}

		var textContent string
		if request.UserInput != "" {
			textContent = request.UserInput
		} else {
			textContent = "请分析这张图片并告诉我你看到了什么。"
		}

		userMessage = openai.ChatCompletionMessage{
			Role: openai.ChatMessageRoleUser,
			MultiContent: []openai.ChatMessagePart{
				{
					Type: openai.ChatMessagePartTypeText,
					Text: textContent,
				},
				{
					Type: openai.ChatMessagePartTypeImageURL,
					ImageURL: &openai.ChatMessageImageURL{
						URL: fmt.Sprintf("data:image/jpeg;base64,%s", base64Image),
					},
				},
			},
		}
	} else {
		// 纯文本消息
		userMessage = openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleUser,
			Content: request.UserInput,
		}
	}

	messages = append(messages, userMessage)

	// 设置模型名称
	model := request.Model
	if model == "" {
		model = "local-model" // 默认模型
	}

	// 创建流式请求
	stream, err := client.CreateChatCompletionStream(
		ctx,
		openai.ChatCompletionRequest{
			Model:    model,
			Messages: messages,
			Stream:   true,
		},
	)

	if err != nil {
		return AIResponse{
			Success: false,
			Content: "",
			Error:   fmt.Sprintf("创建流式聊天完成时出错: %v", err),
		}
	}
	defer stream.Close()

	// 收集AI的完整回复
	var fullResponseContent strings.Builder
	for {
		response, err := stream.Recv()
		if err != nil {
			if err.Error() != "EOF" {
				return AIResponse{
					Success: false,
					Content: "",
					Error:   fmt.Sprintf("读取流时出错: %v", err),
				}
			}
			break
		}

		contentChunk := response.Choices[0].Delta.Content
		if contentChunk != "" {
			fullResponseContent.WriteString(contentChunk)
		}
	}

	return AIResponse{
		Success: true,
		Content: fullResponseContent.String(),
		Error:   "",
	}
}

// SimpleChat 简化的聊天接口，只处理纯文本
func SimpleChat(userInput string, historyMessages []openai.ChatCompletionMessage) AIResponse {
	request := AIRequest{
		Messages:  historyMessages,
		UserInput: userInput,
	}
	return ChatWithAI(request)
}

// ChatWithSystemPrompt 带系统提示词的聊天
func ChatWithSystemPrompt(userInput string, systemPrompt string, historyMessages []openai.ChatCompletionMessage) AIResponse {
	request := AIRequest{
		Messages:     historyMessages,
		UserInput:    userInput,
		SystemPrompt: systemPrompt,
	}
	return ChatWithAI(request)
}
