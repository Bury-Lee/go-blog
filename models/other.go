package models

// RemoveRequest 批量删除请求结构体
// 用于接收需要删除的多个ID列表
type RemoveRequest struct {
	IDList []uint `json:"IDList" binding:"required"` // ID列表，需要删除的记录ID数组
}

// IDRequest 单ID请求结构体
// 用于接收单个ID参数的请求，通常用于查询、删除等操作
type IDRequest struct {
	ID uint `uri:"id" binding:"required"` // 记录ID，通过URI路径参数传递
}

type OptionsResponse[T any] struct {
	Key   string `json:"key"`   // 选项键，用于唯一标识选项
	Value T      `json:"value"` // 选项值，存储具体的选项内容
}
