/**
 * Description：
 * FileName：dict.go
 * Author：CJiaの用心
 * Create：2025/12/3 11:29:51
 * Remark：
 */

package tools

import (
	"context"
	"errors"
	"fmt"
	domainTools "github.com/carefuly/careful-admin-go-gin/internal/domain/careful/tools"
	"github.com/carefuly/careful-admin-go-gin/internal/model/careful/tools"
	"github.com/carefuly/careful-admin-go-gin/pkg/ginx/filters"
	"gorm.io/gorm"
	"time"
)

var (
	ErrDictNotFound             = gorm.ErrRecordNotFound
	ErrDictNameDuplicate        = errors.New("字典名称已存在")
	ErrDictCodeDuplicate        = errors.New("字典编码已存在")
	ErrDictDisabled             = errors.New("字典已被禁用，无法在其下创建字典项")
	ErrDictHasType              = errors.New("字典下仍有字典项，无法删除")
	ErrDictVersionInconsistency = errors.New("数据已被修改，请刷新后重试")
)

type DictDAO interface {
	Insert(ctx context.Context, model tools.Dict) (*tools.Dict, error)
	Delete(ctx context.Context, id string) error
	BatchDelete(ctx context.Context, ids []string) error
	Update(ctx context.Context, model tools.Dict) error

	FindById(ctx context.Context, id string) (*tools.Dict, error)
	FindByName(ctx context.Context, name string) (*tools.Dict, error)
	FindDictTypeCount(ctx context.Context, id string) (int64, error)
	FindListPage(ctx context.Context, filter domainTools.DictFilter) ([]*tools.Dict, int64, error)
	FindListAll(ctx context.Context, filter domainTools.DictFilter) ([]*tools.Dict, error)

	CheckExistByName(ctx context.Context, name, excludeId string) (bool, error)
	CheckExistByCode(ctx context.Context, code, excludeId string) (bool, error)
}

type GORMDictDAO struct {
	db *gorm.DB
}

func NewGORMDictDAO(db *gorm.DB) DictDAO {
	return &GORMDictDAO{
		db: db,
	}
}

// Insert 新增
func (dao *GORMDictDAO) Insert(ctx context.Context, model tools.Dict) (*tools.Dict, error) {
	return &model, dao.db.WithContext(ctx).Create(&model).Error
}

// Delete 删除
func (dao *GORMDictDAO) Delete(ctx context.Context, id string) error {
	return dao.db.WithContext(ctx).Where("id = ?", id).Delete(&tools.Dict{}).Error
}

// BatchDelete 批量删除
func (dao *GORMDictDAO) BatchDelete(ctx context.Context, ids []string) error {
	return dao.db.WithContext(ctx).Where("id IN ?", ids).Delete(&tools.Dict{}).Error
}

// Update 更新
func (dao *GORMDictDAO) Update(ctx context.Context, model tools.Dict) error {
	result := dao.db.WithContext(ctx).
		Model(&model).
		Where("id = ? AND timestamp = ?", model.Id, model.Timestamp).
		Updates(map[string]any{
			"status":      model.Status,
			"code":        model.Code,
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
			Model(&tools.Dict{}).
			Select("1").
			Where("id = ?", model.Id).
			Limit(1).
			Find(&exists)

		if !exists {
			return ErrDictNotFound
		}
		return ErrDictVersionInconsistency
	}
	return result.Error
}

// FindById 根据id获取详情
func (dao *GORMDictDAO) FindById(ctx context.Context, id string) (*tools.Dict, error) {
	var model tools.Dict
	err := dao.db.WithContext(ctx).Where("id = ?", id).First(&model).Error
	return &model, err
}

// FindByName 根据name获取详情
func (dao *GORMDictDAO) FindByName(ctx context.Context, name string) (*tools.Dict, error) {
	var model tools.Dict
	err := dao.db.WithContext(ctx).Where("name = ?", name).First(&model).Error
	return &model, err
}

// FindDictTypeCount 获取字典下的字典项数量
func (dao *GORMDictDAO) FindDictTypeCount(ctx context.Context, id string) (int64, error) {
	var count int64
	err := dao.db.WithContext(ctx).
		Model(&tools.DictType{}).
		Where("dict_id = ?", id).
		Count(&count).
		Error
	return count, err
}

// FindListPage 分页查询
func (dao *GORMDictDAO) FindListPage(ctx context.Context, filter domainTools.DictFilter) ([]*tools.Dict, int64, error) {
	var total int64
	var models []*tools.Dict

	query := dao.buildQuery(ctx, filter)

	err := query.Count(&total).
		Offset((filter.Page - 1) * filter.PageSize).
		Limit(filter.PageSize).
		Find(&models).Error

	return models, total, err
}

// FindListAll 获取所有列表
func (dao *GORMDictDAO) FindListAll(ctx context.Context, filter domainTools.DictFilter) ([]*tools.Dict, error) {
	var models []*tools.Dict
	err := dao.buildQuery(ctx, filter).Find(&models).Error
	return models, err
}

// buildQuery 构建查询条件
func (dao *GORMDictDAO) buildQuery(ctx context.Context, filter domainTools.DictFilter) *gorm.DB {
	builder := &domainTools.DictFilter{
		Filters: filters.Filters{
			Creator:    filter.Creator,
			Modifier:   filter.Modifier,
			BelongDept: filter.BelongDept,
		},
		Status:    filter.Status,
		Name:      filter.Name,
		Code:      filter.Code,
		Type:      filter.Type,
		ValueType: filter.ValueType,
	}
	return builder.QueryFilter(ctx, dao.db.WithContext(ctx).Model(&tools.Dict{}))
}

// CheckExistByName 检查name是否存在
func (dao *GORMDictDAO) CheckExistByName(ctx context.Context, name, excludeId string) (bool, error) {
	return dao.checkByField(ctx, "name", name, excludeId)
}

// CheckExistByCode 检查code是否存在
func (dao *GORMDictDAO) CheckExistByCode(ctx context.Context, code, excludeId string) (bool, error) {
	return dao.checkByField(ctx, "code", code, excludeId)
}

// checkByField 检查指定字段
func (dao *GORMDictDAO) checkByField(ctx context.Context, field string, value interface{}, excludeId string) (bool, error) {
	query := dao.db.WithContext(ctx).
		Model(&tools.Dict{}).
		Select("1"). // 只需检查存在性
		Where(fmt.Sprintf("%s = ?", field), value)

	if excludeId != "" {
		query = query.Where("id != ?", excludeId)
	}

	var exists bool
	err := query.Limit(1).Find(&exists).Error
	if err != nil {
		return false, err
	}
	return exists, nil
}
