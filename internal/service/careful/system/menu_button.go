/**
 * Description：
 * FileName：menu_button.go
 * Author：CJiaの用心
 * Create：2025/12/05 14:50:08
 * Remark：
 */

package system

import (
	"context"
	"errors"
	"fmt"
	domainSystem "github.com/carefuly/careful-admin-go-gin/internal/domain/careful/system"
	"github.com/carefuly/careful-admin-go-gin/internal/model/careful/system"
	repositorySystem "github.com/carefuly/careful-admin-go-gin/internal/repository/repository/careful/system"
	"github.com/carefuly/careful-admin-go-gin/pkg/models"
	"strings"
)

var (
	ErrMenuButtonNotFound             = repositorySystem.ErrMenuButtonNotFound
	ErrMenuButtonVersionInconsistency = repositorySystem.ErrMenuButtonVersionInconsistency
)

type MenuButtonService interface {
	Create(ctx context.Context, domain domainSystem.MenuButton) error
	QuickCreate(ctx context.Context, menuId string, user domainSystem.User) error
	Delete(ctx context.Context, id string) error
	BatchDelete(ctx context.Context, ids []string) error
	Update(ctx context.Context, domain domainSystem.MenuButton) error

	GetById(ctx context.Context, id string) (domainSystem.MenuButton, error)
	GetListPage(ctx context.Context, filters domainSystem.MenuButtonFilter) ([]domainSystem.MenuButton, int64, error)
	GetListAll(ctx context.Context, filters domainSystem.MenuButtonFilter) ([]domainSystem.MenuButton, error)
}

type menuButtonService struct {
	repo     repositorySystem.MenuButtonRepository
	menuRepo repositorySystem.MenuRepository
}

func NewMenuButtonService(repo repositorySystem.MenuButtonRepository, menuRepo repositorySystem.MenuRepository) MenuButtonService {
	return &menuButtonService{
		repo:     repo,
		menuRepo: menuRepo,
	}
}

// Create 创建
func (svc *menuButtonService) Create(ctx context.Context, domain domainSystem.MenuButton) error {
	_, err := svc.repo.Create(ctx, domain)
	return err
}

// QuickCreate 快速添加
func (svc *menuButtonService) QuickCreate(ctx context.Context, menuId string, user domainSystem.User) error {
	domain, err := svc.menuRepo.GetById(ctx, menuId)
	if err != nil {
		if errors.Is(err, repositorySystem.ErrMenuNotFound) {
			return repositorySystem.ErrMenuNotFound
		}
		return err
	}
	if domain.Id == "" {
		return repositorySystem.ErrMenuNotFound
	}

	buttons := svc.publicMenuButton(ctx, domain, user)

	// 循环新增按钮，一个报错，直接终止
	for _, button := range buttons {
		err := svc.Create(ctx, button)
		if err != nil {
			return err
		}
	}

	return nil
}

// Delete 删除
func (svc *menuButtonService) Delete(ctx context.Context, id string) error {
	return svc.repo.Delete(ctx, id)
}

// BatchDelete 批量删除
func (svc *menuButtonService) BatchDelete(ctx context.Context, ids []string) error {
	return svc.repo.BatchDelete(ctx, ids)
}

// Update 更新
func (svc *menuButtonService) Update(ctx context.Context, domain domainSystem.MenuButton) error {
	err := svc.repo.Update(ctx, domain)
	if err != nil {
		if errors.Is(err, repositorySystem.ErrMenuButtonVersionInconsistency) {
			return repositorySystem.ErrMenuButtonVersionInconsistency
		}
		return err
	}
	return nil
}

// GetById 获取详情
func (svc *menuButtonService) GetById(ctx context.Context, id string) (domainSystem.MenuButton, error) {
	domain, err := svc.repo.GetById(ctx, id)
	if err != nil {
		if errors.Is(err, repositorySystem.ErrMenuButtonNotFound) {
			return domain, repositorySystem.ErrMenuButtonNotFound
		}
		return domain, err
	}
	if domain.Id == "" {
		return domain, repositorySystem.ErrMenuButtonNotFound
	}
	return domain, err
}

// GetListPage 分页查询列表
func (svc *menuButtonService) GetListPage(ctx context.Context, filters domainSystem.MenuButtonFilter) ([]domainSystem.MenuButton, int64, error) {
	return svc.repo.GetListPage(ctx, filters)
}

