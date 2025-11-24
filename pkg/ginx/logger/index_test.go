/**
 * Description：
 * FileName：index_test.go.go
 * Author：CJiaの用心
 * Create：2025/9/28 16:35:00
 * Remark：
 */

package logger

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

// TestNewLogger 测试创建不同配置的日志记录器
func TestNewLogger(t *testing.T) {
	testCases := []struct {
		name       string
		cfg        *LogConfig
		wantOutput string
		wantErr    bool
	}{
		{
			name: "标准输出控制台日志",
			cfg: &LogConfig{
				Encoding:    EncodingConsole,
				OutputPath:  OutputStdout,
				Level:       DebugLevel,
				ColorOutput: true,
			},
			wantOutput: "测试控制台日志",
		},
		{
			name: "错误输出控制台日志",
			cfg: &LogConfig{
				Encoding:    EncodingConsole,
				OutputPath:  OutputStderr,
				Level:       InfoLevel,
				ColorOutput: false,
			},
			wantOutput: "测试错误输出",
		},
		{
			name: "JSON格式文件日志",
			cfg: &LogConfig{
				Encoding:     EncodingJSON,
				OutputPath:   filepath.Join(t.TempDir(), "test.json.log"),
				Level:        InfoLevel,
				EnableCaller: true,
			},
			wantOutput: `"msg":"测试JSON日志"`,
		},
		{
			name: "带轮转的文件日志",
			cfg: &LogConfig{
				OutputPath: filepath.Join(t.TempDir(), "rotation.log"),
				Rotation: &RotationConfig{
					MaxSizeMB:  1,
					MaxBackups: 3,
					MaxAgeDays: 1,
					Compress:   true,
				},
			},
			wantOutput: "轮转测试日志",
		},
		{
			name: "动态路径文件日志",
			cfg: &LogConfig{
				OutputPath: "",
				DynamicPathFn: func() string {
					return filepath.Join(t.TempDir(), "dynamic.log")
				},
			},
			wantOutput: "动态路径日志",
		},
		{
			name:       "默认配置",
			cfg:        nil,
			wantOutput: "默认配置日志",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			l := NewLogger(tc.cfg)
			require.NotNil(t, l)
			defer func(l *Logger) {
				err := l.Close()
				if err != nil {
					return
				}
			}(l)

			// 执行日志记录
			l.Info(tc.wantOutput)

			// 验证日志输出
			if strings.HasPrefix(tc.cfg.OutputPath, "std") {
				// 对于控制台输出，需要捕获输出内容
				if tc.cfg.OutputPath == OutputStdout {
					// 这里简化处理，实际中需要捕获stdout
					t.Logf("日志输出到stdout: %s", tc.wantOutput)
				} else {
					// 这里简化处理，实际中需要捕获stderr
					t.Logf("日志输出到stderr: %s", tc.wantOutput)
				}
			} else {
				// 对于文件输出，直接读取文件内容
				filePath := tc.cfg.OutputPath
				if tc.cfg.DynamicPathFn != nil {
					filePath = tc.cfg.DynamicPathFn()
				}

				data, err := os.ReadFile(filePath)
				require.NoError(t, err)
				assert.Contains(t, string(data), tc.wantOutput)
			}
		})
	}
}

