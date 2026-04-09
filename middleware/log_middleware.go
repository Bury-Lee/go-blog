// middleware/log_middleware.go
package middleware

import (
	"StarDreamerCyberNook/service/log_service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ResponseWriter struct {
	gin.ResponseWriter
	Body []byte
	Head http.Header
}

func (w *ResponseWriter) Write(data []byte) (int, error) { //重写Write方法以捕获响应体
	w.Body = append(w.Body, data...)
	return w.ResponseWriter.Write(data)
}

func (w *ResponseWriter) Header() http.Header {
	// 先返回原始 Writer 的 Header
	// 这样 c.Writer.Header().Set(...) 才能真正修改即将发送的响应头
	return w.Head
}

func LogMiddleware(c *gin.Context) {
	log := log_service.NewActionLog(c) // 创建日志实例
	log.SetRequest(c)

	c.Set("log", log)

	// 3. 替换 c.Writer 为我们的自定义响应写入器
	res := &ResponseWriter{
		ResponseWriter: c.Writer,
		Head:           make(http.Header),
	}
	c.Writer = res
	// 4. 继续执行下一个中间件或路由处理函数
	c.Next()
	// 5. 处理响应后：打印响应体
	log.SetResponse(res.Body)
	log.SetResponseHeader(res.Head)
	log.MiddlewareSave()

}
