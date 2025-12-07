/**
 * Description：
 * FileName：system.go
 * Author：CJiaの用心
 * Create：2025/11/26 10:16:17
 * Remark：
 */

package careful

import (
	"github.com/carefuly/careful-admin-go-gin/config"
	cacheSystem "github.com/carefuly/careful-admin-go-gin/internal/repository/cache/careful/system"
	cacheDecoratorSystem "github.com/carefuly/careful-admin-go-gin/internal/repository/cache/decorator/careful/system"
	cacheRecord "github.com/carefuly/careful-admin-go-gin/internal/repository/cache/decorator/record"
	daoSystem "github.com/carefuly/careful-admin-go-gin/internal/repository/dao/careful/system"
	repositorySystem "github.com/carefuly/careful-admin-go-gin/internal/repository/repository/careful/system"
	serviceSystem "github.com/carefuly/careful-admin-go-gin/internal/service/careful/system"
	handlerSystem "github.com/carefuly/careful-admin-go-gin/internal/web/handler/careful/system"
	"github.com/gin-gonic/gin"
)

type SystemRouter struct {
	rely   config.RelyConfig
	router *gin.RouterGroup
}

func NewSystemRouter(rely config.RelyConfig, router *gin.RouterGroup) *SystemRouter {
	return &SystemRouter{
		rely:   rely,
		router: router,
	}
}

func (r *SystemRouter) RegisterRouter() {
	baseRouter := r.router.Group("/system")

	// 用户
	userCache := cacheSystem.NewRedisUserCache(r.rely.Redis)
	userCacheLogger := cacheRecord.NewCacheLogger(r.rely.Db.Careful)
	userCacheLoggingDecorator := cacheDecoratorSystem.NewUserCacheLoggingDecorator(userCache, userCacheLogger)
	userDAO := daoSystem.NewGORMUserDAO(r.rely.Db.Careful)
	userRepository := repositorySystem.NewUserRepository(userDAO, userCacheLoggingDecorator)
	userService := serviceSystem.NewUserService(userRepository)

	// 菜单按钮
	menuButtonCache := cacheSystem.NewRedisMenuButtonCache(r.rely.Redis)
	menuButtonCacheLogger := cacheRecord.NewCacheLogger(r.rely.Db.Careful)
	menuButtonCacheLoggingDecorator := cacheDecoratorSystem.NewMenuButtonCacheLoggingDecorator(menuButtonCache, menuButtonCacheLogger)
	menuButtonDAO := daoSystem.NewGORMMenuButtonDAO(r.rely.Db.Careful)
	menuButtonRepository := repositorySystem.NewMenuButtonRepository(menuButtonDAO, menuButtonCacheLoggingDecorator)

	// 菜单
	menuCache := cacheSystem.NewRedisMenuCache(r.rely.Redis)
	menuCacheLogger := cacheRecord.NewCacheLogger(r.rely.Db.Careful)
	menuCacheLoggingDecorator := cacheDecoratorSystem.NewMenuCacheLoggingDecorator(menuCache, menuCacheLogger)
	menuDAO := daoSystem.NewGORMMenuDAO(r.rely.Db.Careful)
	menuRepository := repositorySystem.NewMenuRepository(menuDAO, menuCacheLoggingDecorator)
	menuService := serviceSystem.NewMenuService(menuRepository, menuButtonRepository)
	menuHandler := handlerSystem.NewMenuHandler(r.rely, menuService, userService)
	menuHandler.RegisterRoutes(baseRouter)

	// 菜单按钮
	// menuButtonService := serviceSystem.NewMenuButtonService(menuButtonRepository)

	// 部门
	deptCache := cacheSystem.NewRedisDeptCache(r.rely.Redis)
	deptCacheLogger := cacheRecord.NewCacheLogger(r.rely.Db.Careful)
	deptCacheLoggingDecorator := cacheDecoratorSystem.NewDeptCacheLoggingDecorator(deptCache, deptCacheLogger)
	deptDAO := daoSystem.NewGORMDeptDAO(r.rely.Db.Careful)
	deptRepository := repositorySystem.NewDeptRepository(deptDAO, deptCacheLoggingDecorator)
	deptService := serviceSystem.NewDeptService(deptRepository)
	deptHandler := handlerSystem.NewDeptHandler(r.rely, deptService, userService)
	deptHandler.RegisterRoutes(baseRouter)

	// 岗位
	postCache := cacheSystem.NewRedisPostCache(r.rely.Redis)
	postCacheLogger := cacheRecord.NewCacheLogger(r.rely.Db.Careful)
	postCacheLoggingDecorator := cacheDecoratorSystem.NewPostCacheLoggingDecorator(postCache, postCacheLogger)
	postDAO := daoSystem.NewGORMPostDAO(r.rely.Db.Careful)
	postRepository := repositorySystem.NewPostRepository(postDAO, postCacheLoggingDecorator)
	postService := serviceSystem.NewPostService(postRepository, deptRepository)
	postHandler := handlerSystem.NewPostHandler(r.rely, postService, userService)
	postHandler.RegisterRoutes(baseRouter)
}
