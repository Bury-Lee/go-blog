// models/article_model.go
// 文章模型定义
// 存储博客系统的文章信息，包括标题、内容、分类、标签、统计等完整字段
package models

import (
	"StarDreamerCyberNook/global"
	"database/sql/driver"
	_ "embed"
	"errors"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type TagList []string

func (this *TagList) Scan(value any) error {
	val, ok := value.([]uint8)
	if ok {
		*this = strings.Split(string(val), ",")
		return nil
	}
	return errors.New(fmt.Sprint("断言失败:", value))
}

func (this TagList) Value() (driver.Value, error) {
	return strings.Join(this, "."), nil
}

// 预备模型:全文搜索模型
type TextModel struct {
	Model
	ArticleID uint   `json:"articleID"`
	Head      string `json:"head"`
	Body      string `json:"body"`
}

// ArticleModel 文章模型
// 用于存储博客系统的文章信息，包含标题、内容、分类、标签、统计等字段
// TODO:文章的审核通过Python微服务移交给AI的api来处理，不想人工审核了，太麻烦了
type ArticleModel struct {
	Model
	Title         string         `gorm:"size:32" json:"title"`                     // 文章标题，最大32字符
	Abstract      string         `gorm:"size:256" json:"abstract"`                 // 文章摘要，最大256字符
	Content       string         `json:"content"`                                  // 文章内容
	CategoryID    *uint          `json:"categoryID"`                               //为0表示无分类                                   // 文章分类ID，关联分类表
	CategoryModel *CategoryModel `gorm:"foreignKey:CategoryID" json:"-"`           //这样可以吗? 以防万一用指针吧             // 分类信息，外键关联，不序列化
	TagList       []string       `gorm:"type:text;serializer:json" json:"tagList"` // 标签列表，JSON序列化存储 //serializer:json要删掉?似乎要换成自己定义的taglist数据类型
	Cover         string         `gorm:"size:256" json:"cover"`                    // 文章封面图片URL
	UserID        uint           `json:"userID"`                                   // 作者用户ID，关联用户表
	UserModel     UserModel      `gorm:"foreignKey:UserID" json:"-"`               // 作者信息，外键关联，不序列化
	LookCount     int            `json:"lookCount"`                                // 浏览次数统计
	DiggCount     int            `json:"diggCount"`                                // 点赞次数统计
	CommentCount  int            `json:"commentCount"`                             // 评论数量统计
	CollectCount  int            `json:"collectCount"`                             // 收藏次数统计//TODO:每个用户对于每篇文章只能收藏一次,不然搞多几个收藏夹就会一直刷收藏统计,所以这个收藏数直接用redis缓存就好了?不需要每次都更新数据库了,有并发问题
	OpenComment   bool           `json:"openComment"`                              // 是否开启评论：true-开启 false-关闭
	Status        Status         `json:"status"`                                   // 文章状态：0-草稿 1-审核中 2-已发布 3-已下线

	AIQuality  string `json:"aiQuality"`  // AI生成的内容质量评级,1-5分
	AIAbstract string `json:"aiAbstract"` // AI生成的内容摘要
}
type Status int8 // 文章状态枚举类型

const (
	StatusDraft     Status = 0 // 草稿//注意的是默认就是零,当前端为传入指定参数时
	StatusPending   Status = 1 // 审核中 (Pending 比 Auditing 更常用于表示“等待处理”的状态)
	StatusPublished Status = 2 // 已发布
	StatusOffline   Status = 3 // 已下线 (或者用 StatusArchived 表示归档/下架)
)

func (s Status) String() string {
	// 定义状态名称映射数组
	// 索引对应 Status 的值
	statusNames := []string{
		"草稿",  // 0
		"审核中", // 1
		"已发布", // 2
		"已下线", // 3
	}

	// 边界检查，防止越界 panic
	if s >= 0 && int(s) < len(statusNames) {
		return statusNames[s]
	}
	return "未知状态"
}

//go:embed mappings/article_mapping.json
var articleMapping string //使用宏把json赋值到这里来

func (ArticleModel) Mapping() string {
	return articleMapping
}
func (ArticleModel) Index() string {
	return "article_index" //返回文章模型的索引名
}

//以下用于ES服务降级,当ES无法使用时,将使用数据库存储全文搜索记录
// func (a *ArticleModel) AfterCreate(tx *gorm.DB) (err error) { //
// 	// 创建文章之后的钩子函数
// 	// 只有发布中的文章会放到全文搜索里面去
// 	if a.Status != StatusPublished {
// 		return nil
// 	}
// 	textList := MDtransform.MdContentTransformation(a.Title, a.Content, a.ID)
// 	var list []TextModel
// 	for _, model := range textList {
// 		list = append(list, TextModel{
// 			ArticleID: model.ArticleID,
// 			Head:      model.Head,
// 			Body:      model.Body,
// 		})
// 	}
// 	err = tx.Create(&list).Error
// 	if err != nil {
// 		logrus.Error(err)
// 		return nil
// 	}
// 	return nil
// }

// func (a *ArticleModel) AfterDelete(tx *gorm.DB) (err error) {
// 	// 删除之后
// 	var textList []TextModel
// 	tx.Find(&textList, "article_id = ?", a.ID)
// 	if len(textList) > 0 {
// 		logrus.Infof("删除全文记录 %d", len(textList))
// 		tx.Delete(&textList)
// 	}
// 	return nil
// }

// func (a *ArticleModel) AfterUpdate(tx *gorm.DB) (err error) {
// 	// 正文发生了变化，才去做转换
// 	a.AfterDelete(tx)
// 	a.AfterCreate(tx)
// 	return nil
// }

func (a *ArticleModel) BeforeDelete(tx *gorm.DB) (err error) { //钩子函数,在删除文章前,先删除关联的评论,点赞,收藏,置顶,浏览记录
	// 评论
	var commentList []CommentModel
	global.DB.Find(&commentList, "article_id = ?", a.ID).Delete(&commentList)
	// 点赞
	var diggList []ArticleDiggModel
	global.DB.Find(&diggList, "article_id = ?", a.ID).Delete(&diggList)
	// 收藏
	var collectList []UserArticleCollectModel
	global.DB.Find(&collectList, "article_id = ?", a.ID).Delete(&collectList)
	// 置顶
	var topList []UserTopArticleModel
	global.DB.Find(&topList, "article_id = ?", a.ID).Delete(&topList)
	// 浏览
	var lookList []UserArticleHistoryModel
	global.DB.Find(&lookList, "article_id = ?", a.ID).Delete(&lookList)

	logrus.Infof("删除关联评论 %d 条", len(commentList))
	logrus.Infof("删除关联点赞 %d 条", len(diggList))
	logrus.Infof("删除关联收藏 %d 条", len(collectList))
	logrus.Infof("删除关联置顶 %d 条", len(topList))
	logrus.Infof("删除关联浏览 %d 条", len(lookList))
	return
}
