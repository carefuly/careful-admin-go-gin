/**
 * Description：
 * FileName：const.go
 * Author：CJiaの用心
 * Create：2025/12/2 14:55:00
 * Remark：
 */

package dict

type Type int // 字典类型

const (
	TypeOrdinary Type = iota + 1 // 普通字典
	TypeSystem                   // 系统字典
	TypeEnum                     // 枚举字典
)

// TypeMapping 字典类型映射
var TypeMapping = map[Type]string{
	TypeOrdinary: "普通字典",
	TypeSystem:   "系统字典",
	TypeEnum:     "枚举字典",
}

// TypeImportMapping 字典类型映射
var TypeImportMapping = map[string]Type{
	"普通字典": TypeOrdinary,
	"系统字典": TypeSystem,
	"枚举字典": TypeEnum,
}

type ValueType int // 数据类型

const (
	ValueTypeStr  ValueType = iota + 1 // 字符串
	ValueTypeInt                       // 整型
	ValueTypeBool                      // 布尔
)

// TypeValueMapping 数据类型映射
var TypeValueMapping = map[ValueType]string{
	ValueTypeStr:  "字符串",
	ValueTypeInt:  "整型",
	ValueTypeBool: "布尔",
}

// TypeValueImportMapping 数据类型映射
var TypeValueImportMapping = map[string]ValueType{
	"字符串": ValueTypeStr,
	"整型":  ValueTypeInt,
	"布尔":  ValueTypeBool,
}
