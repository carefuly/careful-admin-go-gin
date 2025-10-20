/**
 * Description：
 * FileName：dict_type.go
 * Author：CJiaの用心
 * Create：2025/10/19 15:01:34
 * Remark：
 */

package tools

import (
	"context"
	"database/sql"
	"errors"
	domainTools "github.com/carefuly/careful-admin-go-gin/internal/domain/careful/tools"
	modelTools "github.com/carefuly/careful-admin-go-gin/internal/model/careful/tools"
	cacheTools "github.com/carefuly/careful-admin-go-gin/internal/repository/cache/careful/tools"
	cacheDecorator "github.com/carefuly/careful-admin-go-gin/internal/repository/cache/decorator/careful/tools"
	daoTools "github.com/carefuly/careful-admin-go-gin/internal/repository/dao/careful/tools"
	"github.com/carefuly/careful-admin-go-gin/pkg/models"
	"go.uber.org/zap"
)

var (
	ErrDictTypeInvalidDictValueType = daoTools.ErrDictTypeInvalidDictValueType
	ErrDictTypeNotFound             = daoTools.ErrDictTypeNotFound
	ErrDictTypeDuplicate            = daoTools.ErrDictTypeDuplicate
	ErrDictTypeVersionInconsistency = daoTools.ErrDictTypeVersionInconsistency
)

type DictTypeRepository interface {
	Create(ctx context.Context, domain domainTools.DictType) (domainTools.DictType, error)
	Delete(ctx context.Context, id string) error
	BatchDelete(ctx context.Context, ids []string) error
	Update(ctx context.Context, domain domainTools.DictType) error

	GetById(ctx context.Context, id string) (domainTools.DictType, error)
	GetByDictNames(ctx context.Context, dictNames []string) ([]domainTools.DictType, error)
	GetListPage(ctx context.Context, filters domainTools.DictTypeFilter) ([]domainTools.DictType, int64, error)
	GetListAll(ctx context.Context, filters domainTools.DictTypeFilter) ([]domainTools.DictType, error)
}

type dictTypeRepository struct {
	dao   daoTools.DictTypeDAO
	cache cacheDecorator.DictTypeCacheLoggingDecorator
}

func NewDictTypeRepository(dao daoTools.DictTypeDAO, cache cacheDecorator.DictTypeCacheLoggingDecorator) DictTypeRepository {
	return &dictTypeRepository{
		dao:   dao,
		cache: cache,
	}
}

// Create 创建
func (repo *dictTypeRepository) Create(ctx context.Context, domain domainTools.DictType) (domainTools.DictType, error) {
	model, err := repo.toEntity(domain)
	if err != nil {
		return domainTools.DictType{}, err
	}
	insert, err := repo.dao.Insert(ctx, model)
	toDomain := repo.toDomain(insert)
	return toDomain, err
}

