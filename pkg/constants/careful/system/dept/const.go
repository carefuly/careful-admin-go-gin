/**
 * Description：
 * FileName：const.go
 * Author：CJiaの用心
 * Create：2025/11/25 22:49:48
 * Remark：
 */

package dept

// Type 部门类型
type Type string

const (
	TypeCompany    Type = "company"    // 公司
	TypeDepartment Type = "department" // 部门
	TypeTeam       Type = "team"       // 小组
	TypeOther      Type = "other"      // 其他
)

// TypeMapping 部门类型映射
var TypeMapping = map[Type]string{
	TypeCompany:    "公司",
	TypeDepartment: "部门",
	TypeTeam:       "小组",
	TypeOther:      "其他",
}

// TypeImportMapping 部门类型映射
var TypeImportMapping = map[string]Type{
	"公司": TypeCompany,
	"部门": TypeDepartment,
	"小组": TypeTeam,
	"其他": TypeOther,
}

// // Type 返回部门类型的字符串表示
// func (t Type) String() string {
// 	switch t {
// 	case TypeCompany:
// 		return "公司"
// 	case TypeDepartment:
// 		return "部门"
// 	case TypeTeam:
// 		return "小组"
// 	case TypeOther:
// 		return "其他"
// 	default:
// 		return "未知"
// 	}
// }