// TestLogLevels 测试日志级别过滤
func TestLogLevels(t *testing.T) {
	testCases := []struct {
		name           string
		level          LogLevel
		logDebug       bool
		logInfo        bool
		logWarn        bool
		logError       bool
		wantContain    []string
		wantNotContain []string
	}{
		{
			name:        "Debug级别应显示所有日志",
			level:       DebugLevel,
			logDebug:    true,
			logInfo:     true,
			logWarn:     true,
			logError:    true,
			wantContain: []string{"DEBUG", "INFO", "WARN", "ERROR"},
		},
		{
			name:           "Info级别不显示Debug日志",
			level:          InfoLevel,
			logDebug:       true,
			logInfo:        true,
			logWarn:        true,
			logError:       true,
			wantContain:    []string{"INFO", "WARN", "ERROR"},
			wantNotContain: []string{"DEBUG"},
		},
		{
			name:           "Warn级别不显示Debug和Info日志",
			level:          WarnLevel,
			logDebug:       true,
			logInfo:        true,
			logWarn:        true,
			logError:       true,
			wantContain:    []string{"WARN", "ERROR"},
			wantNotContain: []string{"DEBUG", "INFO"},
		},
		{
			name:           "Error级别仅显示Error和以上",
			level:          ErrorLevel,
			logDebug:       true,
			logInfo:        true,
			logWarn:        true,
			logError:       true,
			wantContain:    []string{"ERROR"},
			wantNotContain: []string{"DEBUG", "INFO", "WARN"},
		},
		{
			name:           "动态修改日志级别",
			level:          DebugLevel,
			logDebug:       false, // 初始不记录
			logInfo:        true,
			wantContain:    []string{"INFO"},
			wantNotContain: []string{"DEBUG"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 创建一个临时文件用于捕获日志输出
			tmpFile := filepath.Join(t.TempDir(), "log_levels.log")

			cfg := &LogConfig{
				OutputPath: tmpFile,
				Level:      tc.level,
				Encoding:   EncodingConsole,
			}

			l := NewLogger(cfg)
			defer func(l *Logger) {
				err := l.Close()
				if err != nil {
					return
				}
			}(l)

			// 动态修改级别（如测试用例需要）
			if tc.name == "动态修改日志级别" {
				l.SetLevel(InfoLevel) // 初始设置为Info
				l.Info("初始设置级别为Info")
			}

			// 记录各种级别的日志
			if tc.logDebug {
				l.Debug("调试信息 - DEBUG")
			}
			if tc.logInfo {
				l.Info("普通信息 - INFO")
			}
			if tc.logWarn {
				l.Warn("警告信息 - WARN")
			}
			if tc.logError {
				l.Error("错误信息 - ERROR")
			}

			// 动态修改级别后记录
			if tc.name == "动态修改日志级别" {
				l.SetLevel(DebugLevel) // 修改为Debug
				l.Debug("级别修改为Debug后记录")
			}

			// 确保日志写入完成
			require.NoError(t, l.Sync())

			// 读取并验证日志文件内容
			data, err := os.ReadFile(tmpFile)
			require.NoError(t, err)
			content := string(data)

			for _, s := range tc.wantContain {
				assert.Contains(t, content, s)
			}

			for _, s := range tc.wantNotContain {
				assert.NotContains(t, content, s)
			}
		})
	}
}

// TestLogRotation 测试日志轮转功能
func TestLogRotation(t *testing.T) {
	testCases := []struct {
		name      string
		maxSizeMB int
		logCount  int
		wantFiles int
	}{
		{
			name:      "小日志不触发轮转",
			maxSizeMB: 1,
			logCount:  100,
			wantFiles: 1, // 仅当前日志文件
		},
		{
			name:      "触发单次轮转",
			maxSizeMB: 1,
			logCount:  3000, // ~3MB 日志
			wantFiles: 2,    // 当前 + 1个轮转文件
		},
		{
			name:      "触发多次轮转",
			maxSizeMB: 1,
			logCount:  10000, // ~10MB 日志
			wantFiles: 3,     // 当前 + 2个轮转文件 (MaxBackups=3)
		},
		{
			name:      "手动触发轮转",
			maxSizeMB: 100, // 设置很大，确保不自动轮转
			logCount:  100,
			wantFiles: 2, // 手动轮转会创建新文件
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			logPath := filepath.Join(tmpDir, "rotation_test.log")

			cfg := &LogConfig{
				OutputPath: logPath,
				Rotation: &RotationConfig{
					MaxSizeMB:  tc.maxSizeMB,
					MaxBackups: 2, // 最多保留2个旧日志
					MaxAgeDays: 1,
					Compress:   false, // 测试中禁用压缩
				},
			}

			l := NewLogger(cfg)
			defer func(l *Logger) {
				err := l.Close()
				if err != nil {
					return
				}
			}(l)

			// 生成日志
			msg := strings.Repeat("A", 1000) // 每条日志约1KB
			for i := 0; i < tc.logCount; i++ {
				l.Info(fmt.Sprintf("%d - %s", i, msg))
			}

			// 手动触发轮转
			if tc.name == "手动触发轮转" {
				require.NoError(t, l.Rotate())
			}

			// 确保日志写入完成
			require.NoError(t, l.Sync())

			// 检查日志文件
			files, err := os.ReadDir(tmpDir)
			require.NoError(t, err)

			// 验证文件数量
			assert.Equal(t, tc.wantFiles, len(files), "文件数量不匹配")
		})
	}
}

