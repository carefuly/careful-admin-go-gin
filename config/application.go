/**
 * Description：
 * FileName：application.go
 * Author：CJiaの用心
 * Create：2025/9/29 01:23:13
 * Remark：
 */

package config

// Application 应用
type Application struct {
	Name        string `yaml:"name"`
	Version     string `yaml:"version"`
	Environment string `yaml:"environment"`
	Debug       bool   `yaml:"debug"`
}
