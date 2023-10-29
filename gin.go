/*
 * @Author:   Administrator
 * @IDE:      GoLand
 * @Date:     2023/10/9 14:29
 * @FilePath: log//gin.go
 */

package log

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/mattn/go-isatty"
	"go.uber.org/zap"
	"io"
	"net/http"
	"net/http/httputil"
	"os"
	"runtime"
	"sync"
	"time"
)

var (
	// 堆栈池
	stackPool = sync.Pool{
		// New 方法用于分配新的堆栈内存
		New: func() interface{} {
			return make([]byte, 64<<10)
		},
	}
	// 日志排除路径
	notlogged = []string{"/favicon.ico"}
)

// ConfigGin defines the config for Logger middleware.
type ConfigGin struct {
	// Output is a writer where logs are written.
	Output io.Writer

	// SkipPaths is a url path array which logs are not written.
	// Optional.
	SkipPaths []string
}

// FormatterParams is the structure any formatter will be handed when time to log comes
type FormatterParams struct {
	Request *http.Request

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
// This function takes in a log, configgin and gin.HandlerFunc as parameters and returns a gin.HandlerFunc
func WithConfig(log *zap.Logger, conf ConfigGin) gin.HandlerFunc {
	// Set the output to the configgin's output, or the default writer if the output is nil
	out := conf.Output
	if out == nil {
		out = gin.DefaultWriter
	}
	// Set the isTerm boolean to true if the output is a file, or if the TERM environment variable is set to "dumb" or if the file descriptor is not a terminal or a cygwin terminal
	isTerm := true
	if w, ok := out.(*os.File); !ok || os.Getenv("TERM") == "dumb" || (!isatty.IsTerminal(w.Fd()) && !isatty.IsCygwinTerminal(w.Fd())) {
		isTerm = false
	}
	// Create a map of paths to skip
	skip := make(map[string]struct{})
	for _, path := range conf.SkipPaths {
		skip[path] = struct{}{}
	}
	// Return a function that takes in a gin.Context as a parameter
	return func(c *gin.Context) {
		// Set the start time to the current time
		start := time.Now()
		// Set the path to the request URL path
		path := c.Request.URL.Path
		// Set the raw query to the request URL raw query
		raw := c.Request.URL.RawQuery
		// Call the next handler
		c.Next()
		// If the path is not in the skip map
		if _, ok := skip[path]; !ok {
			// Create a FormatterParams struct
			param := FormatterParams{
				Request: c.Request,
				isTerm:  isTerm,
				Keys:    c.Keys,
			}
			// Set the latency to the difference between the current time and the start time
			param.Latency = time.Since(start)
			// Set the client IP to the request client IP
			param.ClientIP = c.ClientIP()
			// Set the method to the request method
			param.Method = c.Request.Method
			// Set the status code to the request writer status
			param.StatusCode = c.Writer.Status()
			// Set the error message to the request errors by type private string
			param.ErrorMessage = c.Errors.ByType(gin.ErrorTypePrivate).String()
			// Set the body size to the request writer size
			param.BodySize = c.Writer.Size()
			// If the raw query is not empty
			if raw != "" {
				// Set the path to the path and raw query
				path = path + "?" + raw
			}
			// Set the path to the parameter
			param.Path = path
			// If the error message is empty
			if len(param.ErrorMessage) == 0 {
				// Log the request path, status code, method, user agent, latency, and the request
				log.Info(fmt.Sprintf("%5s %-5s", "", ""),
					zap.String("Path", path),
					zap.Int("Code", param.StatusCode),
					zap.String("Method", param.Method),
					zap.String("User-Agent", c.Request.UserAgent()),
					zap.String("Latency", param.Latency.String()),
				)
				// If the error message is not empty
			} else {
				// Log the error message
				log.Error(fmt.Sprintf("%5s %-5s", "[GIN]", ""),
					zap.String("Path", c.Request.URL.Path),
					zap.String("Error", param.ErrorMessage))
			}
		}
	}
}

// Recovery This function is used to recover from panic and log the error
func Recovery(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get a stack from the pool
		stack := stackPool.Get().([]byte)
		// Defer the put of the stack back into the pool until the function returns
		defer stackPool.Put(stack[:0])
		// Defer the execution of the code until the function returns
		defer func() {
			// Create a byte array to store the request
			var rawReq []byte
			// If the request is not nil, dump it into the byte array
			if c.Request != nil {
				rawReq, _ = httputil.DumpRequest(c.Request, true)
			}
			// If there is a panic, log the error
			if err := recover(); err != nil {
				stack = stack[:runtime.Stack(stack, false)]
				logger.Error("[Recovery]",
					zap.String("Path", c.Request.RequestURI),
					zap.Any("Error", err),
					zap.ByteString("Request", rawReq),
					zap.String("Stack", string(stack)),
				)
				// Abort the request with an internal server error
				c.AbortWithStatus(http.StatusInternalServerError)
			}
		}()
		// Execute the next handler
		c.Next()
	}
}
