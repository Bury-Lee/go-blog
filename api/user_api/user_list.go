package user_api

import (
	"StarDreamerCyberNook/common"
	"StarDreamerCyberNook/common/response"
	"StarDreamerCyberNook/models"
	"StarDreamerCyberNook/models/enum"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type UserListRequest struct {
	common.PageInfo
}

func (UserApi) UserListView(c *gin.Context) {
	var req UserListRequest
	if err := c.ShouldBind(&req); err != nil {
		response.FailWithMsg("参数错误", c)
		return
	}
	var options common.Options
	options.PageInfo = req.PageInfo
	options.Likes = []string{"nick_name", "abstract"}
	options.DefaultOrder = "id desc"
	list, count, err := common.ListQuery[models.UserModel](models.UserModel{}, options)
	if err != nil {
		logrus.Errorf("查询用户列表失败 %s", err)
		response.FailWithMsg("查询失败", c)
		return
	}
	for i := 0; i < len(list)-1; i++ {
		list[i].Password = ""
		list[i].RegisterSource = enum.RegisterSource(0)
		list[i].Email = ""
		list[i].OpenID = ""
	}
	response.OkWithList(list, count, c)
}
