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
	var cr LogListRequest
	if err := c.ShouldBindQuery(&cr); err != nil {
		response.FailWithError(err, c)
		return
	}

	cr.PageInfo.Key = "created_at desc"
	list, count, err := common.ListQuery[models.LogModel](models.LogModel{
		UserID:      cr.UserID,
		LogType:     cr.LogType,
		Level:       cr.Level,
		IP:          cr.IP,
		LoginStatus: cr.LoginStatus,
		ServiceName: cr.ServiceName,
	}, common.Options{
		PageInfo: cr.PageInfo,
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
	var cr models.IDRequest
	if err := c.ShouldBindUri(&cr); err != nil {
		response.FailWithError(err, c)
		return
	}
	var log models.LogModel
	if err := global.DB.Take(&log, cr.ID).Error; err != nil {
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
	var cr models.RemoveRequest
	if err := c.ShouldBindJSON(&cr); err != nil {
		response.FailWithError(err, c)
		return
	}
	log := log_service.GetLog(c)
	log.ShowRequest()
	log.ShowResponse()

	var logList []models.LogModel
	global.DB.Find(&logList, "id IN ?", cr.IDList)
	if len(logList) > 0 {
		global.DB.Delete(&logList)
	}

	response.OkWithMsg(fmt.Sprintf("日志删除成功,共删除%d条", len(logList)), c)
}
