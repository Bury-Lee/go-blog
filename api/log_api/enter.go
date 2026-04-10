package log_api

import (
	"StarDreamerCyberNook/common"
	"StarDreamerCyberNook/common/response"
	"StarDreamerCyberNook/global"
	"StarDreamerCyberNook/models"
	"StarDreamerCyberNook/models/enum"
	"StarDreamerCyberNook/service/log_service"
	"fmt"

	"github.com/gin-gonic/gin"
)

type LogApi struct { //在这里注册路由
}

type LogListRequest struct {
	common.PageInfo
	LogType     enum.LogType  `form:"logType"`
	Level       enum.LogLevel `form:"level"`
	IP          string        `form:"ip"`
	LoginStatus bool          `form:"loginStatus"`
	ServiceName string        `form:"serviceName"`
	UserID      uint          `form:"userID"`
}

type LogListResponse struct { //还需要什么就自己加
	models.LogModel
	UserNickName string `json:"userNickName"`
	UserAvatar   string `json:"userAvatar"`
}

func (LogApi) LogListView(c *gin.Context) {
	//支持分页查询和模糊匹配&精确查询
	var req LogListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithError(err, c)
		return
	}

	req.PageInfo.Key = "created_at desc"
	list, count, err := common.ListQuery[models.LogModel](models.LogModel{
		UserID:      req.UserID,
		LogType:     req.LogType,
		Level:       req.Level,
		IP:          req.IP,
		LoginStatus: req.LoginStatus,
		ServiceName: req.ServiceName,
	}, common.Options{
		PageInfo: req.PageInfo,
		Likes:    []string{"Title"},
		Preloads: []string{"UserModel"},
	})
	if err != nil {
		response.FailWithError(err, c)
		return
	}

	response.OkWithList(list, int(count), c)
}

func (LogApi) LogReadView(c *gin.Context) {
	var req models.IDRequest
	if err := c.ShouldBindUri(&req); err != nil {
		response.FailWithError(err, c)
		return
	}
	var log models.LogModel
	if err := global.DB.Take(&log, req.ID).Error; err != nil {
		response.FailWithMsg("日志不存在", c)
		return
	}
	if !log.IsRead {
		global.DB.Model(&log).Update("is_read", true)
		response.OkWithMsg("日志已读取", c) //TODO:也许要返回日志详情?
		return
	}

	response.OkWithMsg("日志已读取", c)
}

func (LogApi) LogRemoveView(c *gin.Context) {
	var req models.RemoveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithError(err, c)
		return
	}
	log := log_service.GetLog(c)
	log.ShowRequest()
	log.ShowResponse()

	var logList []models.LogModel
	global.DB.Find(&logList, "id IN ?", req.IDList)
	if len(logList) > 0 {
		global.DB.Delete(&logList)
	}

	response.OkWithMsg(fmt.Sprintf("日志删除成功,共删除%d条", len(logList)), c)
}
