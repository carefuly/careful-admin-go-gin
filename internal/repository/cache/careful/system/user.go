/**
 * Description：
 * FileName：user.go
 * Author：CJiaの用心
 * Create：2025/10/10 10:47:48
 * Remark：
 */

package system

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	domainSystem "github.com/carefuly/careful-admin-go-gin/internal/domain/careful/system"
	"github.com/redis/go-redis/v9"
	"time"
)

var (
	ErrUserNotExist = redis.Nil
	ErrUserKey      = "careful:system:user:info"
)

type UserCache interface {
	Get(ctx context.Context, id string) (*domainSystem.User, error)
	Set(ctx context.Context, domain domainSystem.User) error
	Del(ctx context.Context, id string) error
	SetNotFound(ctx context.Context, id string) error // 防止缓存穿透
	Key(id string) string
}

type RedisUserCache struct {
	cmd        redis.Cmdable
	expiration time.Duration
}

func NewRedisUserCache(cmd redis.Cmdable) UserCache {
	return &RedisUserCache{
		cmd:        cmd,
		expiration: time.Minute * 15,
	}
}

func (c *RedisUserCache) Get(ctx context.Context, id string) (*domainSystem.User, error) {
	key := c.Key(id)

	data, err := c.cmd.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, ErrUserNotExist
		}
		return nil, err
	}

	// 检查是否是防穿透标记
	if data == "not_found" {
		return nil, nil
	}

	var doMain domainSystem.User
	err = json.Unmarshal([]byte(data), &doMain)
	return &doMain, err
}

func (c *RedisUserCache) Set(ctx context.Context, domain domainSystem.User) error {
	key := c.Key(domain.Id)
	data, err := json.Marshal(domain)
	if err != nil {
		return err
	}
	return c.cmd.Set(ctx, key, data, c.expiration).Err()
}

func (c *RedisUserCache) Del(ctx context.Context, id string) error {
	key := c.Key(id)
	return c.cmd.Del(ctx, key).Err()
}

func (c *RedisUserCache) SetNotFound(ctx context.Context, id string) error {
	key := c.Key(id)
	// 设置短暂的有效期防止缓存穿透
	return c.cmd.Set(ctx, key, "not_found", time.Minute).Err()
}

func (c *RedisUserCache) Key(id string) string {
	return fmt.Sprintf("%s:%s", ErrUserKey, id)
}
