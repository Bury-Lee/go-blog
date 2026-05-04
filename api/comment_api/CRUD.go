package comment_api

import (
	"StarDreamerCyberNook/common"
	"StarDreamerCyberNook/common/response"
	"StarDreamerCyberNook/global"
	"StarDreamerCyberNook/models"
	"StarDreamerCyberNook/models/enum"
	"StarDreamerCyberNook/service/ai_service"
	"StarDreamerCyberNook/service/message_service"
	"StarDreamerCyberNook/service/redis_service/redis_count"
	jwts "StarDreamerCyberNook/utils/jwts"
	utils_other "StarDreamerCyberNook/utils/other"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// 注:评论没有修改这一说,只有增删查
type CommentCreateRequest struct {
	Content   string `json:"content" binding:"required"`
	ArticleID uint   `json:"articleID" binding:"required"`
	ParentID  uint   `json:"parentID"` // 父评论ID
}

// 对于更新就由前端来做吧,收到成功响应之后直接更新
// TODO:创建和删除评论文章的评论数统计在redis进行,需要在创建和删除评论时更新缓存
func (CommentApi) CommentCreateView(c *gin.Context) {
	//流程:参数和文章合规性检验
	//填充内容
	//设置评论关系+发消息
	var req CommentCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMsg("参数错误", c)
		return
	}

	var article models.ArticleModel                                                                       //必须是发布的文章才能发评论
	err := global.DB.Take(&article, "id = ? and status = ?", req.ArticleID, models.StatusPublished).Error //TODO:调试时用,发布后要去掉Debug()
	if err != nil {
		response.FailWithMsg("文章不存在", c)
		return
	}
	claims := jwts.GetClaims(c)

	model := models.CommentModel{ //当前评论模型
		Content:   req.Content,
		UserID:    claims.UserID,
		ArticleID: req.ArticleID,
	}

	//ai审核环节
	/*
		但是这样依然可能出现问题,例如说,发布评论为
		第一条:我
		第二条:干
		第三条:你
		第四条:妈
		像这样大概率可以通过ai审核,但是确实是违规评论,得考虑这样的问题
		如果说要查询评论的历史信息,把上下文评论也查询出来加入到审核内容中,又很麻烦而且显得不合理,也会给数据库带来比较大的压力
	*/
	if global.Config.AI.Enable { //启用ai审核
		reply, err := ai_service.CreateSingleReply(
			"评论内容:"+req.Content,
			global.SystemPromptComment.String(),
		)
		if err != nil {
			response.FailWithMsg("ai审核失败,已经自动创建为待审核状态: "+err.Error(), c)
			return
		}
		switch reply { //TODO:这里无论成功还是失败都应该插入消息,告知原因
		case "通过":
			//通过就正常执行流程
		case "拒绝":
			response.FailWithMsg("评论可能含有违规信息,已拒绝", c)
			return
		default:
			logrus.Errorf("ai审核出错,回复内容:%s,评论详情:%s\n,评论不发布", reply, fmt.Sprintf("%#v", req))
			return
		}
	}

	if req.ParentID == 0 {
		// 父评论路径为空,说明这个评论是一级评论
		model.RootParentID = nil
		model.ParentPath = ""
	}
	// 否则这是二级评论
	if req.ParentID != 0 {
		//不对,现在无论是二级评论怎么样,都应该是:model.RootParentID=Parentmodel.Parentpath的/.../,而父评论是最后一个路径,这样哪怕是次级回复也可以有正确的逻辑
		//为了节省空间,入库时应该以base64编码计入
		var parentModel models.CommentModel
		err := global.DB.Take(&parentModel, "id = ? and article_id = ?", req.ParentID, req.ArticleID).Error //TODO:调试时用,发布后要去掉Debug()
		if err != nil {
			response.FailWithMsg("评论不存在", c) // 回复的评论得是这个文章里的
			return
		}
		model.ParentPath = utils_other.EncodePath(parentModel.ParentPath, parentModel.ID)
		//给父评论发消息
		message_service.InsertReplyMessage(model, parentModel.UserID)
		if parentModel.RootParentID == nil {
			// 如果父评论本身是根评论 (RootParentID 为 nil)，则当前评论的根即为父评论,发布的新评论为二级评论
			//发消息:给根评论发消息
			model.RootParentID = &parentModel.ID
		} else {
			// 如果父评论不是根评论，则直接继承其根评论ID，无需重新解码,发布的新评论为二级的次级评论
			model.RootParentID = parentModel.RootParentID

			//这里只能再查一次数据库看看根评论的用户ID来获取根评论的用户ID
			var rootParentModel models.CommentModel
			err = global.DB.Take(&rootParentModel, "id = ? and article_id = ?", parentModel.RootParentID, req.ArticleID).Error //走到这一步,说明这是二级评论,二级评论的根评论不可能不存在
			if err != nil {
				logrus.Errorf("系统消息发送失败,无法查询.父评论ID: %v, 文章ID: %v", parentModel.RootParentID, req.ArticleID)
			} else if rootParentModel.UserID != parentModel.UserID { //如果根评论的用户ID和父评论的用户ID不一样,说明是不同的人,就给根评论发消息,如果是同一个人就不发了,避免重复发消息了
				err = message_service.InsertReplyMessage(model, rootParentModel.UserID)
				if err != nil {
					logrus.Error("系统消息发送失败")
				}
			}
		}
	}
	//无论是几级评论都要给文章作者发消息
	message_service.InsertCommentMessage(model, article.UserID)

	err = global.DB.Create(&model).Error
	if err != nil {
		response.FailWithMsg("发布评论失败", c)
		return
	}

	// 要给作者,一级评论,父评论发评论消息,在不同的位置插入.已在上面插入了发消息的代码,所以这里就不需要再发一次了,避免重复发消息了
	// err = message_service.InsertCommentMessage(model, article.UserID) //一定给文章作者发消息
	// if err != nil {
	// 	logrus.Errorf("系统消息发送失败:%s", err.Error())
	// }

	// //以后可以考虑一下优化查询流程,避免重复查询数据库
	// if model.RootParentID != nil { //存在根评论,说明这是二级评论,给根评论发消息
	// 	//要先查一遍评论表得到根评论的用户ID,然后才能给根评论发消息
	// 	var RootParent models.CommentModel
	// 	err = global.DB.Find(&RootParent, "id = ? and article_id = ?", model.RootParentID, req.ArticleID).Error
	// 	if err != nil {
	// 		logrus.Errorf("系统消息发送失败:%s", err.Error())
	// 	}
	// 	message_service.InsertReplyMessage(model, RootParent.UserID) //给根评论发消息
	// 	if req.ParentID != 0 {                                       //如果父评论也是根评论同一个人,就不重复发消息了
	// 		//如果有父评论,给父评论发消息
	// 		var Parent models.CommentModel
	// 		ID, _ := model.Decode()
	// 		err = global.DB.Find(&Parent, "id = ? and article_id = ?", ID, req.ArticleID).Error
	// 		if err != nil {
	// 			logrus.Errorf("系统消息发送失败:%s", err.Error())
	// 		}
	// 		if Parent.UserID != RootParent.UserID {
	// 			message_service.InsertReplyMessage(model, RootParent.UserID) //给父评论发消息
	// 		}
	// 	}
	// }

	//旧方法,直接打到数据库,现在改为先更新缓存,然后定时任务再批量更新到数据库
	// if global.DB.Model(&models.ArticleModel{}).Where("id = ?", req.ArticleID).Select("comment_count").Updates(map[string]interface{}{
	// 	"comment_count": gorm.Expr("comment_count + ?", 1),
	// }).Error != nil {
	// 	logrus.Error("文章评论数更新失败,文章ID:", req.ArticleID)
	// }
	// 发布评论后,需要更新文章的评论数

	redis_count.SetCacheComment(req.ArticleID, true) //增量更新缓存

	response.OkWithMsg("发布评论成功", c)
}

