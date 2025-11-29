/**
 * Description：
 * FileName：post.go
 * Author：CJiaの用心
 * Create：2025/11/29 01:30:30
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
	ErrPostNotExist = redis.Nil
	ErrPostKey      = "careful:system:post:info"
)

type PostCache interface {
	Get(ctx context.Context, id string) (*domainSystem.Post, error)
	Set(ctx context.Context, domain domainSystem.Post) error
	Del(ctx context.Context, id string) error
	SetNotFound(ctx context.Context, id string) error // 防止缓存穿透
	Key(id string) string
}

type RedisPostCache struct {
	cmd        redis.Cmdable
	expiration time.Duration
}

func NewRedisPostCache(cmd redis.Cmdable) PostCache {
	return &RedisPostCache{
		cmd:        cmd,
		expiration: time.Minute * 15,
	}
}

func (c *RedisPostCache) Get(ctx context.Context, id string) (*domainSystem.Post, error) {
	key := c.Key(id)

	data, err := c.cmd.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, ErrPostNotExist
		}
		return nil, err
	}

	// 检查是否是防穿透标记
	if data == "not_found" {
		return nil, nil
	}

	var doMain domainSystem.Post
	err = json.Unmarshal([]byte(data), &doMain)
	return &doMain, err
}

func (c *RedisPostCache) Set(ctx context.Context, domain domainSystem.Post) error {
	key := c.Key(domain.Id)
	data, err := json.Marshal(domain)
	if err != nil {
		return err
	}
	return c.cmd.Set(ctx, key, data, c.expiration).Err()
}

func (c *RedisPostCache) Del(ctx context.Context, id string) error {
	key := c.Key(id)
	return c.cmd.Del(ctx, key).Err()
}

func (c *RedisPostCache) SetNotFound(ctx context.Context, id string) error {
	key := c.Key(id)
	// 设置短暂的有效期防止缓存穿透
	return c.cmd.Set(ctx, key, "not_found", time.Minute).Err()
}

func (c *RedisPostCache) Key(id string) string {
	return fmt.Sprintf("%s:%s", ErrPostKey, id)
}
