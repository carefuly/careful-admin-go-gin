/**
 * Description：
 * FileName：role.go
 * Author：CJiaの用心
 * Create：2025/12/19 22:10:58
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
)

var (
	ErrRoleNotFound             = daoSystem.ErrRoleNotFound
	ErrRoleCodeDuplicate        = daoSystem.ErrRoleCodeDuplicate
	ErrRoleVersionInconsistency = daoSystem.ErrRoleVersionInconsistency
)

type RoleRepository interface {
	Create(ctx context.Context, domain domainSystem.Role) (domainSystem.Role, error)
	Delete(ctx context.Context, id string) error
	BatchDelete(ctx context.Context, ids []string) error
	Update(ctx context.Context, domain domainSystem.Role) error

	GetById(ctx context.Context, id string) (domainSystem.Role, error)
	GetListPage(ctx context.Context, filters domainSystem.RoleFilter) ([]domainSystem.Role, int64, error)
	GetListAll(ctx context.Context, filters domainSystem.RoleFilter) ([]domainSystem.Role, error)

	CheckExistByCode(ctx context.Context, code, excludeId string) (bool, error)
}

type roleRepository struct {
	dao   daoSystem.RoleDAO
	cache cacheDecorator.RoleCacheLoggingDecorator
}

func NewRoleRepository(dao daoSystem.RoleDAO, cache cacheDecorator.RoleCacheLoggingDecorator) RoleRepository {
	return &roleRepository{
		dao:   dao,
		cache: cache,
	}
}

// Create 创建
func (repo *roleRepository) Create(ctx context.Context, domain domainSystem.Role) (domainSystem.Role, error) {
	model, err := repo.dao.Insert(ctx, repo.toEntity(domain))
	return repo.toDomain(model), err
}

// Delete 删除
func (repo *roleRepository) Delete(ctx context.Context, id string) error {
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
func (repo *roleRepository) BatchDelete(ctx context.Context, ids []string) error {
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
func (repo *roleRepository) Update(ctx context.Context, domain domainSystem.Role) error {
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

// GetById 根据ID获取
func (repo *roleRepository) GetById(ctx context.Context, id string) (domainSystem.Role, error) {
	domain, err := repo.cache.Get(ctx, id)
	if err == nil && domain != nil {
		return *domain, nil // 命中缓存
	}
	if err != nil && !errors.Is(err, cacheSystem.ErrRoleNotExist) {
		// 缓存查询出错但不是"不存在"错误，记录日志但继续查DB
		zap.L().Error("缓存获取错误:", zap.Error(err))
	}

	entity, err := repo.dao.FindById(ctx, id)
	if err != nil {
		if errors.Is(err, daoSystem.ErrRoleNotFound) {
			// 数据库不存在，设置防穿透标记
			_ = repo.cache.SetNotFound(ctx, id)
			return domainSystem.Role{}, nil
		}
		return domainSystem.Role{}, err
	}

	toDomain := repo.toDomain(entity)
	if err := repo.cache.Set(ctx, toDomain); err != nil {
		// 网络崩了，也可能是 redis 崩了
		zap.L().Error("Redis异常", zap.Error(err))
	}

	return toDomain, nil
}

// GetListPage 分页查询列表
func (repo *roleRepository) GetListPage(ctx context.Context, filters domainSystem.RoleFilter) ([]domainSystem.Role, int64, error) {
	list, row, err := repo.dao.FindListPage(ctx, filters)
	if err != nil {
		return []domainSystem.Role{}, row, err
	}

	if len(list) == 0 {
		return []domainSystem.Role{}, row, nil
	}

	var domain []domainSystem.Role
	for _, v := range list {
		domain = append(domain, repo.toDomain(v))
	}

	return domain, row, nil
}

// GetListAll 查询所有列表
func (repo *roleRepository) GetListAll(ctx context.Context, filters domainSystem.RoleFilter) ([]domainSystem.Role, error) {
	list, err := repo.dao.FindListAll(ctx, filters)
	if err != nil {
		return []domainSystem.Role{}, err
	}

	if len(list) == 0 {
		return []domainSystem.Role{}, nil
	}

	var toDomain []domainSystem.Role
	for _, v := range list {
		toDomain = append(toDomain, repo.toDomain(v))
	}

	return toDomain, nil
}

// CheckExistByCode 检查code是否存在
func (repo *roleRepository) CheckExistByCode(ctx context.Context, code, excludeId string) (bool, error) {
	return repo.dao.CheckExistByCode(ctx, code, excludeId)
}

// toEntity 转换为实体模型
func (repo *roleRepository) toEntity(domain domainSystem.Role) modelSystem.Role {
	return modelSystem.Role{
		CoreModels: models.CoreModels{
			Id:         domain.Id,
			Sort:       domain.Sort,
			Timestamp:  domain.Timestamp,
			Creator:    domain.Creator,
			Modifier:   domain.Modifier,
			BelongDept: domain.BelongDept,
			Remark:     domain.Remark,
		},
		Status:        domain.Status,
		Name:          domain.Name,
		Code:          domain.Code,
		DataScope:     domain.DataScope,
		Description:   domain.Description,
		DeptIDs:       domain.DeptIDs,
		MenuIDs:       domain.MenuIDs,
		MenuButtonIDs: domain.MenuButtonIDs,
	}
}

// toDomain 转换为领域模型
func (repo *roleRepository) toDomain(entity *modelSystem.Role) domainSystem.Role {
	domain := domainSystem.Role{
		Role: *entity,
	}

	if entity.CreateTime != nil {
		domain.CreateTime = entity.CreateTime.Format("2006-01-02 15:04:05")
	}
	if entity.UpdateTime != nil {
		domain.UpdateTime = entity.UpdateTime.Format("2006-01-02 15:04:05")
	}

	return domain
}
