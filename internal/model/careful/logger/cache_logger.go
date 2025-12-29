/**
 * Description：
 * FileName：cache_logger.go
 * Author：CJiaの用心
 * Create：2025/11/24 17:01:38
 * Remark：
 */

package logger

import (
	"github.com/carefuly/careful-admin-go-gin/pkg/models"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// CacheLogger 缓存日志表
type CacheLogger struct {
	models.CoreModels

	Status        bool   `gorm:"type:boolean;index;column:status;comment:状态【true-启用 false-停用】" json:"status"` // 状态
	CacheHost     string `gorm:"size:64;column:cache_host;comment:当前主机地址" json:"cache_host"`                  // 当前主机地址
	CacheIp       string `gorm:"size:64;column:cache_ip;comment:缓存者IP" json:"cache_ip"`                       // 缓存者IP
	CacheUsername string `gorm:"size:32;index;column:cache_username;comment:缓存用户名" json:"cache_username"`     // 缓存用户名
	CacheMethod   string `gorm:"size:16;index;column:cache_method;comment:缓存请求方式" json:"cache_method"`        // 缓存请求方式
	CachePath     string `gorm:"size:256;column:cache_path;comment:缓存请求地址" json:"cache_path"`                 // 缓存请求地址
	CacheTime     string `gorm:"size:256;column:cache_time;comment:缓存记录时间" json:"cache_time"`                 // 缓存记录时间
	CacheKey      string `gorm:"size:256;column:cache_key;comment:缓存key键" json:"cache_key"`                   // 缓存请求地址
	CacheValue    string `gorm:"type:mediumtext;column:cache_value;comment:缓存value值" json:"cache_value"`      // 缓存value值
	CacheError    string `gorm:"size:256;column:cache_error;comment:缓存Error错误" json:"cache_error"`            // 缓存Error错误
}

func NewCacheLogger() *CacheLogger {
	return &CacheLogger{}
}

func (l *CacheLogger) TableName() string {
	return "careful_logger_cache_log"
}

func (l *CacheLogger) AutoMigrate(db *gorm.DB) {
	err := db.Set("gorm:table_options", "ENGINE=InnoDB,COMMENT='缓存日志表'").AutoMigrate(&CacheLogger{})
	if err != nil {
		zap.L().Error("CacheLogger表模型迁移失败", zap.Error(err))
	}
}
