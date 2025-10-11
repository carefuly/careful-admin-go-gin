/**
 * Description：
 * FileName：dict.go
 * Author：CJiaの用心
 * Create：2025/10/11 11:54:48
 * Remark：
 */

package tools

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	domainTools "github.com/carefuly/careful-admin-go-gin/internal/domain/careful/tools"
	"github.com/redis/go-redis/v9"
	"time"
)

var (
	ErrDictNotExist = redis.Nil
	ErrDictKey      = "careful:tools:dict:info"
)

type DictCache interface {
	Get(ctx context.Context, id string) (*domainTools.Dict, error)
	Set(ctx context.Context, domain domainTools.Dict) error
	Del(ctx context.Context, id string) error
	SetNotFound(ctx context.Context, id string) error // 防止缓存穿透
	Key(id string) string
}

type RedisDictCache struct {
	cmd        redis.Cmdable
	expiration time.Duration
}

func NewRedisDictCache(cmd redis.Cmdable) DictCache {
	return &RedisDictCache{
		cmd:        cmd,
		expiration: time.Minute * 15,
	}
}

func (c *RedisDictCache) Get(ctx context.Context, id string) (*domainTools.Dict, error) {
	key := c.Key(id)

	data, err := c.cmd.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, ErrDictNotExist
		}
		return nil, err
	}

	// 检查是否是防穿透标记
	if data == "not_found" {
		return nil, nil
	}

	var doMain domainTools.Dict
	err = json.Unmarshal([]byte(data), &doMain)
	return &doMain, err
}

func (c *RedisDictCache) Set(ctx context.Context, domain domainTools.Dict) error {
	key := c.Key(domain.Id)
	data, err := json.Marshal(domain)
	if err != nil {
		return err
	}
	return c.cmd.Set(ctx, key, data, c.expiration).Err()
}

func (c *RedisDictCache) Del(ctx context.Context, id string) error {
	key := c.Key(id)
	return c.cmd.Del(ctx, key).Err()
}

func (c *RedisDictCache) SetNotFound(ctx context.Context, id string) error {
	key := c.Key(id)
	// 设置短暂的有效期防止缓存穿透
	return c.cmd.Set(ctx, key, "not_found", time.Minute).Err()
}

func (c *RedisDictCache) Key(id string) string {
	return fmt.Sprintf("%s:%s", ErrDictKey, id)
}
