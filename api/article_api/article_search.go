package article_api

import (
	"StarDreamerCyberNook/common"
	"StarDreamerCyberNook/common/response"
	"StarDreamerCyberNook/global"
	"StarDreamerCyberNook/models"
	"StarDreamerCyberNook/models/enum"
	jwts "StarDreamerCyberNook/utils/jwts"
	"StarDreamerCyberNook/utils/sql"
	"context"
	"encoding/json"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/olivere/elastic/v7"
	"github.com/sirupsen/logrus"
)

// ArticleSearchRequest 搜索请求结构体，包含分页信息、标签和排序类型
type ArticleSearchRequest struct {
	common.PageInfo
	Tag  string `form:"tag"`  // 按标签筛选
	Type int8   `form:"type"` // 排序类型: 0 最新发布 1 猜你喜欢    2最多回复 3最多点赞 4最多收藏
}

// ArticleBaseInfo 搜索结果的基础信息结构体
type ArticleBaseInfo struct {
	ID       uint   `json:"id"`
	Title    string `json:"title"`
	Abstract string `json:"abstract"`
}

// ArticleSearchListResponse 搜索结果详情结构体，继承了文章模型，并增加了关联信息
type ArticleSearchListResponse struct {
	models.ArticleModel
	AdminTop      bool    `json:"adminTop"`      // 是否是管理员置顶
	CategoryTitle *string `json:"categoryTitle"` // 所属分类标题
	UserNickname  string  `json:"userNickname"`  // 发布用户昵称
	UserAvatar    string  `json:"userAvatar"`    // 发布用户头像
}

