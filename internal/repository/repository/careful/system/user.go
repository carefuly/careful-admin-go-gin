/**
 * Description：
 * FileName：user.go
 * Author：CJiaの用心
 * Create：2025/11/24 17:10:19
 * Remark：
 */

package system

import (
	"context"
	"errors"
	domainSystem "github.com/carefuly/careful-admin-go-gin/internal/domain/careful/system"
	modelSystem "github.com/carefuly/careful-admin-go-gin/internal/model/careful/system"
	cacheSystem "github.com/carefuly/careful-admin-go-gin/internal/repository/cache/careful/system"
	cacheDecorator "github.com/carefuly/careful-admin-go-gin/internal/repository/cache/decorator/careful/system"
	daoSystem "github.com/carefuly/careful-admin-go-gin/internal/repository/dao/careful/system"
	"github.com/carefuly/careful-admin-go-gin/pkg/models"
	"go.uber.org/zap"
	"time"
)

var (
	ErrUserNotFound             = daoSystem.ErrUserNotFound
	ErrUserUsernameDuplicate    = daoSystem.ErrUserUsernameDuplicate
	ErrUserVersionInconsistency = daoSystem.ErrUserVersionInconsistency
)

type UserRepository interface {
	Create(ctx context.Context, domain domainSystem.User) (domainSystem.User, error)
	Delete(ctx context.Context, id string) error
	BatchDelete(ctx context.Context, ids []string) error
	Update(ctx context.Context, domain domainSystem.User) error
	UpdateLoginField(ctx context.Context, lastLogin string, lastLoginIp string, domain domainSystem.User) error

	GetById(ctx context.Context, id string) (domainSystem.User, error)
	GetByUsername(ctx context.Context, username string) (domainSystem.User, error)
	GetListPage(ctx context.Context, filters domainSystem.UserFilter) ([]domainSystem.User, int64, error)
	GetListAll(ctx context.Context, filters domainSystem.UserFilter) ([]domainSystem.User, error)

	CheckExistByUsername(ctx context.Context, code, excludeId string) (bool, error)
}

type userRepository struct {
	dao   daoSystem.UserDAO
	cache cacheDecorator.UserCacheLoggingDecorator
}

func NewUserRepository(dao daoSystem.UserDAO, cache cacheDecorator.UserCacheLoggingDecorator) UserRepository {
	return &userRepository{
		dao:   dao,
		cache: cache,
	}
}

// Create 创建
func (repo *userRepository) Create(ctx context.Context, domain domainSystem.User) (domainSystem.User, error) {
	model, err := repo.dao.Insert(ctx, repo.toEntity(domain))
	return repo.toDomain(model), err
}

// Delete 删除
func (repo *userRepository) Delete(ctx context.Context, id string) error {
	if err := repo.dao.Delete(ctx, id); err != nil {
		return err
	}

	// 删除缓存
	err := repo.cache.Del(ctx, id)
	if err != nil {
		// 网络崩了，也可能是 redis 崩了
		zap.S().Error("Redis异常", zap.Error(err))
		return err
	}

	return err
}

// BatchDelete 批量删除
func (repo *userRepository) BatchDelete(ctx context.Context, ids []string) error {
	err := repo.dao.BatchDelete(ctx, ids)
	if err != nil {
		return err
	}

	// 删除缓存
	for _, val := range ids {
		err = repo.cache.Del(ctx, val)
		if err != nil {
			// 网络崩了，也可能是 redis 崩了
			zap.L().Error("Redis异常", zap.Error(err))
			return err
		}
	}

	return err
}

// Update 更新
func (repo *userRepository) Update(ctx context.Context, domain domainSystem.User) error {
	err := repo.dao.Update(ctx, repo.toEntity(domain))
	if err != nil {
		return err
	}

	// 删除缓存
	err = repo.cache.Del(ctx, domain.Id)
	if err != nil {
		// 网络崩了，也可能是 redis 崩了
		zap.L().Error("Redis异常", zap.Error(err))
		return err
	}

	return nil
}

// UpdateLoginField 更新登录字段
func (repo *userRepository) UpdateLoginField(ctx context.Context, lastLogin string, lastLoginIp string, domain domainSystem.User) error {
	return repo.dao.UpdateLoginField(ctx, lastLogin, lastLoginIp, repo.toEntity(domain))
}

