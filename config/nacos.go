/**
 * Description：
 * FileName：nacos.go
 * Author：CJiaの用心
 * Create：2025/9/29 01:23:37
 * Remark：
 */

package config

type NaCos struct {
	Host      string `yaml:"host"`
	Port      uint64 `yaml:"port"`
	Namespace string `yaml:"namespace"`
	User      string `yaml:"user"`
	Password  string `yaml:"password"`
	DataId    string `yaml:"dataId"`
	Group     string `yaml:"group"`
}
