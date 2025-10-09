/**
 * Description：
 * FileName：server.go
 * Author：CJiaの用心
 * Create：2025/9/29 01:22:50
 * Remark：
 */

package config

// Server 服务
type Server struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}
