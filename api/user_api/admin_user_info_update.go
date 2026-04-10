package user_api

import (
	"StarDreamerCyberNook/common/response"
	"StarDreamerCyberNook/global"
	"StarDreamerCyberNook/models"
	"StarDreamerCyberNook/models/enum"
	utils_other "StarDreamerCyberNook/utils/other"

	"github.com/gin-gonic/gin"
)

type AdminUserInfoUpdateRequest struct { //这些是最容易出现违规的地方
	UserID   uint           `json:"userID" binding:"required"`
	Username *string        `json:"username" s-u:"username"`
	Nickname *string        `json:"nickname" s-u:"nickname"`
	Avatar   *string        `json:"avatar" s-u:"avatar"`
	Abstract *string        `json:"abstract" s-u:"abstract"`
	Role     *enum.RoleType `json:"role" s-u:"role"`
}

func (UserApi) AdminUserInfoUpdateView(c *gin.Context) {
	var req AdminUserInfoUpdateRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		response.FailWithError(err, c)
		return
	}
	userMap := utils_other.StructToMap(req, "s-u")
	var user models.UserModel
	err = global.DB.Take(&user, req.UserID).Error
	if err != nil {
		response.FailWithMsg("用户不存在", c)
		return
	}

	err = global.DB.Model(&user).Updates(userMap).Error
	if err != nil {
		response.FailWithMsg("用户信息修改失败", c)
		return
	}

	response.OkWithMsg("用户信息修改成功", c)
}
