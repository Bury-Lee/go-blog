package models

type UserArticleHistoryModel struct { //同一文章可多次阅读,所以不使用联合主键与同一ID,而是单独的ID
	Model
	UserID       uint         `gorm:"index" json:"userID"`           //用户ID
	UserModel    UserModel    `gorm:"foreignKey:UserID" json:"-"`    //关联用户表,不返回给前端
	ArticleID    uint         `gorm:"index" json:"articleID"`        //文章ID
	ArticleModel ArticleModel `gorm:"foreignKey:ArticleID" json:"-"` //关联文章表,不返回给前端
}
