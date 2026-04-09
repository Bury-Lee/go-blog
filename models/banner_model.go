// models/banner_model.go
// 轮播图模型定义
// 管理首页轮播图展示，支持图片链接和跳转链接配置
package models

// BannerModel 轮播图模型
// 存储轮播图信息，包括展示状态、图片链接和跳转链接
// 用于首页轮播图展示功能
type BannerModel struct {
	Model
	IsShow bool   `json:"isShow"` // 是否展示：true-展示 false-隐藏
	Cover  string `json:"cover"`  // 轮播图片链接
	Href   string `json:"href"`   // 跳转链接地址
}
