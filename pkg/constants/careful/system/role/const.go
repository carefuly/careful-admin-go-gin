/**
 * Description：
 * FileName：const.go
 * Author：CJiaの用心
 * Create：2025/12/19 21:48:48
 * Remark：
 */

package role

// DataScope 数据权限范围
type DataScope int

const (
	DataScopeOnly    DataScope = iota + 1 // 仅本人数据
	DataScopeDept                         // 本部门数据
	DataScopeDeptSub                      // 本部门及下级部门数据
	DataScopeAll                          // 全部数据
	DataScopeCustom                       // 自定义数据
)

// DataScopeMapping 数据权限范围映射
var DataScopeMapping = map[DataScope]string{
	DataScopeOnly:    "仅本人数据",
	DataScopeDept:    "本部门数据",
	DataScopeDeptSub: "本部门及下级部门数据",
	DataScopeAll:     "全部数据",
	DataScopeCustom:  "自定义数据",
}

// DataScopeImportMapping 数据权限范围映射
var DataScopeImportMapping = map[string]DataScope{
	"仅本人数据":      DataScopeOnly,
	"本部门数据":      DataScopeDept,
	"本部门及下级部门数据": DataScopeDeptSub,
	"全部数据":       DataScopeAll,
	"自定义数据":      DataScopeCustom,
}
