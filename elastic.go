/*
 * @Author:   Administrator
 * @IDE:      GoLand
 * @Date:     2023/11/16 14:43
 * @FilePath: log//elastic.go
 */

package log

import (
	"bufio"
	"bytes"
	"fmt"
	"go.uber.org/zap"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type ElasticsearchLog struct {
	// Logger to use for logging
	*zap.Logger
	// Whether to log request body
	RequestBody bool
	// Whether to log response body
	ResponseBody bool
}

// NewElasticLogger Create a new ElasticsearchLog struct with the given zapLogger, requestBody and responseBody
func NewElasticLogger(zapLogger *zap.Logger, requestBody, responseBody bool) *ElasticsearchLog {
	// Create a new ElasticsearchLog struct
	return &ElasticsearchLog{
		// Set the Logger field to the given zapLogger
		Logger: zapLogger,
		// Set the RequestBody field to the given requestBody
		RequestBody: requestBody,
		// Set the ResponseBody field to the given responseBody
		ResponseBody: responseBody,
	}
}

// LogRoundTrip Function to log roundtrip information for an Elasticsearch request
func (l *ElasticsearchLog) LogRoundTrip(req *http.Request, res *http.Response, err error, start time.Time, dur time.Duration) error {
	// Unescape the query string
	query, _ := url.QueryUnescape(req.URL.RawQuery)
	// If there is a query string, add it to the URL
	if query != "" {
		query = "?" + query
	}

	// Set the status to the response status
	var status = res.Status

	// Check the status code and set the status accordingly
	switch {
	case res.StatusCode > 0 && res.StatusCode < 300:
	case res.StatusCode > 299 && res.StatusCode < 500:
	case res.StatusCode > 499:
	default:
		status = "ERROR"
	}
	// If the request method is not a HEAD request, log the request information
	if req.Method != http.MethodHead {
		l.Info("[ELASTIC]",
			zap.String("Method", req.Method),
			zap.String("Scheme", req.URL.Scheme),
			zap.String("Host", req.URL.Host),
			zap.String("Path", req.URL.Path+query),
			zap.String("Status", status),
			zap.String("Time", dur.Truncate(time.Millisecond).String()),
		)
	}

	// If the request body is enabled and the request body is not empty, log the request body
	if l.RequestBodyEnabled() && req != nil && req.Body != nil && req.Body != http.NoBody {
		var buf bytes.Buffer
		if req.GetBody != nil {
			b, _ := req.GetBody()
			_, _ = buf.ReadFrom(b)
		} else {
			_, _ = buf.ReadFrom(req.Body)
		}
		l.Info(fmt.Sprintf("[ES-REQUEST]  %-6s", logBodyAsText(&buf)))
	}

	// If the response body is enabled and the response body is not empty, log the response body
	if l.ResponseBodyEnabled() && res != nil && res.Body != nil && res.Body != http.NoBody {
		defer res.Body.Close()
		var buf bytes.Buffer
		_, _ = buf.ReadFrom(res.Body)
		l.Info(fmt.Sprintf("[ES-RESPONSE] %-6s ", logBodyAsText(&buf)))
	}

	// If there is an error, log the error
	if err != nil {
		return err
	}

	// Return nil
	return nil
}

// RequestBodyEnabled This function checks if the request body is enabled for the ElasticsearchLog
func (l *ElasticsearchLog) RequestBodyEnabled() bool {
	return l.RequestBody
}

// ResponseBodyEnabled This function checks if the response body is enabled for the ElasticsearchLog
func (l *ElasticsearchLog) ResponseBodyEnabled() bool {
	return l.ResponseBody
}

// logBodyAsText takes an io.Reader and returns a string
func logBodyAsText(body io.Reader) (str string) {
	// Create a new scanner to read the body
	scanner := bufio.NewScanner(body)
	// Loop through each line in the body
	for scanner.Scan() {
		// Trim any whitespace from the line
		s := scanner.Text()
		if s != "" {
			// Add the line to the string
			str += strings.TrimSpace(s)
		}
	}
	// Return the string
	return
}
