package follow_api

import (
	"StarDreamerCyberNook/common"
	"StarDreamerCyberNook/common/response"
	"StarDreamerCyberNook/global"
	"StarDreamerCyberNook/models"
	jwts "StarDreamerCyberNook/utils/jwts"
	"time"

	"github.com/gin-gonic/gin"
)

type FollowUserListRequest struct { //这样在体验上,如果不传入userID,就默认查我的关注,如果传了userID,就忽略UserID查这个人关注的用户列表
	common.PageInfo
	UserID uint `form:"userID"` // 查用户的关注
}
type FollowUserListResponse struct {
	FocusUserID       uint      `json:"focusUserID"`
	FocusUserNickName string    `json:"focusUserNickname"`
	FocusUserAvatar   string    `json:"focusUserAvatar"`
	FocusUserAbstract string    `json:"focusUserAbstract"`
	CreatedAt         time.Time `json:"createdAt"`
}

// FollowUserListView 我的关注和用户的关注
func (FollowApi) FollowUserListView(c *gin.Context) {
	var req FollowUserListRequest
	// 绑定参数
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithMsg("参数错误", c)
		return
	}
	var claims = jwts.GetClaims(c)
	if req.UserID == 0 {
		var err error
		claims, err = jwts.ParseTokenByGin(c)
		if err != nil || claims == nil {
			response.FailWithMsg("请登录", c)
			return
		}
		req.UserID = claims.UserID
	}
	var userConf models.UserConfModel
	err := global.DB.Take(&userConf, "user_id = ?", req.UserID).Error
	if err != nil {
		response.FailWithMsg("用户配置信息不存在", c)
		return
	}
	if !userConf.OpenFollow && req.UserID != claims.UserID {
		response.FailWithMsg("此用户未公开我的关注", c)
		return
	}

	_list, count, _ := common.ListQuery[models.UserFollowModel](models.UserFollowModel{
		UserID: req.UserID, //是这样吗?
	}, common.Options{
		PageInfo: req.PageInfo,
		Preloads: []string{"FocusUserModel"},
	})

	var list = make([]FollowUserListResponse, 0)
	for _, model := range _list {
		list = append(list, FollowUserListResponse{
			FocusUserID:       model.FocusUserID,
			FocusUserNickName: model.FocusUserModel.NickName,
			FocusUserAvatar:   model.FocusUserModel.Avatar,
			FocusUserAbstract: model.FocusUserModel.Abstract,
			CreatedAt:         model.CreatedAt,
		})
	}

	response.OkWithList(list, count, c)
}
