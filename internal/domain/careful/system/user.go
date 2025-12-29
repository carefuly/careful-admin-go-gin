/**
 * Description：
 * FileName：user.go
 * Author：CJiaの用心
 * Create：2025/11/24 16:59:03
 * Remark：
 */

package system

import (
	"context"
	"github.com/carefuly/careful-admin-go-gin/internal/model/careful/system"
	"github.com/carefuly/careful-admin-go-gin/pkg/ginx/filters"
	"gorm.io/gorm"
)

type User struct {
	system.User

	Buttons    []string `json:"buttons"`     // 按钮列表
	PostIDs    []string `json:"post_ids"`    // 岗位ids
	RoleIDs    []string `json:"role_ids"`    // 角色ids
	CreateTime string   `json:"create_time"` // 创建时间
	UpdateTime string   `json:"update_time"` // 更新时间
}

type UserFilter struct {
	filters.Filters
	filters.Pagination
	Status   bool   `json:"status"`   // 状态
	Username string `json:"username"` // 用户名
	Email    string `json:"email"`    // 邮箱
	Mobile   string `json:"mobile"`   // 手机号
}

func (f *UserFilter) QueryFilter(ctx context.Context, query *gorm.DB) *gorm.DB {
	query = f.Filters.QueryFilter(ctx, query).
		Where("status = ?", f.Status).
		Order("sort ASC, update_time DESC")

	if f.Username != "" {
		query = query.Where("username LIKE ?", "%"+f.Username+"%")
	}
	if f.Email != "" {
		query = query.Where("email LIKE ?", "%"+f.Email+"%")
	}
	if f.Mobile != "" {
		query = query.Where("mobile LIKE ?", "%"+f.Mobile+"%")
	}

	return query
}
