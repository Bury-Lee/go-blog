package follow_api

import (
	"StarDreamerCyberNook/common"
	"StarDreamerCyberNook/common/response"
	"StarDreamerCyberNook/global"
	"StarDreamerCyberNook/models"
	jwts "StarDreamerCyberNook/utils/jwts"

	"github.com/gin-gonic/gin"
)

// FollowerListRequest 定义了获取粉丝列表接口的请求参数结构体
type FollowerListRequest struct {
	common.PageInfo      // 匿名嵌入分页信息
	UserID          uint `form:"userID"` // 被查询的用户ID。如果为0，则查询当前登录用户的粉丝列表；否则查询指定用户的粉丝列表。
}

// FollowerListView 是处理获取粉丝列表请求的Gin路由处理器
// 支持查询指定用户或当前登录用户的粉丝列表，并包含权限校验和分页功能。
func (FollowApi) FollowerListView(c *gin.Context) { //迟点检查一下
	// 1. 绑定并验证URL查询参数
	var req FollowerListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithMsg("参数错误", c)
		return
	}

	// 2. 获取当前登录用户的身份信息 (JWT Claims)
	// 如果 req.UserID 为 0，表示查询的是当前登录用户自己的粉丝，需要验证登录状态。
	var claim *jwts.MyClaims
	if req.UserID == 0 {
		claim = jwts.GetClaims(c)
		if claim == nil {
			response.FailWithMsg("请登录", c)
			return
		}
		// 将当前登录用户ID赋值给 req.UserID，方便后续查询
		req.UserID = claim.UserID
	}

	// 3. 权限校验：检查被查询用户的粉丝列表是否对当前用户开放
	var userConf models.UserConfModel
	global.DB.Where("user_id = ?", req.UserID).First(&userConf) // 使用 First 更精确，只取一条记录

	// 如果目标用户关闭了粉丝列表的公开访问，并且当前访问者不是用户本人，则拒绝访问
	if !userConf.OpenFans && (claim == nil || claim.UserID != req.UserID) {
		response.FailWithMsg("此用户未公开我的粉丝", c)
		return
	}

	// 4. 构建查询选项并执行数据库查询
	option := common.Options{
		PageInfo:     req.PageInfo,          // 分页参数
		DefaultOrder: "created_at desc",     // 默认按关注创建时间倒序排列
		Preloads:     []string{"UserModel"}, // 预加载关联的用户信息（例如粉丝的昵称、头像等）
	}

	// 执行通用列表查询方法
	// 查询条件是 UserFocusModel 中的 UserID 字段等于目标用户的 ID
	list, count, err := common.ListQuery[models.UserFollowModel](
		models.UserFollowModel{UserID: req.UserID},
		option,
	)

	if err != nil {
		response.FailWithMsg("查询粉丝列表失败", c)
		return
	}

	// 5. 返回成功响应
	// list: 粉丝列表数据
	// count: 总粉丝数，用于前端分页
	response.OkWithList(list, count, c)
}
