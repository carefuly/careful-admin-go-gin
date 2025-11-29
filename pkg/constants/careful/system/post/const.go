/**
 * Description：
 * FileName：const.go
 * Author：CJiaの用心
 * Create：2025/11/29 01:10:23
 * Remark：
 */

package post

// Type 岗位类型
type Type int

const (
	TypeManagement Type = iota + 1 // 管理岗
	TypeTechnology                 // 技术岗
	TypeBusiness                   // 业务岗
	TypeFunctional                 // 职能岗
	TypeOther                      // 其他
)

// TypeMapping 岗位类型映射
var TypeMapping = map[Type]string{
	TypeManagement: "管理岗",
	TypeTechnology: "技术岗",
	TypeBusiness:   "业务岗",
	TypeFunctional: "职能岗",
	TypeOther:      "其他",
}

// TypeImportMapping 岗位类型映射
var TypeImportMapping = map[string]Type{
	"管理岗": TypeManagement,
	"技术岗": TypeTechnology,
	"业务岗": TypeBusiness,
	"职能岗": TypeFunctional,
	"其他":  TypeOther,
}

// Level 岗位级别
type Level int

const (
	LevelHigh         Level = iota + 1 // 高层
	LevelMiddleLayer                   // 中层
	LevelGrassroots                    // 基层
	LevelGeneralStaff                  // 一般员工
)

// LevelMapping 岗位级别映射
var LevelMapping = map[Level]string{
	LevelHigh:         "高层",
	LevelMiddleLayer:  "中层",
	LevelGrassroots:   "基层",
	LevelGeneralStaff: "一般员工",
}

// LevelImportMapping 岗位级别映射
var LevelImportMapping = map[string]Level{
	"高层":   LevelHigh,
	"中层":   LevelMiddleLayer,
	"基层":   LevelGrassroots,
	"一般员工": LevelGeneralStaff,
}
