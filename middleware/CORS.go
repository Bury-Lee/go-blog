// middleware/CORS.go
package middleware

import "github.com/gin-gonic/gin"

func CORS(c *gin.Context) {
	// 1. 获取请求头的 Origin
	origin := c.Request.Header.Get("Origin")
	if origin != "" {
		// 动态设置 Origin，解决与 AllowCredentials: true 的冲突
		c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
	} else {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	}

	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

	// 2. 专门处理 OPTIONS 预检请求
	if c.Request.Method == "OPTIONS" {
		c.AbortWithStatus(204) // 直接返回 204，不进入后续逻辑
		return
	}
	c.Next()
}
