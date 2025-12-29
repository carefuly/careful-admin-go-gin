/**
 * Description：
 * FileName：dict.go
 * Author：CJiaの用心
 * Create：2025/12/2 14:58:58
 * Remark：
 */

package tools

import (
	"context"
	"github.com/carefuly/careful-admin-go-gin/internal/model/careful/tools"
	"github.com/carefuly/careful-admin-go-gin/pkg/constants/careful/tools/dict"
	"github.com/carefuly/careful-admin-go-gin/pkg/ginx/filters"
	"gorm.io/gorm"
)

type Dict struct {
	tools.Dict

	CreateTime string `json:"create_time"` // 创建时间
	UpdateTime string `json:"update_time"` // 更新时间
}

type DictFilter struct {
	filters.Filters
	filters.Pagination
	Status    bool           `json:"status"`     // 状态
	Name      string         `json:"name"`       // 字典名称
	Code      string         `json:"code"`       // 字典编码
	Type      dict.Type      `json:"type"`       // 字典类型
	ValueType dict.ValueType `json:"value_type"` // 数据类型
}

func (f *DictFilter) QueryFilter(ctx context.Context, query *gorm.DB) *gorm.DB {
	query = f.Filters.QueryFilter(ctx, query).
		Where("status = ?", f.Status).
		Order("sort ASC").
		Order("update_time DESC")

	if f.Name != "" {
		query = query.Where("name LIKE ?", "%"+f.Name+"%")
	}
	if f.Code != "" {
		query = query.Where("code LIKE ?", "%"+f.Code+"%")
	}
	if f.Type > 0 {
		query = query.Where("type = ?", f.Type)
	}
	if f.ValueType > 0 {
		query = query.Where("value_type = ?", f.ValueType)
	}

	return query
}
