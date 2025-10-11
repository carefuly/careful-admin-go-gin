/**
 * Description：
 * FileName：conv_test.go.go
 * Author：CJiaの用心
 * Create：2025/10/11 14:10:53
 * Remark：
 */

package enumconv

import (
	"fmt"
	"testing"
)

// 数值 ↔ 字符串转换
type DictType int

const (
	NormalDict DictType = iota // 普通字典
	SystemDict                 // 系统字典
	EnumDict                   // 枚举字典
)

func TestNumericToString(t *testing.T) {
	// 定义映射关系
	forwardMap := map[DictType]string{
		NormalDict: "普通字典",
		SystemDict: "系统字典",
		EnumDict:   "枚举字典",
	}

	reverseMap := map[string]DictType{
		"普通字典": NormalDict,
		"系统字典": SystemDict,
		"枚举字典": EnumDict,
	}

	validValues := []string{"普通字典", "系统字典", "枚举字典"}

	// 创建转换器
	converter := NewEnumConverter(forwardMap, reverseMap, validValues, "字典类型")

	// 测试正向转换
	str, err := converter.FromEnum(SystemDict)
	fmt.Printf("String: %s, Error: %v\n", str, err)

	// 测试反向转换
	enum, err := converter.ToEnum("枚举字典")
	fmt.Printf("Enum: %d, Error: %v\n", enum, err)

	// 测试无效值
	_, err = converter.ToEnum("无效字典")
	fmt.Println("Error:", err)
}

// 布尔 ↔ 字符串转换
type BoolStatus int

const (
	StatusDisabled BoolStatus = 0
	StatusEnabled  BoolStatus = 1
)

func TestBoolToString(t *testing.T) {
	// 定义映射关系
	forwardMap := map[bool]string{
		true:  "启用",
		false: "禁用",
	}

	reverseMap := map[string]bool{
		"启用": true,
		"禁用": false,
	}

	validValues := []string{"启用", "禁用"}

	// 创建转换器
	converter := NewEnumConverter(forwardMap, reverseMap, validValues, "布尔状态")

	// 测试正向转换
	str, err := converter.FromEnum(true)
	fmt.Printf("String: %s, Error: %v\n", str, err)

	// 测试反向转换
	b, err := converter.ToEnum("禁用")
	fmt.Printf("Bool: %t, Error: %v\n", b, err)
}

// 字符串 ↔ 字符串转换
type LogLevel string

const (
	LogDebug LogLevel = "DEBUG"
	LogInfo  LogLevel = "INFO"
	LogWarn  LogLevel = "WARN"
	LogError LogLevel = "ERROR"
)

func TestStringToString(t *testing.T) {
	// 定义映射关系
	forwardMap := map[LogLevel]string{
		LogDebug: "调试",
		LogInfo:  "信息",
		LogWarn:  "警告",
		LogError: "错误",
	}

	reverseMap := map[string]LogLevel{
		"调试": LogDebug,
		"信息": LogInfo,
		"警告": LogWarn,
		"错误": LogError,
	}

	validValues := []string{"调试", "信息", "警告", "错误"}

	// 创建转换器
	converter := NewEnumConverter(forwardMap, reverseMap, validValues, "日志级别")

	// 测试正向转换
	str, err := converter.FromEnum(LogWarn)
	fmt.Printf("Chinese: %s, Error: %v\n", str, err)

	// 测试反向转换
	level, err := converter.ToEnum("错误")
	fmt.Printf("Level: %s, Error: %v\n", level, err)
}
