package article_api

import (
	"StarDreamerCyberNook/common"
	"StarDreamerCyberNook/common/response"
	"StarDreamerCyberNook/global"
	"StarDreamerCyberNook/models"
	"StarDreamerCyberNook/models/enum"
	"StarDreamerCyberNook/utils"
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

func (ArticleApi) ArticleSearchView(c *gin.Context) {
	// 1. 解析并验证请求参数
	var req ArticleSearchRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithMsg("参数绑定失败", c)
		return
	}

	// 2. 根据请求的Type确定Elasticsearch排序字段
	var esSortMap = map[int8]string{
		0: "created_at",    // 最新发布：按创建时间排序
		1: "_score",        // 猜你喜欢：按相关性评分排序
		2: "comment_count", // 最多回复：按评论数排序
		3: "digg_count",    // 最多点赞：按点赞数排序
		4: "collect_count", // 最多收藏：按收藏数排序
	}

	sortKey, ok := esSortMap[req.Type]
	if !ok { // 如果传入的Type不在map中，则返回错误
		response.FailWithMsg("搜索类型错误", c)
		return
	}

	// 获取管理员置顶文章信息
	articleTopMap := map[uint]bool{}
	var topArticleIDList []uint
	{
		var userIDList []uint
		global.DB.Model(models.UserModel{}).Where("role = ?", enum.AdminRole).Select("id").Scan(&userIDList)
		if len(userIDList) > 0 {
			global.DB.Model(models.UserTopArticleModel{}).Where("user_id in ?", userIDList).Select("article_id").Scan(&topArticleIDList)
			for _, articleID := range topArticleIDList {
				articleTopMap[articleID] = true
			}
		}
	}

	// ES不可用时，使用数据库全文搜索表降级
	if global.ES == nil {
		// 降级搜索

		var Options common.Options
		Options.PageInfo = req.PageInfo
		Options.Likes = []string{"title", "abstract"}

		List, count, err := common.ListQuery(&models.ArticleSearchModel{}, Options)

		var total int64
		if count == 0 {
			response.OkWithList([]ArticleSearchListResponse{}, 0, c)
			return
		}

		var articleIDList []uint
		searchArticleMap := map[uint]ArticleBaseInfo{}
		for _, SearchModel := range List {
			articleIDList = append(articleIDList, SearchModel.ArticleID)
			searchArticleMap[SearchModel.ID] = ArticleBaseInfo{
				ID:       SearchModel.ID,
				Title:    SearchModel.Title,
				Abstract: SearchModel.Abstract,
			}
		}

		// ---- 获取搜索结果详情 ----
		{
			// 1. 准备Redis缓存键名
			keyList := []string{}
			for _, article := range List {
				idStr := strconv.FormatUint(uint64(article.ID), 10)
				keyList = append(keyList, "ArticleID"+idStr) // 缓存键格式：ArticleID{文章ID}
			}

			ctx := context.Background()
			// 2. 批量从Redis获取缓存数据
			res, cacheErr := global.RedisHotPool.MGet(ctx, keyList...).Result()

			var cacheMissIDList []uint                              // 未命中缓存的文章ID列表
			cacheHitMap := make(map[uint]ArticleSearchListResponse) // 缓存命中的结果映射

			// 3. 处理Redis返回的数据
			if cacheErr == nil {
				for index, cacheData := range res {
					// 3.1 缓存未命中
					if cacheData == nil {
						cacheMissIDList = append(cacheMissIDList, articleIDList[index])
						continue
					}

					// 3.2 解析缓存的JSON数据
					var cached ArticleDetailResponse
					err = json.Unmarshal([]byte(cacheData.(string)), &cached)
					if err != nil {
						logrus.Warnf("缓存数据解析错误: %s", err)
						cacheMissIDList = append(cacheMissIDList, articleIDList[index])
						continue
					}

					// 3.3 构建响应对象
					item := ArticleSearchListResponse{
						ArticleModel:  cached.ArticleModel,
						AdminTop:      articleTopMap[cached.ID], // 是否管理员置顶
						CategoryTitle: cached.CategoryTitle,     // 分类标题
						UserNickname:  cached.NickName,          // 用户昵称
						UserAvatar:    cached.UserAvatar,        // 用户头像
					}
					// 补充搜索相关的标题和摘要（可能与原文章不同）
					if article, ok := searchArticleMap[cached.ID]; ok {
						item.Title = article.Title       // 使用搜索匹配时的标题
						item.Abstract = article.Abstract // 使用搜索匹配时的摘要
					}
					cacheHitMap[cached.ID] = item
				}
			} else {
				// Redis出错时，降级为全部从数据库查询
				cacheMissIDList = articleIDList
			}

			// 4. 处理未命中缓存的文章（从数据库查询）
			if len(cacheMissIDList) > 0 {
				// 4.1 构建查询条件
				where := global.DB.Where("id in ?", cacheMissIDList)
				modelList, _, err := common.ListQuery(models.ArticleModel{}, common.Options{
					Where:        where,
					Preloads:     []string{"CategoryModel", "UserModel"},    // 预加载关联表
					DefaultOrder: sql.ConvertSliceOrderSql(cacheMissIDList), // 保持与传入ID顺序一致
				})
				if err != nil {
					logrus.Errorf("降级搜索失败 %s", err)
					response.FailWithMsg("搜索失败", c)
					return
				}

				// 4.2 将数据库查询结果转换为响应格式
				for _, model := range modelList {
					item := ArticleSearchListResponse{
						ArticleModel: model,
						AdminTop:     articleTopMap[model.ID],
						UserNickname: model.UserModel.NickName,
						UserAvatar:   model.UserModel.Avatar,
					}
					// 处理可能为nil的分类
					if model.CategoryModel != nil {
						item.CategoryTitle = &model.CategoryModel.Title
					}
					// 补充搜索匹配的内容
					if article, ok := searchArticleMap[model.ID]; ok {
						item.Title = article.Title
						item.Abstract = article.Abstract
					}
					cacheHitMap[model.ID] = item
				}
			}

			// 5. 按原始顺序组装最终结果
			list := make([]ArticleSearchListResponse, 0, len(articleIDList))
			for _, id := range articleIDList {
				if item, ok := cacheHitMap[id]; ok {
					// 5.1 高亮处理关键词
					item.Abstract = utils.HighlightKeyword(item.Abstract, req.Key)
					item.Title = utils.HighlightKeyword(item.Title, req.Key)
					item.Content = utils.HighlightKeyword(item.Content, req.Key)
					list = append(list, item)
				}
			}
			// 6. 返回搜索结果
			response.OkWithList(list, int(total), c)
		}
		return
	}

	// 构建Elasticsearch Bool Query
	query := elastic.NewBoolQuery()
	// 如果有关键词，则在标题、摘要、内容中进行模糊匹配（Should关系）
	if req.Key != "" {
		query.Should(
			elastic.NewMatchQuery("title", req.Key),
			elastic.NewMatchQuery("abstract", req.Key),
			elastic.NewMatchQuery("content", req.Key),
		)
	}
	// 如果指定了标签，则必须匹配该标签（Must关系）
	if req.Tag != "" {
		query.Must(
			elastic.NewTermQuery("tag_list", req.Tag),
		)
	}

	// 只查询已发布的文章（Must关系）
	query.Must(elastic.NewTermQuery("status", int(models.StatusPublished)))

	// 处理管理员置顶逻辑
	var articleIDList []uint // 用于存储最终要返回的文章ID列表
	if len(topArticleIDList) > 0 {
		var topArticleIDListAny []interface{}
		for _, u := range topArticleIDList {
			topArticleIDListAny = append(topArticleIDListAny, u)
		}
		// 只给命中的置顶文章加权，不绕过搜索条件
		query.Should(elastic.NewTermsQuery("id", topArticleIDListAny...).Boost(10))
	}

	// 如果是"猜你喜欢"（Type=1），则加入用户兴趣标签查询
	if req.Type == 1 {
		// 尝试从JWT Token中解析用户信息
		likeTags, err := func() ([]string, error) {
			claims, err := jwts.ParseTokenByGin(c)
			if err != nil || claims == nil {
				return nil, nil
			}
			var user models.UserModel
			err = global.DB.Select("id", "like_tags").Take(&user, claims.UserID).Error
			if err != nil {
				return nil, err
			}
			return user.LikeTags, nil
		}()
		if err != nil {
			response.FailWithMsg("用户信息不存在", c)
			return
		}
		// 如果用户配置中有感兴趣的文章标签
		if len(likeTags) > 0 {
			tagQuery := elastic.NewBoolQuery()
			var tagAnyList []interface{}
			for _, tag := range likeTags {
				tagAnyList = append(tagAnyList, tag)
			}
			// 兴趣标签至少命中一个
			tagQuery.Should(elastic.NewTermsQuery("tag_list", tagAnyList...)).
				MinimumNumberShouldMatch(1)
			query.Must(tagQuery)
		}
	}

	// 设置高亮显示，对标题和摘要字段进行高亮
	highlight := elastic.NewHighlight()
	highlight.Field("title")
	highlight.Field("abstract")
	highlight.Field("content") //TODO:不知道会不会出现bug

	// 执行Elasticsearch搜索
	result, err := global.ES.
		Search(models.ArticleModel{}.Index()). // 指定搜索的索引
		Query(query).                          // 设置查询DSL
		Highlight(highlight).                  // 设置高亮
		From(req.GetOffset()).                 // 设置分页偏移量
		Size(req.GetLimit()).                  // 设置分页大小
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

	// ---- 获取搜索结果详情 ----
	{
		keyList := []string{}
		for _, id := range articleIDList {
			idStr := strconv.FormatUint(uint64(id), 10)
			keyList = append(keyList, "ArticleID"+idStr)
		}

		ctx := context.Background()
		res, cacheErr := global.RedisHotPool.MGet(ctx, keyList...).Result()

		var cacheMissIDList []uint
		cacheHitMap := make(map[uint]ArticleSearchListResponse)

		if cacheErr == nil {
			for index, cacheData := range res {
				if cacheData == nil {
					cacheMissIDList = append(cacheMissIDList, articleIDList[index])
					continue
				}

				var cached ArticleDetailResponse
				err = json.Unmarshal([]byte(cacheData.(string)), &cached)
				if err != nil {
					logrus.Warnf("缓存数据解析错误: %s", err)
					cacheMissIDList = append(cacheMissIDList, articleIDList[index])
					continue
				}

				item := ArticleSearchListResponse{
					ArticleModel:  cached.ArticleModel,
					AdminTop:      articleTopMap[cached.ID],
					CategoryTitle: cached.CategoryTitle,
					UserNickname:  cached.NickName,
					UserAvatar:    cached.UserAvatar,
				}
				if article, ok := searchArticleMap[cached.ID]; ok {
					item.Title = article.Title
					item.Abstract = article.Abstract
				}
				cacheHitMap[cached.ID] = item
			}
		} else {
			cacheMissIDList = articleIDList
		}

		if len(cacheMissIDList) > 0 {
			where := global.DB.Where("id in ?", cacheMissIDList)
			modelList, _, err := common.ListQuery(models.ArticleModel{}, common.Options{
				Where:        where,
				Preloads:     []string{"CategoryModel", "UserModel"},
				DefaultOrder: sql.ConvertSliceOrderSql(cacheMissIDList),
			})
			if err != nil {
				logrus.Errorf("查询文章详情失败 %s", err)
				response.FailWithMsg("查询失败", c)
				return
			}

			for _, model := range modelList {
				item := ArticleSearchListResponse{
					ArticleModel: model,
					AdminTop:     articleTopMap[model.ID],
					UserNickname: model.UserModel.NickName,
					UserAvatar:   model.UserModel.Avatar,
				}
				if model.CategoryModel != nil {
					item.CategoryTitle = &model.CategoryModel.Title
				}
				if article, ok := searchArticleMap[model.ID]; ok {
					item.Title = article.Title
					item.Abstract = article.Abstract
				}
				cacheHitMap[model.ID] = item
			}
		}

		list := make([]ArticleSearchListResponse, 0, len(articleIDList))
		for _, id := range articleIDList {
			if item, ok := cacheHitMap[id]; ok {
				list = append(list, item)
			}
		}

		//TODO:以后加入带点赞数和评论数的响应字段
		// 返回成功响应，包含文章列表和总数
		response.OkWithList(list, int(count), c)
	}
}