// 分页获取指定文章的一级评论
type CommentDetailRequest struct {
	common.PageInfo
	ArticleID uint `form:"articleID" binding:"required"` // 文章ID
}

/*
方案：只缓存“评论 ID 列表” (推荐)
不要把完整的评论内容存入 ZSET，ZSET 只存 排序索引。
Redis ZSET 结构：
Key: article:comments:hot:{article_id}
Member: comment_id (例如: 1001, 1005)
Score: digg_count (点赞数) 或者 timestamp (如果是按时间排序)
注意：如果是按点赞数排序，Score 会动态变化，需要频繁更新 ZSET 的 Score。
缓存策略：
第一步： 请求进来，先去 Redis 执行 ZREVRANGE key 0 19 (获取前20个热门评论ID)。
第二步： 拿到 ID 列表后，去 Redis 的 String/Hash 缓存中批量获取评论详情（HMGET comments:detail {id1} {id2}...）。
第三步： 如果详情缓存未命中，再去数据库查，查完后回填缓存。
优点：
ZSET 非常轻量，只存 ID 和分数。
更新点赞数时，只需 ZINCRBY，不需要移动整个对象。
*/
func (CommentApi) CommentListlView(c *gin.Context) { //获取某文章的一级评论(分页)
	//TODO:redis缓存
	var req CommentDetailRequest
	if err := c.ShouldBind(&req); err != nil {
		response.FailWithMsg("参数错误", c)
		return
	}
	//先检查文章状态
	var article models.ArticleModel
	err := global.DB.Take(&article, "id = ? and status = ?", req.ArticleID, models.StatusPublished).Error //TODO:调试时用,发布后要去掉Debug()
	if err != nil {
		response.FailWithMsg("文章不存在", c)
		return
	}

	var comments models.CommentModel
	var options common.Options
	options.PageInfo = req.PageInfo
	options.Preloads = []string{"UserModel"}                                                    //预加载用户信息
	options.Where = global.DB.Where("article_id = ? and root_parent_id is null", req.ArticleID) //查询一级评论
	options.DefaultOrder = "digg_count desc"                                                    //默认按点赞数降序排序
	List, count, err := common.ListQuery(comments, options)
	if err != nil { //TODO:加一个点赞增量也加上的逻辑
		response.FailWithMsg("查询评论失败", c)
		return
	}
	for i, v := range List {
		var UserModel = models.UserModel{
			Model:         v.UserModel.Model,
			NickName:      v.UserModel.NickName,
			Avatar:        v.UserModel.Avatar,
			LastLoginTime: v.UserModel.LastLoginTime,
			Age:           v.UserModel.Age,
			LikeTags:      v.UserModel.LikeTags,
		}

		List[i].UserModel = UserModel
	}
	response.OkWithList(List, count, c)
}

