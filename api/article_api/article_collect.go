package article_api

import (
	"StarDreamerCyberNook/common"
	"StarDreamerCyberNook/common/response"
	"StarDreamerCyberNook/global"
	"StarDreamerCyberNook/models"
	"StarDreamerCyberNook/models/enum"
	"StarDreamerCyberNook/service/message_service"
	"StarDreamerCyberNook/service/redis_service/redis_count"
	jwts "StarDreamerCyberNook/utils/jwts"
	utils_other "StarDreamerCyberNook/utils/other"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type ArticleCollectRequest struct {
	ArticleID uint `json:"articleID" binding:"required"`
	CollectID uint `json:"collectID"`
}

func (ArticleApi) ArticleCollectView(c *gin.Context) {
	var req ArticleCollectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMsg("参数错误", c)
		return
	}

	var article models.ArticleModel
	err := global.DB.Take(&article, "status = ? and id = ?", models.StatusPublished, req.ArticleID).Error
	if err != nil {
		response.FailWithMsg("文章不存在", c)
		return
	}
	var collectModel models.CollectModel
	claims := jwts.GetClaims(c)
	if req.CollectID == 0 {
		// 是默认收藏夹
		err = global.DB.Take(&collectModel, "user_id = ? and is_default = ?", claims.UserID, 1).Error
		if err != nil {
			// 创建一个默认收藏夹
			collectModel.Title = "默认收藏夹"
			collectModel.UserID = claims.UserID
			collectModel.IsDefault = true
			global.DB.Create(&collectModel)
		}
		req.CollectID = collectModel.ID
	} else {
		// 判断收藏夹是否存在，并且是否是自己创建的
		err = global.DB.Take(&collectModel, "user_id = ? ", claims.UserID).Error
		if err != nil {
			response.FailWithMsg("收藏夹不存在", c)
			return
		}
	}

	// 判断是否收藏
	var articleCollect models.UserArticleCollectModel
	err = global.DB.Where(models.UserArticleCollectModel{
		UserID:    claims.UserID,
		ArticleID: req.ArticleID,
		CollectID: req.CollectID,
	}).Take(&articleCollect).Error

	if err != nil {
		// 收藏
		err = global.DB.Create(&models.UserArticleCollectModel{
			UserID:    claims.UserID,
			ArticleID: req.ArticleID,
			CollectID: req.CollectID,
		}).Error
		if err != nil {
			response.FailWithMsg("收藏失败", c)
			return
		}
		response.OkWithMsg("收藏成功", c)
		// 发送收藏消息
		err = message_service.InsertCollectMessage(articleCollect)
		if err != nil {
			logrus.Error("发送收藏消息失败", err.Error())
			return
		}
		// 对收藏夹进行加1
		redis_count.SetCacheCollect(req.ArticleID, true)
		return
	}
	// 取消收藏
	err = global.DB.Where(models.UserArticleCollectModel{
		UserID:    claims.UserID,
		ArticleID: req.ArticleID,
		CollectID: req.CollectID,
	}).Delete(&models.UserArticleCollectModel{}).Error

	if err != nil {
		response.FailWithMsg("取消收藏失败", c)
		return
	}
	response.OkWithMsg("取消收藏成功", c)
	//TODO:收藏统计改为使用redis缓存,然后redis定时任务和数据库同步更新,而不要直接更新数据库了,有并发问题
	// global.DB.Model(&collectModel).Update("article_count", gorm.Expr("article_count - 1"))
	redis_count.SetCacheCollect(req.ArticleID, false)
}

type CollectCreateRequest struct { //创建收藏夹请求参数,请求创建时不用传id参数,除了创建也可以用于更新收藏夹
	Title    string `json:"title" binding:"required,max=32" s:"title"`
	Abstract string `json:"abstract" s:"abstract"`
	Cover    string `json:"cover" s:"cover"`
}

