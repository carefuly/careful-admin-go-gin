/**
 * Description：
 * FileName：login_jwt_middleware.go
 * Author：CJiaの用心
 * Create：2025/10/10 11:47:27
 * Remark：
 */

package middleware

import (
	"errors"
	"github.com/carefuly/careful-admin-go-gin/config"
	"github.com/carefuly/careful-admin-go-gin/pkg/ginx/response"
	"github.com/carefuly/careful-admin-go-gin/pkg/utils/jwt"
	"github.com/carefuly/careful-admin-go-gin/pkg/utils/request_utils"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"net/http"
	"strings"
	"time"
)

var (
	UnauthorizedNotFound = "请求未携带token，无权限访问"
	UnauthorizedInvalid  = "无效的Token"
)

// LoginJWTMiddlewareBuilder JWT 登录校验
type LoginJWTMiddlewareBuilder struct {
	rely           config.RelyConfig
	ignorePaths    []string
	jwtService     *jwt.DefaultJWTService
	tokenBlacklist *jwt.TokenBlacklist
}

// NewLoginJWTMiddlewareBuilder 创建JWT中间件
func NewLoginJWTMiddlewareBuilder(rely config.RelyConfig) *LoginJWTMiddlewareBuilder {
	// 创建JWT服务
	jwtConfig := jwt.TokenConfig{
		Secret:      rely.Token.Secret,
		ExpireHours: rely.Token.Expire,
		Issuer:      "careful@用心",
		Audience:    []string{"careful-admin"},
		MaxRefresh:  24 * time.Hour, // 允许在24小时内刷新
	}

	jwtService := jwt.NewJWTService(jwtConfig)
	tokenBlacklist := jwt.NewTokenBlacklist(rely.Redis)

	return &LoginJWTMiddlewareBuilder{
		rely:           rely,
		jwtService:     jwtService,
		tokenBlacklist: tokenBlacklist,
	}
}

// IgnorePaths 添加忽略路径
func (l *LoginJWTMiddlewareBuilder) IgnorePaths(path string) *LoginJWTMiddlewareBuilder {
	l.ignorePaths = append(l.ignorePaths, path)
	return l
}

// FailedWithStatus 响应失败并设置HTTP状态码
func (l *LoginJWTMiddlewareBuilder) FailedWithStatus(ctx *gin.Context, httpStatus, code int, msg string) {
	ctx.JSON(httpStatus, gin.H{
		"code":    code,
		"message": msg,
		"data":    nil,
	})
}

// Unauthorized 响应未授权
func (l *LoginJWTMiddlewareBuilder) Unauthorized(ctx *gin.Context, msg string) {
	if msg == "" {
		msg = "未授权，请先登录"
	}
	l.FailedWithStatus(ctx, http.StatusUnauthorized, http.StatusUnauthorized, msg)
}

// Build JWT认证中间件
func (l *LoginJWTMiddlewareBuilder) Build() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// swagger文档
		if l.containsAnySubstring(ctx.Request.URL.Path, []string{"swagger", "static"}) {
			return
		}
		// 不需要登录校验的
		for _, path := range l.ignorePaths {
			if ctx.Request.URL.Path == path {
				return
			}
		}

		// 获取Authorization头
		authHeader := ctx.GetHeader("Authorization")
		if authHeader == "" {
			response.NewResponse().Error(ctx, http.StatusUnauthorized, UnauthorizedNotFound, nil)
			ctx.Abort()
			return
		}

		seg := strings.Split(authHeader, " ")
		if len(seg) != 2 {
			// 没登录，有人瞎搞
			response.NewResponse().Error(ctx, http.StatusUnauthorized, UnauthorizedInvalid, nil)
			ctx.Abort()
			return
		}

		tokenStr := seg[1]

		// 检查token是否在黑名单中
		tokenBlacklist := jwt.NewTokenBlacklist(l.rely.Redis)
		blacklisted, err := tokenBlacklist.IsBlacklisted(ctx, tokenStr)
		if err != nil {
			zap.L().Error("检查token黑名单失败", zap.Error(err))
			response.NewResponse().Error(ctx, http.StatusInternalServerError, "服务器内部错误", nil)
			ctx.Abort()
			return
		}

		if blacklisted {
			response.NewResponse().Error(ctx, http.StatusUnauthorized, "Token已失效，请重新登录", nil)
			ctx.Abort()
			return
		}

		// 解析token
		claims, err := l.jwtService.ParseToken(tokenStr)
		if err != nil {
			switch {
			case errors.Is(err, jwt.ErrExpiredToken):
				response.NewResponse().Error(ctx, http.StatusUnauthorized, "Token已过期", nil)
			case errors.Is(err, jwt.ErrInvalidToken):
				response.NewResponse().Error(ctx, http.StatusUnauthorized, "Token无效", nil)
			default:
				response.NewResponse().Error(ctx, http.StatusUnauthorized, "Token认证失败", nil)
			}
			ctx.Abort()
			return
		}

		// gin.Context.Set() 方法将数据存储到上下文，可以在后续的中间件或处理程序中访问。
		// 通过 gin.Context.Get() 方法获取存储在上下文中的数据。
		// 通过 gin.Context.Set() 方法存储数据时，需要指定一个键，以便在后续的中间件或处理程序中访问该数据。
		// 通过 gin.Context.Get() 方法获取数据时，需要指定相同的键。

		ctx.Set("requestIp", request_utils.NormalizeIP(ctx))
		ctx.Set("request", ctx.Request)

		// 将用户信息存储到上下文
		ctx.Set("claims", claims)
		ctx.Set("userId", claims.UserId)
		ctx.Set("username", claims.UserInfo["username"])
		ctx.Set("deptId", claims.UserInfo["deptId"])
		ctx.Set("userInfo", claims.UserInfo)
		ctx.Set("X-Request-ID", claims.UserId)

		// 刷新token(如果接近过期)
		l.maybeRefreshToken(ctx, claims, tokenStr)

		ctx.Next()
	}
}

// containsAnySubstring 检查字符串是否包含切片中的任意一个子串
func (l *LoginJWTMiddlewareBuilder) containsAnySubstring(str string, subs []string) bool {
	for _, sub := range subs {
		if strings.Contains(str, sub) {
			return true
		}
	}
	return false
}

// maybeRefreshToken 如果token接近过期，则刷新它
func (l *LoginJWTMiddlewareBuilder) maybeRefreshToken(ctx *gin.Context, claims *jwt.Claims, currentToken string) {
	// 如果token在30分钟内过期，则刷新它
	// if time.Until(claims.ExpiresAt.Time) < 30*time.Minute {
	// 	newToken, err := l.jwtService.RefreshToken(ctx, currentToken)
	// 	if err == nil {
	// 		// 将新token设置在响应头中
	// 		ctx.Header("X-Refreshed-Token", newToken)
	// 	}
	// }
}

// GetUserIDFromContext 从上下文中获取用户ID
func (l *LoginJWTMiddlewareBuilder) GetUserIDFromContext(ctx *gin.Context) (string, error) {
	userID, exists := ctx.Get("userId")
	if !exists {
		return "", errors.New("用户ID未找到")
	}

	return userID.(string), nil
}

// GetUserInfoFromContext 从上下文中获取用户信息
func (l *LoginJWTMiddlewareBuilder) GetUserInfoFromContext(ctx *gin.Context) (map[string]interface{}, error) {
	userInfo, exists := ctx.Get("userInfo")
	if !exists {
		return nil, errors.New("用户信息未找到")
	}

	return userInfo.(map[string]interface{}), nil
}
