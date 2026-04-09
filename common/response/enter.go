package response

import (
	"StarDreamerCyberNook/global"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

var EmptyData map[string]any = map[string]any{}

type Code int //错误代码
const (
	SuccessCode          Code = 200 // 操作成功。
	FailCode             Code = 201 // 一般操作失败。
	FailServiceCode      Code = 203 // 服务端错误，例如，服务端遇到意外情况。
	FailValid            Code = 422 // 验证失败。
	BadRequest           Code = 400 // 客户端错误，例如，请求语法错误，无效的请求消息帧，或欺骗性的请求路由。
	ServerError          Code = 500 // 服务器端错误，例如，服务器遇到意外情况。
	DatabaseError        Code = 500 // 数据库操作失败。
	Unauthorized         Code = 401 // 需要认证或认证失败。
	Forbidden            Code = 403 // 客户端无权访问内容。
	UnknownError         Code = 500 // 发生未知错误。
	NotFound             Code = 404 // 请求的资源或数据未找到。
	InvalidData          Code = 400 // 提供的数据无效或格式错误。
	NotLoggedIn          Code = 401 // 用户未登录。
	AuthenticationFailed Code = 401 // 登录尝试因凭据无效而失败。
	ServiceUnavailable   Code = 503 // 服务器目前无法处理请求，因为临时过载或计划维护。
)

func (c Code) String() string {
	switch c {
	case SuccessCode:
		return "SuccessCode"
	case FailCode:
		return "FailCode"
	case FailValid:
		return "FailValid"
	case FailServiceCode:
		return "FailServiceCode"
	case NotFound:
		return "NotFound"
	case Forbidden:
		return "Forbidden"
	case ServiceUnavailable:
		return "ServiceUnavailable"

	case 400: // BadRequest, InvalidData
		return "BadRequest | InvalidData"
	case 401: // Unauthorized, NotLoggedIn, AuthenticationFailed
		return "Unauthorized | NotLoggedIn | AuthenticationFailed"
	case 500: // ServerError, DatabaseError, UnknownError
		return "ServerError | DatabaseError | UnknownError"

	default:
		return "UnknownCode"
	}
}

type Response struct {
	Code Code   `json:"code"`    //错误代码
	Data any    `json:"data"`    //数据
	Msg  string `json:"message"` //错误信息
}

func (this Response) Json(c *gin.Context) {
	c.JSON(int(this.Code), this)
}

func Ok(data any, msg string, c *gin.Context) { //返回数据与消息
	if global.Config.System.RunMode == "debug" {
		logrus.Info(SuccessCode, data, msg)
	}
	Response{SuccessCode, data, msg}.Json(c)
}

func OkWithData(data any, c *gin.Context) { //返回数据
	Response{Code: SuccessCode, Data: data, Msg: "成功"}.Json(c)
}
func OkWithMsg(msg string, c *gin.Context) { //返回消息
	Response{Code: SuccessCode, Data: EmptyData, Msg: msg}.Json(c)
}

func OkWithList(list any, count int, c *gin.Context) {
	Response{SuccessCode, map[string]any{"list": list, "cout": count}, "成功"}.Json(c)
}

func Fail(code Code, msg string, data any, c *gin.Context) {
	if global.Config.System.RunMode == "debug" {
		logrus.Info(SuccessCode, data, msg)
	}
	Response{Code: code, Data: data, Msg: msg}.Json(c)
}

func FailWithMsg(msg string, c *gin.Context) {
	Response{Code: FailValid, Data: EmptyData, Msg: msg}.Json(c)
}
func FailWithCode(code Code, c *gin.Context) {
	Response{Code: FailServiceCode, Data: EmptyData, Msg: code.String()}.Json(c)
}
func FailWithData(data any, msg string, c *gin.Context) {
	Response{Code: FailCode, Data: data, Msg: msg}.Json(c)
}
func FailWithError(err error, c *gin.Context) { //返回错误信息
	// data, msg := validate.ValidateError(err)
	Response{Code: FailCode, Data: EmptyData, Msg: err.Error()}.Json(c)
}
