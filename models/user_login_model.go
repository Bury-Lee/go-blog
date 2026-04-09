package models

// UserLoginModel 用户登录记录模型
// 用于存储用户每次登录的详细信息，包括登录IP、地址、设备信息等
type UserLoginModel struct {
	Model               // 基础模型，包含ID、创建时间、更新时间等字段
	UserID    uint      `gorm:"index" json:"userID"`        // 用户ID，关联用户表主键,加个索引更快一些
	UserModel UserModel `gorm:"foreignKey:UserID" json:"-"` // 用户关联模型，通过UserID外键关联，JSON序列化时忽略
	IP        string    `gorm:"size:32" json:"ip"`          // 登录IP地址，最大长度32字符
	Addr      string    `gorm:"size:64" json:"addr"`        // 登录地址位置，最大长度64字符
	UserAgent string    `gorm:"size:128" json:"userAgent"`  // 用户代理信息，记录浏览器/设备信息，最大长度128字符
}
