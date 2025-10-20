/**
 * Description：
 * FileName：dict_type.go
 * Author：CJiaの用心
 * Create：2025/10/19 15:21:32
 * Remark：
 */

package tools

import (
	"context"
	"errors"
	domainTools "github.com/carefuly/careful-admin-go-gin/internal/domain/careful/tools"
	repositoryTools "github.com/carefuly/careful-admin-go-gin/internal/repository/repository/careful/tools"
	"github.com/go-sql-driver/mysql"
)

var (
	ErrDictTypeInvalidDictValueType = repositoryTools.ErrDictTypeInvalidDictValueType
	ErrDictTypeNotFound             = repositoryTools.ErrDictTypeNotFound
	ErrDictTypeDuplicate            = repositoryTools.ErrDictTypeDuplicate
	ErrDictTypeVersionInconsistency = repositoryTools.ErrDictTypeVersionInconsistency
)

type DictTypeService interface {
	Create(ctx context.Context, domain domainTools.DictType) error
	Delete(ctx context.Context, id string) error
	BatchDelete(ctx context.Context, ids []string) error
	Update(ctx context.Context, domain domainTools.DictType) error

	GetById(ctx context.Context, id string) (domainTools.DictType, error)
	GetByDictNames(ctx context.Context, dictNames []string) (map[string][]domainTools.DictType, error)
	GetListPage(ctx context.Context, filter domainTools.DictTypeFilter) ([]domainTools.DictType, int64, error)
	GetListAll(ctx context.Context, filter domainTools.DictTypeFilter) ([]domainTools.DictType, error)
}

type dictTypeService struct {
	repo     repositoryTools.DictTypeRepository
	dictRepo repositoryTools.DictRepository
}

func NewDictTypeService(repo repositoryTools.DictTypeRepository, dictRepo repositoryTools.DictRepository) DictTypeService {
	return &dictTypeService{
		repo:     repo,
		dictRepo: dictRepo,
	}
}

// Create 创建
func (svc *dictTypeService) Create(ctx context.Context, domain domainTools.DictType) error {
	// 获取字典详情
	dict, err := svc.dictRepo.GetById(ctx, domain.DictId)
	if err != nil {
		if errors.Is(err, repositoryTools.ErrDictNotFound) {
			return repositoryTools.ErrDictNotFound
		}
		return err
	}

	if dict.Id == "" {
		return repositoryTools.ErrDictNotFound
	}

	// 设置DictName和TypeValue
	domain.DictName = dict.Name
	domain.ValueType = dict.ValueType

	// 唯一性校验
	// 逻辑较为复杂，暂时不实现，默认使用mysql唯一性约束
	if _, err := svc.repo.Create(ctx, domain); err != nil {
		if svc.IsDuplicateEntryError(err) {
			return repositoryTools.ErrDictTypeDuplicate
		}
		return err
	}

	return nil
}

// Import 导入
func (svc *dictTypeService) Import() {}

// Delete 删除
func (svc *dictTypeService) Delete(ctx context.Context, id string) error {
	return svc.repo.Delete(ctx, id)
}

// BatchDelete 批量删除
func (svc *dictTypeService) BatchDelete(ctx context.Context, ids []string) error {
	return svc.repo.BatchDelete(ctx, ids)
}

// Update 更新
func (svc *dictTypeService) Update(ctx context.Context, domain domainTools.DictType) error {
	// 获取字典详情
	dict, err := svc.dictRepo.GetById(ctx, domain.DictId)
	if err != nil {
		if errors.Is(err, repositoryTools.ErrDictNotFound) {
			return repositoryTools.ErrDictNotFound
		}
		return err
	}

	if dict.Id == "" {
		return repositoryTools.ErrDictNotFound
	}

	// 设置DictName和TypeValue
	domain.DictName = dict.Name
	domain.ValueType = dict.ValueType

	// 唯一性校验
	// 逻辑较为复杂，暂时不实现，默认使用mysql唯一性约束

	err = svc.repo.Update(ctx, domain)

	switch {
	case err == nil:
		return err
	case errors.Is(err, repositoryTools.ErrDictTypeNotFound):
		return repositoryTools.ErrDictTypeNotFound
	case errors.Is(err, repositoryTools.ErrDictTypeVersionInconsistency):
		return repositoryTools.ErrDictTypeVersionInconsistency
	default:
		return err
	}
}

// GetById 获取详情
func (svc *dictTypeService) GetById(ctx context.Context, id string) (domainTools.DictType, error) {
	domain, err := svc.repo.GetById(ctx, id)
	if err != nil {
		if errors.Is(err, repositoryTools.ErrDictTypeNotFound) {
			return domain, repositoryTools.ErrDictTypeNotFound
		}
		return domain, err
	}
	if domain.Id == "" {
		return domain, repositoryTools.ErrDictTypeNotFound
	}
	return domain, err
}

// GetByDictNames 根据多个dictName获取详情
func (svc *dictTypeService) GetByDictNames(ctx context.Context, dictNames []string) (map[string][]domainTools.DictType, error) {
	return nil, nil
	// if len(dictNames) == 0 {
	// 	return map[string][]domainTools.DictType{}, nil
	// }
	//
	// dictTypes, err := svc.repo.GetByDictNames(ctx, dictNames)
	// if err != nil {
	// 	return nil, err
	// }
	//
	// // 按dictName分组
	// resultMap := make(map[string][]domainTools.DictType)
	// for _, dt := range dictTypes {
	// 	switch {
	// 	case dt.ValueType == 1:
	// 		dt.Label = dt.Name
	// 		dt.Value = dt.StrValue
	// 	case dt.ValueType == 2:
	// 		dt.Label = dt.Name
	// 		dt.Value = dt.IntValue
	// 	case dt.ValueType == 3:
	// 		dt.Label = dt.Name
	// 		dt.Value = dt.BoolValue
	// 	}
	// 	resultMap[dt.DictName] = append(resultMap[dt.DictName], dt)
	// }
	//
	// // 确保所有请求的键都存在（即使没有数据）
	// for _, name := range dictNames {
	// 	if _, ok := resultMap[name]; !ok {
	// 		resultMap[name] = []domainTools.DictType{}
	// 	}
	// }
	//
	// return resultMap, nil
}

// GetListPage 分页查询列表
func (svc *dictTypeService) GetListPage(ctx context.Context, filter domainTools.DictTypeFilter) ([]domainTools.DictType, int64, error) {
	return svc.repo.GetListPage(ctx, filter)
}

// GetListAll 查询所有列表
func (svc *dictTypeService) GetListAll(ctx context.Context, filter domainTools.DictTypeFilter) ([]domainTools.DictType, error) {
	return svc.repo.GetListAll(ctx, filter)
}

// IsDuplicateEntryError 判断是否是唯一冲突错误
func (svc *dictTypeService) IsDuplicateEntryError(err error) bool {
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) {
		// MySQL 错误码 1062 表示唯一冲突
		return mysqlErr.Number == 1062
	}
	return false
}
