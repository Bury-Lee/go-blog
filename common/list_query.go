// common/list_query.go
package common

import (
	"StarDreamerCyberNook/global"
	"fmt"

	"gorm.io/gorm"
)

type PageInfo struct {
	EndId uint   `form:"endId"`              //某页末尾的游标ID,使用后可以减轻数据库压力
	Limit int    `form:"limit" default:"10"` //一页的数量,默认10条
	Page  int    `form:"page" default:"1"`   //分页查询,默认第1页
	Key   string `form:"key"`                //模糊匹配的参数
	Order string `form:"order"`              //其他查询参数,主要是排序
}

// GetOffset 获取查询偏移量
// 返回:int - 偏移量
// 说明:根据页码和每页数量计算偏移量
func (p *PageInfo) GetOffset() int {
	page := p.GetPage()
	limit := p.GetLimit()
	return (page - 1) * limit
}

// GetLimit 获取每页查询数量
// 返回:int - 每页数量限制
// 说明:限制范围在1到40之间,超出默认10
func (p *PageInfo) GetLimit() int {
	if p.Limit < 1 || p.Limit > 40 {
		return 10
	}
	return p.Limit
}

// GetPage 获取当前页码
// 返回:int - 当前页码
// 说明:页码范围在1到20之间,超出默认1
func (p *PageInfo) GetPage() int {
	if p.Page < 1 || p.Page > 30 {
		return 1
	}
	return p.Page
}

type Options struct { //可用选项,目前有模糊匹配和预加载,以后可以根据需要添加其他选项
	PageInfo     PageInfo //分页查询参数
	Likes        []string //模糊匹配的字段
	Preloads     []string //预加载的关联表
	Where        *gorm.DB //定制化查询
	DefaultOrder string   //其他查询参数,主要是排序,前端没有传入就使用这个默认排序参数
}

// ListQuery 通用分页查询函数
// 参数:model - 模型实例,用于指定查询的表
// 参数:option - 查询的配置选项
// 返回:list - 查询结果列表
// 返回:count - 满足条件的总记录数
// 返回:err - 错误信息
// 说明:支持基础查询,模糊匹配,定制化查询,预加载,排序分页
func ListQuery[T any](model T, option Options) (list []T, count int, err error) {
	query := global.DB.Model(model).Where(model)

	// 模糊匹配
	if len(option.Likes) > 0 && option.PageInfo.Key != "" {
		likes := global.DB
		for i, column := range option.Likes {
			if i == 0 {
				likes = likes.Where(fmt.Sprintf("%s LIKE ?", column), "%"+option.PageInfo.Key+"%")
			} else {
				likes = likes.Or(fmt.Sprintf("%s LIKE ?", column), "%"+option.PageInfo.Key+"%")
			}
		}
		query = query.Where(likes)
	}

	// 定制查询
	if option.Where != nil {
		query = query.Where(option.Where)
	}

	// 预加载
	for _, preLoad := range option.Preloads {
		query = query.Preload(preLoad)
	}

	// 排序（保证游标分页稳定）
	if option.PageInfo.Order != "" {
		query = query.Order(option.PageInfo.Order)
	} else if option.DefaultOrder != "" {
		query = query.Order(option.DefaultOrder)
	} else {
		query = query.Order("id DESC") // 默认兜底
	}

	limit := option.PageInfo.GetLimit()

	if option.PageInfo.EndId > 0 {
		query = query.Where("id < ?", option.PageInfo.EndId)

		err = query.Limit(limit).Find(&list).Error
		return list, 0, err
	}

	// 查总数
	var c int64
	query.Count(&c)
	count = int(c)

	offset := option.PageInfo.GetOffset()

	err = query.Offset(offset).Limit(limit).Find(&list).Error
	return list, count, err
}
