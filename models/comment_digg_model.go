// models/article_digg_model.go
package models

type CommentDiggModel struct { //UserID和CommentID创建复合索引,共同使用同一个索引
	Model
	UserID       uint         `gorm:"uniqueIndex:idx_comment_model_userid" json:"userID"`       // 用户ID
	CommentID    uint         `gorm:"uniqueIndex:idx_comment_model_commentid" json:"commentID"` // 文章ID
	UserModel    UserModel    `gorm:"foreignKey:UserID" json:"-"`                               // 用户信息（关联）
	ArticleID    uint         `gorm:"uniqueIndex:idx_comment_model_articleid" json:"articleID"` // 文章ID
	ArticleModel ArticleModel `gorm:"foreignKey:ArticleID" json:"-"`                            // 文章信息（关联）
	CommentModel CommentModel `gorm:"foreignKey:CommentID" json:"-"`                            // 评论信息（关联）
}
