/**
 * Description：
 * FileName：blacklist.go
 * Author：CJiaの用心
 * Create：2025/10/10 10:35:25
 * Remark：
 */

package jwt

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"time"
)

const (
	// TokenBlacklistPrefix Redis中存储已登出token的前缀
	TokenBlacklistPrefix = "token:blacklist:"
	// UserTokensPrefix Redis中存储用户关联token的前缀
	UserTokensPrefix = "user:tokens:"
)

// TokenBlacklist JWT Token黑名单实现
type TokenBlacklist struct {
	rdb redis.Cmdable
}

// NewTokenBlacklist 创建一个Token黑名单实例
func NewTokenBlacklist(rdb redis.Cmdable) *TokenBlacklist {
	return &TokenBlacklist{
		rdb: rdb,
	}
}

// Add 将Token加入黑名单
// tokenStr: JWT token字符串
// expiresIn: token的剩余有效期（秒）
func (b *TokenBlacklist) Add(ctx context.Context, tokenStr string, userID string, expiresIn time.Duration) error {
	// 使用Redis SET命令将token加入黑名单
	key := fmt.Sprintf("%s%s", TokenBlacklistPrefix, tokenStr)
	err := b.rdb.Set(ctx, key, "1", expiresIn).Err()
	if err != nil {
		return err
	}

	// 同时记录用户与token的关联(用于用户登出所有设备)
	userTokensKey := fmt.Sprintf("%s%s", UserTokensPrefix, userID)
	err = b.rdb.SAdd(ctx, userTokensKey, tokenStr).Err()
	if err != nil {
		return err
	}

	// 设置用户token集合的过期时间(略长于token过期时间)
	return b.rdb.Expire(ctx, userTokensKey, expiresIn+time.Hour).Err()
}

// IsBlacklisted 检查Token是否在黑名单中
func (b *TokenBlacklist) IsBlacklisted(ctx context.Context, tokenStr string) (bool, error) {
	key := fmt.Sprintf("%s%s", TokenBlacklistPrefix, tokenStr)
	result, err := b.rdb.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return result > 0, nil
}

// RemoveFromBlacklist 从黑名单中移除Token(用于特殊情况下撤销黑名单)
func (b *TokenBlacklist) RemoveFromBlacklist(ctx context.Context, tokenStr string, userID string) error {
	// 从黑名单中移除
	key := fmt.Sprintf("%s%s", TokenBlacklistPrefix, tokenStr)
	err := b.rdb.Del(ctx, key).Err()
	if err != nil {
		return err
	}

	// 从用户token集合中移除
	userTokensKey := fmt.Sprintf("%s%s", UserTokensPrefix, userID)
	return b.rdb.SRem(ctx, userTokensKey, tokenStr).Err()
}

// LogoutAllUserTokens 登出用户的所有token
func (b *TokenBlacklist) LogoutAllUserTokens(ctx context.Context, userID string) error {
	userTokensKey := fmt.Sprintf("%s%s", UserTokensPrefix, userID)

	// 获取用户所有token
	tokens, err := b.rdb.SMembers(ctx, userTokensKey).Result()
	if err != nil {
		return err
	}

	// 将所有token加入黑名单(使用管道提高性能)
	pipe := b.rdb.Pipeline()
	for _, token := range tokens {
		key := fmt.Sprintf("%s%s", TokenBlacklistPrefix, token)
		pipe.Set(ctx, key, "1", 24*time.Hour) // 设置为24小时
	}

	// 删除用户token集合
	pipe.Del(ctx, userTokensKey)

	_, err = pipe.Exec(ctx)
	return err
}
