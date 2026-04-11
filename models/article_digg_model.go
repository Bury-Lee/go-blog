// models/article_digg_model.go
package models

import "time"

// ArticleDiggModel 点赞记录模型
type ArticleDiggModel struct { //UserID和ArticleID创建复合索引,共同使用同一个索引
	Model
	UserID       uint         `gorm:"uniqueIndex:idx_article_model_userid" json:"userID"`       // 用户ID
	ArticleID    uint         `gorm:"uniqueIndex:idx_article_model_articleid" json:"articleID"` // 文章ID
	UserModel    UserModel    `gorm:"foreignKey:UserID" json:"-"`                               // 用户信息（关联）
	ArticleModel ArticleModel `gorm:"foreignKey:ArticleID" json:"-"`                            // 文章信息（关联）
	CreatedAt    time.Time    `json:"createdAt"`                                                // 点赞时间
}
