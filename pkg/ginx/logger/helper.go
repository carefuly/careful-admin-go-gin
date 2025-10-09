/**
 * Description：
 * FileName：helper.go
 * Author：CJiaの用心
 * Create：2025/9/28 21:55:17
 * Remark：
 */

package logger

import (
	"fmt"
	"runtime"
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// ColorCode 颜色代码常量
const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorPurple = "\033[35m"
	ColorCyan   = "\033[36m"
	ColorGray   = "\033[90m"
	ColorWhite  = "\033[97m"

	// ColorBgRed 背景色
	ColorBgRed    = "\033[41m"
	ColorBgGreen  = "\033[42m"
	ColorBgYellow = "\033[43m"
	ColorBgBlue   = "\033[44m"

	// ColorBold 样式
	ColorBold      = "\033[1m"
	ColorUnderline = "\033[4m"
)

// LogHelper 日志辅助工具
type LogHelper struct {
	logger *zap.Logger
}

// NewLogHelper 创建日志辅助工具
func NewLogHelper(logger *zap.Logger) *LogHelper {
	return &LogHelper{logger: logger}
}

// Colorize 为文本添加颜色
func Colorize(text, color string) string {
	return fmt.Sprintf("%s%s%s", color, text, ColorReset)
}

// LogWithColor 带颜色的日志输出
func (h *LogHelper) LogWithColor(level zapcore.Level, msg string, fields ...zap.Field) {
	var coloredMsg string
	switch level {
	case zapcore.DebugLevel:
		coloredMsg = Colorize(msg, ColorCyan)
	case zapcore.InfoLevel:
		coloredMsg = Colorize(msg, ColorGreen)
	case zapcore.WarnLevel:
		coloredMsg = Colorize(msg, ColorYellow)
	case zapcore.ErrorLevel:
		coloredMsg = Colorize(msg, ColorRed)
	default:
		coloredMsg = msg
	}

	h.logger.Log(level, coloredMsg, fields...)
}

// Success 成功日志 - 绿色高亮
func (h *LogHelper) Success(msg string, fields ...zap.Field) {
	successMsg := fmt.Sprintf("%s✅ %s%s", ColorGreen, msg, ColorReset)
	h.logger.Info(successMsg, fields...)
}

// Warning 警告日志 - 黄色高亮
func (h *LogHelper) Warning(msg string, fields ...zap.Field) {
	warningMsg := fmt.Sprintf("%s⚠️  %s%s", ColorYellow, msg, ColorReset)
	h.logger.Warn(warningMsg, fields...)
}

// Error 错误日志 - 红色高亮
func (h *LogHelper) Error(msg string, err error, fields ...zap.Field) {
	errorMsg := fmt.Sprintf("%s❌ %s%s", ColorRed, msg, ColorReset)
	allFields := append(fields, zap.Error(err))
	h.logger.Error(errorMsg, allFields...)
}

// Fatal 致命错误日志 - 红色背景高亮
func (h *LogHelper) Fatal(msg string, err error, fields ...zap.Field) {
	fatalMsg := fmt.Sprintf("%s💀 %s%s", ColorBgRed+ColorWhite, msg, ColorReset)
	allFields := append(fields, zap.Error(err))
	h.logger.Fatal(fatalMsg, allFields...)
}

// Progress 进度日志 - 蓝色高亮
func (h *LogHelper) Progress(msg string, current, total int, fields ...zap.Field) {
	percentage := float64(current) / float64(total) * 100
	progressMsg := fmt.Sprintf("%s🔄 %s [%d/%d] %.1f%%%s",
		ColorBlue, msg, current, total, percentage, ColorReset)
	allFields := append(fields,
		zap.Int("current", current),
		zap.Int("total", total),
		zap.Float64("percentage", percentage),
	)
	h.logger.Info(progressMsg, allFields...)
}

// Performance 性能监控日志
func (h *LogHelper) Performance(operation string, duration time.Duration, fields ...zap.Field) {
	var color string
	var icon string

	switch {
	case duration < 100*time.Millisecond:
		color = ColorGreen
		icon = "🚀"
	case duration < 500*time.Millisecond:
		color = ColorYellow
		icon = "⚡"
	default:
		color = ColorRed
		icon = "🐌"
	}

	perfMsg := fmt.Sprintf("%s%s %s 耗时: %v%s",
		color, icon, operation, duration, ColorReset)
	allFields := append(fields,
		zap.String("operation", operation),
		zap.Duration("duration", duration),
	)
	h.logger.Info(perfMsg, allFields...)
}

// Memory 内存使用日志
func (h *LogHelper) Memory(operation string, fields ...zap.Field) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	memMsg := fmt.Sprintf("%s💾 %s - 内存使用情况%s", ColorPurple, operation, ColorReset)
	allFields := append(fields,
		zap.String("operation", operation),
		zap.String("alloc", formatBytes(m.Alloc)),
		zap.String("totalAlloc", formatBytes(m.TotalAlloc)),
		zap.String("sys", formatBytes(m.Sys)),
		zap.Uint32("numGC", m.NumGC),
	)
	h.logger.Info(memMsg, allFields...)
}

// API 接口日志
func (h *LogHelper) API(method, path string, statusCode int, duration time.Duration, fields ...zap.Field) {
	var color string
	var icon string

	switch {
	case statusCode >= 200 && statusCode < 300:
		color = ColorGreen
		icon = "✅"
	case statusCode >= 300 && statusCode < 400:
		color = ColorYellow
		icon = "🔄"
	case statusCode >= 400 && statusCode < 500:
		color = ColorRed
		icon = "❌"
	default:
		color = ColorRed
		icon = "💥"
	}

	apiMsg := fmt.Sprintf("%s%s %s %s [%d] %v%s",
		color, icon, method, path, statusCode, duration, ColorReset)
	allFields := append(fields,
		zap.String("method", method),
		zap.String("path", path),
		zap.Int("statusCode", statusCode),
		zap.Duration("duration", duration),
	)
	h.logger.Info(apiMsg, allFields...)
}

// Database 数据库操作日志
func (h *LogHelper) Database(operation, table string, affected int64, duration time.Duration, fields ...zap.Field) {
	var color string
	var icon string

	switch {
	case duration < 50*time.Millisecond:
		color = ColorGreen
		icon = "🚀"
	case duration < 200*time.Millisecond:
		color = ColorYellow
		icon = "⚡"
	default:
		color = ColorRed
		icon = "🐌"
	}

	dbMsg := fmt.Sprintf("%s%s DB %s [%s] 影响行数: %d, 耗时: %v%s",
		color, icon, operation, table, affected, duration, ColorReset)
	allFields := append(fields,
		zap.String("operation", operation),
		zap.String("table", table),
		zap.Int64("affected", affected),
		zap.Duration("duration", duration),
	)
	h.logger.Info(dbMsg, allFields...)
}

// formatBytes 格式化字节数
func formatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := uint64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// TimerFunc 计时器函数 - 用于测量函数执行时间
func (h *LogHelper) TimerFunc(name string, fn func()) {
	start := time.Now()
	fn()
	duration := time.Since(start)
	h.Performance(name, duration)
}

// Separator 分隔符日志 - 用于分隔不同的日志段落
func (h *LogHelper) Separator(title string) {
	separator := strings.Repeat("=", 50)
	msg := fmt.Sprintf("%s\n%s %s %s\n%s%s",
		ColorCyan, separator, title, separator, separator, ColorReset)
	h.logger.Info(msg)
}
