package article_api

import (
	"StarDreamerCyberNook/common"
	"StarDreamerCyberNook/common/response"
	"StarDreamerCyberNook/global"
	"StarDreamerCyberNook/models"
	"StarDreamerCyberNook/models/enum"
	jwts "StarDreamerCyberNook/utils/jwts"
	"StarDreamerCyberNook/utils/sql"
	"fmt"

	"github.com/gin-gonic/gin"
)

//TODO:写好注释到时候好好看看
/*
可以查某个用户发布的文章，只能查已发布的，不需要登录，支持分类查询

	可以查某个人用户收藏的文章，前提是这个用户开了对应隐私设置

用户侧  能查自己发布的文章，只能查已发布的，需要登录， 支持分类查询

	也能查自己收藏的文章，不会受到自己的隐私设置

	支持按照状态查询，已发布，草稿箱，待审核

管理员侧 查全部，支持按照用户搜索，状态过滤，文章标题模糊匹配，分类过滤
*/

type ArticleListRequest struct {
	common.PageInfo
	Type       string        `form:"type" binding:"required"`
	UserID     uint          `form:"userID"`
	CategoryID *uint         `form:"categoryID"`
	Status     models.Status `form:"status"`
}

type ArticleListResponse struct {
	models.ArticleModel
	UserTop       bool    `json:"userTop"`
	AdminTop      bool    `json:"adminTop"`
	CategoryTitle *string `json:"categoryTitle"`
	UserNickName  string  `json:"userNickName"`
	Avatar        string  `json:"avatar"`
}

func (ArticleApi) ArticleListView(c *gin.Context) {
	var req ArticleListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithMsg("参数错误", c)
		return
	}

	var TopArticleIDList []uint

	var orderColumnMap = map[string]bool{
		"look_count desc":    true,
		"digg_count desc":    true,
		"comment_count desc": true,
		"collect_count desc": true,
		"look_count asc":     true,
		"digg_count asc":     true,
		"comment_count asc":  true,
		"collect_count asc":  true,
	}

	if req.Order != "" {
		_, ok := orderColumnMap[req.Order]
		if !ok {
			response.FailWithMsg("不支持的排序方式", c)
			return
		}
	}

	switch req.Type {
	case "other":
		// 查别人,用户id就是必填的
		// if req.UserID == 0 {
		// 	response.FailWithMsg("用户id是必填项", c)
		// 	return
		// }
		//啊算了,去除这个限制来支持查询最新文章

		// if req.Page > 2 || req.Limit > 10 {
		// 	response.FailWithMsg("查询更多，请登录", c)
		// 	return
		// }
		req.Status = models.StatusPublished // 查别人只能查已发布的
	case "self":
		// 查自己的
		claims, err := jwts.ParseTokenByGin(c)
		if err != nil || claims.UserID == 0 {
			response.FailWithMsg("请登录", c)
			return
		}
		req.UserID = claims.UserID
	case "admin":
		// 管理员
		claims, err := jwts.ParseTokenByGin(c)
		if err != nil || claims.Role != enum.AdminRole {
			response.FailWithMsg("角色错误", c)
			return
		}
	default:
		response.FailWithMsg("请求错误", c)
		return
	}
	var userTopMap = make(map[uint]bool)
	var adminTopMap = make(map[uint]bool)
	if req.UserID != 0 { // 查询用户置顶文章
		var userTopArticleList []models.UserTopArticleModel
		global.DB.Preload("UserModel").Order("created_at desc").Find(&userTopArticleList, "user_id = ?", req.UserID)
		for _, item := range userTopArticleList {
			TopArticleIDList = append(TopArticleIDList, item.ArticleID)
			if item.UserModel.Role == enum.AdminRole {
				adminTopMap[item.ArticleID] = true
			}
			userTopMap[item.ArticleID] = true
		}
	}

	var options = common.Options{
		Likes:        []string{"title"},
		PageInfo:     req.PageInfo,
		Preloads:     []string{"UserModel", "CategoryModel"}, //预加载用户和分类
		DefaultOrder: "created_at desc",
	}
	if len(TopArticleIDList) > 0 {
		options.DefaultOrder = fmt.Sprintf("%s, created_at desc", sql.ConvertSliceOrderSql(TopArticleIDList))
	}

	_list, count, _ := common.ListQuery[models.ArticleModel](models.ArticleModel{
		UserID:     req.UserID,
		CategoryID: req.CategoryID,
		Status:     req.Status,
	}, options)

	var list = make([]ArticleListResponse, 0)

	articleIDs := make([]uint, 0, len(_list))
	idMap := make(map[uint]struct{}, len(_list))
	for _, item := range _list {
		if _, ok := idMap[item.ID]; ok {
			continue
		}
		idMap[item.ID] = struct{}{}
		articleIDs = append(articleIDs, item.ID)
	}
	//这样的话性能消耗可能过大,也许停用会好一些,虽然这样就减弱了一致性
	// collectCountMap := redis_count.GetAllCacheCollect(articleIDs)
	// lookCountMap := redis_count.GetAllCacheLook(articleIDs)
	// diggCountMap := redis_count.GetAllCacheDigg(articleIDs)

	for _, model := range _list {
		// model.Content = ""//前台写个文章预览吧
		// model.DiggCount += diggCountMap[model.ID]
		// model.LookCount += lookCountMap[model.ID]
		// model.CollectCount += collectCountMap[model.ID]
		data := ArticleListResponse{
			ArticleModel: model,
			UserTop:      userTopMap[model.ID],
			AdminTop:     adminTopMap[model.ID],
			UserNickName: model.UserModel.NickName,
			Avatar:       model.UserModel.Avatar,
		}
		if model.CategoryID != nil && model.CategoryModel != nil {
			data.CategoryTitle = &model.CategoryModel.Title
		}
		list = append(list, data)
	}
	response.OkWithList(list, count, c)
}
