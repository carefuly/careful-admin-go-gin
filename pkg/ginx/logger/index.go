/**
 * Description：高性能结构化日志记录器
 * FileName：index.go
 * Author：CJiaの用心
 * Create：2025/9/28 20:27:00
 * Remark：
 */

/**
 * 相关技术官方文档
 * 1. Zap 高性能日志库
 * · GitHub: https://github.com/uber-go/zap
 * · 文档: https://pkg.go.dev/go.uber.org/zap
 * · 最佳实践: https://github.com/uber-go/zap/blob/master/FAQ.md
 * 2. Zapcore (Zap 核心组件)
 * · 文档: https://pkg.go.dev/go.uber.org/zap/zapcore
 * · 编码器文档: https://pkg.go.dev/go.uber.org/zap/zapcore#Encoder
 * 3. Lumberjack 日志轮转库
 * · GitHub: https://github.com/natefinch/lumberjack
 * · 文档: https://pkg.go.dev/gopkg.in/natefinch/lumberjack.v2
 * 4. Go Sync 包 (并发控制)
 * · 官方文档: https://pkg.go.dev/sync
 * · Once 类型: https://pkg.go.dev/sync#Once
 */

package logger

import (
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// 日志配置常量
const (
	EncodingConsole = "console"
	EncodingJSON    = "json"

	OutputStdout = "stdout"
	OutputStderr = "stderr"

	TimeKey    = "time"
	CallerKey  = "caller"
	LevelKey   = "level"
	MessageKey = "msg"
)

// 日志编码器变量
var (
	TimeEncoder     = zapcore.ISO8601TimeEncoder
	CallerEncoder   = zapcore.ShortCallerEncoder
	DurationEncoder = zapcore.SecondsDurationEncoder
)

// LogLevel 日志级别类型
type LogLevel = zapcore.Level

// 预定义日志级别
const (
	DebugLevel = zapcore.DebugLevel
	InfoLevel  = zapcore.InfoLevel
	WarnLevel  = zapcore.WarnLevel
	ErrorLevel = zapcore.ErrorLevel
	PanicLevel = zapcore.PanicLevel
	FatalLevel = zapcore.FatalLevel
)

// RotationConfig 日志轮转配置
type RotationConfig struct {
	MaxSizeMB  int  // 单个日志文件最大大小(MB)
	MaxBackups int  // 保留的旧日志文件数
	MaxAgeDays int  // 日志保留天数
	Compress   bool // 是否压缩旧日志
}

// LogConfig 日志配置
type LogConfig struct {
	Encoding      string          // 编码格式: console 或 json
	OutputPath    string          // 输出目标: stdout/stderr 或 文件路径
	Level         LogLevel        // 日志级别
	ColorOutput   bool            // 是否彩色输出(仅console有效)
	EnableCaller  bool            // 是否记录调用位置
	Rotation      *RotationConfig // 日志轮转配置(文件输出时有效)
	DynamicPathFn func() string   // 动态生成文件路径的函数
}

// Logger 日志记录器
type Logger struct {
	*zap.Logger
	config   *LogConfig
	once     sync.Once // 确保安全关闭
	atom     zap.AtomicLevel
	fileSink *lumberjack.Logger // 文件输出源的引用(如果是文件日志)
}

// NewLogger 创建新日志实例
func NewLogger(cfg *LogConfig) *Logger {
	// 参数校验与默认值设置
	if cfg == nil {
		cfg = &LogConfig{}
	}

	if cfg.OutputPath == "" {
		cfg.OutputPath = OutputStdout
	}

	if cfg.Rotation == nil {
		cfg.Rotation = &RotationConfig{
			MaxSizeMB:  5,
			MaxBackups: 5,
			MaxAgeDays: 30,
			Compress:   true,
		}
	}

	// 创建原子级别，允许运行时动态修改日志级别
	atom := zap.NewAtomicLevelAt(cfg.Level)

	// 创建写入目标
	ws, fileSink := createWriteSyncer(cfg)

	// 构建核心
	core := zapcore.NewCore(
		createEncoder(cfg),
		ws,
		atom,
	)

	// 构建选项
	opts := []zap.Option{zap.AddCallerSkip(1)}
	if cfg.EnableCaller {
		opts = append(opts, zap.AddCaller())
	}

	// 创建基础logger
	zapLogger := zap.New(core, opts...)

	return &Logger{
		Logger:   zapLogger,
		config:   cfg,
		atom:     atom,
		fileSink: fileSink,
	}
}

// CustomColorLevelEncoder 自定义彩色级别编码器 - 提供更丰富的颜色显示
func CustomColorLevelEncoder(level zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
	switch level {
	case zapcore.DebugLevel:
		enc.AppendString("\033[36m[DEBUG]\033[0m") // 青色
	case zapcore.InfoLevel:
		enc.AppendString("\033[32m[INFO]\033[0m") // 绿色
	case zapcore.WarnLevel:
		enc.AppendString("\033[33m[WARN]\033[0m") // 黄色
	case zapcore.ErrorLevel:
		enc.AppendString("\033[31m[ERROR]\033[0m") // 红色
	case zapcore.PanicLevel:
		enc.AppendString("\033[35m[PANIC]\033[0m") // 紫色
	case zapcore.FatalLevel:
		enc.AppendString("\033[41m[FATAL]\033[0m") // 红色背景
	default:
		enc.AppendString(fmt.Sprintf("[%s]", level.CapitalString()))
	}
}

// CustomTimeEncoder 自定义时间编码器 - 提供更友好的时间显示格式
func CustomTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	// 使用带颜色的时间格式
	timeStr := fmt.Sprintf("\033[97m%s\033[0m", t.Format("2006-01-02 15:04:05.000"))
	enc.AppendString(timeStr)
}

