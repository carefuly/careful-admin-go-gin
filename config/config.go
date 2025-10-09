/**
 * Description：
 * FileName：config.go
 * Author：CJiaの用心
 * Create：2025/9/29 00:57:36
 * Remark：
 */

package config

import (
	"gorm.io/gorm"
	"time"
)

// Database 数据库配置
type Database struct {
	Careful     *gorm.DB
	TableConfig *gorm.DB
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

// Email 邮箱配置
type Email struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	From     string `yaml:"from"`
	To       string `yaml:"to"`
}

// AliYun 阿里云OSS配置
type AliYun struct {
	Endpoint        string `yaml:"endpoint"`
	BucketName      string `yaml:"bucketName"`
	OssPrefix       string `yaml:"ossPrefix"`
	AccessKeyID     string `yaml:"accessKeyID"`
	AccessKeySecret string `yaml:"accessKeySecret"`
}

// Config 总配置结构体
type Config struct {
	Server      Server      `yaml:"server"`
	Application Application `yaml:"application"`
	Database    Database    `yaml:"database"`
	Cache       Cache       `yaml:"cache"`
	Token       Token       `yaml:"token"`
	Email       Email       `yaml:"email"`
	AliYun      AliYun      `yaml:"aliYun"`
}