// GetListAll 查询所有列表
func (svc *menuButtonService) GetListAll(ctx context.Context, filters domainSystem.MenuButtonFilter) ([]domainSystem.MenuButton, error) {
	return svc.repo.GetListAll(ctx, filters)
}

// publicMenuButton 生成公共按钮权限
func (svc *menuButtonService) publicMenuButton(ctx context.Context, menu domainSystem.Menu, user domainSystem.User) []domainSystem.MenuButton {
	title := strings.Split(menu.Title, ".")

	return []domainSystem.MenuButton{
		{
			MenuButton: system.MenuButton{
				CoreModels: models.CoreModels{
					Sort:       1,
					Creator:    user.Id,
					Modifier:   user.Id,
					BelongDept: user.DeptID,
					Remark:     "",
				},
				Status:   true,
				Title:    "新增",
				AuthMark: fmt.Sprintf("%s:create", menu.Path),
				Method:   2,
				Api:      fmt.Sprintf("/%s/%s/create", title[1], title[2]),
				MenuID:   menu.Id,
			},
		},
		{
			MenuButton: system.MenuButton{
				CoreModels: models.CoreModels{
					Sort:       2,
					Creator:    user.Id,
					Modifier:   user.Id,
					BelongDept: user.DeptID,
					Remark:     "",
				},
				Status:   true,
				Title:    "导入",
				AuthMark: fmt.Sprintf("%s:import", menu.Path),
				Method:   2,
				Api:      fmt.Sprintf("/%s/%s/import", title[1], title[2]),
				MenuID:   menu.Id,
			},
		},
		{
			MenuButton: system.MenuButton{
				CoreModels: models.CoreModels{
					Sort:       3,
					Creator:    user.Id,
					Modifier:   user.Id,
					BelongDept: user.DeptID,
					Remark:     "",
				},
				Status:   true,
				Title:    "删除",
				AuthMark: fmt.Sprintf("%s:delete", menu.Path),
				Method:   4,
				Api:      fmt.Sprintf("/%s/%s/delete/:id", title[1], title[2]),
				MenuID:   menu.Id,
			},
		},
		{
			MenuButton: system.MenuButton{
				CoreModels: models.CoreModels{
					Sort:       4,
					Creator:    user.Id,
					Modifier:   user.Id,
					BelongDept: user.DeptID,
					Remark:     "",
				},
				Status:   true,
				Title:    "批量删除",
				AuthMark: fmt.Sprintf("%s:batchDelete", menu.Path),
				Method:   2,
				Api:      fmt.Sprintf("/%s/%s/batchDelete", title[1], title[2]),
				MenuID:   menu.Id,
			},
		},
		{
			MenuButton: system.MenuButton{
				CoreModels: models.CoreModels{
					Sort:       5,
					Creator:    user.Id,
					Modifier:   user.Id,
					BelongDept: user.DeptID,
					Remark:     "",
				},
				Status:   true,
				Title:    "更新",
				AuthMark: fmt.Sprintf("%s:update", menu.Path),
				Method:   3,
				Api:      fmt.Sprintf("/%s/%s/update", title[1], title[2]),
				MenuID:   menu.Id,
			},
		},
		{
			MenuButton: system.MenuButton{
				CoreModels: models.CoreModels{
					Sort:       6,
					Creator:    user.Id,
					Modifier:   user.Id,
					BelongDept: user.DeptID,
					Remark:     "",
				},
				Status:   true,
				Title:    "查询",
				AuthMark: fmt.Sprintf("%s:query", menu.Path),
				Method:   1,
				Api:      fmt.Sprintf("/%s/%s", title[1], title[2]),
				MenuID:   menu.Id,
			},
		},
		{
			MenuButton: system.MenuButton{
				CoreModels: models.CoreModels{
					Sort:       7,
					Creator:    user.Id,
					Modifier:   user.Id,
					BelongDept: user.DeptID,
					Remark:     "",
				},
				Status:   true,
				Title:    "导出",
				AuthMark: fmt.Sprintf("%s:export", menu.Path),
				Method:   1,
				Api:      fmt.Sprintf("/%s/%s/export", title[1], title[2]),
				MenuID:   menu.Id,
			},
		},
	}
}
