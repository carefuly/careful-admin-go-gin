/**
 * Description：
 * FileName：user.go
 * Author：CJiaの用心
 * Create：2025/10/9 16:44:47
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
	"go.uber.org/zap"
)

var (
	ErrUserNotFound = daoSystem.ErrUserNotFound
)

type UserRepository interface {
	GetById(ctx context.Context, id string) (domainSystem.User, error)
	GetByUsername(ctx context.Context, username string) (domainSystem.User, error)
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

// toDomain 转换为领域模型
func (repo *userRepository) toDomain(entity *modelSystem.User) domainSystem.User {
	model := domainSystem.User{
		User:   *entity,
		DeptId: entity.DeptId.String,
	}

	if entity.CreateTime != nil {
		model.CreateTime = entity.CreateTime.Format("2006-01-02 15:04:05")
	}
	if entity.UpdateTime != nil {
		model.UpdateTime = entity.UpdateTime.Format("2006-01-02 15:04:05")
	}

	return model
}
