/**
 * Description：
 * FileName：user.go
 * Author：CJiaの用心
 * Create：2025/11/24 16:59:03
 * Remark：
 */

package system

import "github.com/carefuly/careful-admin-go-gin/internal/model/careful/system"

type User struct {
	system.User

	CreateTime string `json:"createTime"` // 创建时间
	UpdateTime string `json:"updateTime"` // 更新时间
}
