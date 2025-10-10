/**
 * Description：
 * FileName：auth.go
 * Author：CJiaの用心
 * Create：2025/10/10 10:30:20
 * Remark：
 */

package auth

import (
	"errors"
	"fmt"
	"github.com/carefuly/careful-admin-go-gin/config"
	domainSystem "github.com/carefuly/careful-admin-go-gin/internal/domain/careful/system"
	serviceSystem "github.com/carefuly/careful-admin-go-gin/internal/service/careful/system"
	"github.com/carefuly/careful-admin-go-gin/pkg/ginx/response"
	"github.com/carefuly/careful-admin-go-gin/pkg/utils/jwt"
	"github.com/carefuly/careful-admin-go-gin/pkg/utils/request_utils"
	"github.com/carefuly/careful-admin-go-gin/pkg/validate"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"net/http"
	"strings"
	"time"
)

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"` // 用户名
	Password string `json:"password" binding:"required,min=6,max=50"` // 密码
}

// LoginResponse 登录响应
type LoginResponse struct {
	Token  string            `json:"token"`  // JWT令牌
	User   domainSystem.User `json:"user"`   // 用户信息
	Expire int               `json:"expire"` // 过期时间(秒)
}

// RefreshTokenRequest 刷新令牌请求
type RefreshTokenRequest struct {
	Token string `json:"token" binding:"required" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6Ikp..."` // 旧的JWT令牌
}

// ChangePasswordRequest 修改密码请求
type ChangePasswordRequest struct {
	OldPassword string `json:"oldPassword" binding:"required"`              // 旧密码
	NewPassword string `json:"newPassword" binding:"required,min=6,max=50"` // 新密码
}

type AuthsHandler interface {
	RegisterRoutes(router *gin.RouterGroup)
	LoginHandler(ctx *gin.Context)
	RefreshTokenHandler(ctx *gin.Context)
	LogoutHandler(ctx *gin.Context)
	ProfileHandler(ctx *gin.Context)
}

type authHandler struct {
	rely         config.RelyConfig
	userSvc      serviceSystem.UserService
	jwtSvc       *jwt.DefaultJWTService
	blacklistSvc *jwt.TokenBlacklist
}

func NewAuthHandler(rely config.RelyConfig, svc serviceSystem.UserService,
	jwtSvc *jwt.DefaultJWTService, blacklistSvc *jwt.TokenBlacklist) AuthsHandler {
	return &authHandler{
		rely:         rely,
		userSvc:      svc,
		jwtSvc:       jwtSvc,
		blacklistSvc: blacklistSvc,
	}
}

// RegisterRoutes 注册路由
func (h *authHandler) RegisterRoutes(router *gin.RouterGroup) {
	router.POST("/login", h.LoginHandler)
	router.POST("/refresh-token", h.RefreshTokenHandler)
	router.POST("/logout", h.LogoutHandler)
	router.GET("/profile", h.ProfileHandler)
}

// LoginHandler
// @Summary 账号密码登录
// @Description 账号密码登录
// @Tags 认证管理
// @Accept application/json
// @Produce application/json
// @Param LoginRequest body LoginRequest true "参数信息"
// @Success 200 {object} LoginResponse
// @Failure 400 {object} response.Response
// @Router /v1/auth/login [post]
func (h *authHandler) LoginHandler(ctx *gin.Context) {
	var req LoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		validate.NewValidatorErrorHandler(h.rely.Trans).Handle(ctx, err)
		return
	}

	// 调用业务逻辑
	domain, err := h.userSvc.Login(ctx, req.Username, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, serviceSystem.ErrUserInvalidCredential):
			response.NewResponse().Error(ctx, http.StatusBadRequest, "用户名或密码错误", nil)
			return
		default:
			ctx.Set("internalError", fmt.Sprintf("用户登录异常 >>> %v", err.Error()))
			zap.S().Error("用户登录异常 >>> ", zap.Error(err))
			response.NewResponse().Error(ctx, http.StatusInternalServerError, "服务器异常", nil)
			return
		}
	}

	// 生成JWT令牌
	token, err := h.jwtSvc.GenerateToken(ctx, domain.Id, domain)
	if err != nil {
		ctx.Set("internalError", fmt.Sprintf("生成令牌异常 >>> %v", err.Error()))
		zap.S().Error("生成令牌异常 >>> ", zap.Error(err))
		response.NewResponse().Error(ctx, http.StatusInternalServerError, "服务器异常", nil)
		return
	}

	// 记录登录日志
	go request_utils.SaveLoginLog(ctx, domain, h.rely.Db.Careful)

	// 返回用户信息和令牌
	response.NewResponse().Success(ctx, "登录成功", LoginResponse{
		Token:  token,
		Expire: h.rely.Token.Expire * 3600,
	})
}

