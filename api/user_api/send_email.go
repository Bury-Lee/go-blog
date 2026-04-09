package user_api

import (
	"StarDreamerCyberNook/common/response"
	"StarDreamerCyberNook/global"
	"StarDreamerCyberNook/middleware"
	"StarDreamerCyberNook/models"
	"StarDreamerCyberNook/service/email_service"
	"StarDreamerCyberNook/utils"
	jwts "StarDreamerCyberNook/utils/jwts"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type SendEmailRequest struct {
	Type  string `json:"type" binding:"required"`  //注册,重置密码,重置邮箱
	Email string `json:"email" binding:"required"` //注册邮箱
}

type SendEmailResponse struct {
	EmailID      string `json:"emailID"`                // 新邮箱的验证码ID
	ResetEmailID string `json:"resetEmailID,omitempty"` // 原邮箱的验证码ID，仅在重置邮箱时返回
}

// SendEmailView 发送验证邮件
// 参数:c - gin.Context
// 返回:无
// 说明:处理注册、重置密码、重置邮箱的邮件发送，生成并缓存验证码到Redis
func (this UserApi) SendEmailView(c *gin.Context) {
	var req SendEmailRequest
	if err := c.ShouldBind(&req); err != nil {
		response.FailWithMsg("参数或邮箱格式错误", c)
		return
	}
	//启用验证码时检查验证码存储的业务是否正确
	if global.Config.Site.Login.Captcha {
		if target, exists := c.Get("target"); exists {
			if target != req.Type { //如果target不是req.Type，说明验证码业务错误
				response.FailWithMsg("验证码业务类型错误", c)
				return
			}
		} else { //如果没找到target，说明验证码业务错误
			response.FailWithMsg("未通过图形验证码校验", c)
			return
		}
	}

	// 前置校验：检查邮箱是否已注册/存在
	switch req.Type {
	case "注册":
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
	case "重置邮箱":
		var exists models.UserModel
		err := global.DB.Where("email = ?", req.Email).Take(&exists).Error
		if err == nil {
			response.FailWithMsg("该邮箱已使用", c)
			return
		}
	default:
		response.FailWithMsg("请选择重置密码,重置邮箱或者注册的其中一种", c)
		return
	}

	ctx := context.Background()
	ttl := 2 // 有效期，单位：分钟

	// 生成验证码（校验通过后再生成，避免无效存储）
	var newEmailID, oldEmailID string
	var sendErr error

	switch req.Type {
	case "注册":
		code := utils.GetRandomString(6, utils.Digits)
		id := utils.GetRandomString(20, utils.AlphaNum)

		// 发送邮件
		sendErr = email_service.SendRegister(req.Email, code)
		if sendErr != nil {
			logrus.Errorf("邮件发送失败: %v", sendErr)
			response.FailWithMsg("邮件发送失败", c)
			return
		}

		// 缓存验证码信息
		jsonData, err := json.Marshal(middleware.EmailVerifyInfo{
			RequstEmail: req.Email,
			EmailID:     id,
			EmailCode:   code,
			Type:        req.Type,
		})
		if err != nil {
			logrus.Error("缓存数据序列化错误:", err)
			response.FailWithMsg("服务器缓存错误", c)
			return
		}

		err = global.RedisTimeCache.Set(ctx, fmt.Sprintf("email:%s", id), string(jsonData), time.Minute*time.Duration(ttl)).Err()
		if err != nil {
			logrus.Error("存入Redis缓存失败:", err)
			response.FailWithMsg("验证码发送失败", c)
			return
		}

		newEmailID = id

	case "重置密码":
		code := utils.GetRandomString(6, utils.Digits)
		id := utils.GetRandomString(20, utils.AlphaNum)

		// 发送邮件
		sendErr = email_service.SendForgetPwd(req.Email, code)
		if sendErr != nil {
			logrus.Errorf("邮件发送失败: %v", sendErr)
			response.FailWithMsg("邮件发送失败", c)
			return
		}

		// 缓存验证码信息
		jsonData, err := json.Marshal(middleware.EmailVerifyInfo{
			RequstEmail: req.Email,
			EmailID:     id,
			EmailCode:   code,
			Type:        req.Type,
		})
		if err != nil {
			logrus.Error("缓存数据序列化错误:", err)
			response.FailWithMsg("服务器缓存错误", c)
			return
		}

		err = global.RedisTimeCache.Set(ctx, fmt.Sprintf("email:%s", id), string(jsonData), time.Minute*time.Duration(ttl)).Err()
		if err != nil {
			logrus.Error("存入Redis缓存失败:", err)
			response.FailWithMsg("验证码发送失败", c)
			return
		}

		newEmailID = id

	case "重置邮箱":
		// 获取当前登录用户信息
		claim, err := jwts.ParseTokenByGin(c)
		if claim == nil || err != nil {
			response.FailWithMsg("请先登录", c)
			return
		}
		user, err := claim.GetUser()
		if err != nil {
			response.FailWithMsg("获取用户信息失败出错", c)
			return
		}

		// 为原邮箱生成验证码
		oldEmailCode := utils.GetRandomString(6, utils.Digits)
		oldEmailID = utils.GetRandomString(20, utils.AlphaNum)

		// 为新邮箱生成验证码
		newEmailCode := utils.GetRandomString(6, utils.Digits)
		newEmailID = utils.GetRandomString(20, utils.AlphaNum)

		// 给原邮箱发送验证邮件
		sendErr = email_service.SendResetEmail(user.Email, oldEmailCode)
		if sendErr != nil {
			logrus.Errorf("原邮箱邮件发送失败: %v", sendErr)
			response.FailWithMsg("邮件发送失败", c)
			return
		}

		// 给新邮箱发送验证邮件
		sendErr = email_service.SendResetEmail(req.Email, newEmailCode)
		if sendErr != nil {
			logrus.Errorf("新邮箱邮件发送失败: %v", sendErr)
			response.FailWithMsg("邮件发送失败", c)
			return
		}

		// 缓存原邮箱的验证码信息
		oldEmailJsonData, err := json.Marshal(middleware.EmailVerifyInfo{
			RequstEmail: user.Email,
			EmailID:     oldEmailID,
			EmailCode:   oldEmailCode,
			Type:        req.Type,
		})
		if err != nil {
			logrus.Error("原邮箱缓存数据序列化错误:", err)
			response.FailWithMsg("服务器缓存错误", c)
			return
		}

		err = global.RedisTimeCache.Set(ctx, fmt.Sprintf("ResetEmail:%s", oldEmailID), string(oldEmailJsonData), time.Minute*time.Duration(ttl)).Err()
		if err != nil {
			logrus.Error("存入原邮箱Redis缓存失败:", err)
			response.FailWithMsg("验证码发送失败", c)
			return
		}

		// 缓存新邮箱的验证码信息
		newEmailJsonData, err := json.Marshal(middleware.EmailVerifyInfo{
			RequstEmail: req.Email,
			EmailID:     newEmailID,
			EmailCode:   newEmailCode,
			Type:        req.Type,
		})
		if err != nil {
			logrus.Error("新邮箱缓存数据序列化错误:", err)
			response.FailWithMsg("服务器缓存错误", c)
			return
		}

		err = global.RedisTimeCache.Set(ctx, fmt.Sprintf("email:%s", newEmailID), string(newEmailJsonData), time.Minute*time.Duration(ttl)).Err()
		if err != nil {
			logrus.Error("存入新邮箱Redis缓存失败:", err)
			response.FailWithMsg("验证码发送失败", c)
			return
		}
	}

	// 返回响应
	if req.Type == "重置邮箱" {
		response.Ok(SendEmailResponse{
			EmailID:      newEmailID,
			ResetEmailID: oldEmailID,
		}, fmt.Sprintf("邮箱验证码有效期为: %d 分钟", ttl), c)
	} else {
		response.Ok(SendEmailResponse{
			EmailID: newEmailID,
		}, fmt.Sprintf("邮箱验证码有效期为: %d 分钟", ttl), c)
	}
}
