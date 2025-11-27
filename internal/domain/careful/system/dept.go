/**
 * Description：
 * FileName：dept.go
 * Author：CJiaの用心
 * Create：2025/11/26 02:30:34
 * Remark：
 */

package system

import (
	"context"
	"github.com/carefuly/careful-admin-go-gin/internal/model/careful/system"
	"github.com/carefuly/careful-admin-go-gin/pkg/constants/careful/system/dept"
	"github.com/carefuly/careful-admin-go-gin/pkg/ginx/filters"
	"gorm.io/gorm"
)

type Dept struct {
	system.Dept

	CreateTime string `json:"createTime"` // 创建时间
	UpdateTime string `json:"updateTime"` // 更新时间
}

type DeptFilter struct {
	filters.Filters
	filters.Pagination
	Status   bool      `json:"status"`    // 状态
	Name     string    `json:"name"`      // 部门名称
	Code     string    `json:"code"`      // 部门编码
	DeptType dept.Type `json:"dept_type"` // 部门类型
	Level    int       `json:"level"`     // 层级深度
}

func (f *DeptFilter) QueryFilter(ctx context.Context, query *gorm.DB) *gorm.DB {
	query = f.Filters.QueryFilter(ctx, query).
		Where("id != ?", "root").
		Where("status = ?", f.Status).
		Order("sort ASC, update_time DESC")

	if f.Name != "" {
		query = query.Where("name LIKE ?", "%"+f.Name+"%")
	}
	if f.Code != "" {
		query = query.Where("code LIKE ?", "%"+f.Code+"%")
	}
	if f.DeptType != "" {
		query = query.Where("dept_type LIKE ?", "%"+f.DeptType+"%")
	}
	if f.Level > 0 {
		query = query.Where("level = ?", f.Level)
	}

	return query
}
