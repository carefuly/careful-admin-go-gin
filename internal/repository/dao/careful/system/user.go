/**
 * Description：
 * FileName：user.go
 * Author：CJiaの用心
 * Create：2025/11/24 17:09:15
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
	ErrUserNotFound             = gorm.ErrRecordNotFound
	ErrUserUsernameDuplicate    = errors.New("用户名已存在")
	ErrUserVersionInconsistency = errors.New("数据已被修改，请刷新后重试")
)

type UserDAO interface {
	Insert(ctx context.Context, model system.User) (*system.User, error)
	Delete(ctx context.Context, id string) error
	BatchDelete(ctx context.Context, ids []string) error
	Update(ctx context.Context, model system.User) error
	UpdateLoginField(ctx context.Context, lastLogin string, lastLoginIp string, model system.User) error

	FindById(ctx context.Context, id string) (*system.User, error)
	FindByUsername(ctx context.Context, username string) (*system.User, error)
	FindListPage(ctx context.Context, filter domainSystem.UserFilter) ([]*system.User, int64, error)
	FindListAll(ctx context.Context, filter domainSystem.UserFilter) ([]*system.User, error)
	CheckExistByUsername(ctx context.Context, username, excludeId string) (bool, error)
}

type GORMUserDAO struct {
	db *gorm.DB
}

func NewGORMUserDAO(db *gorm.DB) UserDAO {
	return &GORMUserDAO{
		db: db,
	}
}

// Insert 新增
func (dao *GORMUserDAO) Insert(ctx context.Context, model system.User) (*system.User, error) {
	return &model, dao.db.WithContext(ctx).Create(&model).Error
}

// Delete 删除
func (dao *GORMUserDAO) Delete(ctx context.Context, id string) error {
	return dao.db.WithContext(ctx).Where("id = ?", id).Delete(&system.User{}).Error
}

// BatchDelete 批量删除
func (dao *GORMUserDAO) BatchDelete(ctx context.Context, ids []string) error {
	return dao.db.WithContext(ctx).Where("id IN ?", ids).Delete(&system.User{}).Error
}

// Update 更新
func (dao *GORMUserDAO) Update(ctx context.Context, model system.User) error {
	// 开启事务
	// tx := dao.db.WithContext(ctx).Begin()
	// defer func() {
	// 	if r := recover(); r != nil {
	// 		tx.Rollback()
	// 	}
	// }()

	result := dao.db.WithContext(ctx).Model(&model).
		Where("id = ? AND timestamp = ?", model.Id, model.Timestamp).
		Updates(map[string]any{
			"status":       model.Status,
			"username":     model.Username,
			"gender":       model.Gender,
			"email":        model.Email,
			"mobile":       model.Mobile,
			"name":         model.Name,
			"avatar":       model.Avatar,
			"birthday":     model.Birthday,
			"city":         model.City,
			"address":      model.Address,
			"bio":          model.Bio,
			"is_superuser": model.IsSuperuser,
			"manager_id":   model.ManagerID,
			"dept_id":      model.DeptID,
			"sort":         model.Sort,
			"timestamp":    time.Now().UnixMicro(),
			"modifier":     model.Modifier,
			"remark":       model.Remark,
		})
	// 处理行影响数为0的情况
	if result.RowsAffected == 0 {
		// 先检查记录是否存在
		var exists bool
		dao.db.WithContext(ctx).
			Model(&system.User{}).
			Select("1").
			Where("id = ?", model.Id).
			Limit(1).
			Find(&exists)

		if !exists {
			return ErrUserNotFound
		}
		return ErrUserVersionInconsistency
	}

	dao.updateAssociations(dao.db, ctx, model)
	// 更新关联关系
	// if err := dao.updateAssociations(tx, ctx, model); err != nil {
	// 	tx.Rollback()
	// 	return err
	// }

	return result.Error
}

// UpdateLoginField 更新登录字段
func (dao *GORMUserDAO) UpdateLoginField(ctx context.Context, lastLogin string, lastLoginIp string, model system.User) error {
	result := dao.db.WithContext(ctx).Model(&model).
		Where("id = ? AND timestamp = ?", model.Id, model.Timestamp).
		Updates(map[string]any{
			"last_login":    lastLogin,
			"last_login_ip": lastLoginIp,
			"timestamp":     time.Now().UnixMicro(),
		})
	// 处理行影响数为0的情况
	if result.RowsAffected == 0 {
		// 先检查记录是否存在
		var exists bool
		dao.db.WithContext(ctx).
			Model(&system.User{}).
			Select("1").
			Where("id = ?", model.Id).
			Limit(1).
			Find(&exists)

		if !exists {
			return ErrUserNotFound
		}
		return ErrUserVersionInconsistency
	}
	return result.Error
}

// updateAssociations 更新所有关联关系
func (dao *GORMUserDAO) updateAssociations(tx *gorm.DB, ctx context.Context, user system.User) error {
	// 更新岗位关联
	// 删除旧关联
	if err := tx.Exec("DELETE FROM careful_system_user_post WHERE user_id = ?", user.Id).Error; err != nil {
		zap.S().Error("删除岗位关联异常：", err)
		return err
	}
	for _, id := range user.PostIDs {
		var post system.Post
		err := dao.db.WithContext(ctx).Where("id = ?", id).First(&post).Error
		if err != nil {
			continue
		}
		if err := tx.Exec("INSERT INTO careful_system_user_post (user_id, post_id) VALUES (?, ?)",
			user.Id, post.Id).Error; err != nil {
			zap.S().Error("更新岗位关联异常：", err)
			return err
		}
	}

	// 更新角色关联
	// 删除旧关联
	if err := tx.Exec("DELETE FROM careful_system_user_role WHERE user_id = ?", user.Id).Error; err != nil {
		zap.S().Error("删除角色关联异常：", err)
		return err
	}
	for _, id := range user.RoleIDs {
		var role system.Role
		err := dao.db.WithContext(ctx).Where("id = ?", id).First(&role).Error
		if err != nil {
			continue
		}
		if err := tx.Exec("INSERT INTO careful_system_user_role (user_id, role_id) VALUES (?, ?)",
			user.Id, role.Id).Error; err != nil {
			zap.S().Error("更新角色关联异常：", err)
			return err
		}
	}

	return nil
}

// FindById 根据id获取详情
func (dao *GORMUserDAO) FindById(ctx context.Context, id string) (*system.User, error) {
	var model system.User
	err := dao.db.WithContext(ctx).
		Preload("Dept").
		Where("id = ?", id).
		First(&model).Error
	return &model, err
}

// FindByUsername 根据用户名获取详情
func (dao *GORMUserDAO) FindByUsername(ctx context.Context, username string) (*system.User, error) {
	var model system.User
	err := dao.db.WithContext(ctx).
		Preload("Dept").
		Where("username = ?", username).
		First(&model).Error
	return &model, err
}

// FindListPage 分页查询
func (dao *GORMUserDAO) FindListPage(ctx context.Context, filter domainSystem.UserFilter) ([]*system.User, int64, error) {
	var total int64
	var models []*system.User

	query := dao.buildQuery(ctx, filter)

	err := query.Count(&total).
		Offset((filter.Page - 1) * filter.PageSize).
		Limit(filter.PageSize).
		Find(&models).Error

	return models, total, err
}

// FindListAll 获取所有列表
func (dao *GORMUserDAO) FindListAll(ctx context.Context, filter domainSystem.UserFilter) ([]*system.User, error) {
	var models []*system.User

	query := dao.buildQuery(ctx, filter)

	if err := query.Find(&models).Error; err != nil {
		return nil, err
	}

	return models, nil
}

// buildQuery 构建查询条件
func (dao *GORMUserDAO) buildQuery(ctx context.Context, filter domainSystem.UserFilter) *gorm.DB {
	builder := &domainSystem.UserFilter{
		Filters: filters.Filters{
			Creator:    filter.Creator,
			Modifier:   filter.Modifier,
			BelongDept: filter.BelongDept,
		},
		Status:   filter.Status,
		Username: filter.Username,
		Email:    filter.Email,
		Mobile:   filter.Mobile,
	}
	return builder.QueryFilter(ctx, dao.db.WithContext(ctx).Model(&system.User{}))
}

// CheckExistByUsername 检查username是否存在
func (dao *GORMUserDAO) CheckExistByUsername(ctx context.Context, username, excludeId string) (bool, error) {
	var model system.User
	query := dao.db.WithContext(ctx).Model(&system.User{}).
		Select("id"). // 只查询必要的字段
		Where("username = ?", username)

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
