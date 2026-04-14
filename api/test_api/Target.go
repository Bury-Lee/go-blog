package test_api

import (
	"StarDreamerCyberNook/common/response"

	"github.com/gin-gonic/gin"
)

type request struct {
	Content string `json:"content"`
}

// 做前端测试
func (TestApi) Print(c *gin.Context) { //响应收到的结构体请求
	// response.FailWithMsg("测试失败", c)
	var req request
	if c.ShouldBindJSON(&req) != nil {
		response.FailWithMsg("参数错误", c)
		return
	}
	response.OkWithData(req, c)
}