// 分页获取某条评论下的子评论详情(多条)
type CommentListRequest struct {
	common.PageInfo
	Root uint `form:"root" binding:"required"` //根评论ID
}

func (CommentApi) CommentChildListView(c *gin.Context) { //可以这样,评论被删除了内容就替换为评论已删除
	var req CommentListRequest
	if err := c.ShouldBind(&req); err != nil {
		response.FailWithMsg("参数错误", c)
		return
	}
	var comments models.CommentModel
	//先检查父评论是否存在
	err := global.DB.Take(&comments, "id = ?", req.Root).Error
	if err != nil {
		response.FailWithMsg("评论不存在", c)
		return
	}

	//检查文章状态
	var article models.ArticleModel
	err = global.DB.Take(&article, "id = ? and status = ?", comments.ArticleID, models.StatusPublished).Error
	if err != nil {
		response.FailWithMsg("文章不存在", c)
		return
	}

	var options common.Options
	options.PageInfo = req.PageInfo
	options.Preloads = []string{"UserModel"} //预加载用户信息
	options.Where = global.DB.Where(
		"article_id = ? and parent_path like ?",
		comments.ArticleID,
		utils_other.EncodePath(comments.ParentPath, comments.ID)+"%", //任意长度的匹配字符串
	) //直接使用父评论的文章ID,查询有相同根评论且父路径包含当前评论路径的所有评论
	//考虑两个问题,
	//1.如果父评论被删除了,那么子评论的parent_path就会失效,但是子评论的root_parent_id还会指向根评论,这时候就需要根据root_parent_id来查询
	//如果用户传入的Rootid是乱写的怎么办,放回err吧
	//用户的
	options.DefaultOrder = "digg_count desc" //默认按点赞数降序排序
	List, count, err := common.ListQuery(models.CommentModel{}, options)
	if err != nil {
		response.FailWithMsg("查询评论失败", c)
		return
	}
	for i, v := range List {
		var UserModel = models.UserModel{
			Model:         v.UserModel.Model,
			NickName:      v.UserModel.NickName,
			Avatar:        v.UserModel.Avatar,
			LastLoginTime: v.UserModel.LastLoginTime,
			Age:           v.UserModel.Age,
			LikeTags:      v.UserModel.LikeTags,
		}

		List[i].UserModel = UserModel
	}
	response.OkWithList(List, count, c)
}

