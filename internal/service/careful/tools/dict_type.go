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
	"fmt"
	domainSystem "github.com/carefuly/careful-admin-go-gin/internal/domain/careful/system"
	domainTools "github.com/carefuly/careful-admin-go-gin/internal/domain/careful/tools"
	modelTools "github.com/carefuly/careful-admin-go-gin/internal/model/careful/tools"
	repositoryTools "github.com/carefuly/careful-admin-go-gin/internal/repository/repository/careful/tools"
	"github.com/carefuly/careful-admin-go-gin/pkg/constants/careful/tools/dict_type"
	"github.com/carefuly/careful-admin-go-gin/pkg/models"
	_string "github.com/carefuly/careful-admin-go-gin/pkg/utils/common/string"
	"github.com/carefuly/careful-admin-go-gin/pkg/utils/enumconv"
	_import "github.com/carefuly/careful-admin-go-gin/pkg/utils/import"
	"github.com/carefuly/careful-admin-go-gin/pkg/utils/json_format"
	"github.com/go-sql-driver/mysql"
	"strconv"
)

var (
	ErrDictTypeInvalidDictValueType = repositoryTools.ErrDictTypeInvalidDictValueType
	ErrDictTypeNotFound             = repositoryTools.ErrDictTypeNotFound
	ErrDictTypeDuplicate            = repositoryTools.ErrDictTypeDuplicate
	ErrDictTypeVersionInconsistency = repositoryTools.ErrDictTypeVersionInconsistency
)

