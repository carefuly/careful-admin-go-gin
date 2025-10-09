/**
 * Description：
 * FileName：nacos.go
 * Author：CJiaの用心
 * Create：2025/10/8 11:35:56
 * Remark：
 */

package ioc

import (
	"github.com/carefuly/careful-admin-go-gin/config"
	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/vo"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

// InitLoadNacosConfig 初始化Nacos远程配置
func InitLoadNacosConfig(localConf *config.LocalConfig) *config.RemoteConfig {
	serverConfig := []constant.ServerConfig{{
		IpAddr: localConf.NaCos.Host,
		Port:   localConf.NaCos.Port,
	}}

	clientConfig := constant.ClientConfig{
		NamespaceId:         localConf.NaCos.Namespace,
		TimeoutMs:           5000,
		NotLoadCacheAtStart: true,
		LogDir:              "tmp/nacos/log",
		CacheDir:            "tmp/nacos/cache",
		LogLevel:            "debug",
		Username:            localConf.NaCos.User,
		Password:            localConf.NaCos.Password,
	}

	configClient, err := clients.CreateConfigClient(map[string]interface{}{
		"serverConfigs": serverConfig,
		"clientConfig":  clientConfig,
	})
	if err != nil {
		zap.S().Panic("创建Nacos客户端失败", zap.Error(err))
	}

	content, err := configClient.GetConfig(vo.ConfigParam{
		DataId: localConf.NaCos.DataId,
		Group:  localConf.NaCos.Group,
	})
	if err != nil || content == "" {
		zap.S().Panicf("获取Nacos配置失败: DataID=%s Group=%s, %v",
			localConf.NaCos.DataId,
			localConf.NaCos.Group,
			err,
		)
	}

	remoteConfig := new(config.RemoteConfig)
	if err = yaml.Unmarshal([]byte(content), remoteConfig); err != nil {
		zap.L().Panic("Nacos配置解析失败", zap.Error(err))
	}

	zap.S().Debugf("Nacos配置加载成功: %+v", remoteConfig)
	return remoteConfig
}