// 删除指定评论及其子评论(?其实子评论不一定要删?
// TODO:创建和删除评论文章的评论数统计在redis进行,需要在创建和删除评论时更新缓存
func (CommentApi) CommentDeleteView(c *gin.Context) {
	var req models.IDRequest
	if err := c.ShouldBindUri(&req); err != nil {
		response.FailWithMsg("参数错误", c)
		return
	}
	var comment models.CommentModel
	err := global.DB.Take(&comment, "id = ?", req.ID).Error
	if err != nil {
		response.FailWithMsg("评论不存在", c)
		return
	}
	//先检查评论是否存在
	//检查文章状态
	var article models.ArticleModel
	err = global.DB.Take(&article, "id = ? and status = ?", comment.ArticleID, models.StatusPublished).Error //TODO:调试时用,发布后要去掉Debug()
	if err != nil {
		response.FailWithMsg("文章不存在", c)
		return
	}
	//鉴权,检查用户是否有删除评论的权限,只能删自己的,管理员可以删除别人的
	claim := jwts.GetClaims(c)
	if claim.UserID != comment.UserID && claim.Role != enum.AdminRole {
		response.FailWithMsg("没有权限删除评论", c)
		return
	}

	if comment.RootParentID == nil {
		//一级评论,连带着二级评论一起删除
		var count int64
		global.DB.Delete(&models.CommentModel{}, "root_parent_id = ?", comment.ID).Count(&count) //TODO:移除DEBUG
		if global.DB.Model(&models.ArticleModel{}).Where("id = ?", comment.ArticleID).Select("comment_count").Updates(map[string]interface{}{
			"comment_count": gorm.Expr("comment_count - ?", count+1), //这里的count是删除的二级评论的数量加上一条本身的数量,所以要加1
		}).Error != nil {
			logrus.Error("文章评论数更新失败,文章ID:", comment.ArticleID)
		}
	} else { //只删除自己
		global.DB.Delete(&comment)
		if global.DB.Model(&models.ArticleModel{}).Where("id = ?", comment.ArticleID).Select("comment_count").Updates(map[string]interface{}{
			"comment_count": gorm.Expr("comment_count - ?", 1),
		}).Error != nil {
			logrus.Error("文章评论数更新失败,文章ID:", comment.ArticleID)
		}
	}

	//统计缓存-1
	redis_count.SetCacheComment(comment.ArticleID, false) //增量更新缓存
	response.OkWithMsg("删除评论成功", c)
}

// 评论点赞
func (CommentApi) CommentDiggView(c *gin.Context) {
	var req models.IDRequest
	if err := c.ShouldBindUri(&req); err != nil {
		response.FailWithMsg("参数错误", c)
		return
	}
	var comment models.CommentModel
	err := global.DB.Take(&comment, req.ID).Error
	if err != nil {
		response.FailWithMsg("评论不存在", c)
		return
	}

	claim := jwts.GetClaims(c) //要先登录才能点赞
	if global.DB.Take(&models.CommentDiggModel{}, "user_id = ? and comment_id = ?", claim.UserID, req.ID).Error == gorm.ErrRecordNotFound {
		//查询不到说明没有点赞过,可以点赞
		if global.DB.Create(&models.CommentDiggModel{
			UserID:    claim.UserID,
			CommentID: req.ID,
		}).Error != nil {
			response.FailWithMsg("点赞失败", c)
			return
		}
		redis_count.SetCacheCommentDigg(req.ID, true) //增量加一
	} else {
		// 查询到说明已经点赞过,取消点赞
		if global.DB.Delete(&models.CommentDiggModel{}, "user_id = ? and comment_id = ?", claim.UserID, req.ID).Error != nil {
			response.FailWithMsg("取消点赞失败", c)
			return
		}
		redis_count.SetCacheCommentDigg(req.ID, false) //增量减一
	}

	response.OkWithData(claim, c)
}
