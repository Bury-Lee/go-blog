// models/comment_model.go
// 评论模型定义
// 支持嵌套评论结构，可构建评论树
package models

import (
	utils_other "StarDreamerCyberNook/utils/other"
)

// CommentModel 评论模型
// 支持多级评论嵌套，包含评论内容、用户信息、文章信息、点赞统计等
// 通过ParentID构建评论层级关系，支持评论回复功能
type CommentModel struct {
	Model
	Content      string       `gorm:"size:1024" json:"content"`                                                  // 评论内容，长度限制宽松
	UserID       uint         `json:"userID"`                                                                    // 评论用户ID
	UserModel    UserModel    `gorm:"foreignKey:UserID" json:"user"`                                             // 评论用户信息，外键关联
	ArticleID    uint         `gorm:"index" json:"articleID"`                                                    // 评论文章ID
	ArticleModel ArticleModel `gorm:"foreignKey:ArticleID" json:"-"`                                             // 评论文章信息，外键关联
	ParentPath   string       `gorm:"type:varchar(512);index;CHARACTER SET ascii COLLATE ascii_bin" json:"path"` // 评论路径，用于快速定位评论层级
	/*
		注:在入库之后,path存储的父级评论路径,而ID才是自己的标识,类似于文件系统的路径
		使用/作为分隔符,将父级评论路径和自己的ID拼接起来
		为""时表示这是根目录,也就是一级评论
	*/
	RootParentID *uint         `json:"rootParentID"`                     // 根评论ID，用于快速定位顶级评论
	RootParent   *CommentModel `gorm:"foreignKey:RootParentID" json:"-"` //预备字段, 根评论信息，外键关联
	DiggCount    int           `json:"diggCount"`                        // 评论点赞数统计
}

// UintEncodeAdd: 将 ParentID 转为 Base62 字符串并追加到 BasePath 后
func (this *CommentModel) UintEncodeAdd(basePath string, parentID uint) {
	this.ParentPath = utils_other.EncodePath(basePath, parentID)

}
func (this *CommentModel) Decode() (ID uint, err error) { //输入路径,输出最后一个路径的解析ID
	if this.ParentPath == "" {
		return 0, nil
	}
	return utils_other.DecodePath(this.ParentPath)
}
