/**
 * Description：
 * FileName：const.go
 * Author：CJiaの用心
 * Create：2025/11/24 15:27:08
 * Remark：
 */

package user

// Status 用户状态
type Status int

const (
	StatusNormal   Status = iota + 1 // 正常
	StatusDisabled                   // 禁用
	StatusLocked                     // 锁定
)

// String 返回状态的字符串表示
func (s Status) String() string {
	switch s {
	case StatusNormal:
		return "正常"
	case StatusDisabled:
		return "禁用"
	case StatusLocked:
		return "锁定"
	default:
		return "未知"
	}
}

// GenderConst 用户性别
type GenderConst int

const (
	GenderConstMale   GenderConst = iota + 1 // 男
	GenderConstFemale                        // 女
	GenderConstSecret                        // 保密
)

// GenderMapping 用户性别映射
var GenderMapping = map[GenderConst]string{
	GenderConstMale:   "男",
	GenderConstFemale: "女",
	GenderConstSecret: "保密",
}

// GenderImportMapping 用户性别映射
var GenderImportMapping = map[string]GenderConst{
	"男":  GenderConstMale,
	"女":  GenderConstFemale,
	"保密": GenderConstSecret,
}

// String 返回性别的字符串表示
func (g GenderConst) String() string {
	switch g {
	case GenderConstMale:
		return "男"
	case GenderConstFemale:
		return "女"
	case GenderConstSecret:
		return "保密"
	default:
		return "未知"
	}
}
