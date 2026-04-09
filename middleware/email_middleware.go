package middleware

import (
	"StarDreamerCyberNook/common/response"
	"StarDreamerCyberNook/global"
	"StarDreamerCyberNook/utils/jwts"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type EmailVerifyInfoRequest struct {
	EmailID   string `json:"emailID" binding:"required"`
	EmailCode string `json:"emailCode" binding:"required"`
}

type EmailVerifyInfo struct {
	RequstEmail string `json:"requestEmail"` //请求的邮箱,预备字段,不从请求体获取,而是从邮箱存储里获取,验证成功后存入上下文
	EmailID     string `json:"emailID" binding:"required"`
	EmailCode   string `json:"emailCode" binding:"required"`
	Type        string `json:"type" binding:"required"` //注册,重置密码,重置邮箱
}

// EmailVerifyMiddleware 邮箱验证中间件
// 参数:c - gin上下文对象
// 说明:验证邮箱验证码,获取请求体并解析,调用邮箱存储验证,验证通过后将邮箱信息存入上下文
func EmailVerifyMiddleware(c *gin.Context) {
	body, err := c.GetRawData()
	if err != nil {
		response.FailWithMsg("获取请求体错误", c)
		c.Abort()
		return
	}
	c.Request.Body = io.NopCloser(bytes.NewReader(body))
	var req EmailVerifyInfoRequest
	err = c.ShouldBindJSON(&req)
	if err != nil {
		logrus.Errorf("邮箱验证失败 %s", err)
		response.FailWithMsg("邮箱验证失败", c)
		c.Abort()
		return
	}

	//在redis里获取
	ctx := context.Background()
	data, err := global.RedisTimeCache.Get(ctx, fmt.Sprintf("email:%s", req.EmailID)).Result()
	if err != nil {
		response.FailWithMsg("数据异常", c)
		c.Abort()
		return
	}

	var info EmailVerifyInfo
	if err := json.Unmarshal([]byte(data), &info); err != nil {
		response.FailWithMsg("数据异常", c)
		c.Abort()
		return
	}

	if info.EmailCode != req.EmailCode {
		// global.RedisTimeCache.Del(ctx, fmt.Sprintf("email:%s", cr.EmailID))//严格的错误限制:验证码错误就删除,防止暴力破解,但可能会有误伤,可以根据实际情况调整
		response.FailWithMsg("验证码错误", c)
		c.Abort()
		return
	}

	// 验证成功后删除验证码，防止重复使用
	global.RedisTimeCache.Del(ctx, fmt.Sprintf("email:%s", req.EmailID))

	c.Set("email", info)

	c.Request.Body = io.NopCloser(bytes.NewReader(body))
}

type ResetEmailVerifyInfoRequest struct { //用于验证原邮箱是否通过的请求体结构体
	EmailID   string `json:"ResetEmailID" binding:"required"`
	EmailCode string `json:"ResetEmailCode" binding:"required"`
}

// ResetEmailVerifyMiddleware 邮箱验证中间件
// 参数:c - gin上下文对象
// 说明:验证邮箱验证码,获取请求体并解析,调用邮箱存储验证,验证通过后将邮箱信息存入上下文
func ResetEmailVerifyMiddleware(c *gin.Context) { //专门给重置邮箱用的中间件,要验证两次,一次验证原邮箱,一次验证新邮箱
	//这里仅检验原邮箱是否通过,不检验新邮箱是否通过
	//只关心请求者是否有原邮箱的验证码,不关心新邮箱的验证码是否正确
	body, err := c.GetRawData()
	if err != nil {
		response.FailWithMsg("获取请求体错误", c)
		c.Abort()
		return
	}
	c.Request.Body = io.NopCloser(bytes.NewReader(body))
	var req ResetEmailVerifyInfoRequest
	err = c.ShouldBindJSON(&req)
	if err != nil {
		logrus.Errorf("邮箱验证失败 %s", err)
		response.FailWithMsg("邮箱验证失败", c)
		c.Abort()
		return
	}

	//在redis里获取
	ctx := context.Background()
	data, err := global.RedisTimeCache.Get(ctx, fmt.Sprintf("ResetEmail:%s", req.EmailID)).Result()
	if err != nil {
		response.FailWithMsg("数据异常", c)
		c.Abort()
		return
	}

	var info EmailVerifyInfo
	if err := json.Unmarshal([]byte(data), &info); err != nil {
		response.FailWithMsg("数据异常", c)
		c.Abort()
		return
	}
	//TODO:判断RequestEmail是否和用户的邮箱一致,不一致就直接返回错误,防止验证码被盗用
	user, err := jwts.GetClaims(c).GetUser()
	if info.RequstEmail != user.Email {
		response.FailWithMsg("邮箱验证出错", c)
		c.Abort()
		return
	}

	if info.EmailCode != req.EmailCode || info.Type != "重置邮箱" {
		// global.RedisTimeCache.Del(ctx, fmt.Sprintf("email:%s", cr.EmailID))//严格的错误限制:验证码错误就删除,防止暴力破解,但可能会有误伤,可以根据实际情况调整
		response.FailWithMsg("验证出错", c)
		c.Abort()
		return
	}

	// 验证成功后删除验证码，防止重复使用
	global.RedisTimeCache.Del(ctx, fmt.Sprintf("ResetEmail:%s", req.EmailID))

	c.Request.Body = io.NopCloser(bytes.NewReader(body))
}