// ArticleSearchView 搜索文章的API处理函数
// 主要逻辑：
// 1. 解析请求参数（分页、关键词、标签、排序类型）
// 2. 构建Elasticsearch查询DSL
// 3. 执行搜索并获取结果（含高亮）
// 4. 获取对应的完整文章数据（从数据库）
// 5. 合并数据并返回搜索结果
func (ArticleApi) ArticleSearchView(c *gin.Context) {
	// 1. 解析并验证请求参数
	var cr ArticleSearchRequest
	if err := c.ShouldBindQuery(&cr); err != nil {
		response.FailWithMsg("参数绑定失败", c)
		return
	}

	//TODO:也许要考虑服务降级的问题?
	// if global.ES == nil {//在这里写降级处理
	// 	response.FailWithMsg("Elasticsearch服务未启动", c)
	// 	return
	// }

	// 2. 根据请求的Type确定Elasticsearch排序字段
	var sortMap = map[int8]string{
		0: "created_at",    // 最新发布：按创建时间排序
		1: "_score",        // 猜你喜欢：按相关性评分排序
		2: "comment_count", // 最多回复：按评论数排序
		3: "digg_count",    // 最多点赞：按点赞数排序
		4: "collect_count", // 最多收藏：按收藏数排序
	}
	sortKey, ok := sortMap[cr.Type]
	if !ok { // 如果传入的Type不在map中，则返回错误
		response.FailWithMsg("搜索类型错误", c)
		return
	}

	// 构建Elasticsearch Bool Query
	query := elastic.NewBoolQuery()
	// 如果有关键词，则在标题、摘要、内容中进行模糊匹配（Should关系）
	if cr.Key != "" {
		query.Should(
			elastic.NewMatchQuery("title", cr.Key),
			elastic.NewMatchQuery("abstract", cr.Key),
			elastic.NewMatchQuery("content", cr.Key),
		)
	}
	// 如果指定了标签，则必须匹配该标签（Must关系）
	if cr.Tag != "" {
		query.Must(
			elastic.NewTermQuery("tag_list", cr.Tag),
		)
	}

	// 只查询已发布的文章（Must关系）
	query.Must(elastic.NewTermQuery("status", int(models.StatusPublished)))

	// 处理管理员置顶逻辑
	var articleIDList []uint    // 用于存储最终要返回的文章ID列表
	var userIDList []uint       // 存储所有管理员用户的ID
	var topArticleIDList []uint // 存储被管理员置顶的文章ID

	//TODO:这里两步查询还是有点低效，以后想想怎么优化吧
	//TODO:这里也应该先去Redis查询置顶文章列表，没有再从数据库查询(预计管理员置顶文章不会太多的情况下)
	//在管理员及其置顶的文章不会太多的情况下,也许可以把这些放Redis里,直接从Redis里取置顶文章列表.记得和top的CRUD函数同步更改
	// 查询所有角色为管理员的用户ID
	global.DB.Model(models.UserModel{}).Where("role = ?", enum.AdminRole).Select("id").Scan(&userIDList) //由于管理员的数量不会超过100人，所以这里可以直接查询所有管理员的ID
	// 查询这些管理员置顶的文章ID
	global.DB.Model(models.UserTopArticleModel{}).Where("user_id in ?", userIDList).Select("article_id").Scan(&topArticleIDList) // 查询所有管理员置顶的文章ID
	var articleTopMap = map[uint]bool{}                                                                                          // 用于快速判断某ID是否为置顶文章
	if len(topArticleIDList) > 0 {
		var topArticleIDListAny []interface{}
		for _, u := range topArticleIDList {
			topArticleIDListAny = append(topArticleIDListAny, u)
			articleTopMap[u] = true // 标记为置顶
		}
		// 只给命中的置顶文章加权，不绕过搜索条件
		query.Should(elastic.NewTermsQuery("id", topArticleIDListAny...).Boost(10))
	}
	//TODO.END

	// 如果是"猜你喜欢"（Type=1），则加入用户兴趣标签查询
	if cr.Type == 1 {
		// 尝试从JWT Token中解析用户信息
		claims, err := jwts.ParseTokenByGin(c)
		if err == nil && claims != nil {
			// 用户已登录
			var user models.UserModel
			// 读取用户兴趣标签
			err = global.DB.Select("id", "like_tags").Take(&user, claims.UserID).Error
			if err != nil {
				response.FailWithMsg("用户信息不存在", c)
				return
			}
			// 如果用户配置中有感兴趣的文章标签
			if len(user.LikeTags) > 0 {
				tagQuery := elastic.NewBoolQuery()
				var tagAnyList []interface{}
				for _, tag := range user.LikeTags {
					tagAnyList = append(tagAnyList, tag)
				}
				// 兴趣标签至少命中一个
				tagQuery.Should(elastic.NewTermsQuery("tag_list", tagAnyList...)).
					MinimumNumberShouldMatch(1)
				query.Must(tagQuery)
			}
		}
	}

	// 设置高亮显示，对标题和摘要字段进行高亮
	highlight := elastic.NewHighlight()
	highlight.Field("title")
	highlight.Field("abstract")

	// 执行Elasticsearch搜索
	result, err := global.ES.
		Search(models.ArticleModel{}.Index()). // 指定搜索的索引
		Query(query).                          // 设置查询DSL
		Highlight(highlight).                  // 设置高亮
		From(cr.GetOffset()).                  // 设置分页偏移量
		Size(cr.GetLimit()).                   // 设置分页大小
		Sort(sortKey, false).                  // 设置排序字段和方向(false表示降序)
		Do(context.Background())               // 执行搜索
	if err != nil {
		// 记录错误日志，包括错误信息和查询语句
		source, _ := query.Source()
		byteData, _ := json.Marshal(source)
		logrus.Errorf("查询失败 %s \n %s", err, string(byteData))
		response.FailWithMsg("查询失败", c)
		return
	}

	// 解析Elasticsearch返回的命中结果
	count := result.Hits.TotalHits.Value              // 获取总命中数
	var searchArticleMap = map[uint]ArticleBaseInfo{} // 用于存储ES返回的精简文章信息（含高亮）
	var articleIDSet = map[uint]struct{}{}            // 用于去重文章ID

	for _, hit := range result.Hits.Hits {
		var art ArticleBaseInfo
		// 将ES返回的JSON Source反序列化为ArticleBaseInfo结构体
		err = json.Unmarshal(hit.Source, &art)
		if err != nil {
			logrus.Warnf("解析失败 %s  %s", err, string(hit.Source))
			continue
		}
		// 如果有高亮结果，则替换原始内容
		if len(hit.Highlight["title"]) > 0 {
			art.Title = hit.Highlight["title"][0]
		}
		if len(hit.Highlight["abstract"]) > 0 {
			art.Abstract = hit.Highlight["abstract"][0]
		}

		searchArticleMap[art.ID] = art // 存入映射表
		if _, ok := articleIDSet[art.ID]; ok {
			continue
		}
		articleIDSet[art.ID] = struct{}{}
		articleIDList = append(articleIDList, art.ID)
	}

	// 没有命中时直接返回，避免无意义的数据库查询
	if len(articleIDList) == 0 {
		response.OkWithList([]ArticleSearchListResponse{}, int(count), c)
		return
	}

	// 根据Elasticsearch返回的文章ID列表，从数据库查询完整的文章对象（包含关联的分类和用户信息）
	//TODO:这里也应该先去Redis查询，没有再从数据库查询(注:这里没查到的文章页不应该放入Redis,查询率不一定代表点击率)

	KeyList := []string{}
	for _, id := range articleIDList {
		idStr := strconv.FormatUint(uint64(id), 10)
		KeyList = append(KeyList, "ArticleID"+idStr)
	}

	ctx := context.Background()
	res, exit := global.RedisHotPool.MGet(ctx, KeyList...).Result()

	var list = make([]ArticleSearchListResponse, 0)
	var cacheMissIDList []uint                                 // 缓存未命中的文章ID
	var cacheHitMap = make(map[uint]ArticleSearchListResponse) // 缓存命中的文章数据

	if exit == nil { //查询成功
		// 处理缓存命中结果
		for i, cacheData := range res {
			if cacheData != nil {
				// 缓存命中
				var cached ArticleDetailResponse
				err := json.Unmarshal([]byte(cacheData.(string)), &cached)
				if err != nil {
					logrus.Warnf("缓存数据解析错误: %s", err)
					// 解析失败时，将文章ID加入未命中列表，后续从数据库查询
					cacheMissIDList = append(cacheMissIDList, articleIDList[i])
					continue
				}

				// 组装缓存数据为搜索结果格式
				item := ArticleSearchListResponse{
					ArticleModel: cached.ArticleModel,
					AdminTop:     articleTopMap[cached.ID], // 设置是否置顶
					UserNickname: cached.NickName,          // 设置用户名
					UserAvatar:   cached.UserAvatar,        // 设置用户头像
				}

				// 使用ES返回的高亮标题和摘要覆盖缓存中的内容
				if art, ok := searchArticleMap[cached.ID]; ok {
					item.Title = art.Title
					item.Abstract = art.Abstract
				}

				cacheHitMap[cached.ID] = item
			} else {
				// 缓存未命中
				cacheMissIDList = append(cacheMissIDList, articleIDList[i])
			}
		}
	} else {
		// Redis查询失败，所有文章ID都从未命中列表处理
		cacheMissIDList = articleIDList
	}

	// 如果有缓存未命中的文章，从数据库查询
	if len(cacheMissIDList) > 0 {
		where := global.DB.Where("id in ?", cacheMissIDList)
		_list, _, err := common.ListQuery[models.ArticleModel](models.ArticleModel{}, common.Options{
			Where:        where,
			Preloads:     []string{"CategoryModel", "UserModel"},    // 预加载分类和用户信息
			DefaultOrder: sql.ConvertSliceOrderSql(cacheMissIDList), // 按照cacheMissIDList的顺序进行排序
		})
		if err != nil {
			logrus.Errorf("查询文章详情失败 %s", err)
			response.FailWithMsg("查询失败", c)
			return
		}

		// 将数据库查询的完整信息与ES返回的高亮信息合并
		for _, model := range _list {
			item := ArticleSearchListResponse{
				ArticleModel: model,
				AdminTop:     articleTopMap[model.ID],  // 设置是否置顶
				UserNickname: model.UserModel.NickName, // 设置用户名
				UserAvatar:   model.UserModel.Avatar,   // 设置用户头像
			}
			// 设置分类标题（如果存在）
			if model.CategoryModel != nil {
				item.CategoryTitle = &model.CategoryModel.Title
			}
			// 使用ES返回的高亮标题和摘要覆盖数据库中的原始内容
			if art, ok := searchArticleMap[model.ID]; ok {
				item.Title = art.Title
				item.Abstract = art.Abstract
			}
			cacheHitMap[model.ID] = item // 存入map，便于后续排序
		}
	}

	// 按照原始文章ID列表的顺序组装最终结果
	for _, id := range articleIDList {
		if item, ok := cacheHitMap[id]; ok {
			list = append(list, item)
		}
	}

	//TODO:以后加入带点赞数和评论数的响应字段
	// 返回成功响应，包含文章列表和总数
	response.OkWithList(list, int(count), c)
}
