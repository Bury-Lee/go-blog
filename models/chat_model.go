package models

import "time"

type ChatModel struct {
	Model
	SendUserID    uint            `json:"sendUserID"`                           // 发送用户ID
	SendUserModel UserModel       `gorm:"foreignKey:SendUserID"  json:"-"`      // 发送用户模型
	RevUserID     uint            `json:"revUserID"`                            // 接收用户ID
	RevUserModel  UserModel       `gorm:"foreignKey:RevUserID"  json:"-"`       // 接收用户模型
	MsgType       ChatMessageType `json:"msgType"`                              // 消息类型,预备以后扩展
	Msg           ChatMsg         `gorm:"type:text;serializer:json" json:"msg"` // 消息内容
}
type ChatMessageType uint8

const ( //使用位操作符表示消息类型,方便以后扩展
	TextMsgType     ChatMessageType = 1 << iota // 1
	ImageMsgType                                // 2
	MarkdownMsgType                             // 4
)

type UserChatActionModel struct { //预备使用的表
	Model
	UserID   uint `json:"userID"`   // 操作者ID
	ChatID   uint `json:"chatID"`   // 聊天记录ID
	IsRead   bool `json:"isRead"`   // 是否已读
	IsDelete bool `json:"isDelete"` // 是否删除
}

type TextMsg struct {
	Content string `json:"content"`
}
type ImageMsg struct {
	Src string `json:"src"`
}

type MarkdownMsg struct {
	Content string `json:"content"`
}

type ChatMsg struct {
	TextMsg     *TextMsg     `json:"textMsg,omitempty" gorm:"serializer:json"`
	ImageMsg    *ImageMsg    `json:"imageMsg,omitempty" gorm:"serializer:json"`
	MarkdownMsg *MarkdownMsg `json:"markdownMsg,omitempty" gorm:"serializer:json"`
}

// 会话表
type SessionModel struct {
	Model
	// 两个会话用户ID拼接就可以唯一标识一个会话
	// 分你对我的会话和我对你的会话,这样查询和统计未读消息的时候就可以很方便的查询到

	UniqueID  string    `json:"uniqueId" gorm:"index:idx_unique_id,unique"` // 唯一索引,组织方式为我对你的会话,所以ID就是我的用户ID+对方用户ID,组织成这样.可以改为定长字符串
	UserID    uint      `json:"userId"`                                     //接收方的用户ID,既然这样,就只查询对方的会话信息就好了
	UserModel UserModel `gorm:"foreignKey:UserID"  json:"-"`                // 用户模型,方便查询用户信息,如昵称,头像等

	LastMessage     ChatModel       `json:"lastMessage"`     // 最后一条消息内容摘要
	LastMessageType ChatMessageType `json:"lastMessageType"` // 消息类型
	LastMessageTime time.Time       `json:"lastMessageTime"` // 最后一条消息时间
	UnreadCount     int             `json:"unreadCount"`     // 未读消息数量
	IsRead          bool            `json:"isRead"`          // 是否已读
}
