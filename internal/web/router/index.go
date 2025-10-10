/**
 * Description：
 * FileName：index.go
 * Author：CJiaの用心
 * Create：2025/10/10 11:14:58
 * Remark：
 */

package router

import (
	"github.com/carefuly/careful-admin-go-gin/config"
	routerCareful "github.com/carefuly/careful-admin-go-gin/internal/web/router/careful"
	"github.com/gin-gonic/gin"
)

// InitRouter 初始化所有路由
func InitRouter(rely config.RelyConfig, router *gin.RouterGroup) {
	// 注册标准系统相关路由
	routerCareful.NewRouter(rely, router).RegisterRoutes()
}
