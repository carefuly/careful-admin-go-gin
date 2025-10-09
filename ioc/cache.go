/**
 * Description：
 * FileName：cache.go
 * Author：CJiaの用心
 * Create：2025/10/8 14:23:57
 * Remark：
 */

package ioc

import (
	"context"
	"fmt"
	"github.com/carefuly/careful-admin-go-gin/config"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"time"
)

// CacheOptions 扩展配置参数
type CacheOptions struct {
	PoolSize       int           // 连接池大小 (默认: 10)
	MinIdleConn    int           // 最小空闲连接数 (默认: 5)
	MaxRetries     int           // 最大重试次数 (默认: 3)
	ConnectTimeout time.Duration // 连接超时 (默认: 5秒)
	ReadTimeout    time.Duration // 读取超时 (默认: 3秒)
}

// DefaultCacheOptions 默认配置
func DefaultCacheOptions() CacheOptions {
	return CacheOptions{
		PoolSize:       10,
		MinIdleConn:    5,
		MaxRetries:     3,
		ConnectTimeout: 5 * time.Second,
		ReadTimeout:    3 * time.Second,
	}
}

func InitCache(cfg config.Cache, opts ...CacheOptions) redis.Cmdable {
	// 合并配置
	opt := DefaultCacheOptions()
	if len(opts) > 0 {
		opt = opts[0]
	}

	// 初始化客户端
	client := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password:     cfg.Password,
		DB:           cfg.DB,
		PoolSize:     opt.PoolSize,
		MinIdleConns: opt.MinIdleConn,
		MaxRetries:   opt.MaxRetries,
		DialTimeout:  opt.ConnectTimeout,
		ReadTimeout:  opt.ReadTimeout,
	})

	// 健康检查
	ctx, cancel := context.WithTimeout(context.Background(), opt.ConnectTimeout)
	defer cancel()

	if _, err := client.Ping(ctx).Result(); err != nil {
		zap.L().Fatal("Redis连接失败",
			zap.String("addr", cfg.Host),
			zap.Int("port", cfg.Port),
			zap.Error(err))
		return nil
	}

	// 注册连接关闭钩子
	registerCleanupHook(client)

	zap.L().Info("Redis连接成功",
		zap.String("addr", fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)),
		zap.Int("db", cfg.DB))

	return client
}

// 注册清理钩子（确保程序退出时关闭连接）
func registerCleanupHook(client *redis.Client) {
	// 当程序退出时关闭连接
	go func() {
		<-context.Background().Done()
		if err := client.Close(); err != nil {
			zap.L().Error("关闭Redis连接失败", zap.Error(err))
		} else {
			zap.L().Info("Redis连接已关闭")
		}
	}()
}
