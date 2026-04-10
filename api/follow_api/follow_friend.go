package follow_api

import (
	"StarDreamerCyberNook/common"
	"StarDreamerCyberNook/common/response"
	"StarDreamerCyberNook/models"
	jwts "StarDreamerCyberNook/utils/jwts"
	"time"

	"github.com/gin-gonic/gin"
)

// 保护隐私,不能看别人的好友列表
type FriendUserListRequest struct { //这样在体验上,如果不传入userID,就默认查我的关注,如果传了userID,就忽略UserID查这个人关注的用户列表
	common.PageInfo
}
type FriendUserListResponse struct {
	FocusUserID       uint      `json:"focusUserID"`
	FocusUserNickName string    `json:"focusUserNickname"`
	FocusUserAvatar   string    `json:"focusUserAvatar"`
	FocusUserAbstract string    `json:"focusUserAbstract"`
	CreatedAt         time.Time `json:"createdAt"`
}

// UserListView 我的关注和用户的关注
func (FollowApi) FriendUserListView(c *gin.Context) {
	var req FriendUserListRequest
	// 绑定参数
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithMsg("参数错误", c)
		return
	}
	claim := jwts.GetClaims(c)
	if claim == nil {
		response.FailWithMsg("请登录", c)
		return
	}
	_list, count, _ := common.ListQuery[models.UserFollowModel](models.UserFollowModel{
		UserID: claim.UserID,
		Friend: true,
	}, common.Options{
		PageInfo: req.PageInfo,
		Preloads: []string{"FocusUserModel"},
	})

	var list = make([]FriendUserListResponse, 0)
	for _, model := range _list {
		list = append(list, FriendUserListResponse{
			FocusUserID:       model.FocusUserID,
			FocusUserNickName: model.FocusUserModel.NickName,
			FocusUserAvatar:   model.FocusUserModel.Avatar,
			FocusUserAbstract: model.FocusUserModel.Abstract,
			CreatedAt:         model.CreatedAt,
		})
	}

	response.OkWithList(list, count, c)
}
