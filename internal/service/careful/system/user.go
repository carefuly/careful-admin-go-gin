/**
 * Description：
 * FileName：user.go
 * Author：CJiaの用心
 * Create：2025/10/9 16:52:52
 * Remark：
 */

package system

import (
	"context"
	"errors"
	domainSystem "github.com/carefuly/careful-admin-go-gin/internal/domain/careful/system"
	repositorySystem "github.com/carefuly/careful-admin-go-gin/internal/repository/repository/careful/system"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserNotFound          = repositorySystem.ErrUserNotFound
	ErrUserInvalidCredential = errors.New("用户名/密码错误")
	ErrUserHasBeen           = errors.New("用户已被禁用")
)

type UserService interface {
	Login(ctx context.Context, username, password string) (domainSystem.User, error)
	GetById(ctx context.Context, id string) (domainSystem.User, error)
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
	if !domain.Status {
		return domainSystem.User{}, ErrUserHasBeen
	}

	return domain, nil
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
