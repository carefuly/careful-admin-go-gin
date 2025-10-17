/**
 * Description：
 * FileName：dict.go
 * Author：CJiaの用心
 * Create：2025/10/11 13:57:44
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
	"github.com/carefuly/careful-admin-go-gin/pkg/constants/careful/tools/dict"
	"github.com/carefuly/careful-admin-go-gin/pkg/models"
	_import "github.com/carefuly/careful-admin-go-gin/pkg/utils/common/import"
	_string "github.com/carefuly/careful-admin-go-gin/pkg/utils/common/string"
	"github.com/carefuly/careful-admin-go-gin/pkg/utils/enumconv"
	"github.com/go-sql-driver/mysql"
	"strconv"
	"strings"
)

var (
	ErrDictNotFound             = repositoryTools.ErrDictNotFound
	ErrDictNameDuplicate        = repositoryTools.ErrDictNameDuplicate
	ErrDictCodeDuplicate        = repositoryTools.ErrDictCodeDuplicate
	ErrDictDuplicate            = repositoryTools.ErrDictDuplicate
	ErrDictVersionInconsistency = repositoryTools.ErrDictVersionInconsistency
)

type DictService interface {
	Create(ctx context.Context, domain domainTools.Dict) error
	Import(ctx context.Context, user domainSystem.User, listMap []map[string]string) _import.ImportResult
	Delete(ctx context.Context, id string) error
	BatchDelete(ctx context.Context, ids []string) error
	Update(ctx context.Context, domain domainTools.Dict) error

	GetById(ctx context.Context, id string) (domainTools.Dict, error)
	GetByName(ctx context.Context, name string) (domainTools.Dict, error)
	GetListPage(ctx context.Context, filters domainTools.DictFilter) ([]domainTools.Dict, int64, error)
	GetListAll(ctx context.Context, filters domainTools.DictFilter) ([]domainTools.Dict, error)
}

type dictService struct {
	repo repositoryTools.DictRepository
}

func NewDictService(repo repositoryTools.DictRepository) DictService {
	return &dictService{
		repo: repo,
	}
}

// Create 创建
func (svc *dictService) Create(ctx context.Context, domain domainTools.Dict) error {
	exists, err := svc.repo.CheckExistByCode(ctx, domain.Code, "")
	if err != nil {
		return err
	}
	if exists {
		return repositoryTools.ErrDictCodeDuplicate
	}

	exists, err = svc.repo.CheckExistByName(ctx, domain.Name, "")
	if err != nil {
		return err
	}
	if exists {
		return repositoryTools.ErrDictNameDuplicate
	}

	if _, err := svc.repo.Create(ctx, domain); err != nil {
		// 分析具体冲突字段
		if field, isDuplicate := svc.IsDuplicateEntryError(err); isDuplicate {
			switch field {
			case "name":
				return repositoryTools.ErrDictNameDuplicate
			case "code":
				return repositoryTools.ErrDictCodeDuplicate
			case "all":
				return repositoryTools.ErrDictDuplicate
			default:
				return err
			}
		} else {
			return err
		}
	}

	return nil
}

// Import 导入
func (svc *dictService) Import(ctx context.Context, user domainSystem.User, listMap []map[string]string) _import.ImportResult {
	result := _import.ImportResult{}

	// 遍历数据
	for index, list := range listMap {
		rowNumber := index + 2

		// 数据清洗
		name := _string.CleanInputString(list["字典名称"])
		code := _string.CleanInputString(list["字典编码"])

		// 字段校验
		if name == "" {
			result.AddError(rowNumber, "【字典名称】不能为空")
			continue
		}
		if code == "" {
			result.AddError(rowNumber, "【字典编码】不能为空")
			continue
		}

		// 唯一性校验
		exists, err := svc.repo.CheckExistByName(ctx, name, "")
		if err != nil {
			result.AddError(rowNumber, fmt.Sprintf("检查【字典名称：%s】唯一性失败：%s", name, err.Error()))
			continue
		}
		if exists {
			result.AddError(rowNumber, fmt.Sprintf("字典名称【%s】已存在", name))
			continue
		}
		exists, err = svc.repo.CheckExistByCode(ctx, code, "")
		if err != nil {
			result.AddError(rowNumber, fmt.Sprintf("检查【字典编码：%s】唯一性失败：%s", code, err.Error()))
			continue
		}
		if exists {
			result.AddError(rowNumber, fmt.Sprintf("字典编码【%s】已存在", code))
			continue
		}

		// 类型转换
		typeValidValues := []string{"普通字典", "系统字典", "枚举字典"}
		converter := enumconv.NewEnumConverter(dict.TypeMapping, dict.TypeImportMapping, typeValidValues, "字典分类")
		dictType, err := converter.ToEnum(list["字典类型"])
		if err != nil {
			result.AddError(rowNumber, fmt.Sprintf("【字典类型】转换失败：%s", err.Error()))
			continue
		}
		valueTypeValidValues := []string{"字符串", "整型", "布尔"}
		valueTypeConverter := enumconv.NewEnumConverter(dict.TypeValueMapping, dict.TypeValueImportMapping, valueTypeValidValues, "字典值类型")
		valueType, err := valueTypeConverter.ToEnum(list["字典类型值"])
		if err != nil {
			result.AddError(rowNumber, fmt.Sprintf("【字典类型值】转换失败：%s", err.Error()))
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
		domain := domainTools.Dict{
			Dict: modelTools.Dict{
				CoreModels: models.CoreModels{
					Creator:    user.Id,
					Modifier:   user.Id,
					BelongDept: user.DeptId,
					Sort:       sort,
					Remark:     list["备注"],
				},
				Status:    true,
				Name:      name,
				Code:      code,
				Type:      dictType,
				ValueType: valueType,
			},
		}

		// 创建记录
		if _, err = svc.repo.Create(ctx, domain); err != nil {
			result.AddError(rowNumber, "创建失败："+err.Error())
			continue
		}

		result.SuccessCount++
	}

	return result
}

// Delete 删除
func (svc *dictService) Delete(ctx context.Context, id string) error {
	return svc.repo.Delete(ctx, id)
}

// BatchDelete 批量删除
func (svc *dictService) BatchDelete(ctx context.Context, ids []string) error {
	return svc.repo.BatchDelete(ctx, ids)
}

// Update 更新
func (svc *dictService) Update(ctx context.Context, domain domainTools.Dict) error {
	exists, err := svc.repo.CheckExistByCode(ctx, domain.Code, domain.Id)
	if err != nil {
		return err
	}
	if exists {
		return repositoryTools.ErrDictCodeDuplicate
	}

	exists, err = svc.repo.CheckExistByName(ctx, domain.Name, domain.Id)
	if err != nil {
		return err
	}
	if exists {
		return repositoryTools.ErrDictNameDuplicate
	}

	err = svc.repo.Update(ctx, domain)
	if err != nil {
		// 分析具体冲突字段
		if field, isDuplicate := svc.IsDuplicateEntryError(err); isDuplicate {
			switch field {
			case "name":
				return repositoryTools.ErrDictNameDuplicate
			case "code":
				return repositoryTools.ErrDictCodeDuplicate
			case "all":
				return repositoryTools.ErrDictDuplicate
			default:
				switch {
				case errors.Is(err, repositoryTools.ErrDictVersionInconsistency):
					return repositoryTools.ErrDictVersionInconsistency
				default:
					return err
				}
			}
		} else {
			return err
		}
	}

	return err
}

// GetById 获取详情
func (svc *dictService) GetById(ctx context.Context, id string) (domainTools.Dict, error) {
	domain, err := svc.repo.GetById(ctx, id)
	if err != nil {
		if errors.Is(err, repositoryTools.ErrDictNotFound) {
			return domain, repositoryTools.ErrDictNotFound
		}
		return domain, err
	}
	if domain.Id == "" {
		return domain, repositoryTools.ErrDictNotFound
	}
	return domain, err
}

// GetByName 根据name获取详情
func (svc *dictService) GetByName(ctx context.Context, name string) (domainTools.Dict, error) {
	domain, err := svc.repo.GetByName(ctx, name)
	if err != nil {
		if errors.Is(err, repositoryTools.ErrDictNotFound) {
			return domain, repositoryTools.ErrDictNotFound
		}
		return domain, err
	}
	if domain.Id == "" {
		return domain, repositoryTools.ErrDictNotFound
	}
	return domain, err
}

// GetListPage 分页查询列表
func (svc *dictService) GetListPage(ctx context.Context, filters domainTools.DictFilter) ([]domainTools.Dict, int64, error) {
	return svc.repo.GetListPage(ctx, filters)
}

// GetListAll 查询所有列表
func (svc *dictService) GetListAll(ctx context.Context, filters domainTools.DictFilter) ([]domainTools.Dict, error) {
	return svc.repo.GetListAll(ctx, filters)
}

// IsDuplicateEntryError 分析错误消息中的索引名
func (svc *dictService) IsDuplicateEntryError(err error) (string, bool) {
	var mysqlErr *mysql.MySQLError
	if !errors.As(err, &mysqlErr) {
		return "", false
	}

	// MySQL 错误码 1062 表示唯一冲突
	if mysqlErr.Number != 1062 {
		return "", false
	}

	// 分析错误消息中的索引名
	switch {
	case strings.Contains(mysqlErr.Message, "uni_careful_tools_dict_name"):
		return "name", true
	case strings.Contains(mysqlErr.Message, "uni_careful_tools_dict_code"):
		return "code", true
	default:
		return "all", true // 未知唯一键冲突
	}
}
