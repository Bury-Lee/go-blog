// models/article_model.go
// 文章模型定义
// 存储博客系统的文章信息，包括标题、内容、分类、标签、统计等完整字段
package models

import (
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
	return strings.Join(this, ","), nil
}

// ArticleModel 文章模型
// 用于存储博客系统的文章信息，包含标题、内容、分类、标签、统计等字段
type ArticleModel struct {
	Model
	Title         string         `gorm:"size:32" json:"title"`                     // 文章标题，最大32字符
	Abstract      string         `gorm:"size:256" json:"abstract"`                 // 文章摘要，最大256字符
	Content       string         `json:"content"`                                  // 文章内容
	CategoryID    *uint          `json:"categoryID"`                               // 文章分类ID，关联分类表
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

// ArticleSearchModel 没有ES时的搜索降级搜索模型
// TODO
type ArticleSearchModel struct {
	Model
	ArticleID uint         `json:"articleID"`                                         // 外键关联文章ID
	Article   ArticleModel `gorm:"foreignKey:ArticleID" json:"-"`                     //预备字段
	Title     string       `gorm:"size:32;index:idx_titleSearch" json:"title"`        // 文章标题，加索引
	Abstract  string       `gorm:"size:256;index:idx_abstractSearch" json:"abstract"` // 文章摘要，加索引
}

// truncateText 按字符数截断文本
// 参数:text - 原始文本
// 参数:maxLen - 最大字符数
// 返回:string - 截断后的文本
// 说明:按rune截断，避免中文等多字节字符被截断
func truncateText(text string, maxLen int) string {
	if maxLen <= 0 {
		return ""
	}
	runes := []rune(strings.TrimSpace(text))
	if len(runes) <= maxLen {
		return string(runes)
	}
	return string(runes[:maxLen])
}

//文章创建时自动创建,并更新全文搜索记录
// 文章更新时,也会更新全文搜索记录
//删除时也会删除全文搜索记录

// 以下用于ES服务降级,当ES无法使用时,将使用数据库存储全文搜索记录

// AfterCreate 文章创建后写入全文搜索记录
// 参数:tx - GORM事务对象
// 返回:err - 错误信息
// 说明:仅发布状态的文章会被提取正文并保存到ArticleSearch中，用于降级搜索
func (this *ArticleModel) AfterCreate(tx *gorm.DB) (err error) {
	if this.Status != StatusPublished {
		return nil
	}
	var Result ArticleSearchModel
	//内容在创建时已经过滤了
	Result.Abstract = truncateText(this.Abstract, 256)
	Result.Title = truncateText(this.Title, 32)
	Result.ArticleID = this.ID
	err = tx.Create(&Result).Error
	if err != nil {
		logrus.Error(err)
		return err
	}
	return nil
}

// AfterDelete 文章删除后清理全文搜索记录
// 参数:tx - GORM事务对象
// 返回:err - 错误信息
// 说明:根据文章ID查找并删除相关的全文搜索记录
func (this *ArticleModel) AfterDelete(tx *gorm.DB) (err error) {
	var Result ArticleSearchModel
	tx.Delete(&Result, "article_id = ?", this.ID)
	return nil
}

// AfterUpdate 文章更新后刷新全文搜索记录
// 参数:tx - GORM事务对象
// 返回:err - 错误信息
// 说明:先删除旧的全文记录，再根据当前内容重新生成全文搜索记录
func (this *ArticleModel) AfterUpdate(tx *gorm.DB) (err error) {
	err = this.AfterDelete(tx)
	if err != nil {
		return err
	}
	return this.AfterCreate(tx)
}

// BeforeDelete 删除文章前清理关联数据
// 参数:tx - GORM事务对象
// 返回:err - 错误信息
// 说明:删除文章前，级联删除评论、点赞、收藏、置顶、浏览等关联记录
func (this *ArticleModel) BeforeDelete(tx *gorm.DB) (err error) {
	// 评论
	var commentList []CommentModel
	tx.Find(&commentList, "article_id = ?", this.ID).Delete(&commentList)
	// 点赞
	var diggList []ArticleDiggModel
	tx.Find(&diggList, "article_id = ?", this.ID).Delete(&diggList)
	// 收藏
	var collectList []UserArticleCollectModel
	tx.Find(&collectList, "article_id = ?", this.ID).Delete(&collectList)
	// 置顶
	var topList []UserTopArticleModel
	tx.Find(&topList, "article_id = ?", this.ID).Delete(&topList)
	// 浏览
	var lookList []UserArticleHistoryModel
	tx.Find(&lookList, "article_id = ?", this.ID).Delete(&lookList)

	logrus.Infof("删除关联评论 %d 条", len(commentList))
	logrus.Infof("删除关联点赞 %d 条", len(diggList))
	logrus.Infof("删除关联收藏 %d 条", len(collectList))
	logrus.Infof("删除关联置顶 %d 条", len(topList))
	logrus.Infof("删除关联浏览 %d 条", len(lookList))
	return nil
}
