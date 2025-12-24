/**
 * Description：
 * FileName：role.go
 * Author：CJiaの用心
 * Create：2025/12/19 22:05:22
 * Remark：
 */

package system

import (
	"context"
	"errors"
	domainSystem "github.com/carefuly/careful-admin-go-gin/internal/domain/careful/system"
	"github.com/carefuly/careful-admin-go-gin/internal/model/careful/system"
	"github.com/carefuly/careful-admin-go-gin/pkg/ginx/filters"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"time"
)

var (
	ErrRoleNotFound             = gorm.ErrRecordNotFound
	ErrRoleCodeDuplicate        = errors.New("角色编码已存在")
	ErrRoleVersionInconsistency = errors.New("数据已被修改，请刷新后重试")
)

type RoleDAO interface {
	Insert(ctx context.Context, model system.Role) (*system.Role, error)
	Delete(ctx context.Context, id string) error
	BatchDelete(ctx context.Context, ids []string) error
	Update(ctx context.Context, model system.Role) error

	FindById(ctx context.Context, id string) (*system.Role, error)
	FindListPage(ctx context.Context, filter domainSystem.RoleFilter) ([]*system.Role, int64, error)
	FindListAll(ctx context.Context, filter domainSystem.RoleFilter) ([]*system.Role, error)

	CheckExistByCode(ctx context.Context, code, excludeId string) (bool, error)
}

type GORMRoleDAO struct {
	db           *gorm.DB
	deptDb       DeptDAO
	menuDb       MenuDAO
	menuButtonDb MenuButtonDAO
}

func NewGORMRoleDAO(db *gorm.DB, deptDb DeptDAO, menuDb MenuDAO, menuButtonDb MenuButtonDAO) RoleDAO {
	return &GORMRoleDAO{
		db:           db,
		deptDb:       deptDb,
		menuDb:       menuDb,
		menuButtonDb: menuButtonDb,
	}
}

// Insert 新增
func (dao *GORMRoleDAO) Insert(ctx context.Context, model system.Role) (*system.Role, error) {
	return &model, dao.db.WithContext(ctx).Create(&model).Error
}

// Delete 删除
func (dao *GORMRoleDAO) Delete(ctx context.Context, id string) error {
	return dao.db.WithContext(ctx).Where("id = ?", id).Delete(&system.Role{}).Error
}

// BatchDelete 批量删除
func (dao *GORMRoleDAO) BatchDelete(ctx context.Context, ids []string) error {
	return dao.db.WithContext(ctx).Where("id IN ?", ids).Delete(&system.Role{}).Error
}

// Update 更新
func (dao *GORMRoleDAO) Update(ctx context.Context, model system.Role) error {
	// 开启事务
	tx := dao.db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	result := dao.db.WithContext(ctx).Model(&model).
		Where("id = ? AND timestamp = ?", model.Id, model.Timestamp).
		Updates(map[string]any{
			"status":      model.Status,
			"name":        model.Name,
			"code":        model.Code,
			"data_scope":  model.DataScope,
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
			Model(&system.Role{}).
			Select("1").
			Where("id = ?", model.Id).
			Limit(1).
			Find(&exists)

		if !exists {
			return ErrRoleNotFound
		}
		return ErrRoleVersionInconsistency
	}

	// 更新关联关系
	if err := dao.updateAssociations(tx, ctx, model); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

// updateAssociations 更新所有关联关系
func (dao *GORMRoleDAO) updateAssociations(tx *gorm.DB, ctx context.Context, role system.Role) error {
	// 更新部门关联
	// 删除旧关联
	if err := tx.Exec("DELETE FROM careful_system_role_dept WHERE role_id = ?", role.Id).Error; err != nil {
		zap.S().Error("删除部门关联异常：", err)
		return err
	}
	for _, id := range role.DeptIDs {
		dept, err := dao.deptDb.FindById(ctx, id)
		if err != nil {
			continue
		}
		if err := tx.Exec("INSERT INTO careful_system_role_dept (role_id, dept_id) VALUES (?, ?)",
			role.Id, dept.Id).Error; err != nil {
			zap.S().Error("更新部门关联异常：", err)
			return err
		}
	}

	// 更新菜单关联
	// 删除旧关联
	if err := tx.Exec("DELETE FROM careful_system_role_menu WHERE role_id = ?", role.Id).Error; err != nil {
		zap.S().Error("删除菜单关联异常：", err)
		return err
	}
	for _, id := range role.MenuIDs {
		menu, err := dao.menuDb.FindById(ctx, id)
		if err != nil {
			continue
		}
		if err := tx.Exec("INSERT INTO careful_system_role_menu (role_id, menu_id) VALUES (?, ?)",
			role.Id, menu.Id).Error; err != nil {
			zap.S().Error("更新菜单关联异常：", err)
			return err
		}
	}
	// 更新菜单按钮关联
	if err := tx.Exec("DELETE FROM careful_system_role_menu_button WHERE role_id = ?", role.Id).Error; err != nil {
		zap.S().Error("删除菜单按钮关联异常：", err)
		return err
	}
	for _, id := range role.MenuButtonIDs {
		menuButton, err := dao.menuButtonDb.FindById(ctx, id)
		if err != nil {
			continue
		}
		if err := tx.Exec("INSERT INTO careful_system_role_menu_button (role_id, menu_button_id) VALUES (?, ?)",
			role.Id, menuButton.Id).Error; err != nil {
			zap.S().Error("更新菜单按钮关联异常：", err)
			return err
		}
	}

	return nil
}

// FindById 根据id获取详情
func (dao *GORMRoleDAO) FindById(ctx context.Context, id string) (*system.Role, error) {
	var model system.Role
	err := dao.db.WithContext(ctx).
		Preload("Dept").
		Preload("Menu").
		Preload("MenuButton").
		Where("id = ?", id).
		First(&model).
		Error
	return &model, err
}

// FindListPage 分页查询
func (dao *GORMRoleDAO) FindListPage(ctx context.Context, filter domainSystem.RoleFilter) ([]*system.Role, int64, error) {
	var total int64
	var models []*system.Role

	query := dao.buildQuery(ctx, filter)

	err := query.Count(&total).
		Offset((filter.Page - 1) * filter.PageSize).
		Limit(filter.PageSize).
		Find(&models).Error

	return models, total, err
}

// FindListAll 获取所有列表
func (dao *GORMRoleDAO) FindListAll(ctx context.Context, filter domainSystem.RoleFilter) ([]*system.Role, error) {
	var models []*system.Role

	query := dao.buildQuery(ctx, filter)

	if err := query.Find(&models).Error; err != nil {
		return nil, err
	}

	return models, nil
}

// buildQuery 构建查询条件
func (dao *GORMRoleDAO) buildQuery(ctx context.Context, filter domainSystem.RoleFilter) *gorm.DB {
	builder := &domainSystem.RoleFilter{
		Filters: filters.Filters{
			Creator:    filter.Creator,
			Modifier:   filter.Modifier,
			BelongDept: filter.BelongDept,
		},
		Status:    filter.Status,
		Name:      filter.Name,
		Code:      filter.Code,
		DataScope: filter.DataScope,
	}
	return builder.QueryFilter(ctx, dao.db.WithContext(ctx).Model(&system.Role{}))
}

// CheckExistByCode 检查code是否存在
func (dao *GORMRoleDAO) CheckExistByCode(ctx context.Context, code, excludeId string) (bool, error) {
	var model system.Role
	query := dao.db.WithContext(ctx).Model(&system.Role{}).
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
