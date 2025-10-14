/**
 * Description：
 * FileName：dict.go
 * Author：CJiaの用心
 * Create：2025/10/11 11:40:11
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
	ErrDictNotFound             = gorm.ErrRecordNotFound
	ErrDictCodeDuplicate        = errors.New("字典编码已存在")
	ErrDictNameDuplicate        = errors.New("字典名称已存在")
	ErrDictDuplicate            = errors.New("字典信息已存在")
	ErrDictVersionInconsistency = errors.New("数据已被修改，请刷新后重试")
)

type DictDAO interface {
	Insert(ctx context.Context, model tools.Dict) (*tools.Dict, error)
	Delete(ctx context.Context, id string) error
	BatchDelete(ctx context.Context, ids []string) error
	Update(ctx context.Context, model tools.Dict) error

	FindById(ctx context.Context, id string) (*tools.Dict, error)
	FindByName(ctx context.Context, name string) (*tools.Dict, error)
	FindListPage(ctx context.Context, filter domainTools.DictFilter) ([]*tools.Dict, int64, error)
	FindListAll(ctx context.Context, filter domainTools.DictFilter) ([]*tools.Dict, error)

	CheckExistByCode(ctx context.Context, code, excludeId string) (bool, error)
	CheckExistByName(ctx context.Context, name, excludeId string) (bool, error)
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
	result := dao.db.WithContext(ctx).Model(&model).
		Where("id = ? AND timestamp = ?", model.Id, model.Timestamp).
		Updates(map[string]any{
			"code":      model.Code,
			"sort":      model.Sort,
			"timestamp": time.Now().UnixMicro(),
			"status":    model.Status,
			"modifier":  model.Modifier,
			"remark":    model.Remark,
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

// FindByName 根据字典名称获取详情
func (dao *GORMDictDAO) FindByName(ctx context.Context, name string) (*tools.Dict, error) {
	var model tools.Dict
	err := dao.db.WithContext(ctx).Where("name = ?", name).First(&model).Error
	return &model, err
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

	query := dao.buildQuery(ctx, filter)

	// 查询
	if err := query.Find(&models).Error; err != nil {
		return nil, err
	}

	return models, nil
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

// CheckExistByCode 检查code是否存在
func (dao *GORMDictDAO) CheckExistByCode(ctx context.Context, code, excludeId string) (bool, error) {
	var model tools.Dict
	query := dao.db.WithContext(ctx).Model(&tools.Dict{}).
		Select("id"). // 只查询必要的字段
		Where("code = ?", code)

	if excludeId != "" {
		query = query.Where("id != ?", excludeId)
	}

	// 使用 LIMIT 1 快速判断是否存在
	err := query.Limit(1).First(&model).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil // 不存在
	}
	return err == nil, err // 存在或查询出错
}

// CheckExistByName 检查name是否存在
func (dao *GORMDictDAO) CheckExistByName(ctx context.Context, name, excludeId string) (bool, error) {
	var model tools.Dict
	query := dao.db.WithContext(ctx).Model(&tools.Dict{}).
		Select("id"). // 只查询必要的字段
		Where("name = ?", name)

	if excludeId != "" {
		query = query.Where("id != ?", excludeId)
	}

	// 使用 LIMIT 1 快速判断是否存在
	err := query.Limit(1).First(&model).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil // 不存在
	}
	return err == nil, err // 存在或查询出错
}
