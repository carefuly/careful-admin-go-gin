/**
 * Description：
 * FileName：menu_button.go
 * Author：CJiaの用心
 * Create：2025/12/05 14:02:29
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
	ErrMenuButtonNotExist = redis.Nil
	ErrMenuButtonKey      = "careful:system:menu_button:info"
)

type MenuButtonCache interface {
	Get(ctx context.Context, id string) (*domainSystem.MenuButton, error)
	Set(ctx context.Context, domain domainSystem.MenuButton) error
	Del(ctx context.Context, id string) error
	SetNotFound(ctx context.Context, id string) error // 防止缓存穿透
	Key(id string) string
}

type RedisMenuButtonCache struct {
	cmd        redis.Cmdable
	expiration time.Duration
}

func NewRedisMenuButtonCache(cmd redis.Cmdable) MenuButtonCache {
	return &RedisMenuButtonCache{
		cmd:        cmd,
		expiration: time.Minute * 15,
	}
}

func (c *RedisMenuButtonCache) Get(ctx context.Context, id string) (*domainSystem.MenuButton, error) {
	key := c.Key(id)

	data, err := c.cmd.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, ErrMenuButtonNotExist
		}
		return nil, err
	}

	// 检查是否是防穿透标记
	if data == "not_found" {
		return nil, nil
	}

	var doMain domainSystem.MenuButton
	err = json.Unmarshal([]byte(data), &doMain)
	return &doMain, err
}

func (c *RedisMenuButtonCache) Set(ctx context.Context, domain domainSystem.MenuButton) error {
	key := c.Key(domain.Id)
	data, err := json.Marshal(domain)
	if err != nil {
		return err
	}
	return c.cmd.Set(ctx, key, data, c.expiration).Err()
}

func (c *RedisMenuButtonCache) Del(ctx context.Context, id string) error {
	key := c.Key(id)
	return c.cmd.Del(ctx, key).Err()
}

func (c *RedisMenuButtonCache) SetNotFound(ctx context.Context, id string) error {
	key := c.Key(id)
	// 设置短暂的有效期防止缓存穿透
	return c.cmd.Set(ctx, key, "not_found", time.Minute).Err()
}

func (c *RedisMenuButtonCache) Key(id string) string {
	return fmt.Sprintf("%s:%s", ErrMenuButtonKey, id)
}
