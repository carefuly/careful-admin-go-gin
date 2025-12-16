/**
 * Description：
 * FileName：post.go
 * Author：CJiaの用心
 * Create：2025/11/29 01:52:02
 * Remark：
 */

package system

import (
	"context"
	"errors"
	"fmt"
	domainSystem "github.com/carefuly/careful-admin-go-gin/internal/domain/careful/system"
	modelSystem "github.com/carefuly/careful-admin-go-gin/internal/model/careful/system"
	repositorySystem "github.com/carefuly/careful-admin-go-gin/internal/repository/repository/careful/system"
	"github.com/carefuly/careful-admin-go-gin/pkg/constants/careful/system/post"
	"github.com/carefuly/careful-admin-go-gin/pkg/models"
	_string "github.com/carefuly/careful-admin-go-gin/pkg/utils/common/string"
	"github.com/carefuly/careful-admin-go-gin/pkg/utils/enumconv"
	_import "github.com/carefuly/careful-admin-go-gin/pkg/utils/import"
	"github.com/go-sql-driver/mysql"
	"strconv"
)

var (
	ErrPostNotFound             = repositorySystem.ErrPostNotFound
	ErrPostCodeDuplicate        = repositorySystem.ErrPostCodeDuplicate
	ErrPostHasUsers             = repositorySystem.ErrPostHasUsers
	ErrPostVersionInconsistency = repositorySystem.ErrPostVersionInconsistency
)

type PostService interface {
	Create(ctx context.Context, domain domainSystem.Post) error
	Import(ctx context.Context, user domainSystem.User, listMap []map[string]string) _import.ImportResult
	Delete(ctx context.Context, id string) error
	BatchDelete(ctx context.Context, ids []string) error
	Update(ctx context.Context, domain domainSystem.Post) error

	GetById(ctx context.Context, id string) (domainSystem.Post, error)
	GetListPage(ctx context.Context, filters domainSystem.PostFilter) ([]domainSystem.Post, int64, error)
	GetListAll(ctx context.Context, filters domainSystem.PostFilter) ([]domainSystem.Post, error)
}

type postService struct {
	repo     repositorySystem.PostRepository
	deptRepo repositorySystem.DeptRepository
}

func NewPostService(repo repositorySystem.PostRepository, deptRepo repositorySystem.DeptRepository) PostService {
	return &postService{
		repo:     repo,
		deptRepo: deptRepo,
	}
}

// Create 创建
func (svc *postService) Create(ctx context.Context, domain domainSystem.Post) error {
	exists, err := svc.repo.CheckExistByCode(ctx, domain.Code, "")
	if err != nil {
		return err
	}
	if exists {
		return repositorySystem.ErrPostCodeDuplicate
	}

	// 创建
	if _, err := svc.repo.Create(ctx, domain); err != nil {
		if svc.IsDuplicateEntryError(err) {
			return repositorySystem.ErrPostCodeDuplicate
		}
		return err
	}

	return nil
}

