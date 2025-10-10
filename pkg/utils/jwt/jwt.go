/**
 * Description：
 * FileName：jwt.go
 * Author：CJiaの用心
 * Create：2025/10/10 10:31:54
 * Remark：
 */

package jwt

import (
	"errors"
	domainSystem "github.com/carefuly/careful-admin-go-gin/internal/domain/careful/system"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"time"
)

var (
	ErrInvalidToken  = errors.New("无效令牌")
	ErrExpiredToken  = errors.New("令牌已过期")
	ErrTokenNotFound = errors.New("未找到令牌")
)

type Claims struct {
	jwt.RegisteredClaims
	UserId    string                 `json:"userId"`    // 用户ID
	Username  string                 `json:"username"`  // 用户名
	DeptId    string                 `json:"DeptId"`    // 部门ID
	UserAgent string                 `json:"userAgent"` // 用户代理
	UserInfo  map[string]interface{} `json:"userInfo"`  // 用户信息(精简版)
}

// TokenConfig JWT配置
type TokenConfig struct {
	Secret      string        `json:"secret"`      // 密钥
	ExpireHours int           `json:"expireHours"` // 过期时间(小时)
	Issuer      string        `json:"issuer"`      // 签发者
	Audience    []string      `json:"audience"`    // 接收方
	MaxRefresh  time.Duration `json:"maxRefresh"`  // 最大刷新时间
}

// TokenService JWT服务接口
type TokenService interface {
	GenerateToken(ctx *gin.Context, userId string, userInfo domainSystem.User) (string, error)
	ParseToken(tokenString string) (*Claims, error)
}

// DefaultJWTService 默认JWT服务实现
type DefaultJWTService struct {
	config TokenConfig
}

// NewJWTService 创建JWT服务实例
func NewJWTService(config TokenConfig) *DefaultJWTService {
	return &DefaultJWTService{config: config}
}

// GenerateToken 生成新的 JWT 令牌
func (s *DefaultJWTService) GenerateToken(ctx *gin.Context, userId string, userInfo domainSystem.User) (string, error) {
	// 创建精简版用户信息
	essentialUserInfo := map[string]interface{}{
		"id":       userInfo.Id,
		"username": userInfo.Username,
		"deptId":   userInfo.DeptId,
		// 只包含必要信息，避免令牌过大
	}

	// 设置声明
	now := time.Now()
	expiresAt := now.Add(time.Hour * time.Duration(s.config.ExpireHours))

	claims := Claims{
		UserId:    userId,
		Username:  userInfo.Username,
		DeptId:    userInfo.DeptId,
		UserAgent: ctx.GetHeader("User-Agent"),
		UserInfo:  essentialUserInfo,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    s.config.Issuer,
			Audience:  s.config.Audience,
		},
	}

	// 创建令牌
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)

	// 使用密钥签名
	return token.SignedString([]byte(s.config.Secret))
}

// ParseToken 解析 JWT 令牌并返回声明
func (s *DefaultJWTService) ParseToken(tokenString string) (*Claims, error) {
	if tokenString == "" {
		return nil, ErrTokenNotFound
	}

	// 解析令牌
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// 验证签名方法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return []byte(s.config.Secret), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	// 验证令牌
	if !token.Valid {
		return nil, ErrInvalidToken
	}

	// 提取声明
	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, ErrInvalidToken
	}

	return claims, nil
}
