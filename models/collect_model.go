// models/collect_model.go
// 收藏夹模型定义
// 用户文章收藏功能，支持创建个人收藏夹
package models

// CollectModel 收藏夹模型
// 存储用户创建的收藏夹信息，支持自定义收藏夹名称、简介和封面
// 每个收藏夹属于特定用户，可收藏多篇文章
type CollectModel struct {
	Model
	Title       string                    `gorm:"size:64" json:"title"`                    // 收藏夹名称
	Abstract    string                    `gorm:"size:512" json:"abstract"`                // 收藏夹简介
	Cover       string                    `gorm:"size:512" json:"cover"`                   // 收藏夹封面图片URL
	ArticleList []UserArticleCollectModel `gorm:"foreignKey:CollectID" json:"articleList"` // 收藏夹内文章ID列表
	UserID      uint                      `json:"userID"`                                  // 收藏夹所属用户ID
	UserModel   UserModel                 `gorm:"foreignKey:UserID" json:"-"`              // 收藏夹所属用户信息

	IsDefault bool `json:"isDefault"` // 是否为默认收藏夹
}