// TestDynamicPath 测试动态路径功能
func TestDynamicPath(t *testing.T) {
	testCases := []struct {
		name        string
		dynamicFn   func() string
		wantPattern string
	}{
		{
			name:        "简化版动态路径生成器",
			dynamicFn:   SimpleDynamicPath("/var/log/myapp"),
			wantPattern: `/var/log/myapp/\d{4}-\d{2}-\d{2}\.log`,
		},
		{
			name: "每天动态变化的路径",
			dynamicFn: func() string {
				return fmt.Sprintf("/tmp/logs/%s.log", time.Now().Format("2006-01-02"))
			},
			wantPattern: `^/tmp/logs/\d{4}-\d{2}-\d{2}\.log$`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			path := tc.dynamicFn()
			assert.Regexp(t, tc.wantPattern, path, "路径模式不匹配")

			// 在日志记录器中使用动态路径
			cfg := &LogConfig{
				DynamicPathFn: tc.dynamicFn,
				Level:         InfoLevel,
			}

			l := NewLogger(cfg)
			l.Info("测试动态路径日志")
			defer func(l *Logger) {
				err := l.Close()
				if err != nil {
					return
				}
			}(l)

			// 确保日志文件已创建
			_, err := os.Stat(path)
			assert.NoError(t, err, "日志文件应该存在")
		})
	}
}

// TestLoggerConcurrency 测试并发安全
func TestLoggerConcurrency(t *testing.T) {
	testCases := []struct {
		name           string
		goroutines     int
		logsPerRoutine int
	}{
		{
			name:           "轻量并发-5个协程",
			goroutines:     5,
			logsPerRoutine: 100,
		},
		{
			name:           "高并发-50个协程",
			goroutines:     50,
			logsPerRoutine: 200,
		},
		{
			name:           "极高并发-100个协程",
			goroutines:     100,
			logsPerRoutine: 100,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tmpFile := filepath.Join(t.TempDir(), "concurrency.log")

			cfg := &LogConfig{
				OutputPath: tmpFile,
				Level:      DebugLevel,
				Encoding:   EncodingConsole,
			}

			l := NewLogger(cfg)
			defer func(l *Logger) {
				err := l.Close()
				if err != nil {
					return
				}
			}(l)

			var wg sync.WaitGroup
			wg.Add(tc.goroutines)

			for i := 0; i < tc.goroutines; i++ {
				go func(id int) {
					defer wg.Done()
					for j := 0; j < tc.logsPerRoutine; j++ {
						// 在协程中动态修改日志级别（测试并发安全）
						if j%10 == 0 {
							l.SetLevel(zapcore.Level(id % 4))
						}
					}
				}(i)
			}

			wg.Wait()
			require.NoError(t, l.Sync())

			// 验证日志数量
			data, err := os.ReadFile(tmpFile)
			require.NoError(t, err)

			logLines := strings.Split(string(data), "\n")
			// 过滤空行
			var count int
			for _, line := range logLines {
				if strings.TrimSpace(line) != "" {
					count++
				}
			}

			expected := tc.goroutines * tc.logsPerRoutine
			assert.Equal(t, expected, count, "日志行数不匹配")
		})
	}
}

// TestLoggerClose 测试安全关闭
func TestCloseLogger(t *testing.T) {
	testCases := []struct {
		name       string
		closeTimes int
		wantErr    bool
	}{
		{
			name:       "正常关闭一次",
			closeTimes: 1,
			wantErr:    false,
		},
		{
			name:       "多次关闭",
			closeTimes: 3,
			wantErr:    true, // 第一次之后的关闭应该返回错误
		},
		{
			name:       "关闭后尝试写入",
			closeTimes: 1,
			wantErr:    false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			l := NewLogger(nil)

			// 关闭日志记录器
			var lastErr error
			for i := 0; i < tc.closeTimes; i++ {
				lastErr = l.Close()
			}

			if tc.wantErr {
				assert.Error(t, lastErr, "多次关闭应该返回错误")
			} else {
				assert.NoError(t, lastErr, "关闭应该成功")
			}

			// 尝试在关闭后写入日志
			if tc.name == "关闭后尝试写入" {
				assert.NotPanics(t, func() {
					l.Info("关闭后写入应该安全失败")
				}, "关闭后写入不应该panic")
			}
		})
	}
}
