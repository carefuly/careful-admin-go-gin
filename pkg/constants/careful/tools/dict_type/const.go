/**
 * Description：
 * FileName：const.go
 * Author：CJiaの用心
 * Create：2025/10/19 14:19:52
 * Remark：
 */

package dict_type

type DictTag string // 标签类型

const (
	DictTagPrimary DictTag = "primary" // primary
	DictTagSuccess DictTag = "success" // success
	DictTagWarning DictTag = "warning" // warning
	DictTagDanger  DictTag = "danger"  // danger
	DictTagInfo    DictTag = "info"    // info
)

// DictTagMapping 标签类型映射
var DictTagMapping = map[DictTag]string{
	DictTagPrimary: "primary",
	DictTagSuccess: "success",
	DictTagWarning: "warning",
	DictTagDanger:  "danger",
	DictTagInfo:    "info",
}

// DictTagImportMapping 标签类型映射
var DictTagImportMapping = map[string]DictTag{
	"primary": DictTagPrimary,
	"success": DictTagSuccess,
	"warning": DictTagWarning,
	"danger":  DictTagDanger,
	"info":    DictTagInfo,
}

// BoolValueImportMapping 布尔值类型映射
var BoolValueImportMapping = map[string]bool{
	"是": true,
	"否": false,
}
