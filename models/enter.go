// models/enter.go
// 基础模型定义
// 提供所有数据库模型的公共字段，作为其他模型的嵌入结构体
package models

import (
	"time"
)

// Model 基础模型结构体
// 所有数据库模型都会嵌入此结构体，包含主键ID和创建更新时间
// ID: 主键ID，自增
// CreatedAt: 记录创建时间
// UpdatedAt: 记录更新时间
type Model struct {
	ID        uint      `gorm:"primarykey" json:"id"` // 主键ID
	CreatedAt time.Time `json:"createdAt"`            // 创建时间
	UpdatedAt time.Time `json:"updatedAt"`            // 更新时间
}
