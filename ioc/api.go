/**
 * Description：
 * FileName：api.go
 * Author：CJiaの用心
 * Create：2025/11/23 01:31:43
 * Remark：
 */

package ioc

import (
	"github.com/carefuly/careful-admin-go-gin/config"
	"github.com/carefuly/careful-admin-go-gin/docs"
	"github.com/carefuly/careful-admin-go-gin/pkg/ginx/response"
	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"net/http"
)

type RouterRegistrar interface {
	RegisterRoutes(group *gin.RouterGroup)
}

func RegisterRoutes(debug bool, engine *gin.Engine, rely config.RelyConfig) {
	// 配置接口前缀
	apiPrefix := "/api"
	docs.SwaggerInfo.BasePath = apiPrefix

	// 注册Swagger文档
	engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	// 创建API路由组
	// apiGroup := engine.Group(apiPrefix)
	{
		// 版本化路由组
		// v1 := apiGroup.Group("/v1")
		// {
		// 	// 注册所有路由
		// 	router.InitRouter(rely, v1)
		// }
	}

	// 注册404处理
	engine.NoRoute(func(c *gin.Context) {
		response.NewResponse().Error(c, http.StatusNotFound, "未找到资源，请重试", nil)
	})
}
