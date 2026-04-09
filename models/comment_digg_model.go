// models/article_digg_model.go
package models

// TODO:CommentDiggModel 点赞记录模型,也许以后会用上
type CommentDiggModel struct { //UserID和CommentID创建复合索引,共同使用同一个索引
	Model
	UserID       uint         `gorm:"uniqueIndex:idx_name" json:"userID"`    // 用户ID
	CommentID    uint         `gorm:"uniqueIndex:idx_name" json:"commentID"` // 文章ID
	UserModel    UserModel    `gorm:"foreignKey:UserID" json:"-"`            // 用户信息（关联）
	ArticleModel ArticleModel `gorm:"foreignKey:ArticleID" json:"-"`         // 文章信息（关联）
}
