/**
 * Description：
 * FileName：menu.go
 * Author：CJiaの用心
 * Create：2025/12/05 14:21:14
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
	ErrMenuNotExist = redis.Nil
	ErrMenuKey      = "careful:system:menu:info"
)

type MenuCache interface {
	Get(ctx context.Context, id string) (*domainSystem.Menu, error)
	Set(ctx context.Context, domain domainSystem.Menu) error
	Del(ctx context.Context, id string) error
	SetNotFound(ctx context.Context, id string) error // 防止缓存穿透
	Key(id string) string
}

type RedisMenuCache struct {
	cmd        redis.Cmdable
	expiration time.Duration
}

func NewRedisMenuCache(cmd redis.Cmdable) MenuCache {
	return &RedisMenuCache{
		cmd:        cmd,
		expiration: time.Minute * 15,
	}
}

func (c *RedisMenuCache) Get(ctx context.Context, id string) (*domainSystem.Menu, error) {
	key := c.Key(id)

	data, err := c.cmd.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, ErrMenuNotExist
		}
		return nil, err
	}

	// 检查是否是防穿透标记
	if data == "not_found" {
		return nil, nil
	}

	var doMain domainSystem.Menu
	err = json.Unmarshal([]byte(data), &doMain)
	return &doMain, err
}

func (c *RedisMenuCache) Set(ctx context.Context, domain domainSystem.Menu) error {
	key := c.Key(domain.Id)
	data, err := json.Marshal(domain)
	if err != nil {
		return err
	}
	return c.cmd.Set(ctx, key, data, c.expiration).Err()
}

func (c *RedisMenuCache) Del(ctx context.Context, id string) error {
	key := c.Key(id)
	return c.cmd.Del(ctx, key).Err()
}

func (c *RedisMenuCache) SetNotFound(ctx context.Context, id string) error {
	key := c.Key(id)
	// 设置短暂的有效期防止缓存穿透
	return c.cmd.Set(ctx, key, "not_found", time.Minute).Err()
}

func (c *RedisMenuCache) Key(id string) string {
	return fmt.Sprintf("%s:%s", ErrMenuKey, id)
}
