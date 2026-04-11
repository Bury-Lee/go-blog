// models/user_top_article_model.go
package models

type UserTopArticleModel struct {
	Model
	UserID       uint         `gorm:"uniqueIndex:idx_user" json:"userID"`
	ArticleID    uint         `gorm:"uniqueIndex:idx_article" json:"articleID"`
	UserModel    UserModel    `gorm:"foreignKey:UserID" json:"-"`
	ArticleModel ArticleModel `gorm:"foreignKey:ArticleID" json:"-"`
}
