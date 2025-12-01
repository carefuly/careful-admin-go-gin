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
	domainSystem "github.com/carefuly/careful-admin-go-gin/internal/domain/careful/system"
	repositorySystem "github.com/carefuly/careful-admin-go-gin/internal/repository/repository/careful/system"
	"github.com/go-sql-driver/mysql"
)

var (
	ErrPostNotFound             = repositorySystem.ErrPostNotFound
	ErrPostCodeDuplicate        = repositorySystem.ErrPostCodeDuplicate
	ErrPostHasUsers             = repositorySystem.ErrPostHasUsers
	ErrPostVersionInconsistency = repositorySystem.ErrPostVersionInconsistency
)

type PostService interface {
	Create(ctx context.Context, domain domainSystem.Post) error
	Import(ctx context.Context)
	Delete(ctx context.Context, id string) error
	BatchDelete(ctx context.Context, ids []string) error
	Update(ctx context.Context, domain domainSystem.Post) error

	GetById(ctx context.Context, id string) (domainSystem.Post, error)
	GetListPage(ctx context.Context, filters domainSystem.PostFilter) ([]domainSystem.Post, int64, error)
	GetListAll(ctx context.Context, filters domainSystem.PostFilter) ([]domainSystem.Post, error)
}

type postService struct {
	repo repositorySystem.PostRepository
}

func NewPostService(repo repositorySystem.PostRepository) PostService {
	return &postService{
		repo: repo,
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
func (svc *postService) Import(ctx context.Context) {

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
