/**
 * Description：
 * FileName：menu.go
 * Author：CJiaの用心
 * Create：2025/12/05 14:20:33
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
	ErrMenuNotFound             = daoSystem.ErrMenuNotFound
	ErrMenuDuplicate            = daoSystem.ErrMenuDuplicate
	ErrMenuParentNotFound       = daoSystem.ErrMenuParentNotFound
	ErrMenuDisabled             = daoSystem.ErrMenuDisabled
	ErrMenuHasChildren          = daoSystem.ErrMenuHasChildren
	ErrMenuVersionInconsistency = daoSystem.ErrMenuVersionInconsistency
)

type MenuRepository interface {
	Create(ctx context.Context, domain domainSystem.Menu) (domainSystem.Menu, error)
	Delete(ctx context.Context, id string) error
	BatchDelete(ctx context.Context, ids []string) error
	Update(ctx context.Context, domain domainSystem.Menu) error

	GetById(ctx context.Context, id string) (domainSystem.Menu, error)
	GetByParentId(ctx context.Context, parentId string) (domainSystem.Menu, error)
	GetChildCount(ctx context.Context, id string) (int64, error)
	GetListPage(ctx context.Context, filters domainSystem.MenuFilter) ([]domainSystem.Menu, int64, error)
	GetListAll(ctx context.Context, filters domainSystem.MenuFilter) ([]domainSystem.Menu, error)

	CheckExistByNameAndPathAndTitleAndParentId(ctx context.Context, name, path, title, parentId, excludeId string) (bool, error)
}

type menuRepository struct {
	dao   daoSystem.MenuDAO
	cache cacheDecorator.MenuCacheLoggingDecorator
}

func NewMenuRepository(dao daoSystem.MenuDAO, cache cacheDecorator.MenuCacheLoggingDecorator) MenuRepository {
	return &menuRepository{
		dao:   dao,
		cache: cache,
	}
}

// Create 创建
func (repo *menuRepository) Create(ctx context.Context, domain domainSystem.Menu) (domainSystem.Menu, error) {
	model, err := repo.dao.Insert(ctx, repo.toEntity(domain))
	return repo.toDomain(model), err
}

// Delete 删除
func (repo *menuRepository) Delete(ctx context.Context, id string) error {
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
func (repo *menuRepository) BatchDelete(ctx context.Context, ids []string) error {
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
func (repo *menuRepository) Update(ctx context.Context, domain domainSystem.Menu) error {
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
func (repo *menuRepository) GetById(ctx context.Context, id string) (domainSystem.Menu, error) {
	domain, err := repo.cache.Get(ctx, id)
	if err == nil && domain != nil {
		return *domain, nil // 命中缓存
	}
	if err != nil && !errors.Is(err, cacheSystem.ErrMenuNotExist) {
		// 缓存查询出错但不是"不存在"错误，记录日志但继续查DB
		zap.L().Error("缓存获取错误:", zap.Error(err))
	}

	entity, err := repo.dao.FindById(ctx, id)
	if err != nil {
		if errors.Is(err, daoSystem.ErrMenuNotFound) {
			// 数据库不存在，设置防穿透标记
			_ = repo.cache.SetNotFound(ctx, id)
			return domainSystem.Menu{}, nil
		}
		return domainSystem.Menu{}, err
	}

	toDomain := repo.toDomain(entity)
	if err := repo.cache.Set(ctx, toDomain); err != nil {
		// 网络崩了，也可能是 redis 崩了
		zap.L().Error("Redis异常", zap.Error(err))
	}

	return toDomain, nil
}

// GetByParentId 根据parent_id获取详情
func (repo *menuRepository) GetByParentId(ctx context.Context, parentId string) (domainSystem.Menu, error) {
	model, err := repo.dao.FindByParentId(ctx, parentId)
	if err != nil {
		return domainSystem.Menu{}, err
	}
	return repo.toDomain(model), nil
}

// GetChildCount 获取子菜单数量
func (repo *menuRepository) GetChildCount(ctx context.Context, id string) (int64, error) {
	return repo.dao.FindChildCount(ctx, id)
}

// GetListPage 分页查询列表
func (repo *menuRepository) GetListPage(ctx context.Context, filters domainSystem.MenuFilter) ([]domainSystem.Menu, int64, error) {
	list, row, err := repo.dao.FindListPage(ctx, filters)
	if err != nil {
		return []domainSystem.Menu{}, row, err
	}

	if len(list) == 0 {
		return []domainSystem.Menu{}, row, nil
	}

	var domain []domainSystem.Menu
	for _, v := range list {
		domain = append(domain, repo.toDomain(v))
	}

	return domain, row, nil
}

// GetListAll 查询所有列表
func (repo *menuRepository) GetListAll(ctx context.Context, filters domainSystem.MenuFilter) ([]domainSystem.Menu, error) {
	list, err := repo.dao.FindListAll(ctx, filters)
	if err != nil {
		return []domainSystem.Menu{}, err
	}

	if len(list) == 0 {
		return []domainSystem.Menu{}, nil
	}

	var toDomain []domainSystem.Menu
	for _, v := range list {
		toDomain = append(toDomain, repo.toDomain(v))
	}

	return toDomain, nil
}

// CheckExistByNameAndPathAndTitleAndParentId 检查name、path、title和parentId是否同时存在
func (repo *menuRepository) CheckExistByNameAndPathAndTitleAndParentId(ctx context.Context, name, path, title, parentId, excludeId string) (bool, error) {
	return repo.dao.CheckExistByNameAndPathAndTitleAndParentId(ctx, name, path, title, parentId, excludeId)
}

// toEntity 转换为实体模型
func (repo *menuRepository) toEntity(domain domainSystem.Menu) modelSystem.Menu {
	return modelSystem.Menu{
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
		Path:          domain.Path,
		Component:     domain.Component,
		Title:         domain.Title,
		Icon:          domain.Icon,
		ShowBadge:     domain.ShowBadge,
		ShowTextBadge: domain.ShowTextBadge,
		IsHide:        domain.IsHide,
		IsHideTab:     domain.IsHideTab,
		Link:          domain.Link,
		IsIframe:      domain.IsIframe,
		KeepAlive:     domain.KeepAlive,
		IsFirstLevel:  domain.IsFirstLevel,
		FixedTab:      domain.FixedTab,
		ActivePath:    domain.ActivePath,
		IsFullPage:    domain.IsFullPage,
		IsAuthButton:  domain.IsAuthButton,
		AuthMark:      domain.AuthMark,
		ParentID:      domain.ParentID,
	}
}

// toDomain 转换为领域模型
func (repo *menuRepository) toDomain(entity *modelSystem.Menu) domainSystem.Menu {
	domain := domainSystem.Menu{
		Menu: *entity,
	}

	if entity.CreateTime != nil {
		domain.CreateTime = entity.CreateTime.Format("2006-01-02 15:04:05")
	}
	if entity.UpdateTime != nil {
		domain.UpdateTime = entity.UpdateTime.Format("2006-01-02 15:04:05")
	}

	return domain
}
