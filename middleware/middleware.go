package middleware

import (
	"bytes"
	"fmt"
	"io"

	"sendMsgGo/logger"

	"github.com/gin-gonic/gin"
)

// HttpLogger 定义自定义日志格式的中间件
func HttpLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		//初始化Http日志
		logger := logger.InitHTTPLogger() //独立日志与DB日志区分

		// 记录请求体
		var requestBody []byte
		if c.Request.Body != nil {
			requestBody, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		// 记录请求头
		requestHeaders := ""
		for key, values := range c.Request.Header {
			for _, value := range values {
				requestHeaders += fmt.Sprintf("%s: %s; ", key, value)
			}
		}

		// 记录响应体
		writer := &responseBodyWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = writer
		// logrus.Infof("[DB] @%s - executing, sql: Query { sql: \"%s\", Params: [%v] }")

		str := fmt.Sprintf("[Client] IP: [%s] -[%+v]", c.ClientIP(), c.Request.Method)
		logger.Info(str)

		// 自定义日志格式
		// _logFormat := fmt.Sprintf("%s - [%s] \"%s %s %s\" %d %s \"%s\" \"%s\" \"%s\" \"%s\" \"%s\" \"%s\"\n",
		// 	c.ClientIP(),
		// 	start.Format(time.RFC3339Nano),
		// 	c.Request.Method,
		// 	c.Request.URL.Path,
		// 	c.Request.Proto,
		// 	c.Writer.Status(),
		// 	latency,
		// 	c.Request.UserAgent(),
		// 	string(requestBody),
		// 	requestHeaders,
		// 	writer.body.String(),
		// 	responseHeaders,
		// 	c.Errors.ByType(gin.ErrorTypePrivate).String(),
		// )

		// 处理请求
		c.Next()

		// 记录响应时间
		// latency := time.Since(start)

		// 记录响应头
		responseHeaders := ""
		for key, values := range c.Writer.Header() {
			for _, value := range values {
				responseHeaders += fmt.Sprintf("%s: %s; ", key, value)
			}
		}

	}
}

// responseBodyWriter 是一个用于记录响应体的自定义 ResponseWriter
type responseBodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (r *responseBodyWriter) Write(b []byte) (int, error) {
	r.body.Write(b)
	return r.ResponseWriter.Write(b)
}
