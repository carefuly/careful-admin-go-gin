/**
 * Description：
 * FileName：dict_type.go
 * Author：CJiaの用心
 * Create：2025/10/19 14:56:46
 * Remark：
 */

package tools

import (
	"context"
	"errors"
	domainTools "github.com/carefuly/careful-admin-go-gin/internal/domain/careful/tools"
	"github.com/carefuly/careful-admin-go-gin/internal/model/careful/tools"
	"github.com/carefuly/careful-admin-go-gin/pkg/ginx/filters"
	"gorm.io/gorm"
	"time"
)

var (
	ErrDictTypeNotFound             = gorm.ErrRecordNotFound
	ErrDictTypeInvalidDictValueType = tools.ErrDictTypeInvalidDictValueType
	ErrDictTypeDuplicate            = tools.ErrDictTypeUniqueIndex
	ErrDictTypeVersionInconsistency = errors.New("数据已被修改，请刷新后重试")
)

type DictTypeDAO interface {
	Insert(ctx context.Context, model tools.DictType) (*tools.DictType, error)
	Delete(ctx context.Context, id string) error
	BatchDelete(ctx context.Context, ids []string) error
	Update(ctx context.Context, model tools.DictType) error

	FindById(ctx context.Context, id string) (*tools.DictType, error)
	FindListPage(ctx context.Context, filter domainTools.DictTypeFilter) ([]*tools.DictType, int64, error)
	FindListAll(ctx context.Context, filter domainTools.DictTypeFilter) ([]*tools.DictType, error)
}

type GORMDictTypeDAO struct {
	db *gorm.DB
}

func NewGORMDictTypeDAO(db *gorm.DB) DictTypeDAO {
	return &GORMDictTypeDAO{db: db}
}

// Insert 新增
func (dao *GORMDictTypeDAO) Insert(ctx context.Context, model tools.DictType) (*tools.DictType, error) {
	return &model, dao.db.WithContext(ctx).Create(&model).Error
}

// Delete 删除
func (dao *GORMDictTypeDAO) Delete(ctx context.Context, id string) error {
	return dao.db.WithContext(ctx).Where("id = ?", id).Delete(&tools.DictType{}).Error
}

// BatchDelete 批量删除
func (dao *GORMDictTypeDAO) BatchDelete(ctx context.Context, ids []string) error {
	return dao.db.WithContext(ctx).Where("id IN ?", ids).Delete(&tools.DictType{}).Error
}

// Update 更新
func (dao *GORMDictTypeDAO) Update(ctx context.Context, model tools.DictType) error {
	result := dao.db.WithContext(ctx).Model(&model).
		Where("id = ? AND timestamp = ?", model.Id, model.Timestamp).
		Updates(map[string]any{
			"status":      model.Status,
			"name":        model.Name,
			"dict_tag":    model.DictTag,
			"dict_color":  model.DictColor,
			"description": model.Description,
			"sort":        model.Sort,
			"timestamp":   time.Now().UnixMicro(),
			"modifier":    model.Modifier,
			"remark":      model.Remark,
		})
	// 处理行影响数为0的情况
	if result.RowsAffected == 0 {
		// 先检查记录是否存在
		var exists bool
		dao.db.WithContext(ctx).
			Model(&tools.DictType{}).
			Select("1").
			Where("id = ?", model.Id).
			Limit(1).
			Find(&exists)

		if !exists {
			return ErrDictTypeNotFound
		}
		return ErrDictTypeVersionInconsistency
	}
	return result.Error
}

// FindById 根据id获取详情
func (dao *GORMDictTypeDAO) FindById(ctx context.Context, id string) (*tools.DictType, error) {
	var model tools.DictType
	err := dao.db.WithContext(ctx).
		Preload("Dict").
		Where("id = ?", id).
		First(&model).
		Error
	return &model, err
}

// FindListPage 分页查询
func (dao *GORMDictTypeDAO) FindListPage(ctx context.Context, filter domainTools.DictTypeFilter) ([]*tools.DictType, int64, error) {
	var total int64
	var models []*tools.DictType

	query := dao.buildQuery(ctx, filter)

	err := query.Count(&total).
		Offset((filter.Page - 1) * filter.PageSize).
		Limit(filter.PageSize).
		Find(&models).Error

	return models, total, err
}

// FindListAll 查询所有字典
func (dao *GORMDictTypeDAO) FindListAll(ctx context.Context, filter domainTools.DictTypeFilter) ([]*tools.DictType, error) {
	var models []*tools.DictType
	err := dao.buildQuery(ctx, filter).Find(&models).Error
	return models, err
}

// buildQuery 构建查询条件
func (dao *GORMDictTypeDAO) buildQuery(ctx context.Context, filter domainTools.DictTypeFilter) *gorm.DB {
	builder := &domainTools.DictTypeFilter{
		Filters: filters.Filters{
			Creator:    filter.Creator,
			Modifier:   filter.Modifier,
			BelongDept: filter.BelongDept,
		},
		Status:    filter.Status,
		Name:      filter.Name,
		DictTag:   filter.DictTag,
		DictName:  filter.DictName,
		ValueType: filter.ValueType,
		DictID:    filter.DictID,
	}
	return builder.QueryFilter(ctx, dao.db.WithContext(ctx).Model(&tools.DictType{}))
}
