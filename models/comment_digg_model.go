// models/article_digg_model.go
package models

type CommentDiggModel struct {
	Model
	UserID       uint         `gorm:"index:idx_comment_digg_user_comment,unique,priority:1" json:"userID"`    // 用户ID
	CommentID    uint         `gorm:"index:idx_comment_digg_user_comment,unique,priority:2" json:"commentID"` // 评论ID
	UserModel    UserModel    `gorm:"foreignKey:UserID" json:"-"`                                             // 用户信息（关联）
	CommentModel CommentModel `gorm:"foreignKey:CommentID" json:"-"`                                          // 评论信息（关联）
}
