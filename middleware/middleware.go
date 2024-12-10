package middleware

import (
	"bytes"
	"fmt"
	"io"
	"time"

	"sendmsggo/logger"
	"sendmsggo/mssql"

	"github.com/gin-gonic/gin"
)
 
func HttpLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 初始化Http日志
		logger := logger.InitHTTPLogger() // 独立日志与DB日志区分

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

		// 记录请求开始时间
		start := time.Now()

		// 处理请求
		c.Next()

		// 记录响应时间
		latency := time.Since(start)

		// 记录响应头
		responseHeaders := ""
		for key, values := range c.Writer.Header() {
			for _, value := range values {
				responseHeaders += fmt.Sprintf("%s: %s; ", key, value)
			}
		}

		var Error string
		// 记录错误信息
		if len(c.Errors) > 0 {
			Error = c.Errors.ByType(gin.ErrorTypePrivate).String()
		} else {
			Error = "None"
		}

		//请求URL如果没有查询参数，则不显示"/"符号
		var sign string
		if c.Request.URL.RawQuery == "" {
			sign = ""
		} else {
			sign = "/"
		}

		// 自定义日志格式
		str := fmt.Sprintf(
			"[Gin] | %s | %d | %4.2vms | %+v | %s | Errors: %s",
			c.ClientIP(),
			c.Writer.Status(),
			latency,
			c.Request.Method,
			c.Request.URL.Path+sign+c.Request.URL.RawQuery,
			Error,
		)
		logger.Info(str)
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

// #endregion 日志中间件

// #region 数据库中间件
// 中间件函数，将数据库连接注入到上下文中
type DB = mssql.DBWrapper

func DBMiddleware(db *DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("db", db)
		c.Next()
	}
}
 
