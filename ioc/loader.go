/**
 * Description：
 * FileName：loader.go
 * Author：CJiaの用心
 * Create：2025/9/28 20:09:07
 * Remark：
 */

package ioc

import (
	"github.com/carefuly/careful-admin-go-gin/config"
	"github.com/fsnotify/fsnotify"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
	"os"
	"sync"
)

// ConfigManager 配置管理器
type ConfigManager struct {
	Config     *config.LocalConfig
	RelyConfig config.RelyConfig
	ConfigFile string
	Watcher    *fsnotify.Watcher
	Mutex      sync.RWMutex
}

func InitConfig(configFile string) *ConfigManager {
	// 初始化配置管理器
	return initLocalConfig(configFile)
}

// 初始化本地配置文件
func initLocalConfig(configFile string) *ConfigManager {
	cm := &ConfigManager{
		ConfigFile: configFile,
		Config:     &config.LocalConfig{},
	}

	// 初始加载配置
	if err := cm.loadConfig(); err != nil {
		zap.S().Errorw("加载配置文件失败", "filePath", cm.ConfigFile, "error", err)
		panic("加载配置文件失败，请检查路径和权限")
	}

	return cm
}

// loadConfig 加载配置文件
func (cm *ConfigManager) loadConfig() error {
	cm.Mutex.Lock()
	defer cm.Mutex.Unlock()

	data, err := os.ReadFile(cm.ConfigFile)
	if err != nil {
		return err
	}

	newConfig := &config.LocalConfig{}
	if err := yaml.Unmarshal(data, newConfig); err != nil {
		return err
	}

	cm.Config = newConfig
	zap.S().Debugf("配置文件已重新加载: %s", cm.ConfigFile)
	return nil
}

// GetConfig 获取当前配置（线程安全）
func (cm *ConfigManager) GetConfig() *config.LocalConfig {
	cm.Mutex.RLock()
	defer cm.Mutex.RUnlock()

	// 返回配置的副本以避免并发修改
	configCopy := *cm.Config
	return &configCopy
}

// StartWatching 开始监听配置文件变化
func (cm *ConfigManager) StartWatching() error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	cm.Watcher = watcher

	// 添加配置文件到监听列表
	err = watcher.Add(cm.ConfigFile)
	if err != nil {
		return err
	}

	// 启动监听协程
	go cm.watchConfigFile()

	zap.S().Debugf("开始监听配置文件变化: %s", cm.ConfigFile)
	return nil
}

// watchConfigFile 监听配置文件变化
func (cm *ConfigManager) watchConfigFile() {
	for {
		select {
		case event, ok := <-cm.Watcher.Events:
			if !ok {
				return
			}

			// 检查是否是写入事件
			if event.Op&fsnotify.Write == fsnotify.Write {
				zap.S().Debug("检测到配置文件变化", event.Name)
				if err := cm.loadConfig(); err != nil {
					zap.S().Errorw("重新加载配置文件失败", "error", err)
				} else {
					zap.S().Debugf("配置文件重新加载成功")
				}
			}

		case err, ok := <-cm.Watcher.Errors:
			if !ok {
				return
			}
			zap.S().Errorw("配置文件监听错误", "error", err)
		}
	}
}

// StopWatching 停止监听配置文件变化
func (cm *ConfigManager) StopWatching() {
	if cm.Watcher != nil {
		err := cm.Watcher.Close()
		if err != nil {
			return
		}
		zap.S().Infof("已停止监听配置文件变化")
	}
}