// GetById 根据ID获取
func (repo *userRepository) GetById(ctx context.Context, id string) (domainSystem.User, error) {
	domain, err := repo.cache.Get(ctx, id)
	if err == nil && domain != nil {
		return *domain, nil // 命中缓存
	}

	if err != nil && !errors.Is(err, cacheSystem.ErrUserNotExist) {
		// 缓存查询出错但不是"不存在"错误，记录日志但继续查DB
		zap.L().Error("缓存获取错误:", zap.Error(err))
	}

	entity, err := repo.dao.FindById(ctx, id)
	if err != nil {
		if errors.Is(err, daoSystem.ErrUserNotFound) {
			// 数据库不存在，设置防穿透标记
			_ = repo.cache.SetNotFound(ctx, id)
			return domainSystem.User{}, daoSystem.ErrUserNotFound
		}
		return domainSystem.User{}, err
	}

	toDomain := repo.toDomain(entity)

	if err := repo.cache.Set(ctx, toDomain); err != nil {
		// 网络崩了，也可能是 redis 崩了
		// 缓存删除失败不影响主流程，记录日志即可
		zap.L().Error("设置缓存失败异常", zap.Error(err))
	}

	return toDomain, nil
}

// GetByUsername 根据用户名获取
func (repo *userRepository) GetByUsername(ctx context.Context, username string) (domainSystem.User, error) {
	user, err := repo.dao.FindByUsername(ctx, username)
	if err != nil {
		return domainSystem.User{}, err
	}
	return repo.toDomain(user), nil
}

// GetListPage 分页查询列表
func (repo *userRepository) GetListPage(ctx context.Context, filters domainSystem.UserFilter) ([]domainSystem.User, int64, error) {
	list, row, err := repo.dao.FindListPage(ctx, filters)
	if err != nil {
		return []domainSystem.User{}, row, err
	}

	if len(list) == 0 {
		return []domainSystem.User{}, row, nil
	}

	var domain []domainSystem.User
	for _, v := range list {
		domain = append(domain, repo.toDomain(v))
	}

	return domain, row, nil
}

// GetListAll 查询所有列表
func (repo *userRepository) GetListAll(ctx context.Context, filters domainSystem.UserFilter) ([]domainSystem.User, error) {
	list, err := repo.dao.FindListAll(ctx, filters)
	if err != nil {
		return []domainSystem.User{}, err
	}

	if len(list) == 0 {
		return []domainSystem.User{}, nil
	}

	var toDomain []domainSystem.User
	for _, v := range list {
		toDomain = append(toDomain, repo.toDomain(v))
	}

	return toDomain, nil
}

// CheckExistByUsername 检查username是否存在
func (repo *userRepository) CheckExistByUsername(ctx context.Context, code, excludeId string) (bool, error) {
	return repo.dao.CheckExistByUsername(ctx, code, excludeId)
}

// toEntity 转换为实体模型
func (repo *userRepository) toEntity(domain domainSystem.User) modelSystem.User {
	return modelSystem.User{
		CoreModels: models.CoreModels{
			Id:         domain.Id,
			Sort:       domain.Sort,
			Timestamp:  domain.Timestamp,
			Creator:    domain.Creator,
			Modifier:   domain.Modifier,
			BelongDept: domain.BelongDept,
			Remark:     domain.Remark,
		},
		Status:      domain.Status,
		Username:    domain.Username,
		Password:    domain.Password,
		Gender:      domain.Gender,
		Email:       domain.Email,
		Mobile:      domain.Mobile,
		Name:        domain.Name,
		Avatar:      domain.Avatar,
		Birthday:    domain.Birthday,
		City:        domain.City,
		Address:     domain.Address,
		Bio:         domain.Bio,
		IsSuperuser: domain.IsSuperuser,
		LastLogin:   domain.LastLogin,
		LastLoginIp: domain.LastLoginIp,
		ManagerID:   domain.ManagerID,
		DeptID:      domain.DeptID,
		PostIDs:     domain.PostIDs,
		RoleIDs:     domain.RoleIDs,
	}
}

// toDomain 转换为领域模型
func (repo *userRepository) toDomain(entity *modelSystem.User) domainSystem.User {
	model := domainSystem.User{
		User: *entity,
	}

	if entity.Birthday != nil {
		birthdayStr := entity.Birthday.Format("2006-01-02")
		birthdayTime, err := time.Parse("2006-01-02", birthdayStr)
		if err != nil {

		}
		model.Birthday = &birthdayTime
	}
	if entity.CreateTime != nil {
		model.CreateTime = entity.CreateTime.Format("2006-01-02 15:04:05")
	}
	if entity.UpdateTime != nil {
		model.UpdateTime = entity.UpdateTime.Format("2006-01-02 15:04:05")
	}

	return model
}
