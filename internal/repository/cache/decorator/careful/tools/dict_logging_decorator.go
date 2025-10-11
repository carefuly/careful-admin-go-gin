/**
 * Description：
 * FileName：dict_logging_decorator.go
 * Author：CJiaの用心
 * Create：2025/10/11 11:58:05
 * Remark：
 */

package tools

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	domainTools "github.com/carefuly/careful-admin-go-gin/internal/domain/careful/tools"
	modelLogger "github.com/carefuly/careful-admin-go-gin/internal/model/careful/logger"
	cacheTools "github.com/carefuly/careful-admin-go-gin/internal/repository/cache/careful/tools"
	cacheRecord "github.com/carefuly/careful-admin-go-gin/internal/repository/cache/decorator/record"
	"github.com/carefuly/careful-admin-go-gin/pkg/models"
	"net/http"
	"time"
)

type DictCacheLoggingDecorator struct {
	cache  cacheTools.DictCache
	logger cacheRecord.CacheLogger
}

func NewDictCacheLoggingDecorator(cache cacheTools.DictCache, logger cacheRecord.CacheLogger) DictCacheLoggingDecorator {
	return DictCacheLoggingDecorator{
		cache:  cache,
		logger: logger,
	}
}

// 通用日志记录函数
func (d *DictCacheLoggingDecorator) logOperation(
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

func (d *DictCacheLoggingDecorator) Get(ctx context.Context, id string) (*domainTools.Dict, error) {
	start := time.Now()
	result, err := d.cache.Get(ctx, id)

	// 特殊处理"未找到"情况
	var value interface{}
	if errors.Is(err, cacheTools.ErrDictNotExist) {
		value = "not_found"
	} else if result != nil {
		value = result
	}

	d.logOperation(ctx, id, value, err, start)
	return result, err
}

func (d *DictCacheLoggingDecorator) Set(ctx context.Context, domain domainTools.Dict) error {
	start := time.Now()
	err := d.cache.Set(ctx, domain)
	d.logOperation(ctx, domain.Id, domain, err, start)
	return err
}

func (d *DictCacheLoggingDecorator) Del(ctx context.Context, id string) error {
	start := time.Now()
	err := d.cache.Del(ctx, id)
	d.logOperation(ctx, id, "not_found", err, start)
	return err
}

func (d *DictCacheLoggingDecorator) SetNotFound(ctx context.Context, id string) error {
	start := time.Now()
	err := d.cache.SetNotFound(ctx, id)
	d.logOperation(ctx, id, "not_found", err, start)
	return err
}

func (d *DictCacheLoggingDecorator) key(id string) string {
	return fmt.Sprintf("%s:%s", cacheTools.ErrDictKey, id)
}
