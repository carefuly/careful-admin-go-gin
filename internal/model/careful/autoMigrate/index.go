/**
 * Description：
 * FileName：index.go
 * Author：CJiaの用心
 * Create：2025/10/9 16:03:19
 * Remark：
 */

package autoMigrate

import (
	"github.com/carefuly/careful-admin-go-gin/internal/model/careful/logger"
	"github.com/carefuly/careful-admin-go-gin/internal/model/careful/system"
	"gorm.io/gorm"
)

// AutoMigrate 迁移表
func AutoMigrate(db *gorm.DB) {
	// initSystem(db)
	// initLogger(db)
}

func initSystem(db *gorm.DB) {
	system.NewUser().AutoMigrate(db) // 用户表
	system.NewDept().AutoMigrate(db) // 部门表
}

func initLogger(db *gorm.DB) {
	logger.NewLoginLogger().AutoMigrate(db)   // 登录日志表
	logger.NewOperateLogger().AutoMigrate(db) // 操作日志表
	logger.NewCacheLogger().AutoMigrate(db)   // 缓存日志表
}
