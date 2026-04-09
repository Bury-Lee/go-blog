package user_api

import (
	"StarDreamerCyberNook/common"
	"StarDreamerCyberNook/common/response"
	"StarDreamerCyberNook/global"
	"StarDreamerCyberNook/models"
	jwts "StarDreamerCyberNook/utils/jwts"
	"time"

	"github.com/gin-gonic/gin"
)

type UserLoginListRequest struct {
	common.PageInfo
	UserID    uint   `form:"userId"`
	Ip        string `form:"ip"`
	Addr      string `form:"addr"`
	StartTime string `form:"startTime"` // 起止时间的 年月日时分秒格式
	EndTime   string `form:"endTime"`
	Type      string `form:"type" binding:"required"` // 用户：只能查自己的  管理员 ：能查全部. 使用字符串,更加稳定可控
}
type UserLoginListResponse struct {
	models.UserLoginModel
	UserNickname string `json:"userNickname,omitempty"`
	UserAvatar   string `json:"userAvatar,omitempty"`
}

func (UserApi) UserLoginListView(c *gin.Context) {
	var req UserLoginListRequest
	err := c.ShouldBindQuery(&req)
	if err != nil {
		response.FailWithMsg("请求参数错误", c)
		return
	}
	if req.Type != "user" && req.Type != "admin" {
		response.FailWithMsg("请求参数错误", c) //TODO:迟点再改为switch了
		return
	}

	claims := jwts.GetClaims(c)
	if req.Type == "user" {
		req.UserID = claims.UserID
	}

	var query = global.DB.Where("")
	if req.StartTime != "" {
		_, err = time.Parse("2006-01-02 15:04:05", req.StartTime)
		if err != nil {
			response.FailWithMsg("开始时间格式错误", c)
			return
		}
		query.Where("created_at >= ?", req.StartTime)
	}
	if req.EndTime != "" {
		_, err = time.Parse("2006-01-02 15:04:05", req.EndTime)
		if err != nil {
			response.FailWithMsg("结束时间格式错误", c)
			return
		}
		query.Where("created_at <= ?", req.EndTime)
	}
	var preloads []string
	if req.Type == "admin" {
		preloads = []string{"UserModel"}
	}

	_list, count, _ := common.ListQuery[models.UserLoginModel](models.UserLoginModel{
		UserID: req.UserID,
		IP:     req.Ip,
		Addr:   req.Addr,
	}, common.Options{
		PageInfo: req.PageInfo,
		Where:    query,
		Preloads: preloads,
	})

	var list = make([]UserLoginListResponse, 0)
	for _, model := range _list {
		list = append(list, UserLoginListResponse{
			UserLoginModel: model,
			UserNickname:   model.UserModel.NickName,
			UserAvatar:     model.UserModel.Avatar,
		})
	}

	response.OkWithList(list, count, c)

}
