package middleware

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"runtime/debug"
	"sendmsggo/controller"
	"sendmsggo/util/logger"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type JsonError = controller.JsonError

func HttpLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		var Error string

		// 初始化Http日志
		logger := logger.InitHTTPLogger() // 独立日志与DB日志区分

		// 记录请求头
		requestHeaders := ""
		for key, values := range c.Request.Header { // 确保这里使用 := range 来遍历 map
			for _, value := range values {
				requestHeaders += fmt.Sprintf("%s: %s; ", key, value)
			}
		}

		// 记录请求体
		var requestBody []byte
		if c.Request.Body != nil {
			requestBody, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		// 记录响应体
		writer := &responseBodyWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = writer

		// 记录请求开始时间
		start := time.Now()

		// 捕捉 panic 错误并记录日志
		defer func() {
			if rec := recover(); rec != nil {
				err := fmt.Errorf("%v", rec)
				c.Error(err) // 将 panic 转换为 gin 错误

				// 获取堆栈跟踪信息并提取重要部分
				stack := debug.Stack()
				importantStack := extractImportantStack(stack)

				// 检查是否已经写入响应头
				if !c.Writer.Written() {
					// 设置状态码为 500
					c.Writer.WriteHeader(http.StatusInternalServerError)
					json := JsonError{
						Code: http.StatusInternalServerError,
						Msg:  fmt.Sprintf("%v | %s", err, importantStack), // 将重要的堆栈跟踪信息添加到错误信息中
					}
					c.AbortWithStatusJSON(c.Writer.Status(), json)
				}

				// 将错误信息记录到 Error 中
				Error = fmt.Sprintf("panic: %v | %s", err, importantStack)
			}

			// 记录错误信息
			if len(c.Errors) > 0 {
				// 记录捕获的错误信息
				if Error == "None" || Error == "" {
					Error = strings.ReplaceAll(c.Errors.ByType(gin.ErrorTypePrivate).String(), "\n", "")
				} else {
					Error = fmt.Sprintf("%s | %s", Error, strings.ReplaceAll(c.Errors.ByType(gin.ErrorTypePrivate).String(), "\n", ""))
				}
			} else if Error == "" {
				Error = "None"
			}

			// 添加非200状态的状态信息
			if c.Writer.Status() != http.StatusOK {
				statusText, exists := statusText[c.Writer.Status()]
				if exists {
					Error = fmt.Sprintf("%s | %d %s", Error, c.Writer.Status(), statusText)
				} else {
					Error = fmt.Sprintf("%s | %d", Error, c.Writer.Status())
				}
			}

			//请求URL如果没有查询参数，则不显示"/"符号
			var sign string
			if c.Request.URL.RawQuery == "" {
				sign = ""
			} else {
				sign = "/"
			}

			// 获取处理函数名称
			handlerName := c.HandlerName()
			// 记录响应时间
			latency := time.Since(start)
			// 自定义日志格式
			str := fmt.Sprintf(
				"[Gin] | %s | %d |%4.2vms | %s | %+v | %s | Errors: %s",
				c.ClientIP(),
				c.Writer.Status(),
				latency,
				c.Request.Method,
				c.Request.URL.Path+sign+c.Request.URL.RawQuery,
				handlerName, // 添加处理函数名称
				Error,
			)
			if c.Writer.Status() != http.StatusOK || len(c.Errors) > 0 {
				logger.Error(str)
			} else {
				logger.Info(str)
			}
		}()

		c.Request.URL.Path = strings.ToLower(c.Request.URL.Path) // 请求地址 path 统一转为小写,防止前端传大写导致路由匹配不正确

		// 处理请求
		c.Next()
	}
}

// extractImportantStack 提取堆栈跟踪信息中的重要部分并格式化成单行
func extractImportantStack(stack []byte) string {
	lines := bytes.Split(stack, []byte{'\n'})
	var importantLines [][]byte

	// 标记开始记录关键堆栈信息的标志
	record := false

	for _, line := range lines {
		if bytes.Contains(line, []byte("panic")) {
			record = true
		}
		if record {
			importantLines = append(importantLines, bytes.TrimSpace(line))
		}
	}
	// 只保留重要的前3行
	if len(importantLines) > 3 {
		importantLines = importantLines[:3]
	}
	return string(bytes.Join(importantLines, []byte(" | ")))
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

// HTTPSTATUS
var statusText = map[int]string{
	100: "Continue",
	101: "Switching Protocols",
	102: "Processing",
	103: "Early Hints",
	200: "OK",
	201: "Created",
	202: "Accepted",
	203: "Non-Authoritative Information",
	204: "No Content",
	205: "Reset Content",
	206: "Partial Content",
	207: "Multi-Status",
	208: "Already Reported",
	226: "IM Used",
	300: "Multiple Choices",
	301: "Moved Permanently",
	302: "Found",
	303: "See Other",
	304: "Not Modified",
	305: "Use Proxy",
	307: "Temporary Redirect",
	308: "Permanent Redirect",
	400: "Bad Request",
	401: "Unauthorized",
	402: "Payment Required",
	403: "Forbidden",
	404: "Not Found",
	405: "Method Not Allowed",
	406: "Not Acceptable",
	407: "Proxy Authentication Required",
	408: "Request Timeout",
	409: "Conflict",
	410: "Gone",
	411: "Length Required",
	412: "Precondition Failed",
	413: "Payload Too Large",
	414: "URI Too Long",
	415: "Unsupported Media Type",
	416: "Range Not Satisfiable",
	417: "Expectation Failed",
	418: "I'm a teapot",
	421: "Misdirected Request",
	422: "Unprocessable Entity",
	423: "Locked",
	424: "Failed Dependency",
	425: "Too Early",
	426: "Upgrade Required",
	428: "Precondition Required",
	429: "Too Many Requests",
	431: "Request Header Fields Too Large",
	451: "Unavailable For Legal Reasons",
	500: "Internal Server Error",
	501: "Not Implemented",
	502: "Bad Gateway",
	503: "Service Unavailable",
	504: "Gateway Timeout",
	505: "HTTP Version Not Supported",
	506: "Variant Also Negotiates",
	507: "Insufficient Storage",
	508: "Loop Detected",
	510: "Not Extended",
	511: "Network Authentication Required",
}
