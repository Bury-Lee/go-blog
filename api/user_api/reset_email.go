package user_api

//TODO:如果没记错的话邮箱验证码只能验证一次,输错一次就直接作废,到时候改一下,改为10次
import (
	"StarDreamerCyberNook/common/response"
	"StarDreamerCyberNook/global"
	"StarDreamerCyberNook/middleware"
	jwts "StarDreamerCyberNook/utils/jwts"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// RegisterEmailView 邮箱注册接口
// 处理用户通过邮箱验证码完成的注册流程
// 流程：验证参数 -> 校验邮箱验证码 -> 生成用户名 -> 创建用户 -> 返回Token
func (UserApi) ResetEmailView(c *gin.Context) {
	_email, _ := c.Get("email")
	email, ok := _email.(middleware.EmailVerifyInfo)
	if !ok {
		logrus.Error("邮箱验证信息类型断言失败")
		response.FailWithMsg("意外错误", c)
		return
	}
	if email.Type != "重置邮箱" {
		response.FailWithMsg("邮箱验证类型错误", c)
		return
	}
	user, err := jwts.GetClaims(c).GetUser()
	if err != nil {
		response.FailWithMsg("不存在的用户", c)
		return
	}

	//更新数据
	if err := global.DB.Model(&user).Update("email", email.RequstEmail).Error; err != nil {
		response.FailWithMsg("邮箱重置失败", c)
		return
	}
	response.OkWithMsg("邮箱重置成功", c)
}
