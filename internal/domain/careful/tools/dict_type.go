/**
 * Description：
 * FileName：dict_type.go
 * Author：CJiaの用心
 * Create：2025/10/19 14:48:14
 * Remark：
 */

package tools

import (
	"context"
	"github.com/carefuly/careful-admin-go-gin/internal/model/careful/tools"
	"github.com/carefuly/careful-admin-go-gin/pkg/constants/careful/tools/dict"
	"github.com/carefuly/careful-admin-go-gin/pkg/constants/careful/tools/dict_type"
	"github.com/carefuly/careful-admin-go-gin/pkg/ginx/filters"
	"gorm.io/gorm"
)

type DictType struct {
	tools.DictType

	Label      string `json:"label"`       // 名称
	Value      any    `json:"value"`       // 值
	StrValue   string `json:"str_value"`   // 字符串-字典项值
	IntValue   int64  `json:"int_value"`   // 整型-字典项值
	BoolValue  bool   `json:"bool_value"`  // 布尔-字典项值
	CreateTime string `json:"create_time"` // 创建时间
	UpdateTime string `json:"update_time"` // 更新时间
}

type DictTypeFilter struct {
	filters.Filters
	filters.Pagination
	Status    bool              `json:"status"`     // 状态
	Name      string            `json:"name"`       // 字典项名称
	DictTag   dict_type.DictTag `json:"dict_tag"`   // 标签类型
	DictName  string            `json:"dict_name"`  // 字典名称
	ValueType dict.ValueType    `json:"value_type"` // 数据类型
	DictID    string            `json:"dict_id"`    // 所属字典ID
}

func (f *DictTypeFilter) QueryFilter(ctx context.Context, query *gorm.DB) *gorm.DB {
	query = f.Filters.QueryFilter(ctx, query).
		Where("status = ?", f.Status).
		Order("sort ASC").
		Order("update_time DESC")

	if f.Name != "" {
		query = query.Where("name LIKE ?", "%"+f.Name+"%")
	}
	if f.DictTag != "" {
		query = query.Where("dict_tag LIKE ?", "%"+f.DictTag+"%")
	}
	if f.DictName != "" {
		query = query.Where("dict_name = ?", f.DictName)
	}
	if f.ValueType > 0 {
		query = query.Where("value_type = ?", f.ValueType)
	}
	if f.DictID != "" {
		query = query.Where("dict_id = ?", f.DictID)
	}

	return query
}
