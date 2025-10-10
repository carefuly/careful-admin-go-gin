/**
 * Description：
 * FileName：user_logging_decorator.go
 * Author：CJiaの用心
 * Create：2025/10/10 10:59:46
 * Remark：
 */

package system

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	domainSystem "github.com/carefuly/careful-admin-go-gin/internal/domain/careful/system"
	modelLogger "github.com/carefuly/careful-admin-go-gin/internal/model/careful/logger"
	cacheSystem "github.com/carefuly/careful-admin-go-gin/internal/repository/cache/careful/system"
	cacheRecord "github.com/carefuly/careful-admin-go-gin/internal/repository/cache/decorator/record"
	"github.com/carefuly/careful-admin-go-gin/pkg/models"
	"net/http"
	"time"
)

type UserCacheLoggingDecorator struct {
	cache  cacheSystem.UserCache
	logger cacheRecord.CacheLogger
}

func NewUserCacheLoggingDecorator(cache cacheSystem.UserCache, logger cacheRecord.CacheLogger) UserCacheLoggingDecorator {
	return UserCacheLoggingDecorator{
		cache:  cache,
		logger: logger,
	}
}

// 通用日志记录函数
func (d *UserCacheLoggingDecorator) logOperation(
	ctx context.Context,
	key string,
	value interface{},
	err error,
	start time.Time,
) {
	request, ok := ctx.Value("request").(*http.Request)
	if !ok {
		return // 没有请求上下文，不记录日志
	}

	entity := &modelLogger.CacheLogger{
		CoreModels: models.CoreModels{
			Creator:    getStringFromContext(ctx, "userId"),
			Modifier:   getStringFromContext(ctx, "userId"),
			BelongDept: getStringFromContext(ctx, "deptId"),
		},
		CacheHost:     request.Host,
		CacheIp:       getStringFromContext(ctx, "requestIp"),
		CacheUsername: getStringFromContext(ctx, "username"),
		CacheMethod:   request.Method,
		CachePath:     request.URL.Path,
		CacheKey:      d.key(key),
		CacheTime:     time.Since(start).String(),
	}

	if err != nil {
		entity.CacheError = err.Error()
	}

	// 处理值
	if value != nil {
		if data, err := json.Marshal(value); err == nil {
			entity.CacheValue = string(data)
		}
	}

	// 异步记录日志
	go d.logger.Log(ctx, entity)
}

// 从上下文中安全获取字符串值
func getStringFromContext(ctx context.Context, key string) string {
	if val, ok := ctx.Value(key).(string); ok {
		return val
	}
	return ""
}

func (d *UserCacheLoggingDecorator) Get(ctx context.Context, id string) (*domainSystem.User, error) {
	start := time.Now()
	result, err := d.cache.Get(ctx, id)

	// 特殊处理"未找到"情况
	var value interface{}
	if errors.Is(err, cacheSystem.ErrUserNotExist) {
		value = "not_found"
	} else if result != nil {
		value = result
	}

	d.logOperation(ctx, id, value, err, start)
	return result, err
}

func (d *UserCacheLoggingDecorator) Set(ctx context.Context, domain domainSystem.User) error {
	start := time.Now()
	err := d.cache.Set(ctx, domain)
	d.logOperation(ctx, domain.Id, domain, err, start)
	return err
}

func (d *UserCacheLoggingDecorator) Del(ctx context.Context, id string) error {
	start := time.Now()
	err := d.cache.Del(ctx, id)
	d.logOperation(ctx, id, "not_found", err, start)
	return err
}

func (d *UserCacheLoggingDecorator) SetNotFound(ctx context.Context, id string) error {
	start := time.Now()
	err := d.cache.SetNotFound(ctx, id)
	d.logOperation(ctx, id, "not_found", err, start)
	return err
}

func (d *UserCacheLoggingDecorator) key(id string) string {
	return fmt.Sprintf("%s:%s", cacheSystem.ErrUserKey, id)
}
