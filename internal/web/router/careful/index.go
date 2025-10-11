/**
 * Description：
 * FileName：index.go
 * Author：CJiaの用心
 * Create：2025/10/10 11:13:00
 * Remark：
 */

package careful

import (
	"github.com/carefuly/careful-admin-go-gin/config"
	"github.com/gin-gonic/gin"
)

type Router struct {
	rely   config.RelyConfig
	router *gin.RouterGroup
}

func NewRouter(rely config.RelyConfig, router *gin.RouterGroup) *Router {
	return &Router{
		rely:   rely,
		router: router,
	}
}

func (r *Router) RegisterRoutes() {
	// 认证管理
	NewAuthRouter(r.rely, r.router).RegisterRouter()
	// 系统工具
	NewToolsRouter(r.rely, r.router).RegisterRouter()
}
