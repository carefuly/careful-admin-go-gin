/**
 * Description：
 * FileName：role.go
 * Author：CJiaの用心
 * Create：2025/12/19 22:02:27
 * Remark：
 */

package system

import (
	"context"
	"github.com/carefuly/careful-admin-go-gin/internal/model/careful/system"
	"github.com/carefuly/careful-admin-go-gin/pkg/constants/careful/system/role"
	"github.com/carefuly/careful-admin-go-gin/pkg/ginx/filters"
	"gorm.io/gorm"
)

type Role struct {
	system.Role

	CreateTime string `json:"create_time"` // 创建时间
	UpdateTime string `json:"update_time"` // 更新时间
}

type RoleFilter struct {
	filters.Filters
	filters.Pagination
	Status    bool           `json:"status"`     // 状态
	Name      string         `json:"name"`       // 角色名称
	Code      string         `json:"code"`       // 角色编码
	DataScope role.DataScope `json:"data_scope"` // 数据权限范围
}

func (f *RoleFilter) QueryFilter(ctx context.Context, query *gorm.DB) *gorm.DB {
	query = f.Filters.QueryFilter(ctx, query).
		Where("status = ?", f.Status).
		Order("sort ASC, update_time DESC")

	if f.Name != "" {
		query = query.Where("name LIKE ?", "%"+f.Name+"%")
	}
	if f.Code != "" {
		query = query.Where("code LIKE ?", "%"+f.Code+"%")
	}
	if f.DataScope > 0 {
		query = query.Where("data_scope = ?", f.DataScope)
	}

	return query
}
