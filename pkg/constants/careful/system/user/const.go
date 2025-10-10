/**
 * Description：
 * FileName：const.go
 * Author：CJiaの用心
 * Create：2025/10/9 15:59:17
 * Remark：
 */

package user

type GenderConst int

const (
	GenderConstMale   GenderConst = iota + 1 // 男
	GenderConstFemale                        // 女
	GenderConstSecret                        // 保密
)