// CustomCallerEncoder 自定义调用者编码器 - 提供更清晰的调用位置显示
func CustomCallerEncoder(caller zapcore.EntryCaller, enc zapcore.PrimitiveArrayEncoder) {
	// 使用带颜色的调用者信息
	callerStr := fmt.Sprintf("\033[94m%s\033[0m", caller.TrimmedPath())
	enc.AppendString(callerStr)
}

// createEncoder 根据配置创建编码器
func createEncoder(cfg *LogConfig) zapcore.Encoder {
	encCfg := zapcore.EncoderConfig{
		TimeKey:        TimeKey,
		LevelKey:       LevelKey,
		MessageKey:     MessageKey,
		CallerKey:      CallerKey,
		EncodeDuration: DurationEncoder,
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeTime:     CustomTimeEncoder,   // 自定义时间编码器
		EncodeCaller:   CustomCallerEncoder, // 自定义调用者编码器
	}

	switch cfg.Encoding {
	case EncodingJSON:
		// JSON 格式不使用颜色
		encCfg.EncodeTime = zapcore.ISO8601TimeEncoder
		encCfg.EncodeCaller = zapcore.ShortCallerEncoder
		encCfg.EncodeLevel = zapcore.LowercaseLevelEncoder
		return zapcore.NewJSONEncoder(encCfg)
	default: // console
		if cfg.ColorOutput {
			encCfg.EncodeLevel = CustomColorLevelEncoder
		} else {
			encCfg.EncodeTime = zapcore.ISO8601TimeEncoder
			encCfg.EncodeCaller = zapcore.ShortCallerEncoder
			encCfg.EncodeLevel = zapcore.CapitalLevelEncoder
		}
		return zapcore.NewConsoleEncoder(encCfg)
	}
}

