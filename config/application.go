/**
 * Description：
 * FileName：application.go
 * Author：CJiaの用心
 * Create：2025/11/23 00:24:11
 * Remark：
 */

package config

import (
	"gorm.io/gorm"
	"time"
)

// Server 服务
type Server struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

// Application 应用
type Application struct {
	Name        string `yaml:"name"`
	Version     string `yaml:"version"`
	Environment string `yaml:"environment"`
	Debug       bool   `yaml:"debug"`
}

// Database 数据库配置
type Database struct {
	Careful     *gorm.DB
}

// DatabaseDetail 数据库详细配置
type DatabaseDetail struct {
	Type            string         `yaml:"type"`
	Host            string         `yaml:"host"`
	Port            int            `yaml:"port"`
	Username        string         `yaml:"username"`
	Password        string         `yaml:"password"`
	DBName          string         `yaml:"dbname"`
	Charset         string         `yaml:"charset"`
	Collation       string         `yaml:"collation"`
	Prefix          string         `yaml:"prefix"`
	MaxIdleConn     int            `yaml:"maxIdleConn"`
	MaxOpenConn     int            `yaml:"maxOpenConn"`
	ConnMaxLifetime *time.Duration `yaml:"connMaxLifetime"`
}

// Cache 缓存配置 (Redis)
type Cache struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

// Token Token配置
type Token struct {
	Secret string `yaml:"secret"`
	Expire int    `yaml:"expire"` // 建议明确单位，如 ExpireHour
}
