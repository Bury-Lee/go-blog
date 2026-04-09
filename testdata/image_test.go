package main

import (
	"github.com/gin-gonic/gin"
	"github.com/mojocn/base64Captcha"
	"github.com/sirupsen/logrus"
)

type CaptchaApi struct {
}

type CaptchaResponse struct {
	CaptchaID string `json:"captchaID"`
	Captcha   string `json:"captcha"`
}

func (CaptchaApi) CaptchaView(c *gin.Context) {
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
	captcha := base64Captcha.NewCaptcha(driver, global.CaptchaStore)
	lid, lb64s, _, err := captcha.Generate()
	if err != nil {
		logrus.Error(err)
		res.FailWithMsg("图片验证码生成失败", c)
		return
	}
	res.OkWithData(CaptchaResponse{
		CaptchaID: lid,
		Captcha:   lb64s,
	}, c)
}
