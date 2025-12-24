/**
 * Description：
 * FileName：user.go
 * Author：CJiaの用心
 * Create：2025/11/24 17:11:36
 * Remark：
 */

package system

import (
	"context"
	"errors"
	domainSystem "github.com/carefuly/careful-admin-go-gin/internal/domain/careful/system"
	repositorySystem "github.com/carefuly/careful-admin-go-gin/internal/repository/repository/careful/system"
	"github.com/carefuly/careful-admin-go-gin/pkg/constants/careful/system/user"
	"github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserNotFound             = repositorySystem.ErrUserNotFound
	ErrUserUsernameDuplicate    = repositorySystem.ErrUserUsernameDuplicate
	ErrUserVersionInconsistency = repositorySystem.ErrUserVersionInconsistency
	ErrUserInvalidCredential    = errors.New("用户名/密码错误")
	ErrUserDisabled             = errors.New("用户已被禁用")
	ErrUserLocked               = errors.New("用户已被锁定")
)

type UserService interface {
	Login(ctx context.Context, username, password string) (domainSystem.User, error)

	Create(ctx context.Context, domain domainSystem.User) error
	Delete(ctx context.Context, id string) error
	BatchDelete(ctx context.Context, ids []string) error
	Update(ctx context.Context, domain domainSystem.User) error
	UpdateLoginField(ctx context.Context, lastLogin string, lastLoginIp string, domain domainSystem.User) error

	GetById(ctx context.Context, id string) (domainSystem.User, error)
	GetListPage(ctx context.Context, filters domainSystem.UserFilter) ([]domainSystem.User, int64, error)
	GetListAll(ctx context.Context, filters domainSystem.UserFilter) ([]domainSystem.User, error)
}

type userService struct {
	repo repositorySystem.UserRepository
}

func NewUserService(repo repositorySystem.UserRepository) UserService {
	return &userService{
		repo: repo,
	}
}

// Login 登录
func (svc *userService) Login(ctx context.Context, username, password string) (domainSystem.User, error) {
	// 根据用户名获取用户
	domain, err := svc.repo.GetByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, repositorySystem.ErrUserNotFound) {
			return domainSystem.User{}, ErrUserInvalidCredential
		}
		return domainSystem.User{}, err
	}

	// 验证密码
	err = bcrypt.CompareHashAndPassword([]byte(domain.Password), []byte(password))
	if err != nil {
		return domainSystem.User{}, ErrUserInvalidCredential
	}

	// 检查用户状态
	if domain.Status == user.StatusDisabled {
		return domainSystem.User{}, ErrUserDisabled
	} else if domain.Status == user.StatusLocked {
		return domainSystem.User{}, ErrUserLocked
	}

	return domain, nil
}

// Create 创建
func (svc *userService) Create(ctx context.Context, domain domainSystem.User) error {
	exists, err := svc.repo.CheckExistByUsername(ctx, domain.Username, "")
	if err != nil {
		return err
	}
	if exists {
		return repositorySystem.ErrUserUsernameDuplicate
	}

	// 创建
	domain, err = svc.repo.Create(ctx, domain)
	if err != nil {
		if svc.IsDuplicateEntryError(err) {
			return repositorySystem.ErrUserUsernameDuplicate
		}
		return err
	}

	// 更新
	err = svc.repo.Update(ctx, domain)
	if err != nil {
		return err
	}

	return nil
}

// Delete 删除
func (svc *userService) Delete(ctx context.Context, id string) error {
	return svc.repo.Delete(ctx, id)
}

// BatchDelete 批量删除
func (svc *userService) BatchDelete(ctx context.Context, ids []string) error {
	return svc.repo.BatchDelete(ctx, ids)
}

// Update 更新
func (svc *userService) Update(ctx context.Context, domain domainSystem.User) error {
	exists, err := svc.repo.CheckExistByUsername(ctx, domain.Username, domain.Id)
	if err != nil {
		return err
	}
	if exists {
		return repositorySystem.ErrUserUsernameDuplicate
	}

	if err := svc.repo.Update(ctx, domain); err != nil {
		switch {
		case svc.IsDuplicateEntryError(err):
			return repositorySystem.ErrUserUsernameDuplicate
		case errors.Is(err, repositorySystem.ErrUserVersionInconsistency):
			return repositorySystem.ErrUserVersionInconsistency
		default:
			return err
		}
	}

	return nil
}

// UpdateLoginField 更新登录字段
func (svc *userService) UpdateLoginField(ctx context.Context, lastLogin string, lastLoginIp string, domain domainSystem.User) error {
	return svc.repo.UpdateLoginField(ctx, lastLogin, lastLoginIp, domain)
}

// GetById 获取详情
func (svc *userService) GetById(ctx context.Context, id string) (domainSystem.User, error) {
	domain, err := svc.repo.GetById(ctx, id)
	if err != nil {
		if errors.Is(err, repositorySystem.ErrUserNotFound) {
			return domainSystem.User{}, repositorySystem.ErrUserNotFound
		}
		return domainSystem.User{}, err
	}
	if domain.Id == "" {
		return domainSystem.User{}, repositorySystem.ErrUserNotFound
	}
	return domain, nil
}

// GetListPage 分页查询列表
func (svc *userService) GetListPage(ctx context.Context, filters domainSystem.UserFilter) ([]domainSystem.User, int64, error) {
	return svc.repo.GetListPage(ctx, filters)
}

// GetListAll 查询所有列表
func (svc *userService) GetListAll(ctx context.Context, filters domainSystem.UserFilter) ([]domainSystem.User, error) {
	return svc.repo.GetListAll(ctx, filters)
}

// IsDuplicateEntryError 判断是否是唯一冲突错误
func (svc *userService) IsDuplicateEntryError(err error) bool {
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) {
		// MySQL 错误码 1062 表示唯一冲突
		return mysqlErr.Number == 1062
	}
	return false
}
