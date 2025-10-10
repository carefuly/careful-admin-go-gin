/**
 * Description：
 * FileName：auth.go
 * Author：CJiaの用心
 * Create：2025/10/10 11:13:24
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
	authSystem "github.com/carefuly/careful-admin-go-gin/internal/web/handler/careful/auth"
	"github.com/carefuly/careful-admin-go-gin/pkg/utils/jwt"
	"github.com/gin-gonic/gin"
	"time"
)

type AuthRouter struct {
	rely   config.RelyConfig
	router *gin.RouterGroup
}

func NewAuthRouter(rely config.RelyConfig, router *gin.RouterGroup) *AuthRouter {
	return &AuthRouter{
		rely:   rely,
		router: router,
	}
}

func (r *AuthRouter) RegisterRouter() {
	baseRouter := r.router.Group("/auth")

	userCache := cacheSystem.NewRedisUserCache(r.rely.Redis)
	userCacheLogger := cacheRecord.NewCacheLogger(r.rely.Db.Careful)
	userCacheLoggingDecorator := cacheDecoratorSystem.NewUserCacheLoggingDecorator(userCache, userCacheLogger)
	userDAO := daoSystem.NewGORMUserDAO(r.rely.Db.Careful)
	userRepository := repositorySystem.NewUserRepository(userDAO, userCacheLoggingDecorator)
	userService := serviceSystem.NewUserService(userRepository)
	// jwt配置
	jwtConfig := jwt.TokenConfig{
		Secret:      r.rely.Token.Secret,
		ExpireHours: r.rely.Token.Expire,
		Issuer:      "careful@用心",
		Audience:    []string{"careful-admin"},
		MaxRefresh:  24 * time.Hour, // 允许在24小时内刷新
	}
	jwtService := jwt.NewJWTService(jwtConfig)
	// 黑名单配置
	blacklistService := jwt.NewTokenBlacklist(r.rely.Redis)
	authHandler := authSystem.NewAuthHandler(r.rely, userService, jwtService, blacklistService)
	authHandler.RegisterRoutes(baseRouter)
}