func (ArticleApi) CollectCreateView(c *gin.Context) {
	var req CollectCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMsg("参数错误", c)
		return
	}

	claims := jwts.GetClaims(c)
	//不允许创建重复名称的收藏夹
	// var model models.CollectModel
	// err := global.DB.Take(&model, "user_id = ? and title = ?", claims.UserID, req.Title).Error
	// if err == nil {
	// 	response.FailWithMsg("收藏夹名称重复", c)
	// 	return
	// }

	// 创建
	if global.DB.Create(&models.CollectModel{
		Title:    req.Title,
		UserID:   claims.UserID,
		Abstract: req.Abstract,
		Cover:    req.Cover,
	}).Error != nil {
		response.FailWithMsg("创建收藏夹失败", c)
		return
	}
	response.OkWithMsg("创建收藏夹成功", c)
}

type CollectUpdateRequest struct { //创建收藏夹请求参数,请求创建时不用传id参数,除了创建也可以用于更新收藏夹
	ID       uint    `json:"id" binding:"required"` //更新时需要传id参数
	Title    *string `json:"title" `
	Abstract *string `json:"abstract"`
	Cover    *string `json:"cover"`
}

func (ArticleApi) CollectUpdateView(c *gin.Context) {
	var req CollectUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMsg("参数错误", c)
		return
	}

	claims := jwts.GetClaims(c)
	var model models.CollectModel
	err := global.DB.Take(&model, "user_id = ? and id = ?", claims.UserID, req.ID).Error
	if err != nil {
		response.FailWithMsg("收藏夹不存在", c)
		return
	}

	updateMap := utils_other.StructToMap(&req, "sql") //把请求参数转换成map,方便后续更新,并且只更新有值的字段
	//也许不允许更新成重复名称的收藏夹?
	//现在来看还允许吧,给用户更高的自由度,毕竟收藏夹名称也不是很重要,而且用户也可以通过id和封面区分不同的收藏夹,所以就不做这个限制了

	err = global.DB.Model(&model).Updates(updateMap).Error
	if err != nil {
		response.FailWithMsg("更新收藏夹失败", c)
		return
	}

	response.OkWithMsg("更新收藏夹成功", c)
}

func (ArticleApi) CollectRemoveView(c *gin.Context) {
	var req models.RemoveRequest
	// 1. 参数绑定
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMsg("参数错误", c)
		return
	}
	//其实用户可能传入0 ID进来,不过问题应该不大,数据库可以处理这个错误
	// 2. 基础查询构建 (注意：不要提前执行 Find)
	// 假设模型名为 CollectModel，表名为 collects
	query := global.DB.Model(&models.CollectModel{}).
		Where("id IN ? AND is_default = ?", req.IDList, false)

	// 3. 权限控制
	claims := jwts.GetClaims(c)
	if claims.Role != enum.AdminRole {
		// 非管理员只能删除自己的收藏夹
		// 将 user_id 条件链式追加到 query 中
		query = query.Where("user_id = ?", claims.UserID)
	}

	// 4. 执行删除操作
	// 直接执行 Delete，不需要先 Find 再 Delete，减少一次数据库交互
	// Delete 会自动根据前面的 Where 条件生成 SQL
	result := query.Delete(&models.CollectModel{})

	if result.Error != nil {
		// 数据库层面错误
		response.FailWithMsg("删除收藏夹失败: "+result.Error.Error(), c)
		return
	}

	// 5. 检查受影响行数
	if result.RowsAffected == 0 {
		// 情况 A: ID 列表为空
		// 情况 B: ID 不存在
		// 情况 C: 所有 ID 都是默认收藏夹 (is_default=true)
		// 情况 D: 非管理员尝试删除他人的收藏夹 (被 user_id 过滤)
		// 根据业务需求，这里可以返回“未找到可删除项”或者直接视为成功（幂等性）
		// 通常如果没有删除任何数据，提示“无符合条件的记录”比报“失败”更友好
		response.OkWithMsg("未找到可删除的收藏夹或无权限", c)
		// 如果业务严格要求必须删掉才算成功，则改为:
		// response.FailWithMsg("未找到可删除的收藏夹或无权限", c)
		return
	}

	response.OkWithMsg("删除收藏夹成功", c)
}