// Delete 删除
func (repo *dictTypeRepository) Delete(ctx context.Context, id string) error {
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
func (repo *dictTypeRepository) BatchDelete(ctx context.Context, ids []string) error {
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
func (repo *dictTypeRepository) Update(ctx context.Context, domain domainTools.DictType) error {
	entity, err := repo.toEntity(domain)
	if err != nil {
		return err
	}

	err = repo.dao.Update(ctx, entity)
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
func (repo *dictTypeRepository) GetById(ctx context.Context, id string) (domainTools.DictType, error) {
	domain, err := repo.cache.Get(ctx, id)
	if err == nil && domain != nil {
		return *domain, nil // 命中缓存
	}
	if err != nil && !errors.Is(err, cacheTools.ErrDictTypeNotExist) {
		// 缓存查询出错但不是"不存在"错误，记录日志但继续查DB
		zap.L().Error("缓存获取错误:", zap.Error(err))
	}

	entity, err := repo.dao.FindById(ctx, id)
	if err != nil {
		if errors.Is(err, daoTools.ErrDictNotFound) {
			// 数据库不存在，设置防穿透标记
			_ = repo.cache.SetNotFound(ctx, id)
			return domainTools.DictType{}, nil
		}
		return domainTools.DictType{}, err
	}

	toDomain := repo.toDomain(entity)
	if err := repo.cache.Set(ctx, toDomain); err != nil {
		// 网络崩了，也可能是 redis 崩了
		zap.L().Error("Redis异常", zap.Error(err))
	}

	return toDomain, nil
}

// GetByDictNames 根据多个dictName获取详情
func (repo *dictTypeRepository) GetByDictNames(ctx context.Context, dictNames []string) ([]domainTools.DictType, error) {
	// list, err := repo.dao.FindByDictNames(ctx, dictNames)
	// if err != nil {
	// 	return []domainTools.DictType{}, err
	// }
	//
	// if len(list) == 0 {
	// 	return []domainTools.DictType{}, nil
	// }
	//
	// var domains []domainTools.DictType
	// for _, v := range list {
	// 	domains = append(domains, repo.toDomain(v))
	// }
	//
	// return domains, nil
	return nil, nil
}

// GetListPage 分页查询列表
func (repo *dictTypeRepository) GetListPage(ctx context.Context, filters domainTools.DictTypeFilter) ([]domainTools.DictType, int64, error) {
	list, row, err := repo.dao.FindListPage(ctx, filters)
	if err != nil {
		return []domainTools.DictType{}, row, err
	}

	if len(list) == 0 {
		return []domainTools.DictType{}, row, nil
	}

	var domains []domainTools.DictType
	for _, v := range list {
		domains = append(domains, repo.toDomain(v))
	}

	return domains, row, nil
}

// GetListAll 查询所有列表
func (repo *dictTypeRepository) GetListAll(ctx context.Context, filters domainTools.DictTypeFilter) ([]domainTools.DictType, error) {
	list, err := repo.dao.FindListAll(ctx, filters)
	if err != nil {
		return []domainTools.DictType{}, err
	}

	if len(list) == 0 {
		return []domainTools.DictType{}, nil
	}

	var domains []domainTools.DictType
	for _, v := range list {
		domains = append(domains, repo.toDomain(v))
	}

	return domains, nil
}

// toEntity 转换为实体模型
func (repo *dictTypeRepository) toEntity(domain domainTools.DictType) (modelTools.DictType, error) {
	model := modelTools.DictType{
		CoreModels: models.CoreModels{
			Id:         domain.Id,
			Sort:       domain.Sort,
			Timestamp:  domain.Timestamp,
			Creator:    domain.Creator,
			Modifier:   domain.Modifier,
			BelongDept: domain.BelongDept,
			Remark:     domain.Remark,
		},
		Status:    domain.Status,
		Name:      domain.Name,
		DictTag:   domain.DictTag,
		DictColor: domain.DictColor,
		DictName:  domain.DictName,
		ValueType: domain.ValueType,
		DictId:    domain.DictId,
	}

	// 根据类型设置值
	switch domain.ValueType {
	case 1: // 字符串
		model.StrValue = sql.NullString{
			Valid:  true,
			String: domain.StrValue,
		}
		model.IntValue = sql.NullInt64{Valid: false}
		model.BoolValue = sql.NullBool{Valid: false}
	case 2: // 整型
		model.StrValue = sql.NullString{Valid: false}
		model.IntValue = sql.NullInt64{
			Valid: true,
			Int64: domain.IntValue,
		}
		model.BoolValue = sql.NullBool{Valid: false}
	case 3: // 布尔
		model.StrValue = sql.NullString{Valid: false}
		model.IntValue = sql.NullInt64{Valid: false}
		model.BoolValue = sql.NullBool{
			Valid: true,
			Bool:  domain.BoolValue,
		}
	default:
		return modelTools.DictType{}, daoTools.ErrDictTypeInvalidDictValueType
	}

	return model, nil
}

// toDomain 转换为领域模型
func (repo *dictTypeRepository) toDomain(entity *modelTools.DictType) domainTools.DictType {
	model := domainTools.DictType{
		DictType:  *entity,
		StrValue:  entity.StrValue.String,
		IntValue:  entity.IntValue.Int64,
		BoolValue: entity.BoolValue.Bool,
	}

	if entity.CreateTime != nil {
		model.CreateTime = entity.CreateTime.Format("2006-01-02 15:04:05")
	}
	if entity.UpdateTime != nil {
		model.UpdateTime = entity.UpdateTime.Format("2006-01-02 15:04:05")
	}

	return model
}
