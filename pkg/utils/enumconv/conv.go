/**
 * Description：
 * FileName：conv.go
 * Author：CJiaの用心
 * Create：2025/10/11 14:10:08
 * Remark：
 */

package enumconv

import (
	"fmt"
	"strings"
)

// EnumConverter 枚举转换器结构体
type EnumConverter[T EnumType, U StringConvertible] struct {
	forwardMap  map[T]U // 正向映射 (枚举值 -> 字符串/其他)
	reverseMap  map[U]T // 反向映射 (字符串/其他 -> 枚举值)
	validValues []U     // 有效值列表
	enumName    string  // 枚举名称(用于错误消息)
}

// NewEnumConverter 创建新的枚举转换器
func NewEnumConverter[T EnumType, U StringConvertible](
	forwardMap map[T]U,
	reverseMap map[U]T,
	validValues []U,
	enumName string,
) *EnumConverter[T, U] {
	return &EnumConverter[T, U]{
		forwardMap:  forwardMap,
		reverseMap:  reverseMap,
		validValues: validValues,
		enumName:    enumName,
	}
}

// FromEnum 从枚举值转换为字符串/其他类型
func (c *EnumConverter[T, U]) FromEnum(input T) (U, error) {
	if val, exists := c.forwardMap[input]; exists {
		return val, nil
	}

	var zero U
	return zero, fmt.Errorf("无效的%s枚举值: %v", c.enumName, input)
}

// ToEnum 从字符串/其他类型转换为枚举值
func (c *EnumConverter[T, U]) ToEnum(input U) (T, error) {
	if val, exists := c.reverseMap[input]; exists {
		return val, nil
	}

	var zero T
	validStr := strings.Join(c.toStringSlice(c.validValues), "/")
	return zero, fmt.Errorf("无效的%s输入值: %v，可选值：%s", c.enumName, input, validStr)
}

// toStringSlice 将任意类型的切片转换为字符串切片
func (c *EnumConverter[T, U]) toStringSlice(values []U) []string {
	result := make([]string, len(values))
	for i, v := range values {
		result[i] = fmt.Sprintf("%v", v)
	}
	return result
}
