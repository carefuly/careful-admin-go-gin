/**
 * Description：
 * FileName：index.go
 * Author：CJiaの用心
 * Create：2025/11/24 16:07:59
 * Remark：
 */

package auto_migrate

import (
	"github.com/carefuly/careful-admin-go-gin/internal/model/careful/logger"
	"github.com/carefuly/careful-admin-go-gin/internal/model/careful/system"
	"github.com/carefuly/careful-admin-go-gin/internal/model/careful/tools"
	"gorm.io/gorm"
)

// AutoMigrate 迁移表
func AutoMigrate(db *gorm.DB) {
	initSystem(db)
	initTools(db)
	initLogger(db)
}

func initSystem(db *gorm.DB) {
	system.NewDept().AutoMigrate(db)       // 部门表
	system.NewMenu().AutoMigrate(db)       // 菜单表
	system.NewMenuButton().AutoMigrate(db) // 菜单按钮表
	system.NewPost().AutoMigrate(db)       // 岗位表
	system.NewUser().AutoMigrate(db)       // 用户表
}

func initTools(db *gorm.DB) {
	tools.NewDict().AutoMigrate(db)     // 字典表
	tools.NewDictType().AutoMigrate(db) // 字典项表
}

func initLogger(db *gorm.DB) {
	logger.NewLoginLogger().AutoMigrate(db)   // 登录日志表
	logger.NewOperateLogger().AutoMigrate(db) // 操作日志表
	logger.NewCacheLogger().AutoMigrate(db)   // 缓存日志表
}
