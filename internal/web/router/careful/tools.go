/**
 * Description：
 * FileName：tools.go
 * Author：CJiaの用心
 * Create：2025/10/11 14:07:06
 * Remark：
 */

package careful

import (
	"github.com/carefuly/careful-admin-go-gin/config"
	cacheSystem "github.com/carefuly/careful-admin-go-gin/internal/repository/cache/careful/system"
	cacheTools "github.com/carefuly/careful-admin-go-gin/internal/repository/cache/careful/tools"
	cacheDecoratorSystem "github.com/carefuly/careful-admin-go-gin/internal/repository/cache/decorator/careful/system"
	cacheDecoratorTools "github.com/carefuly/careful-admin-go-gin/internal/repository/cache/decorator/careful/tools"
	cacheRecord "github.com/carefuly/careful-admin-go-gin/internal/repository/cache/decorator/record"
	daoSystem "github.com/carefuly/careful-admin-go-gin/internal/repository/dao/careful/system"
	daoTools "github.com/carefuly/careful-admin-go-gin/internal/repository/dao/careful/tools"
	repositorySystem "github.com/carefuly/careful-admin-go-gin/internal/repository/repository/careful/system"
	repositoryTools "github.com/carefuly/careful-admin-go-gin/internal/repository/repository/careful/tools"
	serviceSystem "github.com/carefuly/careful-admin-go-gin/internal/service/careful/system"
	serviceTools "github.com/carefuly/careful-admin-go-gin/internal/service/careful/tools"
	handlerTools "github.com/carefuly/careful-admin-go-gin/internal/web/handler/careful/tools"
	"github.com/gin-gonic/gin"
)

type ToolsRouter struct {
	rely   config.RelyConfig
	router *gin.RouterGroup
}

func NewToolsRouter(rely config.RelyConfig, router *gin.RouterGroup) *ToolsRouter {
	return &ToolsRouter{
		rely:   rely,
		router: router,
	}
}

func (r *ToolsRouter) RegisterRouter() {
	baseRouter := r.router.Group("/tools")

	// 用户
	userCache := cacheSystem.NewRedisUserCache(r.rely.Redis)
	userCacheLogger := cacheRecord.NewCacheLogger(r.rely.Db.Careful)
	userCacheLoggingDecorator := cacheDecoratorSystem.NewUserCacheLoggingDecorator(userCache, userCacheLogger)
	userDAO := daoSystem.NewGORMUserDAO(r.rely.Db.Careful)
	userRepository := repositorySystem.NewUserRepository(userDAO, userCacheLoggingDecorator)
	userService := serviceSystem.NewUserService(userRepository)

	// 数据字典
	dictCache := cacheTools.NewRedisDictCache(r.rely.Redis)
	dictCacheLogger := cacheRecord.NewCacheLogger(r.rely.Db.Careful)
	dictDAO := daoTools.NewGORMDictDAO(r.rely.Db.Careful)
	dictCacheLoggingDecorator := cacheDecoratorTools.NewDictCacheLoggingDecorator(dictCache, dictCacheLogger)
	dictRepository := repositoryTools.NewDictRepository(dictDAO, dictCacheLoggingDecorator)
	dictService := serviceTools.NewDictService(dictRepository)
	dictHandler := handlerTools.NewDictHandler(r.rely, dictService, userService)
	dictHandler.RegisterRoutes(baseRouter)
}
