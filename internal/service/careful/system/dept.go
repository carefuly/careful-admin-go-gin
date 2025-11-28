/**
 * Description：
 * FileName：dept.go
 * Author：CJiaの用心
 * Create：2025/11/26 09:30:53
 * Remark：
 */

package system

import (
	"context"
	"errors"
	domainSystem "github.com/carefuly/careful-admin-go-gin/internal/domain/careful/system"
	repositorySystem "github.com/carefuly/careful-admin-go-gin/internal/repository/repository/careful/system"
	"github.com/go-sql-driver/mysql"
	"strings"
)

var (
	ErrDeptNotFound             = repositorySystem.ErrDeptNotFound
	ErrDeptCodeDuplicate        = repositorySystem.ErrDeptCodeDuplicate
	ErrDeptNameParentDuplicate  = repositorySystem.ErrDeptNameParentDuplicate
	ErrDeptParentNotFound       = repositorySystem.ErrDeptParentNotFound
	ErrDeptDisabled             = repositorySystem.ErrDeptDisabled
	ErrDeptYourParent           = errors.New("不能将自己设置为父部门")
	ErrDeptCycleReference       = errors.New("不能将子部门设置为父部门，会形成循环引用")
	ErrDeptHasChildren          = repositorySystem.ErrDeptHasChildren
	ErrDeptHasUsers             = repositorySystem.ErrDeptHasUsers
	ErrDeptVersionInconsistency = repositorySystem.ErrDeptVersionInconsistency
)

// DeptTree 部门树形结构
type DeptTree struct {
	domainSystem.Dept             // 父部门
	Children          []*DeptTree `json:"children"` // 子部门列表
}

type DeptService interface {
	Create(ctx context.Context, domain domainSystem.Dept) error
	Delete(ctx context.Context, id string) error
	BatchDelete(ctx context.Context, ids []string) error
	Update(ctx context.Context, domain domainSystem.Dept) error

	GetById(ctx context.Context, id string) (domainSystem.Dept, error)
	GetListTree(ctx context.Context, filters domainSystem.DeptFilter) ([]DeptTree, error)
	GetListAll(ctx context.Context, filters domainSystem.DeptFilter) ([]domainSystem.Dept, error)
}

type deptService struct {
	repo repositorySystem.DeptRepository
}

func NewDeptService(repo repositorySystem.DeptRepository) DeptService {
	return &deptService{
		repo: repo,
	}
}

// Create 创建
func (svc *deptService) Create(ctx context.Context, domain domainSystem.Dept) error {
	root := "root"
	// 判空处理
	if domain.ParentID == nil || *domain.ParentID == "" {
		domain.ParentID = &root
	}

	exists, err := svc.repo.CheckExistByCode(ctx, domain.Code, "")
	if err != nil {
		return err
	}
	if exists {
		return repositorySystem.ErrDeptCodeDuplicate
	}
	exists, err = svc.repo.CheckExistByNameAndParentId(ctx, domain.Name, *domain.ParentID, "")
	if err != nil {
		return err
	}
	if exists {
		return repositorySystem.ErrDeptNameParentDuplicate
	}
	// 查询父部门
	parent, err := svc.repo.GetByParentId(ctx, *domain.ParentID)
	if err != nil {
		if errors.Is(err, repositorySystem.ErrDeptParentNotFound) {
			return repositorySystem.ErrDeptParentNotFound
		}
		return err
	}
	if parent.Id == "" {
		return repositorySystem.ErrDeptParentNotFound
	}
	if !parent.Status {
		return repositorySystem.ErrDeptDisabled
	}
	// 保存时自动计算层级和路径
	domain.Level = parent.Level + 1
	domain.Path = parent.Path + parent.Id + "/"

	if _, err := svc.repo.Create(ctx, domain); err != nil {
		// 分析具体冲突字段
		if field, isDuplicate := svc.IsDuplicateEntryError(err); isDuplicate {
			switch field {
			case "code":
				return repositorySystem.ErrDeptCodeDuplicate
			case "name_parent":
				return repositorySystem.ErrDeptNameParentDuplicate
			}
		}
		return err
	}

	return nil
}

// Delete 删除
func (svc *deptService) Delete(ctx context.Context, id string) error {
	canDelete, err := svc.canDelete(ctx, id)
	if err != nil {
		return err
	}

	if !canDelete {
		leaf, _ := svc.isLeaf(ctx, id)
		if !leaf {
			return ErrDeptHasChildren
		}
		count, _ := svc.GetUserCount(ctx, id)
		if count > 0 {
			return ErrDeptHasUsers
		}
	}

	return svc.repo.Delete(ctx, id)
}

// BatchDelete 批量删除
func (svc *deptService) BatchDelete(ctx context.Context, ids []string) error {
	var failedIds []string

	for _, id := range ids {
		canDelete, _ := svc.canDelete(ctx, id)
		if canDelete {
			failedIds = append(failedIds, id)
			continue
		}
	}

	return svc.repo.BatchDelete(ctx, failedIds)
}

