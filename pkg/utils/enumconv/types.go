/**
 * Description：
 * FileName：types.go
 * Author：CJiaの用心
 * Create：2025/10/11 14:09:35
 * Remark：
 */

package enumconv

// EnumType 基础枚举类型约束
type EnumType interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 | ~bool | ~string
}

// StringConvertible 可转换为字符串的类型约束
type StringConvertible interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 | ~bool | ~string
}
