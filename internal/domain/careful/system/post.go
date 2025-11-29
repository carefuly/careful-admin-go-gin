/**
 * Description：
 * FileName：post.go
 * Author：CJiaの用心
 * Create：2025/11/29 01:26:15
 * Remark：
 */

package system

import (
	"context"
	"github.com/carefuly/careful-admin-go-gin/internal/model/careful/system"
	"github.com/carefuly/careful-admin-go-gin/pkg/constants/careful/system/post"
	"github.com/carefuly/careful-admin-go-gin/pkg/ginx/filters"
	"gorm.io/gorm"
)

type Post struct {
	system.Post

	CreateTime string `json:"createTime"` // 创建时间
	UpdateTime string `json:"updateTime"` // 更新时间
}

type PostFilter struct {
	filters.Filters
	filters.Pagination
	Status   bool       `json:"status"`    // 状态
	Name     string     `json:"name"`      // 岗位名称
	Code     string     `json:"code"`      // 岗位编码
	PostType post.Type  `json:"post_type"` // 岗位类型
	Level    post.Level `json:"level"`     // 岗位级别
	DeptID   string     `json:"dept_id"`   // 所属部门ID
}

func (f *PostFilter) QueryFilter(ctx context.Context, query *gorm.DB) *gorm.DB {
	query = f.Filters.QueryFilter(ctx, query).
		Where("status = ?", f.Status).
		Order("sort ASC, update_time DESC")

	if f.Name != "" {
		query = query.Where("name LIKE ?", "%"+f.Name+"%")
	}
	if f.Code != "" {
		query = query.Where("code LIKE ?", "%"+f.Code+"%")
	}
	if f.PostType > 0 {
		query = query.Where("post_type = ?", f.PostType)
	}
	if f.Level > 0 {
		query = query.Where("level = ?", f.Level)
	}
	if f.DeptID != "" {
		query = query.Where("dept_id = ?", f.DeptID)
	}

	return query
}
