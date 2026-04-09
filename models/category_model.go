// models/category_model.go
// 分类模型定义
// 用于文章分类管理，支持用户自定义分类
package models

// CategoryModel 分类模型
// 存储文章分类信息，每个分类属于特定用户
// 支持用户创建和管理个人文章分类
type CategoryModel struct {
	Model
	Title     string    `gorm:"size:32" json:"title"`       // 分类标题，最大32字符
	UserID    uint      `json:"userID"`                     // 所属用户ID
	UserModel UserModel `gorm:"foreignKey:UserID" json:"-"` // 所属用户信息，外键关联

	ArticleList []ArticleModel `gorm:"foreignKey:CategoryID" json:"-"` // 关联文章，外键关联，不序列化
}
