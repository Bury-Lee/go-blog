package test_api

import (
	"StarDreamerCyberNook/common/response"
	"StarDreamerCyberNook/global"
	"StarDreamerCyberNook/models"
	"StarDreamerCyberNook/service/email_service"
	"StarDreamerCyberNook/utils"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type TestApi struct {
}

func (TestApi) TestView(c *gin.Context) {
	// response.FailWithMsg("测试失败", c)
	c.JSON(200, gin.H{
		"code": 200,
		"msg":  "测试成功",
	})
}

type SendEmailRequest struct {
	Type  string `json:"type" binding:"required"`  //1注册2重置密码 TODO:改为枚举
	Email string `json:"email" binding:"required"` //注册邮箱
}
type SendEmailResponse struct {
	EmailID string `json:"emailID"`
}

func (TestApi) SendEmailView(c *gin.Context) {
	var req SendEmailRequest
	if err := c.ShouldBind(&req); err != nil {
		logrus.Debug(req)
		response.FailWithMsg("参数错误", c)
		return
	}
	// 前置校验：检查邮箱是否已注册/存在
	switch req.Type {
	case "注册": //TODO:注册成功之后要入库
		var exists models.UserModel
		err := global.DB.Where("email = ?", req.Email).Take(&exists).Error
		if err == nil {
			response.FailWithMsg("该邮箱已注册", c)
			return
		}
	case "重置密码":
		var user models.UserModel
		if err := global.DB.Take(&user, "email = ?", req.Email).Error; err != nil {
			response.FailWithMsg("该邮箱未注册", c)
			return
		}
	default:
		response.FailWithMsg("请选择重置密码或者注册的其中一种", c)
		return
	}

	// 生成验证码（校验通过后再生成，避免无效存储）
	// code := base64Captcha.RandText(6, "1234567890") //TODO:到时候换成自己的
	// id := base64Captcha.RandomId()

	code := utils.GetRandomString(6, "1234567890")
	id := utils.GetRandomString(20, utils.AlphaNum)
	// 发送邮件
	var sendErr error
	switch req.Type {
	case "注册":
		sendErr = email_service.SendRegister(req.Email, code)
	case "重置密码":
		sendErr = email_service.SendForgetPwd(req.Email, code)
	}

	if sendErr != nil {
		logrus.Errorf("邮件发送失败: %v", sendErr)
		response.FailWithMsg("邮件发送失败", c)
		return
	}

	// 存储验证码（发送成功后再存储）
	global.CaptchaStore.Set(id, code) // TODO: 迁移至Redis

	response.OkWithData(SendEmailResponse{EmailID: id}, c)
}
