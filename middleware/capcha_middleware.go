package middleware

import (
	"StarDreamerCyberNook/common/response"
	"StarDreamerCyberNook/global"
	"bytes"
	"io"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// CaptchaMiddlewareRequest 验证码中间件请求参数
// 参数:CaptchaID - 验证码ID, CaptchaCode - 验证码代码, Target - 业务标识符
// 说明:用于接收和验证用户输入的验证码信息
type CaptchaMiddlewareRequest struct {
	CaptchaID   string `json:"captchaID" binding:"required"`
	CaptchaCode string `json:"captchaCode" binding:"required"`
}

// CaptchaMiddleware 验证码中间件
// 参数:c - gin上下文对象
// 说明:验证图形验证码,检查是否启用验证码功能,获取并解析请求体,调用验证码存储验证,验证失败则中止请求
// 如果需要更严格的安全限制,还可以改一下,记录发送IP和场景,只有发送ip和场景都一致的时候才生效
func CaptchaMiddleware(c *gin.Context) {
	if !global.Config.Site.Login.Captcha {
		return
	}
	body, err := c.GetRawData() //获取请求体
	if err != nil {
		response.FailWithMsg("获取请求体失败", c)
		c.Abort()
		return
	}

	// 第一次构造: ShouldBindJSON会读取并消费body数据,需要重新构造供后续使用
	c.Request.Body = io.NopCloser(bytes.NewReader(body))
	var req CaptchaMiddlewareRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMsg("验证码参数校验失败", c)
		c.Abort()
		return
	}

	result, err := global.RedisTimeCache.Get(c, req.CaptchaID).Result()
	if err != nil {
		response.FailWithMsg("验证码错误", c)
		c.Abort()
		return
	}
	parts := strings.Split(result, "/")
	if len(parts) != 2 {
		logrus.Errorf("验证码格式错误:%s", result)
		response.FailWithMsg("验证码错误", c)
		c.Abort()
		return
	}
	// 验证码缓存格式为 target/answer
	answer := parts[1]
	if answer != req.CaptchaCode {
		response.FailWithMsg("验证码错误", c)
		c.Abort()
		return
	}
	// 第二次构造: 验证通过后重新构造body,确保后续处理器能正常读取请求体
	c.Request.Body = io.NopCloser(bytes.NewReader(body))
	//把业务标识添加到上下文,给后面的函数使用

	c.Set("target", parts[0])
}