// createWriteSyncer 根据配置创建写入目标
func createWriteSyncer(cfg *LogConfig) (zapcore.WriteSyncer, *lumberjack.Logger) {
	switch strings.ToLower(cfg.OutputPath) {
	case OutputStdout:
		return zapcore.AddSync(os.Stdout), nil
	case OutputStderr:
		return zapcore.AddSync(os.Stderr), nil
	default:
		// 动态路径处理
		filePath := cfg.OutputPath
		if cfg.DynamicPathFn != nil {
			filePath = cfg.DynamicPathFn()
		}

		// 确保目录存在
		if dir := filepath.Dir(filePath); dir != "" {
			if err := os.MkdirAll(dir, 0755); err != nil {
				fmt.Printf("无法创建日志目录 %s: %v\n", dir, err)
			}
		}

		// 创建 lumberjack 实例
		lj := &lumberjack.Logger{
			Filename:   filePath,
			MaxSize:    cfg.Rotation.MaxSizeMB,
			MaxBackups: cfg.Rotation.MaxBackups,
			MaxAge:     cfg.Rotation.MaxAgeDays,
			Compress:   cfg.Rotation.Compress,
		}

		// 返回包装后的写入目标及原始引用
		return zapcore.AddSync(lj), lj
	}
}

// SetLevel 动态设置日志级别
func (l *Logger) SetLevel(level LogLevel) {
	l.atom.SetLevel(level)
}

// GetLevel 获取当前日志级别
func (l *Logger) GetLevel() LogLevel {
	return l.atom.Level()
}

// Rotate 手动触发日志轮转(文件日志有效)
func (l *Logger) Rotate() error {
	if l.fileSink != nil {
		return l.fileSink.Rotate()
	}
	return fmt.Errorf("当前输出不是文件日志，不支持轮转")
}

// Close 安全关闭日志
func (l *Logger) Close() error {
	var err error
	l.once.Do(func() {
		err = l.Sync()
	})
	return err
}

// Sync 同步缓冲区的日志条目
func (l *Logger) Sync() error {
	return l.Logger.Sync()
}

// NewDevelopmentLogger 创建开发环境日志器 - 带有丰富的颜色和调试信息
func NewDevelopmentLogger() *Logger {
	cfg := &LogConfig{
		Encoding:     EncodingConsole,
		OutputPath:   OutputStdout,
		Level:        DebugLevel,
		ColorOutput:  true,
		EnableCaller: true,
	}
	return NewLogger(cfg)
}

// NewProductionLogger 创建生产环境日志器 - JSON 格式，适合日志收集
func NewProductionLogger(filePath string) *Logger {
	cfg := &LogConfig{
		Encoding:     EncodingJSON,
		OutputPath:   filePath,
		Level:        InfoLevel,
		ColorOutput:  false,
		EnableCaller: false,
		Rotation: &RotationConfig{
			MaxSizeMB:  50, // 生产环境使用更大的文件
			MaxBackups: 10, // 保留更多备份
			MaxAgeDays: 90, // 保留更长时间
			Compress:   true,
		},
	}
	return NewLogger(cfg)
}

// NewTestLogger 创建测试环境日志器 - 简洁输出，便于测试
func NewTestLogger() *Logger {
	cfg := &LogConfig{
		Encoding:     EncodingConsole,
		OutputPath:   OutputStdout,
		Level:        WarnLevel, // 测试时只显示警告和错误
		ColorOutput:  false,     // 测试环境不使用颜色
		EnableCaller: false,     // 测试时不显示调用位置
	}
	return NewLogger(cfg)
}

// 以下为动态路径生成函数 -------------------------------------------------

// DefaultDynamicPath 默认动态路径生成函数
func DefaultDynamicPath(path string) string {
	now := time.Now()
	paths := strings.Split(path, "/")
	return fmt.Sprintf(
		"./tmp/admin/%s/%s/%s/%s.log",
		paths[2],
		now.Format("2006-01-02"),
		paths[3],
		paths[4],
	)
}

// SimpleDynamicPath 简化版动态路径生成
func SimpleDynamicPath(baseDir string) func() string {
	return func() string {
		now := time.Now()
		return fmt.Sprintf(
			"%s/%s.log",
			baseDir,
			now.Format("2006-01-02"),
		)
	}
}