// Import 导入
func (svc *postService) Import(ctx context.Context, user domainSystem.User, listMap []map[string]string) _import.ImportResult {
	result := _import.ImportResult{}

	// 遍历数据
	for _, list := range listMap {
		// 数据清洗
		name := _string.CleanInputString(list["岗位名称"])
		code := _string.CleanInputString(list["岗位编码"])

		// 字段校验
		if name == "" {
			list["导入状态"] = "400"
			list["导入结果"] = "❌【岗位名称】不能为空"
			continue
		}
		if code == "" {
			list["导入状态"] = "400"
			list["导入结果"] = "❌【岗位编码】不能为空"
			continue
		}

		// 校验参数
		typeValidValues := []string{"管理岗", "技术岗", "业务岗", "职能岗", "其他"}
		converter := enumconv.NewEnumConverter(post.TypeMapping, post.TypeImportMapping, typeValidValues, "岗位类型")
		postType, err := converter.ToEnum(list["岗位类型"])
		if err != nil {
			list["导入状态"] = "400"
			list["导入结果"] = fmt.Sprintf("❌【岗位类型】转换失败：%s", err.Error())
			continue
		}
		levelValidValues := []string{"高层", "中层", "基层", "一般员工"}
		levelConverter := enumconv.NewEnumConverter(post.LevelMapping, post.LevelImportMapping, levelValidValues, "岗位级别")
		level, err := levelConverter.ToEnum(list["岗位级别"])
		if err != nil {
			list["导入状态"] = "400"
			list["导入结果"] = fmt.Sprintf("❌【岗位级别】转换失败：%s", err.Error())
			continue
		}

		// 唯一性校验
		exists, err := svc.repo.CheckExistByCode(ctx, code, "")
		if err != nil {
			list["导入状态"] = "400"
			list["导入结果"] = fmt.Sprintf("❌检查【岗位编码：%s】唯一性失败：%s", code, err.Error())
			continue
		}
		if exists {
			list["导入状态"] = "400"
			list["导入结果"] = fmt.Sprintf("❌岗位编码【%s】已存在", code)
			continue
		}

		// 处理所属部门
		var deptId *string
		dept, err := svc.deptRepo.GetByCode(ctx, list["部门编码"])
		if err != nil {
			deptId = nil
			if errors.Is(err, repositorySystem.ErrDeptNotFound) {
				list["导入状态"] = "400"
				list["导入结果"] = fmt.Sprintf("❌【部门编码】不存在：%s", err.Error())
			}
			list["导入状态"] = "400"
			list["导入结果"] = fmt.Sprintf("❌【部门编码】查询异常：%s", err.Error())
		}
		if dept.Id == "" {
			deptId = nil
			list["导入状态"] = "400"
			list["导入结果"] = "❌【部门编码】不存在"
		}
		deptId = &dept.Id

		// 处理字段
		var sort int
		if list["排序"] == "" {
			sort = 1
		} else {
			sort, _ = strconv.Atoi(list["排序"])
		}

		// 构建领域模型
		domain := domainSystem.Post{
			Post: modelSystem.Post{
				CoreModels: models.CoreModels{
					Sort:       sort,
					Creator:    user.Id,
					Modifier:   user.Id,
					BelongDept: user.DeptID,
					Remark:     list["备注"],
				},
				Status:      true,
				Name:        name,
				Code:        code,
				PostType:    postType,
				Level:       level,
				Description: list["描述"],
				DeptID:      deptId,
			},
		}

		// 创建记录
		if _, err = svc.repo.Create(ctx, domain); err != nil {
			list["导入状态"] = "400"
			list["导入结果"] = fmt.Sprintf("创建失败：%s", err.Error())
			continue
		}
	}

	result.Result = listMap

	return result
}

// Delete 删除
func (svc *postService) Delete(ctx context.Context, id string) error {
	count := svc.repo.GetUserCount(ctx, id)
	if count > 0 {
		return repositorySystem.ErrPostHasUsers
	}
	return svc.repo.Delete(ctx, id)
}

// BatchDelete 批量删除
func (svc *postService) BatchDelete(ctx context.Context, ids []string) error {
	var failedIds []string

	for _, id := range ids {
		count := svc.repo.GetUserCount(ctx, id)
		if count == 0 {
			failedIds = append(failedIds, id)
			continue
		}
	}

	return svc.repo.BatchDelete(ctx, failedIds)
}

// Update 更新
func (svc *postService) Update(ctx context.Context, domain domainSystem.Post) error {
	exists, err := svc.repo.CheckExistByCode(ctx, domain.Code, domain.Id)
	if err != nil {
		return err
	}
	if exists {
		return repositorySystem.ErrPostCodeDuplicate
	}

	if err := svc.repo.Update(ctx, domain); err != nil {
		switch {
		case svc.IsDuplicateEntryError(err):
			return repositorySystem.ErrPostCodeDuplicate
		case errors.Is(err, repositorySystem.ErrPostVersionInconsistency):
			return repositorySystem.ErrPostVersionInconsistency
		default:
			return err
		}
	}

	return nil
}

// GetById 获取详情
func (svc *postService) GetById(ctx context.Context, id string) (domainSystem.Post, error) {
	domain, err := svc.repo.GetById(ctx, id)
	if err != nil {
		if errors.Is(err, repositorySystem.ErrPostNotFound) {
			return domain, repositorySystem.ErrPostNotFound
		}
		return domain, err
	}
	if domain.Id == "" {
		return domain, repositorySystem.ErrPostNotFound
	}
	return domain, err
}

// GetListPage 分页查询列表
func (svc *postService) GetListPage(ctx context.Context, filters domainSystem.PostFilter) ([]domainSystem.Post, int64, error) {
	return svc.repo.GetListPage(ctx, filters)
}

// GetListAll 查询所有列表
func (svc *postService) GetListAll(ctx context.Context, filters domainSystem.PostFilter) ([]domainSystem.Post, error) {
	return svc.repo.GetListAll(ctx, filters)
}

// IsDuplicateEntryError 判断是否是唯一冲突错误
func (svc *postService) IsDuplicateEntryError(err error) bool {
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) {
		// MySQL 错误码 1062 表示唯一冲突
		return mysqlErr.Number == 1062
	}
	return false
}
