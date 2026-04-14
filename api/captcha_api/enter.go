package captcha_api

import (
	"StarDreamerCyberNook/common/response"
	"StarDreamerCyberNook/global"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mojocn/base64Captcha"
	"github.com/sirupsen/logrus"
)

type CaptchaApi struct {
}

type CaptchaRequest struct {
	Target string `json:"target" form:"target" query:"target"` //目标业务
}
type CaptchaResponse struct {
	CaptchaID string `json:"captchaID"` //验证码会话标识ID
	Captcha   string `json:"captcha"`   //验证码
}

// TODO:把参数写在配置里,实现更个性化的生成
func (CaptchaApi) CaptchaView(c *gin.Context) {
	var req CaptchaRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithMsg("参数错误", c)
		return
	}
	target := req.Target //历史遗留代码问题

	fmt.Println(target) //debug
	//注册,重置密码,重置邮箱,用户名密码,邮箱
	mapTarget := map[string]bool{
		"注册":    true,
		"重置密码":  true,
		"重置邮箱":  true,
		"用户名密码": true,
		"邮箱":    true,
	}
	if _, ok := mapTarget[target]; !ok {
		response.FailWithMsg("目标业务不存在", c)
		return
	}

	//TODO:存储ip和业务代码实现验证码的隔离,不能让验证码变成业务通用验证码.否则会出现一台机器申请识别并发送验证码,另一台进行非法操作的问题,还有要加上时间限制,超过一定时间自动作废
	//修正:实现验证码专码专用,似乎应该是要在业务内再做一次验证码识别和使用,从Redis中取出kv对+业务标识符(),然后在业务内检查
	var driver base64Captcha.Driver
	var driverString base64Captcha.DriverString

	captchaConfig := base64Captcha.DriverString{
		Height:          60,
		Width:           200,
		NoiseCount:      1,
		ShowLineOptions: 2 | 4,
		Length:          4,
		Source:          "1234567890",
	}
	driverString = captchaConfig
	driver = driverString.ConvertFonts()
	store := &NoStore{}
	captcha := base64Captcha.NewCaptcha(driver, store) //创建验证码,由于已经使用了redis,所以不需要存储验证码到内存,定义空实现Store
	//取消存储
	lid, lb64s, answer, err := captcha.Generate() //第三个是键值对的answer,到时候就是在Redis里存ip,业务,id,answer
	if err != nil {
		logrus.Error(err)
		response.FailWithMsg("图片验证码生成失败", c)
		return
	}
	global.RedisTimeCache.Set(c, lid, target+"/"+answer, 3*time.Minute) //存储业务标识和验证码,过期时间为3分钟
	response.OkWithData(CaptchaResponse{
		CaptchaID: lid,
		Captcha:   lb64s,
	}, c)
}

type NoStore struct{}

// Set方法需要返回error
func (s *NoStore) Set(id string, value string) error {
	// 不做任何操作，返回nil表示成功
	return nil
}

// Get方法也需要实现
func (s *NoStore) Get(id string, clear bool) string {
	// 不做任何操作，返回空字符串
	return ""
}
func (s *NoStore) Verify(id, answer string, clear bool) bool {
	// 不做任何操作，返回false
	return false
}