type DictTypeService interface {
	Create(ctx context.Context, domain domainTools.DictType) (domainTools.DictType, error)
	Import(ctx context.Context, user domainSystem.User, listMap []map[string]string) _import.ImportResult
	Delete(ctx context.Context, id string) error
	BatchDelete(ctx context.Context, ids []string) error
	Update(ctx context.Context, domain domainTools.DictType) error

	GetById(ctx context.Context, id string) (domainTools.DictType, error)
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
func (svc *dictTypeService) Create(ctx context.Context, domain domainTools.DictType) (domainTools.DictType, error) {
	// 获取字典详情
	dict, err := svc.dictRepo.GetById(ctx, domain.DictID)
	if err != nil {
		if errors.Is(err, repositoryTools.ErrDictNotFound) {
			return domainTools.DictType{}, repositoryTools.ErrDictNotFound
		}
		return domainTools.DictType{}, err
	}

	if dict.Id == "" {
		return domainTools.DictType{}, repositoryTools.ErrDictNotFound
	}

	if !dict.Status {
		return domainTools.DictType{}, repositoryTools.ErrDictDisabled
	}

	// 设置DictName和TypeValue
	domain.DictName = dict.Name
	domain.ValueType = dict.ValueType

	// 唯一性校验
	// 逻辑较为复杂，暂时不实现，默认使用mysql唯一性约束
	domain, err = svc.repo.Create(ctx, domain)
	if err != nil {
		if svc.IsDuplicateEntryError(err) {
			return domainTools.DictType{}, repositoryTools.ErrDictTypeDuplicate
		}
		return domainTools.DictType{}, err
	}

	return domain, err
}

// Import 导入
func (svc *dictTypeService) Import(ctx context.Context, user domainSystem.User, listMap []map[string]string) _import.ImportResult {
	result := _import.ImportResult{}

	// 遍历数据
	for index, list := range listMap {
		// 数据清洗
		name := _string.CleanInputString(list["字典项名称"])
		dictName := _string.CleanInputString(list["所属字典"])

		// 字段校验
		if name == "" {
			list["导入状态"] = "400"
			list["导入结果"] = "❌【字典项名称】不能为空"
			continue
		}
		if dictName == "" {
			list["导入状态"] = "400"
			list["导入结果"] = "❌【所属字典】不能为空"
			continue
		}

		var intValue int
		if list["整型值"] == "" {
			intValue = 0
		} else {
			intValue, _ = strconv.Atoi(list["整型值"])
		}

		var boolValue bool
		if list["布尔值"] == "是" {
			boolValue = true
		} else {
			boolValue = false
		}

		// 类型转换
		dictTagValues := []string{"primary", "success", "warning", "danger", "info"}
		converter := enumconv.NewEnumConverter(dict_type.DictTagMapping, dict_type.DictTagImportMapping, dictTagValues, "标签类型")
		dictTag, err := converter.ToEnum(list["标签类型"])
		if err != nil {
			list["导入状态"] = "400"
			list["导入结果"] = fmt.Sprintf("❌【标签类型】转换失败：%s", err.Error())
			continue
		}

		// 所属字典校验
		dict, err := svc.dictRepo.GetByName(ctx, dictName)
		if err != nil {
			if errors.Is(err, repositoryTools.ErrDictNotFound) {
				list["导入状态"] = "400"
				list["导入结果"] = fmt.Sprintf("❌所属字典【%s】不存在", dictName)
				continue
			}
			list["导入状态"] = "400"
			list["导入结果"] = err.Error()
			continue
		}
		if dict.Id == "" {
			list["导入状态"] = "400"
			list["导入结果"] = fmt.Sprintf("❌所属字典【%s】不存在", dictName)
			continue
		}
		if !dict.Status {
			list["导入状态"] = "400"
			list["导入结果"] = fmt.Sprintf("❌所属字典【%s】已被禁用，无法在其下创建字典项", dictName)
			continue
		}

		// 处理字段
		var sort int
		if list["排序"] == "" {
			sort = 1
		} else {
			sort, _ = strconv.Atoi(list["排序"])
		}

		// 构建领域模型
		domain := domainTools.DictType{
			DictType: modelTools.DictType{
				CoreModels: models.CoreModels{
					Creator:    user.Id,
					Modifier:   user.Id,
					BelongDept: user.DeptID,
					Sort:       sort,
					Remark:     list["备注"],
				},
				Status:      true,
				Name:        name,
				DictTag:     dictTag,
				DictColor:   list["标签颜色"],
				DictName:    dict.Name,
				ValueType:   dict.ValueType,
				Description: list["字典项描述"],
				DictID:      dict.Id,
			},
			StrValue:  list["字符串值"],
			IntValue:  int64(intValue),
			BoolValue: boolValue,
		}

		fmt.Println("导入行数据 >>> ", index+2)
		json_format.PrintFormattedJSON(domain)

		// 创建记录
		if _, err = svc.repo.Create(ctx, domain); err != nil {
			list["导入状态"] = "400"
			list["导入结果"] = fmt.Sprintf("❌创建失败：%s", err.Error())
			continue
		}

		list["导入状态"] = "200"
		list["导入结果"] = "✅创建成功"
	}

	result.Result = listMap

	return result
}

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
	dict, err := svc.dictRepo.GetById(ctx, domain.DictID)
	if err != nil {
		if errors.Is(err, repositoryTools.ErrDictNotFound) {
			return repositoryTools.ErrDictNotFound
		}
		return err
	}

	if dict.Id == "" {
		return repositoryTools.ErrDictNotFound
	}

	if !dict.Status {
		return repositoryTools.ErrDictDisabled
	}

	// 设置DictName和TypeValue
	domain.DictName = dict.Name
	domain.ValueType = dict.ValueType

	// 唯一性校验
	// 逻辑较为复杂，暂时不实现，默认使用mysql唯一性约束
	err = svc.repo.Update(ctx, domain)
	if err != nil {
		if svc.IsDuplicateEntryError(err) {
			return repositoryTools.ErrDictTypeDuplicate
		}
		switch {
		case errors.Is(err, repositoryTools.ErrDictTypeVersionInconsistency):
			return repositoryTools.ErrDictTypeVersionInconsistency
		default:
			return err
		}
	}

	return nil
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

// GetListPage 分页查询列表
func (svc *dictTypeService) GetListPage(ctx context.Context, filters domainTools.DictTypeFilter) ([]domainTools.DictType, int64, error) {
	list, row, err := svc.repo.GetListPage(ctx, filters)
	if err != nil {
		return []domainTools.DictType{}, row, err
	}
	if list == nil || len(list) == 0 || row == 0 {
		return []domainTools.DictType{}, row, nil
	}
	return list, row, nil
}

// GetListAll 查询所有列表
func (svc *dictTypeService) GetListAll(ctx context.Context, filter domainTools.DictTypeFilter) ([]domainTools.DictType, error) {
	list, err := svc.repo.GetListAll(ctx, filter)
	if err != nil {
		return []domainTools.DictType{}, err
	}
	if list == nil || len(list) == 0 {
		return []domainTools.DictType{}, nil
	}
	return list, nil
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
