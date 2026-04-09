// models/global_notification_model.go
// 全局通知模型定义
// 管理系统全局通知消息，支持图标、标题、内容和跳转链接
package models

// GlobalNotificationModel 全局通知模型
// 存储系统全局通知信息，用于向所有用户推送重要消息
// 包含标题、图标、内容文本和跳转链接等完整信息
type GlobalNotificationModel struct {
	Model
	Title   string `gorm:"size:64" json:"title"`    // 通知标题
	Icon    string `gorm:"size:512" json:"icon"`    // 通知图标URL
	Content string `gorm:"size:128" json:"content"` // 通知内容
	Href    string `gorm:"size:512" json:"href"`    // 跳转链接，用户点击通知后跳转
}
