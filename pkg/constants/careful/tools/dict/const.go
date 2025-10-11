/**
 * Description：
 * FileName：const.go
 * Author：CJiaの用心
 * Create：2025/10/11 11:27:43
 * Remark：
 */

package dict

type TypeConst int // 字典类型

const (
	TypeConstOrdinary TypeConst = iota + 1 // 普通字典
	TypeConstSystem                        // 系统字典
	TypeConstEnum                          // 枚举字典
)

// TypeMapping 字典类型映射
var TypeMapping = map[TypeConst]string{
	TypeConstOrdinary: "普通字典",
	TypeConstSystem:   "系统字典",
	TypeConstEnum:     "枚举字典",
}

// TypeImportMapping 字典类型映射
var TypeImportMapping = map[string]TypeConst{
	"普通字典": TypeConstOrdinary,
	"系统字典": TypeConstSystem,
	"枚举字典": TypeConstEnum,
}

type ValueTypeConst int // 数据类型

const (
	ValueTypeConstStr  ValueTypeConst = iota + 1 // 字符串
	ValueTypeConstInt                            // 整型
	ValueTypeConstBool                           // 布尔
)

// TypeValueMapping 数据类型映射
var TypeValueMapping = map[ValueTypeConst]string{
	ValueTypeConstStr:  "字符串",
	ValueTypeConstInt:  "整型",
	ValueTypeConstBool: "布尔",
}

// TypeValueImportMapping 数据类型映射
var TypeValueImportMapping = map[string]ValueTypeConst{
	"字符串": ValueTypeConstStr,
	"整型":  ValueTypeConstInt,
	"布尔":  ValueTypeConstBool,
}
