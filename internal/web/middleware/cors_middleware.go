/**
 * Description：
 * FileName：cors_middleware.go
 * Author：CJiaの用心
 * Create：2025/10/9 15:14:58
 * Remark：
 */

package middleware

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"strings"
	"time"
)

// CorsMiddlewareBuilder 跨域中间件
type CorsMiddlewareBuilder struct {
}

func NewCorsMiddlewareBuilder() *CorsMiddlewareBuilder {
	return &CorsMiddlewareBuilder{}
}

func (c *CorsMiddlewareBuilder) Build() gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"POST", "DELETE", "PUT", "GET", "OPTIONS", "UPDATE"},
		AllowHeaders: []string{"Origin", "Accept", "Content-Type", "Authorization", "X-Requested-Id", "X-Requested-Sign"},
		// 响应头
		ExposeHeaders: []string{"Content-Length", "Access-Control-Allow-Origin", "Access-Control-Allow-Headers", "x-jwt-token"},
		// 是否允许带 cookie 之类的东西
		AllowCredentials: true,
		AllowOriginFunc: func(origin string) bool {
			if strings.HasPrefix(origin, "http://localhost") {
				// 开发环境
				return true
			}
			// return strings.Contains(origin, "your_url.com")
			return true
		},
		MaxAge: 12 * time.Hour,
	})
}
