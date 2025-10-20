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
	"github.com/carefuly/careful-admin-go-gin/pkg/ginx/filters"
	"gorm.io/gorm"
)

type DictType struct {
	tools.DictType
	Label      string `json:"label"`      // 名称
	Value      any    `json:"value"`      // 值
	StrValue   string `json:"strValue"`   // 字符串-字典信息值
	IntValue   int64  `json:"intValue"`   // 整型-字典信息值
	BoolValue  bool   `json:"boolValue"`  // 布尔-字典信息值
	CreateTime string `json:"createTime"` // 创建时间
	UpdateTime string `json:"updateTime"` // 更新时间
}

type DictTypeFilter struct {
	filters.Filters
	filters.Pagination
	Status    bool   `json:"status"`    // 状态
	Name      string `json:"name"`      // 字典信息名称
	DictTag   string `json:"dictTag"`   // 标签类型
	DictName  string `json:"dictName"`  // 数据字典名称
	ValueType int    `json:"valueType"` // 数据类型
	DictId    string `json:"dict_id"`   // 字典ID
}

func (f *DictTypeFilter) QueryFilter(ctx context.Context, query *gorm.DB) *gorm.DB {
	query = f.Filters.QueryFilter(ctx, query).
		Where("status = ?", f.Status).
		Order("sort ASC, update_time DESC")

	if f.Name != "" {
		query = query.Where("name LIKE ?", "%"+f.Name+"%")
	}
	if f.DictTag != "" {
		query = query.Where("dictTag LIKE ?", "%"+f.DictTag+"%")
	}
	if f.DictName != "" {
		query = query.Where("dictName LIKE ?", "%"+f.DictName+"%")
	}
	if f.ValueType > 0 {
		query = query.Where("valueType = ?", f.ValueType)
	}
	if f.DictId != "" {
		query = query.Where("dict_id = ?", f.DictId)
	}

	return query
}
