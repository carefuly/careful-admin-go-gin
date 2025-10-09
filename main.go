/**
 * Description：
 * FileName：main.go
 * Author：CJiaの用心
 * Create：2025/9/28 20:20:44
 * Remark：
 */

package main

import (
	"github.com/carefuly/careful-admin-go-gin/config"
	"github.com/carefuly/careful-admin-go-gin/ioc"
	"go.uber.org/zap"
)

// @title CarefulAdmin 后台管理系统 API
// @version 1.0.0
// @description CarefulAdmin 是一个高效、安全的企业级后台管理系统，提供完整的权限管理和数据可视化功能。
// @description 功能模块包括: 用户管理、角色管理、部门管理、菜单管理、操作日志和数据字典等。

// @termsOfService http://swagger.io/terms
// @contact.name CJiaの用心 - 技术支持
// @contact.url http://www.swagger.io/support
// @contact.email 2224693191@qq.com
// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /dev-api
// @schemes http

// @securityDefinitions.apikey  BearerAuth
// @in                          header
// @name                        Authorization
// @description                 JWT认证令牌，格式: "Bearer {token}"

// @externalDocs.description    开源代码库
// @externalDocs.url            https://github.com/carefuly/carefuly-admin-go-gin
func main() {
	// 初始化日志
	loggerManager := ioc.InitLogger()
	// 初始化配置管理器
	configManager := ioc.InitConfig("./application.yaml")
	configManager.RelyConfig.Logger = loggerManager.GetLogger()
	// 启动配置文件监听
	if err := configManager.StartWatching(); err != nil {
		zap.S().Fatal("启动配置文件监听失败", err)
	}
	defer configManager.StopWatching()
	// 初始化远程配置
	remoteConfig := ioc.InitLoadNacosConfig(configManager.Config)
	// 初始化数据库池
	dbPool := ioc.NewDbPool(remoteConfig.DatabaseConfig)
	configManager.RelyConfig.Db = config.Database{
		Careful: dbPool.CarefulDB,
		// Table:   dbPool.TableDB,
	}
	// 初始化缓存
	configManager.RelyConfig.Redis = ioc.InitCache(remoteConfig.CacheConfig)
	// Token密钥
	configManager.RelyConfig.Token = remoteConfig.TokenConfig

	server := ioc.NewServer(configManager.RelyConfig, "zh")
	// 初始化翻译器
	if err := server.InitTranslator(); err != nil {
		zap.L().Fatal("翻译器初始化失败", zap.Error(err))
	}
	configManager.RelyConfig.Trans = server.Translator
	// 初始化中间件
	middlewares := server.InitGinMiddlewares(configManager.RelyConfig)
	// 初始化Web服务器
	engine := server.InitWebServer(middlewares, configManager.Config.Application.Debug)
	// 注册API路由
	ioc.RegisterRoutes(true, engine, configManager.RelyConfig)
	// 启动服务
	if err := server.Run(configManager.Config.Server.Host, configManager.Config.Server.Port); err != nil {
		zap.L().Error("服务运行失败", zap.Error(err))
	}
}
