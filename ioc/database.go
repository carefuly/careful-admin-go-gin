/**
 * Description：
 * FileName：database.go
 * Author：CJiaの用心
 * Create：2025/10/8 14:16:29
 * Remark：
 */

package ioc

import (
	"fmt"
	"github.com/carefuly/careful-admin-go-gin/config"
	carefulAutoMigrate "github.com/carefuly/careful-admin-go-gin/internal/model/careful/autoMigrate"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"log"
	"os"
	"time"
)

type DbPool struct {
	CarefulDB *gorm.DB
	TableDB   *gorm.DB
}

func NewDbPool(databases map[string]config.DatabaseDetail) *DbPool {
	pool := &DbPool{}
	pool.initDatabases(databases) // 整合初始化逻辑
	return pool
}

// 私有方法避免外部误调用
func (p *DbPool) initDatabases(databases map[string]config.DatabaseDetail) {
	for name, dbConfig := range databases {
		db, err := initDatabase(dbConfig)
		if err != nil {
			zap.L().Fatal("数据库初始化失败",
				zap.String("name", name),
				zap.Error(err))
			continue
		}

		switch name {
		case "careful":
			p.CarefulDB = db
			configureConnectionPool(p.CarefulDB, dbConfig)
		case "table":
			p.TableDB = db
			configureConnectionPool(p.TableDB, dbConfig)
		default:
			zap.L().Warn("未知的数据库配置", zap.String("name", name))
		}
	}
}

func initDatabase(database config.DatabaseDetail) (*gorm.DB, error) {
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logger.Info, // 生产环境建议Warn级别
			Colorful:                  true,
			IgnoreRecordNotFoundError: true, // 忽略记录未找到错误
		},
	)

	switch database.Type {
	case "mysql":
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
			database.Username, database.Password, database.Host,
			database.Port, database.DBName, database.Charset)

		gormConfig := &gorm.Config{
			NamingStrategy: schema.NamingStrategy{
				// TablePrefix: database.Prefix, // 表前缀
			},
			Logger: newLogger,
		}

		db, err := gorm.Open(mysql.Open(dsn), gormConfig)
		if err != nil {
			return nil, fmt.Errorf("数据库连接失败: %w", err)
		}

		// 实际迁移操作应该在此处调用
		// 迁移系统表
		carefulAutoMigrate.AutoMigrate(db)

		return db, nil
	default:
		return nil, fmt.Errorf("不支持的数据库类型: %s", database.Type)
	}
}

// 配置连接池参数
func configureConnectionPool(db *gorm.DB, cfg config.DatabaseDetail) {
	sqlDB, err := db.DB()
	if err != nil {
		zap.L().Error("获取数据库连接池失败", zap.Error(err))
		return
	}

	// 设置连接池参数 (使用配置值或默认值)
	if cfg.MaxIdleConn > 0 {
		sqlDB.SetMaxIdleConns(cfg.MaxIdleConn)
	} else {
		sqlDB.SetMaxIdleConns(10) // 默认值
	}

	if cfg.MaxOpenConn > 0 {
		sqlDB.SetMaxOpenConns(cfg.MaxOpenConn)
	} else {
		sqlDB.SetMaxOpenConns(100) // 默认值
	}

	if cfg.ConnMaxLifetime != nil {
		sqlDB.SetConnMaxLifetime(*cfg.ConnMaxLifetime)
	} else {
		sqlDB.SetConnMaxLifetime(30 * time.Minute) // 默认值
	}
}
