package middleware

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// RequestLogMiddleware 用于打印完整的请求信息（Header, Cookie, Body）
func RequestLogMiddleware(c *gin.Context) {
	var logBuilder strings.Builder

	// 构建请求日志内容
	logBuilder.WriteString("\n== 收到新请求 ==\n")
	logBuilder.WriteString(fmt.Sprintf("Method: %s\n", c.Request.Method))
	logBuilder.WriteString(fmt.Sprintf("URL: %s\n", c.Request.URL.String()))
	logBuilder.WriteString(fmt.Sprintf("Remote Addr: %s\n", c.ClientIP()))

	// 添加请求头信息
	logBuilder.WriteString("--- 请求头 ---\n")
	for key, values := range c.Request.Header {
		logBuilder.WriteString(fmt.Sprintf("%s: %s\n", key, strings.Join(values, ", ")))
	}

	// 读取并记录请求体 (Body)
	var bodyBytes []byte
	var err error

	if c.Request.Body != nil {
		bodyBytes, err = io.ReadAll(c.Request.Body)
		if err != nil {
			logrus.Errorf("读取请求体失败: %v", err)
			c.Next()
			return
		}

		// 添加请求体内容
		logBuilder.WriteString("--- 请求体 ---\n")
		if len(bodyBytes) > 0 {
			bodyStr := string(bodyBytes)
			if len(bodyStr) > 2000 {
				logBuilder.WriteString(fmt.Sprintf("%s\n... (内容过长，已截断) 长度: %d\n", bodyStr[:2000], len(bodyStr)))
			} else {
				logBuilder.WriteString(fmt.Sprintf("%s\n", bodyStr))
			}
		} else {
			logBuilder.WriteString("(空)\n")
		}

		// 【关键步骤】将读取过的数据重新写回 Body，否则后续的 gin handler 无法读取
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	} else {
		logBuilder.WriteString("--- 请求体 ---\n")
		logBuilder.WriteString("(无请求体)\n")
	}

	logBuilder.WriteString("============================")

	// 使用 Logrus 一次性输出完整日志
	logrus.Debug(logBuilder.String())

	// 继续处理后续的逻辑
	c.Next()
}
