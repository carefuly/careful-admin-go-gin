/**
 * Description：
 * FileName：config.go
 * Author：CJiaの用心
 * Create：2025/11/23 00:42:11
 * Remark：
 */

package config

import (
	ut "github.com/go-playground/universal-translator"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type LocalConfig struct {
	Server      Server                    `yaml:"server"`
	Application Application               `yaml:"application"`
	Database    map[string]DatabaseDetail `yaml:"database" json:"database"`
	Cache       Cache                     `yaml:"cache"`
	Token       Token                     `yaml:"token"`
}

type RelyConfig struct {
	Logger *zap.Logger
	Db     Database
	Redis  redis.Cmdable
	Trans  ut.Translator
	Token  Token
}
