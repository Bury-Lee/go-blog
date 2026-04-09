package log_service

import (
	"StarDreamerCyberNook/global"
	"StarDreamerCyberNook/models"
	"StarDreamerCyberNook/models/enum"
	"StarDreamerCyberNook/utils/ip"
	jwts "StarDreamerCyberNook/utils/jwts"
	"fmt"

	"github.com/gin-gonic/gin"
)

func NewLoginSuccess(c *gin.Context, LoginType enum.LoginType) {
	ipAdd := c.ClientIP()
	addr := ip.GetIpAddr(ipAdd)
	// token := c.GetHeader("token")
	UserID := 1 //TODO:通过JWT从token中解析出用户ID
	global.DB.Create(&models.LogModel{
		LogType:     enum.LoginLogType,
		Title:       "登录成功",
		Content:     fmt.Sprintf("用户 %d 登录成功", UserID),
		UserID:      uint(UserID),
		IP:          ipAdd,
		Addr:        addr,
		LoginStatus: true,
		Username:    "", //登录成功后就没必要记录错误的用户账号了
		Pwd:         "", //登录成功后就没必要记录错误的密码了
		LoginType:   LoginType,
	})
}

func NewLoginFail(c *gin.Context, LoginType enum.LoginType, msg string, username string, pwd string) {
	ipAdd := c.ClientIP()
	addr := ip.GetIpAddr(ipAdd)
	// token := c.GetHeader("token")
	var userID uint = 0
	var userName string = ""
	claims, err := jwts.ParseTokenByGin(c)
	if err == nil && claims != nil {
		userID = claims.UserID
	}
	global.DB.Create(&models.LogModel{
		LogType:     enum.LoginLogType,
		UserID:      uint(userID),
		Title:       "登录失败",
		Content:     fmt.Sprintf("用户 %s 登录失败: %s", username, msg),
		IP:          ipAdd,
		Addr:        addr,
		LoginStatus: false,
		Username:    userName,
		Pwd:         pwd, //登录失败时记录用户输入的密码
		LoginType:   LoginType,
	})
}