// Update 更新
func (svc *deptService) Update(ctx context.Context, domain domainSystem.Dept) error {
	root := "root"
	// 判空处理
	if domain.ParentID == nil || *domain.ParentID == "" {
		domain.ParentID = &root
	}

	exists, err := svc.repo.CheckExistByCode(ctx, domain.Code, domain.Id)
	if err != nil {
		return err
	}
	if exists {
		return repositorySystem.ErrDeptCodeDuplicate
	}
	exists, err = svc.repo.CheckExistByNameAndParentId(ctx, domain.Name, *domain.ParentID, domain.Id)
	if err != nil {
		return err
	}
	if exists {
		return repositorySystem.ErrDeptNameParentDuplicate
	}
	// 不能将自己设置为父部门
	if *domain.ParentID == domain.Id {
		return ErrDeptYourParent
	}
	// 查询父部门
	parent, err := svc.repo.GetByParentId(ctx, *domain.ParentID)
	if err != nil {
		if errors.Is(err, repositorySystem.ErrDeptParentNotFound) {
			return repositorySystem.ErrDeptParentNotFound
		}
		return err
	}
	if parent.Id == "" {
		return repositorySystem.ErrDeptParentNotFound
	}
	if !parent.Status {
		return repositorySystem.ErrDeptDisabled
	}
	// 不能将子部门设置为父部门，会形成循环引用
	ancestors, err := svc.repo.GetAncestors(ctx, domain)
	if err != nil {
		return err
	}

	// 目标部门是否在父部门的祖先链中 → 是则循环引用
	for _, ancestor := range ancestors {
		if ancestor.Id == domain.Id {
			return ErrDeptCycleReference
		}
	}
	// 保存时自动计算层级和路径
	domain.Level = parent.Level + 1
	domain.Path = parent.Path + parent.Id + "/"

	err = svc.repo.Update(ctx, domain)
	if err != nil {
		// 分析具体冲突字段
		if field, isDuplicate := svc.IsDuplicateEntryError(err); isDuplicate {
			switch field {
			case "code":
				return repositorySystem.ErrDeptCodeDuplicate
			case "name_parent":
				return repositorySystem.ErrDeptNameParentDuplicate
			}
		}
		switch {
		case errors.Is(err, repositorySystem.ErrDeptVersionInconsistency):
			return repositorySystem.ErrDeptVersionInconsistency
		default:
			return err
		}
	}

	return err
}

// GetById 获取详情
func (svc *deptService) GetById(ctx context.Context, id string) (domainSystem.Dept, error) {
	domain, err := svc.repo.GetById(ctx, id)
	if err != nil {
		if errors.Is(err, repositorySystem.ErrDeptNotFound) {
			return domain, repositorySystem.ErrDeptNotFound
		}
		return domain, err
	}
	if domain.Id == "" {
		return domain, repositorySystem.ErrDeptNotFound
	}
	return domain, err
}

// GetUserCount 获取部门下的用户数量
func (svc *deptService) GetUserCount(ctx context.Context, id string) (int64, error) {
	return svc.repo.GetUserCount(ctx, id)
}

// GetListTree 查询部门树
func (svc *deptService) GetListTree(ctx context.Context, filters domainSystem.DeptFilter) ([]DeptTree, error) {
	list, err := svc.repo.GetListAll(ctx, filters)
	if err != nil {
		return nil, err
	}
	// 构建树
	deptMap := make(map[string]*DeptTree)
	var roots []DeptTree

	if len(list) == 0 {
		return []DeptTree{}, nil
	}

	// 第一遍遍历，创建所有节点
	for _, dept := range list {
		deptMap[dept.Id] = &DeptTree{
			Dept:     dept,
			Children: []*DeptTree{},
		}
	}
	// 第二遍遍历，构建树结构
	for _, dept := range list {
		node := deptMap[dept.Id]
		if dept.ParentID == nil || *dept.ParentID == "root" || deptMap[*dept.ParentID] == nil {
			roots = append(roots, *node)
		} else {
			parent := deptMap[*dept.ParentID]
			parent.Children = append(parent.Children, node)
		}
	}

	return roots, nil
}

// GetListAll 查询所有列表
func (svc *deptService) GetListAll(ctx context.Context, filters domainSystem.DeptFilter) ([]domainSystem.Dept, error) {
	return svc.repo.GetListAll(ctx, filters)
}

// IsDuplicateEntryError 分析错误消息中的索引名
func (svc *deptService) IsDuplicateEntryError(err error) (string, bool) {
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
	case strings.Contains(mysqlErr.Message, "uni_name_parent"):
		return "name_parent", true
	case strings.Contains(mysqlErr.Message, "idx_careful_system_dept_code"):
		return "code", true
	default:
		return "unknown", true // 未知唯一键冲突
	}
}

// isLeaf 判断是否为叶子节点（没有子部门）
func (svc *deptService) isLeaf(ctx context.Context, id string) (bool, error) {
	count, err := svc.repo.GetChildCount(ctx, id)
	return count == 0, err
}

// canDelete 判断是否可以删除（没有子部门和用户）
func (svc *deptService) canDelete(ctx context.Context, id string) (bool, error) {
	isLeaf, err := svc.isLeaf(ctx, id)
	if err != nil || !isLeaf {
		return false, err
	}

	userCount, err := svc.GetUserCount(ctx, id)
	if err != nil || userCount > 0 {
		return false, err
	}

	return true, nil
}
