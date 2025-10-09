/**
 * Description：
 * FileName：types.go
 * Author：CJiaの用心
 * Create：2025/9/29 01:24:46
 * Remark：
 */

package config

import (
	ut "github.com/go-playground/universal-translator"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type LocalConfig struct {
	Server      Server      `yaml:"server"`
	Application Application `yaml:"application"`
	NaCos       NaCos       `yaml:"nacos"`
}

type RemoteConfig struct {
	DatabaseConfig map[string]DatabaseDetail `yaml:"database" json:"database"`
	CacheConfig    Cache                     `yaml:"cache" json:"cache"`
	TokenConfig    Token                     `yaml:"token" json:"token"`
}

type RelyConfig struct {
	Logger *zap.Logger
	Db     Database
	Redis  redis.Cmdable
	Trans  ut.Translator
	Token  Token
}
