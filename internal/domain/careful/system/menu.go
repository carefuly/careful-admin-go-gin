/**
 * Description：
 * FileName：menu.go
 * Author：CJiaの用心
 * Create：2025/12/05 12:00:40
 * Remark：
 */

package system

import (
	"context"
	"github.com/carefuly/careful-admin-go-gin/internal/model/careful/system"
	"github.com/carefuly/careful-admin-go-gin/pkg/ginx/filters"
	"gorm.io/gorm"
)

type Menu struct {
	system.Menu

	Children   []Menu `json:"children"`   // 子菜单
	CreateTime string `json:"createTime"` // 创建时间
	UpdateTime string `json:"updateTime"` // 更新时间

	// Meta       map[string]any `json:"meta,omitempty"` // 元信息
	// ParentId   string         `json:"parent_id"`  // 上级菜单
}

type MenuFilter struct {
	filters.Filters
	filters.Pagination
	Status bool   `json:"status"` // 状态
	Title  string `json:"title"`  // 路由标题
}

func (f *MenuFilter) QueryFilter(ctx context.Context, query *gorm.DB) *gorm.DB {
	query = f.Filters.QueryFilter(ctx, query).
		Where("id != ?", "root").
		Where("status = ?", f.Status).
		Order("sort ASC, update_time DESC")

	if f.Title != "" {
		query = query.Where("title LIKE ?", "%"+f.Title+"%")
	}

	return query
}
