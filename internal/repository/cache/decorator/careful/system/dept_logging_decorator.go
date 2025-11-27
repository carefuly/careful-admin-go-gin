/**
 * Description：
 * FileName：dept_logging_decorator.go
 * Author：CJiaの用心
 * Create：2025/11/26 02:29:33
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

type DeptCacheLoggingDecorator struct {
	cache  cacheSystem.DeptCache
	logger cacheRecord.CacheLogger
}

func NewDeptCacheLoggingDecorator(cache cacheSystem.DeptCache, logger cacheRecord.CacheLogger) DeptCacheLoggingDecorator {
	return DeptCacheLoggingDecorator{
		cache:  cache,
		logger: logger,
	}
}

// 通用日志记录函数
func (d *DeptCacheLoggingDecorator) logOperation(
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

	deptId := d.getStringFromContext(ctx, "dept_id")

	entity := &modelLogger.CacheLogger{
		CoreModels: models.CoreModels{
			Creator:    d.getStringFromContext(ctx, "user_id"),
			Modifier:   d.getStringFromContext(ctx, "user_id"),
			BelongDept: &deptId,
		},
		CacheHost:     request.Host,
		CacheIp:       d.getStringFromContext(ctx, "request_ip"),
		CacheUsername: d.getStringFromContext(ctx, "username"),
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
func (d *DeptCacheLoggingDecorator) getStringFromContext(ctx context.Context, key string) string {
	if val, ok := ctx.Value(key).(string); ok {
		return val
	}
	return ""
}

func (d *DeptCacheLoggingDecorator) Get(ctx context.Context, id string) (*domainSystem.Dept, error) {
	start := time.Now()
	result, err := d.cache.Get(ctx, id)

	// 特殊处理"未找到"情况
	var value interface{}
	if errors.Is(err, cacheSystem.ErrDeptNotExist) {
		value = "not_found"
	} else if result != nil {
		value = result
	}

	d.logOperation(ctx, id, value, err, start)
	return result, err
}

func (d *DeptCacheLoggingDecorator) Set(ctx context.Context, domain domainSystem.Dept) error {
	start := time.Now()
	err := d.cache.Set(ctx, domain)
	d.logOperation(ctx, domain.Id, domain, err, start)
	return err
}

func (d *DeptCacheLoggingDecorator) Del(ctx context.Context, id string) error {
	start := time.Now()
	err := d.cache.Del(ctx, id)
	d.logOperation(ctx, id, "not_found", err, start)
	return err
}

func (d *DeptCacheLoggingDecorator) SetNotFound(ctx context.Context, id string) error {
	start := time.Now()
	err := d.cache.SetNotFound(ctx, id)
	d.logOperation(ctx, id, "not_found", err, start)
	return err
}

func (d *DeptCacheLoggingDecorator) key(id string) string {
	return fmt.Sprintf("%s:%s", cacheSystem.ErrDeptKey, id)
}
