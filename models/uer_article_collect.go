// models/user_article_collect_model.go
package models

import (
	"StarDreamerCyberNook/service/redis_service/redis_count"
	"time"

	"gorm.io/gorm"
)

type UserArticleCollectModel struct { //用户收藏文章表.目前只能收藏到一个收藏夹，后续可能会改为多对多关系
	UserID       uint         `gorm:"uniqueIndex:idx_name" json:"userID"`       // 用户id
	UserModel    UserModel    `gorm:"foreignKey:UserID" json:"-"`               // 收藏者
	ArticleID    uint         `gorm:"uniqueIndex:idx_name" json:"articleID"`    // 文章id
	ArticleModel ArticleModel `gorm:"foreignKey:ArticleID" json:"-"`            // 被收藏的文章
	CollectID    uint         `gorm:"uniqueIndex:idx_name" json:"collectID"`    // 收藏夹的id
	CollectModel CollectModel `gorm:"foreignKey:CollectID" json:"collectModel"` // 属于哪一个收藏夹
	CreatedAt    time.Time    `json:"createdAt"`                                // 收藏的时间
}

func (u *UserArticleCollectModel) AfterCreate(db *gorm.DB) error { //创建时对应文章的收藏数加1
	redis_count.SetCacheCollect(u.ArticleID, true)
	return nil
}

func (u *UserArticleCollectModel) BeforeDelete(db *gorm.DB) error { //删除时对应文章的收藏数减1
	redis_count.SetCacheCollect(u.ArticleID, false)
	return nil
}
