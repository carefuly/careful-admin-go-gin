/**
 * Description：
 * FileName：request_logger_middleware.go
 * Author：CJiaの用心
 * Create：2025/10/10 11:49:30
 * Remark：
 */

package middleware

import (
	"bytes"
	"fmt"
	loggerMiddleware "github.com/carefuly/careful-admin-go-gin/pkg/ginx/middleware/logger"
	_string "github.com/carefuly/careful-admin-go-gin/pkg/utils/common/string"
	"github.com/carefuly/careful-admin-go-gin/pkg/utils/request_utils"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"io/ioutil"
	"time"
)

type RequestLogger struct {
	zap *zap.Logger
}

func NewLogger(zap *zap.Logger) *RequestLogger {
	return &RequestLogger{
		zap: zap,
	}
}

// Build 请求日志中间件
func (m *RequestLogger) Build() gin.HandlerFunc {
	return func(c *gin.Context) {
		if _string.ContainsAnySubstring(
			c.Request.URL.Path,
			"health",
			"swagger",
			"static",
			"export") {
			c.Next()
		} else {
			// 开始时间
			startTime := time.Now()

			// 获取请求信息
			requestIP := request_utils.NormalizeIP(c)
			requestPath := c.Request.URL.Path
			requestMethod := c.Request.Method
			requestQuery := c.Request.URL.RawQuery

			// 创建自定义读取器
			buffer := &bytes.Buffer{}
			loggingReader := &loggerMiddleware.LoggingReader{
				Reader: c.Request.Body,
				Buffer: buffer,
			}
			c.Request.Body = ioutil.NopCloser(loggingReader)

			c.Next()

			// 结束时间
			timeStamp := time.Now()
			requestTime := timeStamp.Sub(startTime)

			m.zap.Debug(
				requestPath,
				zap.String("time", fmt.Sprintf("%v", requestTime)),
				zap.String("ip", requestIP),
				zap.String("path", requestPath),
				zap.String("method", requestMethod),
				zap.Any("query", requestQuery),
				zap.Int("status", c.Writer.Status()),
				zap.Any("body", loggingReader.Format()),
				zap.String("errors", c.Errors.ByType(gin.ErrorTypePrivate).String()),
				zap.String("internal", m.GetResValue(c, "internalError")),
			)
		}
	}
}

// GetResValue 获取请求头值
func (m *RequestLogger) GetResValue(c *gin.Context, key string) string {
	value, exists := c.Get(key)
	if !exists {
		return ""
	}
	return value.(string)
}
