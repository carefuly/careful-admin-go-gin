/**
 * Description：
 * FileName：logger.go
 * Author：CJiaの用心
 * Create：2025/11/23 00:51:37
 * Remark：
 */

package ioc

import (
	"github.com/carefuly/careful-admin-go-gin/pkg/ginx/logger"
	"go.uber.org/zap"
)

// LoggerManager 日志管理器
type LoggerManager struct {
	logger *zap.Logger
}

// InitLogger 初始化日志系统 - 创建带彩色输出的控制台日志记录器
// 这是应用程序启动时的第一个初始化步骤
func InitLogger() *LoggerManager {
	// 创建控制台日志配置 - 启用彩色输出和调用位置记录
	logCfg := &logger.LogConfig{
		Encoding:     logger.EncodingConsole, // 使用控制台格式以获得更好的可读性
		OutputPath:   logger.OutputStdout,    // 输出到标准输出
		Level:        logger.DebugLevel,      // 设置为调试级别，便于开发调试
		ColorOutput:  true,                   // 启用彩色输出，提升日志可读性
		EnableCaller: true,                   // 记录调用位置，便于问题定位
	}

	// 创建新的日志器实例
	customLogger := logger.NewLogger(logCfg)

	// 获取原始的 zap.Logger 实例
	zapLogger := customLogger.Logger

	// 替换全局日志器，使其他模块可以直接使用 zap.L()
	zap.ReplaceGlobals(zapLogger)

	// 记录初始化成功信息
	zapLogger.Info("🚀 日志系统初始化完成",
		zap.String("encoding", logCfg.Encoding),
		zap.String("output", logCfg.OutputPath),
		zap.String("level", logCfg.Level.String()),
		zap.Bool("colorOutput", logCfg.ColorOutput),
	)

	return &LoggerManager{
		logger: zapLogger,
	}
}

// InitFileLogger 初始化文件日志记录器 - 用于生产环境的文件日志
func InitFileLogger(filePath string) *LoggerManager {
	// 创建文件日志配置
	logCfg := &logger.LogConfig{
		Encoding:     logger.EncodingJSON, // JSON 格式便于日志分析
		OutputPath:   filePath,            // 指定文件路径
		Level:        logger.InfoLevel,    // 生产环境使用 Info 级别
		ColorOutput:  false,               // 文件输出不需要颜色
		EnableCaller: true,                // 保留调用位置信息
		Rotation: &logger.RotationConfig{ // 配置日志轮转
			MaxSizeMB:  10,   // 单文件最大 10MB
			MaxBackups: 5,    // 保留 5 个备份文件
			MaxAgeDays: 30,   // 保留 30 天
			Compress:   true, // 压缩旧日志文件
		},
	}

	// 创建文件日志器
	customLogger := logger.NewLogger(logCfg)
	zapLogger := customLogger.Logger

	// 记录文件日志初始化信息
	zapLogger.Info("📁 文件日志系统初始化完成",
		zap.String("filePath", filePath),
		zap.String("level", logCfg.Level.String()),
	)

	return &LoggerManager{
		logger: zapLogger,
	}
}

// GetLogger 获取日志器实例
func (lm *LoggerManager) GetLogger() *zap.Logger {
	return lm.logger
}

// Close 关闭日志器，确保所有日志都被写入
func (lm *LoggerManager) Close() error {
	if lm.logger != nil {
		return lm.logger.Sync()
	}
	return nil
}
