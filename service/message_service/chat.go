package message_service

import (
	"StarDreamerCyberNook/global"
	"StarDreamerCyberNook/models"
	xss_filter "StarDreamerCyberNook/utils/XSSfilter"

	"github.com/sirupsen/logrus"
)

// ToChat A给B发消息
func ToChat(A, B uint, msgType models.ChatMessageType, msg models.ChatMsg) {
	err := global.DB.Create(&models.ChatModel{
		SendUserID: A,
		RevUserID:  B,
		MsgType:    msgType,
		Msg:        msg,
	}).Error
	if err != nil {
		logrus.Errorf("对话创建失败 %s", err)
	}
}

func ToTextChat(A, B uint, content string) {
	ToChat(A, B, models.TextMsgType, models.ChatMsg{
		TextMsg: &models.TextMsg{
			Content: content,
		},
	})
}

func ToImageChat(A, B uint, src string) {
	ToChat(A, B, models.ImageMsgType, models.ChatMsg{
		ImageMsg: &models.ImageMsg{
			Src: src,
		},
	})
}

func ToMarkdownChat(A, B uint, content string) {
	// 过滤xss
	xssFilter := xss_filter.NewXSSFilter()
	filterContent := xssFilter.Sanitize(content)
	ToChat(A, B, models.MarkdownMsgType, models.ChatMsg{
		MarkdownMsg: &models.MarkdownMsg{
			Content: filterContent,
		},
	})
}
