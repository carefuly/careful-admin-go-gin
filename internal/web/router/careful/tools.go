/**
 * Description：
 * FileName：tools.go
 * Author：CJiaの用心
 * Create：2025/12/3 23:46:02
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

	// cache
	userCache := cacheSystem.NewRedisUserCache(r.rely.Redis)
	userCacheLogger := cacheRecord.NewCacheLogger(r.rely.Db.Careful)
	userCacheLoggingDecorator := cacheDecoratorSystem.NewUserCacheLoggingDecorator(userCache, userCacheLogger)
	dictCache := cacheTools.NewRedisDictCache(r.rely.Redis)
	dictCacheLogger := cacheRecord.NewCacheLogger(r.rely.Db.Careful)
	dictCacheLoggingDecorator := cacheDecoratorTools.NewDictCacheLoggingDecorator(dictCache, dictCacheLogger)
	dictTypeCache := cacheTools.NewRedisDictTypeCache(r.rely.Redis)
	dictTypeCacheLogger := cacheRecord.NewCacheLogger(r.rely.Db.Careful)
	dictTypeCacheLoggingDecorator := cacheDecoratorTools.NewDictTypeCacheLoggingDecorator(dictTypeCache, dictTypeCacheLogger)

	// dao
	userDAO := daoSystem.NewGORMUserDAO(r.rely.Db.Careful)
	dictDAO := daoTools.NewGORMDictDAO(r.rely.Db.Careful)
	dictTypeDAO := daoTools.NewGORMDictTypeDAO(r.rely.Db.Careful)

	// repository
	userRepository := repositorySystem.NewUserRepository(userDAO, userCacheLoggingDecorator)
	dictRepository := repositoryTools.NewDictRepository(dictDAO, dictCacheLoggingDecorator)
	dictTypeRepository := repositoryTools.NewDictTypeRepository(dictTypeDAO, dictTypeCacheLoggingDecorator)

	// service
	userService := serviceSystem.NewUserService(userRepository)
	dictService := serviceTools.NewDictService(dictRepository)
	dictTypeService := serviceTools.NewDictTypeService(dictTypeRepository, dictRepository)

	// web
	dictHandler := handlerTools.NewDictHandler(r.rely, dictService, userService)
	dictTypeHandler := handlerTools.NewDictTypeHandler(r.rely, dictTypeService, userService)

	// 数据字典
	dictHandler.RegisterRoutes(baseRouter)
	// 字典项
	dictTypeHandler.RegisterRoutes(baseRouter)
}