// RefreshTokenHandler
// @Summary 刷新令牌
// @Description 使用旧的JWT令牌获取新的令牌
// @Tags 认证管理
// @Accept application/json
// @Produce application/json
// @Param RefreshTokenRequest body RefreshTokenRequest true "刷新令牌参数"
// @Success 200 {object} LoginResponse
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /v1/auth/refresh-token [post]
func (h *authHandler) RefreshTokenHandler(ctx *gin.Context) {
	var req RefreshTokenRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		validate.NewValidatorErrorHandler(h.rely.Trans).Handle(ctx, err)
		return
	}

	// 解析旧令牌
	claims, err := h.jwtSvc.ParseToken(req.Token)
	if err != nil {
		switch {
		case errors.Is(err, jwt.ErrExpiredToken):
			response.NewResponse().Error(ctx, http.StatusUnauthorized, "令牌已过期，请重新登录", nil)
		case errors.Is(err, jwt.ErrInvalidToken):
			response.NewResponse().Error(ctx, http.StatusUnauthorized, "无效的令牌", nil)
		default:
			response.NewResponse().Error(ctx, http.StatusUnauthorized, "认证失败", nil)
		}
		return
	}

	// 检查token是否在黑名单中
	isBlacklisted, err := h.blacklistSvc.IsBlacklisted(ctx, req.Token)
	if err != nil {
		response.NewResponse().Error(ctx, http.StatusUnauthorized, "检查令牌状态失败", nil)
		return
	}
	if isBlacklisted {
		response.NewResponse().Error(ctx, http.StatusUnauthorized, "令牌已被加入黑名单", nil)
		return
	}

	// 获取用户信息
	domain, err := h.userSvc.GetById(ctx, claims.UserId)
	if err != nil {
		if errors.Is(err, serviceSystem.ErrUserNotFound) {
			response.NewResponse().Error(ctx, http.StatusUnauthorized, "用户不存在", nil)
			return
		}
		ctx.Set("internalError", fmt.Sprintf("刷新令牌获取用户信息异常 >>> %v", err.Error()))
		zap.S().Error("刷新令牌获取用户信息异常 >>> ", zap.Error(err))
		response.NewResponse().Error(ctx, http.StatusInternalServerError, "服务器异常", nil)
		return
	}

	// 生成JWT令牌
	token, err := h.jwtSvc.GenerateToken(ctx, domain.Id, domain)
	if err != nil {
		ctx.Set("internalError", fmt.Sprintf("生成令牌异常 >>> %v", err.Error()))
		zap.S().Error("生成令牌异常 >>> ", zap.Error(err))
		response.NewResponse().Error(ctx, http.StatusInternalServerError, "服务器异常", nil)
		return
	}

	// 返回用户信息和令牌
	response.NewResponse().Success(ctx, "登录成功", LoginResponse{
		Token:  token,
		User:   domain,
		Expire: h.rely.Token.Expire * 3600,
	})
}

// LogoutHandler
// @Summary 退出登录
// @Description 用户退出登录
// @Tags 认证管理
// @Accept application/json
// @Produce application/json
// @Security BearerAuth
// @Success 200 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /v1/auth/logout [post]
// @Security LoginToken
func (h *authHandler) LogoutHandler(ctx *gin.Context) {
	// 从上下文中获取登录信息
	claims, ok := ctx.MustGet("claims").(*jwt.Claims)
	if !ok {
		zap.S().Error("未找到用户认证信息 >>> ", zap.Error(errors.New(claims.UserId)))
		response.NewResponse().Error(ctx, http.StatusInternalServerError, "服务器异常", nil)
		return
	}

	// 获取Authorization头信息中的token
	authHeader := ctx.GetHeader("Authorization")
	if authHeader == "" {
		response.NewResponse().Error(ctx, http.StatusUnauthorized, "退出登录失败：未携带令牌", nil)
		return
	}

	// 提取token
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		response.NewResponse().Error(ctx, http.StatusUnauthorized, "退出登录失败：无效的令牌格式", nil)
		return
	}
	tokenStr := parts[1]

	// 解析token以获取过期时间
	jwtClaims, err := h.jwtSvc.ParseToken(tokenStr)
	if err != nil {
		zap.S().Error("解析token失败 >>> ", zap.Error(err))
		response.NewResponse().Error(ctx, http.StatusUnauthorized, "退出登录失败：无效的令牌", nil)
		return
	}

	// 计算token的剩余有效期
	expirationTime := jwtClaims.ExpiresAt.Time
	remainingTime := time.Until(expirationTime)
	if remainingTime <= 0 {
		// 如果token已过期，直接返回成功
		response.NewResponse().Success(ctx, "令牌已过期，登出成功", nil)
		return
	}

	// 将token加入黑名单
	if err := h.blacklistSvc.Add(ctx, tokenStr, claims.UserId, remainingTime); err != nil {
		ctx.Set("internalError", fmt.Sprintf("将token加入黑名单失败 >>> %v", err.Error()))
		zap.S().Error("将token加入黑名单失败 >>> ", zap.Error(err))
		response.NewResponse().Error(ctx, http.StatusInternalServerError, "服务器异常", nil)
		return
	}

	// 返回成功信息
	response.NewResponse().Success(ctx, "注销成功", nil)
}

// ProfileHandler
// @Summary 获取当前登录用户信息
// @Description 获取当前登录用户的详细信息
// @Tags 认证管理
// @Accept application/json
// @Produce application/json
// @Security BearerAuth
// @Success 200 {object} domainSystem.User
// @Failure 401 {object} response.Response
// @Router /v1/auth/profile [get]
// @Security LoginToken
func (h *authHandler) ProfileHandler(ctx *gin.Context) {
	// 从上下文中获取登录信息
	claims, ok := ctx.MustGet("claims").(*jwt.Claims)
	if !ok {
		zap.S().Error("未找到用户认证信息 >>> ", zap.Error(errors.New(claims.UserId)))
		response.NewResponse().Error(ctx, http.StatusInternalServerError, "服务器异常", nil)
		return
	}

	// 根据id获取用户信息
	domain, err := h.userSvc.GetById(ctx, claims.UserId)
	if err != nil {
		if errors.Is(err, serviceSystem.ErrUserNotFound) {
			response.NewResponse().Error(ctx, http.StatusBadRequest, "用户不存在", nil)
			return
		}
		ctx.Set("internalError", fmt.Sprintf("获取用户信息异常 >>> %v", err.Error()))
		zap.S().Error("获取用户信息异常 >>> ", zap.Error(err))
		response.NewResponse().Error(ctx, http.StatusInternalServerError, "服务器异常", nil)
		return
	}

	// 返回用户信息
	response.NewResponse().Success(ctx, "获取成功", domain)
}
