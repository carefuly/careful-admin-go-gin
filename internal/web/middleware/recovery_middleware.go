/**
 * Description：
 * FileName：recovery_middleware.go
 * Author：CJiaの用心
 * Create：2025/10/10 11:53:40
 * Remark：
 */

package middleware

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"net/http"
	"runtime"
	"time"
)

type ProductionRecoveryMiddleware struct {
}

func NewProductionRecoveryMiddleware() *ProductionRecoveryMiddleware {
	return &ProductionRecoveryMiddleware{}
}

// Build 生产环境恢复中间件
func (p *ProductionRecoveryMiddleware) Build() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				// 生成错误ID
				errorID := fmt.Sprintf("ERR_%d", time.Now().UnixNano())

				// 获取堆栈信息
				stack := make([]byte, 4096)
				length := runtime.Stack(stack, false)
				stackStr := string(stack[:length])

				// 记录错误
				zap.L().Error("服务器内部错误",
					zap.String("error_id", errorID),
					zap.Any("error", r),
					zap.String("path", c.Request.URL.Path),
					zap.String("method", c.Request.Method),
					zap.String("stack", stackStr),
				)

				// 检查响应是否已经开始写入
				if !c.Writer.Written() {
					c.JSON(http.StatusInternalServerError, gin.H{
						"code":     500,
						"message":  "服务器内部错误",
						"error_id": errorID,
					})
				} else {
					zap.L().Warn("响应已开始写入，无法发送错误响应",
						zap.String("error_id", errorID),
						zap.Int("status", c.Writer.Status()),
					)
				}

				c.Next()
			}
		}()

		c.Next()
	}
}
