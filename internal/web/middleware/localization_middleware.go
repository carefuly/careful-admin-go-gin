/**
 * Description：
 * FileName：localization_middleware.go
 * Author：CJiaの用心
 * Create：2025/10/10 11:51:26
 * Remark：
 */

package middleware

import (
	"bytes"
	"fmt"
	"github.com/carefuly/careful-admin-go-gin/config"
	loggerModel "github.com/carefuly/careful-admin-go-gin/internal/model/careful/logger"
	"github.com/carefuly/careful-admin-go-gin/pkg/ginx/logger"
	loggerMiddleware "github.com/carefuly/careful-admin-go-gin/pkg/ginx/middleware/logger"
	"github.com/carefuly/careful-admin-go-gin/pkg/models"
	_string "github.com/carefuly/careful-admin-go-gin/pkg/utils/common/string"
	"github.com/carefuly/careful-admin-go-gin/pkg/utils/request_utils"
	"github.com/gin-gonic/gin"
	"github.com/mssola/user_agent"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"io/ioutil"
	"time"
)

type Storage struct {
	rely config.RelyConfig
}

func NewStorage(rely config.RelyConfig) *Storage {
	return &Storage{
		rely: rely,
	}
}

// Build 本地化中间件
func (l *Storage) Build() gin.HandlerFunc {
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
			userAgent := c.Request.UserAgent()
			ua := user_agent.New(userAgent)
			browserName, _ := ua.Browser()
			osName := ua.OS()

			// 创建自定义读取器
			buffer := &bytes.Buffer{}
			loggingReader := &loggerMiddleware.LoggingReader{
				Reader: c.Request.Body,
				Buffer: buffer,
			}
			c.Request.Body = ioutil.NopCloser(loggingReader)

			// 创建自定义响应写入器
			crw := &loggerMiddleware.CustomGinResponseWriter{
				ResponseWriter: c.Writer,
				Body:           bytes.NewBuffer(nil),
			}
			c.Writer = crw

			c.Next()

			// 结束时间
			endTime := time.Now()
			requestTime := endTime.Sub(startTime)

			// 获取响应数据
			responseBody := crw.Body.String()
			responseJson := crw.Format(responseBody)

			model := loggerModel.OperateLogger{
				CoreModels:      models.CoreModels{},
				RequestUsername: request_utils.GetRequestUser(c),
				RequestTime:     fmt.Sprintf("%v", requestTime),
				RequestStatus:   c.Writer.Status(),
				RequestMethod:   requestMethod,
				RequestIp:       requestIP,
				RequestPath:     requestPath,
				RequestQuery:    requestQuery,
				RequestBody:     buffer.String(),
				RequestOs:       osName,
				RequestBrowser:  browserName,
				UserAgent:       userAgent,
				RequestCode:     responseJson.Code,
				RequestResult:   responseBody,
				RequestInternal: l.GetResValue(c, "internalError"),
			}

			// 记录日志
			model.Insert(c, l.rely.Db.Careful, model)

			// GET请求不持久化日志
			if requestMethod != "GET" {
				logs := l.StorageFileLog(c, requestPath)
				logs.Info(
					requestPath,
					zap.String("requestUsername", model.RequestUsername),
					zap.String("requestTime", model.RequestTime),
					zap.Int("requestStatus", model.RequestStatus),
					zap.String("requestMethod", model.RequestMethod),
					zap.String("requestIp", model.RequestIp),
					zap.String("requestPath", model.RequestPath),
					zap.Any("requestQuery", model.RequestQuery),
					zap.Any("requestBody", loggingReader.Format()),
					zap.String("requestOs", model.RequestOs),
					zap.String("requestBrowser", model.RequestBrowser),
					zap.String("userAgent", model.UserAgent),
					zap.Int("requestCode", model.RequestCode),
					zap.Any("requestResult", responseJson),
					zap.String("requestInternal", model.RequestInternal),
				)
			}
		}
	}
}

// GetResValue 获取请求头值
func (l *Storage) GetResValue(c *gin.Context, key string) string {
	value, exists := c.Get(key)
	if !exists {
		return ""
	}
	return value.(string)
}

// StorageFileLog 持久化文件日志
func (l *Storage) StorageFileLog(c *gin.Context, path string) *logger.Logger {
	logCfg := &logger.LogConfig{
		Encoding:     logger.EncodingJSON,             // 使用JSON格式
		OutputPath:   logger.DefaultDynamicPath(path), // 自定义文件路径
		Level:        zapcore.InfoLevel,               // 日志级别
		EnableCaller: true,                            // 记录调用位置
		Rotation: &logger.RotationConfig{ // 文件轮转配置
			MaxSizeMB:  10,   // 10MB文件大小限制
			MaxBackups: 7,    // 保留7个备份
			MaxAgeDays: 30,   // 保留30天
			Compress:   true, // 压缩旧日志
		},
	}

	return logger.NewLogger(logCfg)
}
