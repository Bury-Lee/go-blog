package models

type MessageType uint8

// 分点赞我的文章,评论我的文章,@我,收藏我的文章,私信和系统通知
const (
	MessageTypeComment MessageType = 1 //评论我的文章
	MessageTypeReply   MessageType = 2 //回复我的评论
	MessageTypeDigg    MessageType = 3 //点赞我的文章/评论
	MessageTypeCollect MessageType = 4 //收藏我的文章
	MessageTypePrivate MessageType = 5 //私信
	MessageTypeSystem  MessageType = 6 //系统通知
	MessageTypeAt      MessageType = 7 //@我
)

func (m MessageType) String() string {
	change := []string{
		"评论通知",
		"回复通知",
		"点赞通知",
		"收藏通知",
		"私信通知",
		"系统通知",
		"有人@我",
	}
	return change[m]
}

type MessageModel struct {
	Model
	Type               MessageType `json:"type"`
	RevUserID          uint        `json:"revUserID"`          // 接收人的id
	ActionUserID       uint        `json:"ActionUserID"`       // 发送人的id
	ActionUserNickname string      `json:"actionUserNickname"` // 发送人昵称
	ActionUserAvatar   string      `json:"actionUserAvatar"`   // 发送人头像
	Title              string      `json:"title"`              // 消息标题
	Content            string      `json:"content"`            // 消息内容
	ArticleID          uint        `json:"articleID"`          // 文章id
	ArticleTitle       string      `json:"articleTitle"`       // 文章标题
	CommentID          uint        `json:"commentID"`          // 评论id
	LinkTitle          string      `json:"linkTitle"`          // 链接标题
	LinkHref           string      `json:"linkHref"`           // 链接href
	IsRead             bool        `json:"isRead"`             // 是否已读
}
