/*
 * @Author:   Administrator
 * @IDE:      GoLand
 * @Date:     2023/10/9 14:29
 * @FilePath: log//gin.go
 */

package log

import (
	"github.com/gin-gonic/gin"
	"github.com/mattn/go-isatty"
	"go.uber.org/zap"
	"io"
	"net/http"
	"net/http/httputil"
	"os"
	"runtime"
	"time"
)

const (
	green   = "\033[97;42m"
	white   = "\033[90;47m"
	yellow  = "\033[90;43m"
	red     = "\033[97;41m"
	blue    = "\033[97;44m"
	magenta = "\033[97;45m"
	cyan    = "\033[97;46m"
	reset   = "\033[0m"
)

var notlogged = []string{"/favicon.ico"}

// ConfigGin defines the config for Logger middleware.
type ConfigGin struct {
	// Optional. Default value is gin.defaultLogFormatter
	Formatter Formatter

	// Output is a writer where logs are written.
	// Optional. Default value is gin.DefaultWriter.
	Output io.Writer

	// SkipPaths is a url path array which logs are not written.
	// Optional.
	SkipPaths []string
}

// Formatter gives the signature of the formatter function passed to LoggerWithFormatter
type Formatter func(params FormatterParams) string

// FormatterParams is the structure any formatter will be handed when time to log comes
type FormatterParams struct {
	Request *http.Request

	// TimeStamp shows the time after the server returns a response.
	TimeStamp time.Time
	// StatusCode is HTTP response code.
	StatusCode int
	// Latency is how much time the server cost to process a certain request.
	Latency time.Duration
	// ClientIP equals Context's ClientIP method.
	ClientIP string
	// Method is the HTTP method given to the request.
	Method string
	// Path is a path the client requests.
	Path string
	// ErrorMessage is set if error has occurred in processing the request.
	ErrorMessage string
	// isTerm shows whether does gin's output descriptor refers to a terminal.
	isTerm bool
	// BodySize is the size of the Response Body
	BodySize int
	// Keys are the keys set on the request's context.
	Keys map[string]interface{}
}

// StatusCodeColor is the ANSI color for appropriately logging http status code to a terminal.
func (p *FormatterParams) StatusCodeColor() string {
	code := p.StatusCode

	switch {
	case code >= http.StatusOK && code < http.StatusMultipleChoices:
		return green
	case code >= http.StatusMultipleChoices && code < http.StatusBadRequest:
		return white
	case code >= http.StatusBadRequest && code < http.StatusInternalServerError:
		return yellow
	default:
		return red
	}
}

// MethodColor is the ANSI color for appropriately logging http method to a terminal.
func (p *FormatterParams) MethodColor() string {
	method := p.Method

	switch method {
	case http.MethodGet:
		return blue
	case http.MethodPost:
		return cyan
	case http.MethodPut:
		return yellow
	case http.MethodDelete:
		return red
	case http.MethodPatch:
		return green
	case http.MethodHead:
		return magenta
	case http.MethodOptions:
		return white
	default:
		return reset
	}
}

// ResetColor resets all escape attributes.
func (p *FormatterParams) ResetColor() string {
	return reset
}

// GinLogger instances a Logger middleware that will write the logs to gin.DefaultWriter.
func GinLogger(logger *zap.Logger) gin.HandlerFunc {
	return WithWriter(logger, gin.DefaultWriter, notlogged...)
}

// WithWriter instance a Logger middleware with the specified writer buffer.
// Example: os.Stdout, a file opened in write mode, a socket...
func WithWriter(logger *zap.Logger, out io.Writer, notlogged ...string) gin.HandlerFunc {
	return WithConfig(logger, ConfigGin{
		Output:    out,
		SkipPaths: notlogged,
	})
}

// WithConfig instance a Logger middleware with config.
func WithConfig(log *zap.Logger, conf ConfigGin) gin.HandlerFunc {
	out := conf.Output
	// 如果没有设置输出，则使用默认的输出
	if out == nil {
		out = gin.DefaultWriter
	}
	// 获取需要跳过的路径
	netlogo := conf.SkipPaths
	// 是否为终端
	isTerm := true
	// 判断输出是否为终端
	if w, ok := out.(*os.File); !ok || os.Getenv("TERM") == "dumb" || (!isatty.IsTerminal(w.Fd()) && !isatty.IsCygwinTerminal(w.Fd())) {
		isTerm = false
	}
	var skip map[string]struct{}
	// 如果需要跳过的路径不为空，则将其存储到 map 中
	if length := len(netlogo); length > 0 {
		skip = make(map[string]struct{}, length)
		for _, path := range netlogo {
			skip[path] = struct{}{}
		}
	}

	return func(c *gin.Context) {
		// 记录开始时间和请求路径
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// 继续处理请求
		c.Next()

		// 只有当路径不被跳过时才记录日志
		if _, ok := skip[path]; !ok {
			// 构造日志参数
			param := FormatterParams{
				Request: c.Request,
				isTerm:  isTerm,
				Keys:    c.Keys,
			}
			// 记录结束时间和请求耗时
			param.TimeStamp = time.Now()
			param.Latency = param.TimeStamp.Sub(start)
			// 记录客户端IP、请求方法、状态码、错误信息和响应体大小
			param.ClientIP = c.ClientIP()
			param.Method = c.Request.Method
			param.StatusCode = c.Writer.Status()
			param.ErrorMessage = c.Errors.ByType(gin.ErrorTypePrivate).String()
			param.BodySize = c.Writer.Size()
			// 如果有查询参数，则将其添加到路径中
			if raw != "" {
				path = path + "?" + raw
			}
			param.Path = path
			// 如果没有错误信息，则记录日志
			if len(param.ErrorMessage) == 0 {
				log.Info("[GIN]",
					zap.String("Path", path),
					zap.Int("Code", param.StatusCode),
					zap.String("Method", param.Method),
					zap.String("User-Agent", c.Request.UserAgent()),
					zap.String("Latency", param.Latency.String()),
				)
			} else {
				for _, e := range c.Errors.Errors() {
					log.Error(e)
				}
			}
		}
	}
}

// Recovery 使用zap替换gin内部的recovery模块
func Recovery(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		var (
			// 存储原始请求内容的字节数组
			rawReq []byte
		)
		if c.Request != nil {
			// 获取原始请求内容
			rawReq, _ = httputil.DumpRequest(c.Request, true)
		}
		defer func() {
			// 捕获 panic
			if err := recover(); err != nil {
				const size = 64 << 10
				stack := make([]byte, size)
				// 获取堆栈信息
				stack = stack[:runtime.Stack(stack, false)]
				// 记录错误日志
				logger.Error("[Recovery]",
					// 记录错误信息
					zap.Any("Error", err),
					// 记录原始请求内容
					zap.String("Request", string(rawReq)),
					// 记录请求 URI
					zap.String("RequestURI", c.Request.RequestURI),
					// 记录堆栈信息
					zap.String("Stack", string(stack)),
				)
				// 中止请求并返回 500 错误
				c.AbortWithStatus(http.StatusInternalServerError)
			}
		}()
		// 继续处理请求
		c.Next()
	}
}
