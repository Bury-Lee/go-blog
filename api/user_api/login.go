package user_api

import (
	"StarDreamerCyberNook/common/response"
	"StarDreamerCyberNook/global"
	"StarDreamerCyberNook/models"
	Hash "StarDreamerCyberNook/utils/hash"
	"StarDreamerCyberNook/utils/ip"
	jwts "StarDreamerCyberNook/utils/jwts"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type LoginRequest struct {
	Type string `json:"type" binding:"required"` //登录类型,用户名或邮箱登录//TODO:qq登录
	Val  string `json:"val" binding:"required"`
	Pwd  string `json:"pwd" binding:"required"`
}

func (UserApi) Login(c *gin.Context) { //用户名-密码登录
	var req LoginRequest
	if err := c.ShouldBind(&req); err != nil {
		response.FailWithMsg("参数校验失败", c)
		return
	}
	//启用验证码时检查验证码业务设置是否正确
	if global.Config.Site.Login.Captcha {
		if target, exists := c.Get("target"); exists {
			if target != req.Type { //如果target不是req.Type，说明验证码业务错误
				response.FailWithMsg("验证码错误", c)
				return
			}
		} else { //如果没找到target，说明验证码业务错误
			response.FailWithMsg("验证码错误", c)
			return
		}
	}

	var usermodel models.UserModel
	switch req.Type {
	case "邮箱":
		if !global.Config.Site.Login.EmailLogin {
			response.FailWithMsg("邮箱登录未开放", c)
			return
		}
		err := global.DB.Take(&usermodel, "email = ?", req.Val).Error
		if err != nil {
			response.FailWithMsg("邮箱密码错误", c)
			return
		}
	case "用户名":
		if !global.Config.Site.Login.UsernamePassword {
			response.FailWithMsg("账密登录未开放", c)
			return
		}
		err := global.DB.Take(&usermodel, "user_name = ?", req.Val).Error
		if err != nil {
			response.FailWithMsg("用户密码错误", c)
			return
		}
	default:
		response.FailWithMsg("请选择正确的登录方式", c)
		return
	}
	if !Hash.CheckPassword(req.Pwd, usermodel.Password) {
		response.FailWithMsg("用户名密码错误", c)
		return
	}

	// TODO:
	// 	创建登录日志和修改最近登录时间
	var loginLog = models.UserLoginModel{
		UserID:    usermodel.ID,
		IP:        c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
	}
	global.DB.Create(&loginLog)

	AccessToken, RefreshToken, err := jwts.GetToken(jwts.Claims{
		UserID:   usermodel.ID,
		Username: usermodel.UserName,
		Role:     usermodel.Role,
	})
	if err != nil {
		response.FailWithMsg("登录失败", c)
		return
	}
	global.DB.Model(&usermodel).Update("last_login_time", time.Now()).Update("ip", ip.GetIpAddr(c.ClientIP()))
	// 13. 返回成功响应
	// 将JWT令牌返回给前端，后续请求携带此Token进行身份验证
	logrus.Info(fmt.Sprintf("用户 %s 登录", req.Val))
	response.OkWithData(map[string]string{
		"AccessToken":  AccessToken,
		"RefreshToken": RefreshToken,
	}, c)
}
