/**
 * Description：
 * FileName：cache_logger.go
 * Author：CJiaの用心
 * Create：2025/10/10 10:58:37
 * Remark：
 */

package record

import (
	"context"
	modelLogger "github.com/carefuly/careful-admin-go-gin/internal/model/careful/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"time"
)

// CacheLogger 缓存日志记录器
type CacheLogger struct {
	db *gorm.DB
}

func NewCacheLogger(db *gorm.DB) CacheLogger {
	return CacheLogger{db: db}
}

// Log 异步记录缓存操作日志
func (l *CacheLogger) Log(ctx context.Context, entity *modelLogger.CacheLogger) {
	// 使用goroutine异步记录日志，不影响主流程
	go func() {
		// 设置上下文超时防止日志写入阻塞
		logCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		// 保存原始日志级别
		originalLogger := l.db.Logger

		// 临时禁用SQL日志
		l.db.Logger = logger.Default.LogMode(logger.Silent)

		// 确保恢复原始日志设置
		defer func() {
			l.db.Logger = originalLogger
		}()

		// 使用数据库连接池
		tx := l.db.WithContext(logCtx).Begin()
		if tx.Error != nil {
			logError(entity, tx.Error)
			return
		}

		if err := tx.Create(entity).Error; err != nil {
			logError(entity, err)
			tx.Rollback()
			return
		}

		if err := tx.Commit().Error; err != nil {
			logError(entity, err)
			tx.Rollback()
		}
	}()
}

// 记录错误日志
func logError(entry *modelLogger.CacheLogger, err error) {
	zap.L().Error("缓存日志记录失败",
		zap.String("key", entry.CacheKey),
		zap.String("path", entry.CachePath),
		zap.String("method", entry.CacheMethod),
		zap.Error(err),
	)
}
