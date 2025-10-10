/**
 * Description：
 * FileName：request_timeout_middleware.go
 * Author：CJiaの用心
 * Create：2025/10/10 11:52:58
 * Remark：
 */

package middleware

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"net/http"
	"strings"
	"time"
)

/**
使用：
config := middleware.TimeoutConfig{
	DefaultTimeout: 15 * time.Second,
	ExactPathTimeouts: map[string]time.Duration{
		"/api/export": 5 * time.Minute,  // 精确匹配导出操作
	},
	PrefixPathTimeouts: map[string]time.Duration{
		"/api/upload": 30 * time.Second, // 前缀匹配所有上传接口
	},
	SuffixPathTimeouts: map[string]time.Duration{
		"/reports": 2 * time.Minute,     // 后缀匹配所有报告接口
	},
}
router.Use(middleware.RequestTimeoutWithConfig(config))
*/

// TimeoutConfig 超时配置
type TimeoutConfig struct {
	DefaultTimeout time.Duration
	// 精确路径匹配
	ExactPathTimeouts map[string]time.Duration
	// 前缀路径匹配（例如 "/api/upload" 会匹配 "/api/upload/avatar" 和 "/api/upload/file"）
	PrefixPathTimeouts map[string]time.Duration
	// 后缀路径匹配（例如 "/upload" 会匹配 "/api/user/upload" 和 "/api/role/upload"）
	SuffixPathTimeouts map[string]time.Duration
}

type RequestTimeoutWithConfig struct {
	config TimeoutConfig
}

func NewRequestTimeoutWithConfig(config TimeoutConfig) *RequestTimeoutWithConfig {
	return &RequestTimeoutWithConfig{
		config: config,
	}
}

// Build 带配置的超时中间件
func (r *RequestTimeoutWithConfig) Build() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 确定当前路径的超时时间
		path := c.Request.URL.Path
		var timeout time.Duration
		var found bool

		// 1. 首先检查精确匹配
		if timeout, found = r.config.ExactPathTimeouts[path]; found {
			// 使用精确匹配的超时
		} else {
			// 2. 检查前缀匹配
			for prefix, prefixTimeout := range r.config.PrefixPathTimeouts {
				if strings.HasPrefix(path, prefix) {
					timeout = prefixTimeout
					found = true
					break
				}
			}

			// 3. 检查后缀匹配
			if !found {
				for suffix, suffixTimeout := range r.config.SuffixPathTimeouts {
					if strings.HasSuffix(path, suffix) {
						timeout = suffixTimeout
						found = true
						break
					}
				}
			}

			// 4. 如果没有匹配到任何特殊规则，使用默认超时
			if !found {
				timeout = r.config.DefaultTimeout
			}
		}

		// 创建超时上下文
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		// 替换请求的上下文
		c.Request = c.Request.WithContext(ctx)

		// 使用通道来跟踪请求完成或超时
		done := make(chan struct{})
		go func() {
			defer close(done)
			c.Next() // 继续处理请求
		}()

		select {
		case <-done:
			// 请求正常完成
			return
		case <-ctx.Done():
			// 请求超时
			if errors.Is(ctx.Err(), context.DeadlineExceeded) {
				// 记录超时日志
				zap.L().Warn("请求超时",
					zap.String("path", path),
					zap.String("method", c.Request.Method),
					zap.Duration("timeout", timeout),
				)

				// 检查是否已经开始写入响应
				if !c.Writer.Written() {
					c.AbortWithStatusJSON(http.StatusGatewayTimeout, gin.H{
						"code":    504,
						"message": "请求处理超时",
						"details": "服务器在规定时间内未能完成请求处理",
					})
				} else {
					// 如果已经开始写入响应，记录警告但无法改变响应
					zap.L().Warn("请求超时，但响应已部分发送",
						zap.String("path", path),
						zap.String("method", c.Request.Method),
					)
				}
			}
		}
	}
}
