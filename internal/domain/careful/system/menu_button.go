/**
 * Description：
 * FileName：menu_button.go
 * Author：CJiaの用心
 * Create：2025/12/05 11:56:47
 * Remark：
 */

package system

import (
	"context"
	"github.com/carefuly/careful-admin-go-gin/internal/model/careful/system"
	"github.com/carefuly/careful-admin-go-gin/pkg/constants/careful/system/menu"
	"github.com/carefuly/careful-admin-go-gin/pkg/ginx/filters"
	"gorm.io/gorm"
)

type MenuButton struct {
	system.MenuButton

	CreateTime string `json:"createTime"` // 创建时间
	UpdateTime string `json:"updateTime"` // 更新时间
}

type MenuButtonFilter struct {
	filters.Filters
	filters.Pagination
	Status   bool        `json:"status"`   // 状态
	Title    string      `json:"title"`    // 按钮名称
	AuthMark string      `json:"authMark"` // 按钮权限值
	Method   menu.Method `json:"method"`   // 方法类型
	Api      string      `json:"api"`      // 接口地址
	MenuID   string      `json:"menu_id"`  // 关联菜单
}

func (f *MenuButtonFilter) QueryFilter(ctx context.Context, query *gorm.DB) *gorm.DB {
	query = f.Filters.QueryFilter(ctx, query).
		Where("status = ?", f.Status).
		Order("sort ASC, update_time DESC")

	if f.Title != "" {
		query = query.Where("title LIKE ?", "%"+f.Title+"%")
	}
	if f.AuthMark != "" {
		query = query.Where("authMark LIKE ?", "%"+f.AuthMark+"%")
	}
	if f.Method > 0 {
		query = query.Where("method = ?", f.Method)
	}
	if f.Api != "" {
		query = query.Where("api LIKE ?", "%"+f.Api+"%")
	}
	if f.MenuID != "" {
		query = query.Where("menu_id LIKE ?", "%"+f.MenuID+"%")
	}

	return query
}
