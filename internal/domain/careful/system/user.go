/**
 * Description：
 * FileName：user.go
 * Author：CJiaの用心
 * Create：2025/10/9 16:13:33
 * Remark：
 */

package system

import "github.com/carefuly/careful-admin-go-gin/internal/model/careful/system"

type User struct {
	system.User
	DeptId     string `json:"dept_id"`    // 部门ID
	CreateTime string `json:"createTime"` // 创建时间
	UpdateTime string `json:"updateTime"` // 更新时间
}
