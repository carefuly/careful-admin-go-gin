/**
 * Description：
 * FileName：dept.go
 * Author：CJiaの用心
 * Create：2025/11/26 09:03:36
 * Remark：
 */

package system

import (
	"context"
	"errors"
	domainSystem "github.com/carefuly/careful-admin-go-gin/internal/domain/careful/system"
	"github.com/carefuly/careful-admin-go-gin/internal/model/careful/system"
	"github.com/carefuly/careful-admin-go-gin/pkg/ginx/filters"
	"gorm.io/gorm"
	"time"
)

var (
	ErrDeptNotFound             = gorm.ErrRecordNotFound
	ErrDeptCodeDuplicate        = errors.New("部门编码已存在")
	ErrDeptNameParentDuplicate  = errors.New("同级别下已存在相同的部门信息")
	ErrDeptParentNotFound       = gorm.ErrRecordNotFound
	ErrDeptDisabled             = errors.New("父部门已被禁用，无法在其下创建子部门")
	ErrDeptHasChildren          = errors.New("部门含有子部门，无法删除")
	ErrDeptHasUsers             = errors.New("部门下仍有用户，无法删除")
	ErrDeptVersionInconsistency = errors.New("数据已被修改，请刷新后重试")
)

type DeptDAO interface {
	Insert(ctx context.Context, model system.Dept) (*system.Dept, error)
	Delete(ctx context.Context, id string) error
	BatchDelete(ctx context.Context, ids []string) error
	Update(ctx context.Context, model system.Dept) error

	FindById(ctx context.Context, id string) (*system.Dept, error)
	FindByParentId(ctx context.Context, parentId string) (*system.Dept, error)
	FindUserCount(ctx context.Context, id string) (int64, error)
	FindChildCount(ctx context.Context, id string) (int64, error)
	FindAncestors(ctx context.Context, model system.Dept) ([]*system.Dept, error)
	FindListAll(ctx context.Context, filter domainSystem.DeptFilter) ([]*system.Dept, error)

	CheckExistByCode(ctx context.Context, code, excludeId string) (bool, error)
	CheckExistByNameAndParentId(ctx context.Context, name, parentId, excludeId string) (bool, error)
}

type GORMDeptDAO struct {
	db *gorm.DB
}

func NewGORMDeptDAO(db *gorm.DB) DeptDAO {
	return &GORMDeptDAO{
		db: db,
	}
}

// Insert 新增
func (dao *GORMDeptDAO) Insert(ctx context.Context, model system.Dept) (*system.Dept, error) {
	return &model, dao.db.WithContext(ctx).Create(&model).Error
}

// Delete 删除
func (dao *GORMDeptDAO) Delete(ctx context.Context, id string) error {
	return dao.db.WithContext(ctx).Where("id = ?", id).Delete(&system.Dept{}).Error
}

// BatchDelete 批量删除
func (dao *GORMDeptDAO) BatchDelete(ctx context.Context, ids []string) error {
	return dao.db.WithContext(ctx).Where("id IN ?", ids).Delete(&system.Dept{}).Error
}

// Update 更新
func (dao *GORMDeptDAO) Update(ctx context.Context, model system.Dept) error {
	result := dao.db.WithContext(ctx).Model(&model).
		Where("id = ? AND timestamp = ?", model.Id, model.Timestamp).
		Updates(map[string]any{
			"status":      model.Status,
			"name":        model.Name,
			"code":        model.Code,
			"dept_type":   model.DeptType,
			"owner":       model.Owner,
			"phone":       model.Phone,
			"email":       model.Email,
			"description": model.Description,
			"parent_id":   model.ParentID,
			"level":       model.Level,
			"path":        model.Path,
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
			Model(&system.Dept{}).
			Select("1").
			Where("id = ?", model.Id).
			Limit(1).
			Find(&exists)

		if !exists {
			return ErrDeptNotFound
		}
		return ErrDeptVersionInconsistency
	}
	return result.Error
}

// FindById 根据id获取详情
func (dao *GORMDeptDAO) FindById(ctx context.Context, id string) (*system.Dept, error) {
	var model system.Dept
	err := dao.db.WithContext(ctx).
		Preload("Parent").
		Where("id = ?", id).
		First(&model).
		Error
	return &model, err
}

// FindByParentId 根据parent_id获取详情
func (dao *GORMDeptDAO) FindByParentId(ctx context.Context, parentId string) (*system.Dept, error) {
	var model system.Dept
	err := dao.db.WithContext(ctx).Where("id = ?", parentId).First(&model).Error
	return &model, err
}

// FindUserCount 获取部门下的用户数量
func (dao *GORMDeptDAO) FindUserCount(ctx context.Context, id string) (int64, error) {
	var userCount int64
	err := dao.db.WithContext(ctx).
		Model(&system.Dept{}).
		Where("dept_id = ?", id).
		Count(&userCount).
		Error
	return userCount, err
}

// FindChildCount 获取子部门数量
func (dao *GORMDeptDAO) FindChildCount(ctx context.Context, id string) (int64, error) {
	var count int64
	err := dao.db.WithContext(ctx).
		Model(&system.Dept{}).
		Where("parent_id = ?", id).
		Count(&count).
		Error
	return count, err
}

// FindAncestors 获取所有祖先部门
func (dao *GORMDeptDAO) FindAncestors(ctx context.Context, model system.Dept) ([]*system.Dept, error) {
	var ancestors []*system.Dept
	currentID := model.ParentID
	for currentID != nil {
		var parent *system.Dept
		if err := dao.db.WithContext(ctx).First(&parent, currentID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				break // 防止因数据不一致导致死循环
			}
			return nil, err
		}
		ancestors = append([]*system.Dept{parent}, ancestors...) //  prepend
		currentID = parent.ParentID
	}
	return ancestors, nil
}

// FindListAll 获取所有列表
func (dao *GORMDeptDAO) FindListAll(ctx context.Context, filter domainSystem.DeptFilter) ([]*system.Dept, error) {
	var models []*system.Dept

	query := dao.buildQuery(ctx, filter)

	if err := query.Find(&models).Error; err != nil {
		return nil, err
	}

	return models, nil
}

// buildQuery 构建查询条件
func (dao *GORMDeptDAO) buildQuery(ctx context.Context, filter domainSystem.DeptFilter) *gorm.DB {
	builder := &domainSystem.DeptFilter{
		Filters: filters.Filters{
			Creator:    filter.Creator,
			Modifier:   filter.Modifier,
			BelongDept: filter.BelongDept,
		},
		Status:   filter.Status,
		Name:     filter.Name,
		Code:     filter.Code,
		DeptType: filter.DeptType,
		Level:    filter.Level,
	}
	return builder.QueryFilter(ctx, dao.db.WithContext(ctx).Model(&system.Dept{}))
}

// CheckExistByCode 检查code是否存在
func (dao *GORMDeptDAO) CheckExistByCode(ctx context.Context, code, excludeId string) (bool, error) {
	var model system.Dept
	query := dao.db.WithContext(ctx).Model(&system.Dept{}).
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

// CheckExistByNameAndParentId 检查name和parent_id是否存在
func (dao *GORMDeptDAO) CheckExistByNameAndParentId(ctx context.Context, name, parentId, excludeId string) (bool, error) {
	var model system.Dept
	query := dao.db.WithContext(ctx).Model(&system.Dept{}).
		Select("id"). // 只查询必要的字段
		Where("name = ? AND parent_id = ?", name, parentId)

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
