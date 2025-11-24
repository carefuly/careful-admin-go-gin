/**
 * Description：
 * FileName：types.go
 * Author：CJiaの用心
 * Create：2025/10/10 11:42:52
 * Remark：
 */

package logger

import (
	"bytes"
	"encoding/json"
	"github.com/carefuly/careful-admin-go-gin/pkg/ginx/response"
	"github.com/gin-gonic/gin"
	"io"
)

// LoggingReader 自定义读取器，用于记录读取的内容
type LoggingReader struct {
	Reader io.Reader
	Buffer *bytes.Buffer
}

func (lr *LoggingReader) Read(p []byte) (n int, err error) {
	n, err = lr.Reader.Read(p)
	if n > 0 {
		lr.Buffer.Write(p[:n])
	}
	return
}

func (lr *LoggingReader) Format() any {
	var result map[string]any
	err := json.Unmarshal([]byte(lr.Buffer.String()), &result)
	if err != nil {
		return nil
	}
	return result
}

// CustomGinResponseWriter 自定义 Gin 响应写入器
type CustomGinResponseWriter struct {
	gin.ResponseWriter
	Body *bytes.Buffer
}

func (crw *CustomGinResponseWriter) Write(b []byte) (int, error) {
	// 记录响应数据
	crw.Body.Write(b)
	return crw.ResponseWriter.Write(b)
}

func (crw *CustomGinResponseWriter) Format(src string) response.Response {
	var result response.Response
	err := json.Unmarshal([]byte(src), &result)
	if err != nil {
		return response.Response{}
	}
	return result
}
