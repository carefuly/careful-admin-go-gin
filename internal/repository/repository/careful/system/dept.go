/**
 * Description：
 * FileName：dept.go
 * Author：CJiaの用心
 * Create：2025/11/26 09:20:29
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
	ErrDeptNotFound             = daoSystem.ErrDeptNotFound
	ErrDeptCodeDuplicate        = daoSystem.ErrDeptCodeDuplicate
	ErrDeptNameParentDuplicate  = daoSystem.ErrDeptNameParentDuplicate
	ErrDeptParentNotFound       = daoSystem.ErrDeptParentNotFound
	ErrDeptDisabled             = daoSystem.ErrDeptDisabled
	ErrDeptHasChildren          = daoSystem.ErrDeptHasChildren
	ErrDeptHasUsers             = daoSystem.ErrDeptHasUsers
	ErrDeptVersionInconsistency = daoSystem.ErrDeptVersionInconsistency
)

type DeptRepository interface {
	Create(ctx context.Context, domain domainSystem.Dept) (domainSystem.Dept, error)
	Delete(ctx context.Context, id string) error
	BatchDelete(ctx context.Context, ids []string) error
	Update(ctx context.Context, domain domainSystem.Dept) error

	GetById(ctx context.Context, id string) (domainSystem.Dept, error)
	GetByParentId(ctx context.Context, parentId string) (domainSystem.Dept, error)
	GetListAll(ctx context.Context, filters domainSystem.DeptFilter) ([]domainSystem.Dept, error)

	CheckExistByCode(ctx context.Context, code, excludeId string) (bool, error)
	CheckExistByNameAndParentId(ctx context.Context, name, parentId, excludeId string) (bool, error)
}

type deptRepository struct {
	dao   daoSystem.DeptDAO
	cache cacheDecorator.DeptCacheLoggingDecorator
}

func NewDeptRepository(dao daoSystem.DeptDAO, cache cacheDecorator.DeptCacheLoggingDecorator) DeptRepository {
	return &deptRepository{
		dao:   dao,
		cache: cache,
	}
}

// Create 创建
func (repo *deptRepository) Create(ctx context.Context, domain domainSystem.Dept) (domainSystem.Dept, error) {
	model, err := repo.dao.Insert(ctx, repo.toEntity(domain))
	return repo.toDomain(model), err
}

// Delete 删除
func (repo *deptRepository) Delete(ctx context.Context, id string) error {
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
func (repo *deptRepository) BatchDelete(ctx context.Context, ids []string) error {
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
func (repo *deptRepository) Update(ctx context.Context, domain domainSystem.Dept) error {
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
func (repo *deptRepository) GetById(ctx context.Context, id string) (domainSystem.Dept, error) {
	domain, err := repo.cache.Get(ctx, id)
	if err == nil && domain != nil {
		return *domain, nil // 命中缓存
	}
	if err != nil && !errors.Is(err, cacheSystem.ErrDeptNotExist) {
		// 缓存查询出错但不是"不存在"错误，记录日志但继续查DB
		zap.L().Error("缓存获取错误:", zap.Error(err))
	}

	entity, err := repo.dao.FindById(ctx, id)
	if err != nil {
		if errors.Is(err, daoSystem.ErrDeptNotFound) {
			// 数据库不存在，设置防穿透标记
			_ = repo.cache.SetNotFound(ctx, id)
			return domainSystem.Dept{}, nil
		}
		return domainSystem.Dept{}, err
	}

	toDomain := repo.toDomain(entity)
	if err := repo.cache.Set(ctx, toDomain); err != nil {
		// 网络崩了，也可能是 redis 崩了
		zap.L().Error("Redis异常", zap.Error(err))
	}

	return toDomain, nil
}

// GetByParentId 根据parent_id获取详情
func (repo *deptRepository) GetByParentId(ctx context.Context, parentId string) (domainSystem.Dept, error) {
	model, err := repo.dao.FindByParentId(ctx, parentId)
	if err != nil {
		return domainSystem.Dept{}, err
	}
	return repo.toDomain(model), nil
}

// GetListAll 查询所有列表
func (repo *deptRepository) GetListAll(ctx context.Context, filters domainSystem.DeptFilter) ([]domainSystem.Dept, error) {
	list, err := repo.dao.FindListAll(ctx, filters)
	if err != nil {
		return []domainSystem.Dept{}, err
	}

	if len(list) == 0 {
		return []domainSystem.Dept{}, nil
	}

	var toDomain []domainSystem.Dept
	for _, v := range list {
		toDomain = append(toDomain, repo.toDomain(v))
	}

	return toDomain, nil
}

// CheckExistByCode 检查code是否存在
func (repo *deptRepository) CheckExistByCode(ctx context.Context, code, excludeId string) (bool, error) {
	return repo.dao.CheckExistByCode(ctx, code, excludeId)
}

// CheckExistByNameAndParentId 检查name是否存在
func (repo *deptRepository) CheckExistByNameAndParentId(ctx context.Context, name, parentId, excludeId string) (bool, error) {
	return repo.dao.CheckExistByNameAndParentId(ctx, name, parentId, excludeId)
}

// toEntity 转换为实体模型
func (repo *deptRepository) toEntity(domain domainSystem.Dept) modelSystem.Dept {
	return modelSystem.Dept{
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
		Name:        domain.Name,
		Code:        domain.Code,
		DeptType:    domain.DeptType,
		Owner:       domain.Owner,
		Phone:       domain.Phone,
		Email:       domain.Email,
		Description: domain.Description,
		ParentID:    domain.ParentID,
		Level:       domain.Level,
		Path:        domain.Path,
	}
}

// toDomain 转换为领域模型
func (repo *deptRepository) toDomain(entity *modelSystem.Dept) domainSystem.Dept {
	model := domainSystem.Dept{
		Dept: *entity,
	}

	if entity.CreateTime != nil {
		model.CreateTime = entity.CreateTime.Format("2006-01-02 15:04:05")
	}
	if entity.UpdateTime != nil {
		model.UpdateTime = entity.UpdateTime.Format("2006-01-02 15:04:05")
	}

	return model
}