type CollectListViewRequest struct {
	common.PageInfo
	Likes []string `form:"likes"`
	ID    uint     `form:"id"`
}

// 先写着吧,一般来说是只有好友才能互看收藏夹的,或者以后给管理员看
func (ArticleApi) CollectListView(c *gin.Context) { //先看看用户有没有公开收藏夹,再查询用户收藏夹列表
	var req CollectListViewRequest //查询用户的收藏夹列表,应该使用分页查询吧

	if err := c.ShouldBind(&req); err != nil {
		response.FailWithMsg("参数错误", c)
		return
	}
	if req.ID == 0 {
		response.FailWithMsg("参数错误", c)
		return
	}

	//检查收藏夹所属用户,查看用户是否开启收藏夹列表,如果没有开启且非用户,则返回错误,通过后分页查询收藏夹列表
	var user models.UserConfModel
	claims, _ := jwts.ParseTokenByGin(c)
	global.DB.Where("user_id = ?", req.ID).First(&user)                         //看看用户有没有开启收藏夹功能
	if user.OpenCollect != true && (claims == nil || claims.UserID != req.ID) { //如果用户没有开启收藏夹功能,并且请求者不是用户本人,则返回错误
		response.FailWithMsg("用户未开启收藏夹功能", c)
		return
	}

	//通过验证,组织数据,分页查询
	var query models.CollectModel

	var option common.Options
	//对req的option进行过滤
	option.Where = global.DB.Where("user_id = ?", req.ID) //只允许查询指定ID的收藏夹
	option.Debug = false
	option.Likes = req.Likes
	option.DefaultOrder = "created_at desc"

	data, count, err := common.ListQuery[models.CollectModel](query, option)
	if err != nil {
		response.FailWithMsg("查询收藏夹列表失败", c)
		return
	}
	// global.DB.Where("user_id = ?", req.ID).Find(&list)
	response.OkWithList(data, count, c)
}

type CollectArticleListViewRequest struct {
	common.PageInfo
	Likes []string `form:"likes"`
	ID    uint     `form:"id"`
}

func (ArticleApi) CollectArticleListView(c *gin.Context) {
	var req CollectArticleListViewRequest //查询用户的收藏夹文章列表
	if err := c.ShouldBind(&req); err != nil {
		response.FailWithMsg("参数错误", c)
		return
	}
	if req.ID == 0 {
		response.FailWithMsg("参数错误", c)
		return
	}

	//检查收藏夹所属用户,查看用户是否开启收藏夹列表,如果没有开启且非用户,则返回错误,通过后分页查询收藏夹文章列表
	var collect models.CollectModel
	global.DB.Where("id = ?", req.ID).First(&collect)
	claims := jwts.GetClaims(c)
	if claims != nil && claims.UserID != collect.UserID {
		var User models.UserConfModel
		global.DB.Where("user_id = ?", collect.UserID).First(&User)
		if User.OpenCollect != true {
			response.FailWithMsg("用户未开启收藏夹功能", c)
			return
		}
	}

	//通过验证,组织数据
	var option common.Options
	option.Where = global.DB.Where("collect_id = ?", req.ID) //只允许查询指定ID的收藏夹文章
	option.Debug = false
	option.Likes = req.Likes
	option.DefaultOrder = "created_at desc"
	data, count, err := common.ListQuery[models.ArticleModel](models.ArticleModel{}, option)
	if err != nil {
		response.FailWithMsg("查询收藏夹文章列表失败", c)
		return
	}
	response.OkWithList(data, count, c)
}
