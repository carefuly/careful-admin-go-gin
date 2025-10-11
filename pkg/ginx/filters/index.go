/**
 * Description：
 * FileName：index.go
 * Author：CJiaの用心
 * Create：2025/10/11 11:32:30
 * Remark：
 */

package filters

import (
	"context"
	"gorm.io/gorm"
)

// QueryFiltersBuilder 查询构建器接口
type QueryFiltersBuilder interface {
	QueryFilter(ctx context.Context, query *gorm.DB) *gorm.DB
}

// Pagination 分页查询构建器
type Pagination struct {
	Page     int `json:"page"`     // 当前页
	PageSize int `json:"pageSize"` // 每页显示的条数
}

// Filters 基础查询构建器
type Filters struct {
	Creator    string `json:"creator"`    // 创建人
	Modifier   string `json:"modifier"`   // 修改人
	BelongDept string `json:"belongDept"` // 数据归属部门
}

func (f *Filters) QueryFilter(ctx context.Context, query *gorm.DB) *gorm.DB {
	// 进入后先查询权限
	// if f.Creator != "" {
	// 	query = query.Where("creator LIKE ?", "%"+f.Creator+"%")
	// }
	// if f.Modifier != "" {
	// 	query = query.Where("modifier LIKE ?", "%"+f.Modifier+"%")
	// }
	return query
}