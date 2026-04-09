package user_api

import (
	"StarDreamerCyberNook/common/response"
	"StarDreamerCyberNook/global"
	"StarDreamerCyberNook/models"
	"StarDreamerCyberNook/models/enum"
	jwts "StarDreamerCyberNook/utils/jwts"
	"time"

	"github.com/gin-gonic/gin"
)

type UserDetailResponse struct { //个人主页的返回
	models.Model
	UserName    string            ` json:"username"` //用户名
	NickName    string            ` json:"nickname"` //昵称
	Avatar      string            ` json:"avatar"`   //头像
	Abstract    string            ` json:"abstract"` //简介
	Age         int               `json:"Age"`       //年龄
	LikeTags    []string          ` json:"likeTags"` //兴趣标签
	ContactInfo map[string]string ` json:"contactInfo"`
	Role        enum.RoleType     `json:"role"`

	//以下为配置表的字段
	UpdateUsernameDate *time.Time `json:"updateUsernameDate"` // 上次修改用户名的时间,因为可能没改过,避免无法区分nil,使用指针
	OpenCollect        bool       `json:"openCollect"`        // 公开我的收藏
	OpenFollow         bool       `json:"openFollow"`         // 公开我的关注
	OpenFans           bool       `json:"openFans"`           // 公开我的粉丝
	HomeStyleID        uint       `json:"homeStyleID"`        // 主页样式的id
}

func (UserApi) UserDetailView(c *gin.Context) {
	claims := jwts.GetClaims(c)
	print(claims) //debug
	if claims == nil {
		response.FailWithMsg("未登录", c)
		return
	}
	var user models.UserModel
	if err := global.DB.Preload("UserConfModel").Take(&user, claims.UserID).Error; err != nil {
		response.FailWithMsg("用户不存在", c)
		return
	}

	var result = UserDetailResponse{
		Model:       user.Model,
		UserName:    user.UserName,
		NickName:    user.NickName,
		Avatar:      user.Avatar,
		Abstract:    user.Abstract,
		Age:         user.Age,
		ContactInfo: user.ContactInfo,
		Role:        user.Role,
	}

	if user.UserConfModel != nil {
		result.UpdateUsernameDate = user.UserConfModel.UpdateUsernameDate
		result.OpenCollect = user.UserConfModel.OpenCollect
		result.OpenFollow = user.UserConfModel.OpenFollow
		result.OpenFans = user.UserConfModel.OpenFans
		result.HomeStyleID = user.UserConfModel.HomeStyleID
	}
	response.OkWithData(result, c)
}
