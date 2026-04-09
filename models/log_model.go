// models/log_model.go
// 日志模型定义
// 记录系统操作日志、用户登录日志等信息
package models

import "StarDreamerCyberNook/models/enum"

// LogModel 日志模型
// 存储系统日志信息，包括操作日志、登录日志等
type LogModel struct {
	Model
	LogType     enum.LogType   `json:"logType"`                    // 日志类型枚举
	Title       string         `gorm:"size:64" json:"title"`       // 日志标题
	Content     string         `json:"content"`                    // 日志内容
	Level       enum.LogLevel  `json:"level"`                      // 日志级别枚举
	UserID      uint           `json:"userID"`                     // 关联用户ID
	UserModel   UserModel      `gorm:"foreignKey:UserID" json:"-"` // 关联用户信息
	IP          string         `gorm:"size:32" json:"ip"`          // IP地址
	Addr        string         `gorm:"size:64" json:"addr"`        // 地理位置
	IsRead      bool           `json:"isRead"`                     // 是否已读
	LoginStatus bool           `json:"loginStatus"`                // 登录状态
	LoginType   enum.LoginType `json:"loginType"`                  // 登录类型枚举
	ServiceName string         `gorm:"size:32" json:"serviceName"` // 服务名称
}
