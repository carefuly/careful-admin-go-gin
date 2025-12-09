/**
 * Description：
 * FileName：menu.go
 * Author：CJiaの用心
 * Create：2025/12/05 14:51:21
 * Remark：
 */

package system

import (
	"context"
	"errors"
	domainSystem "github.com/carefuly/careful-admin-go-gin/internal/domain/careful/system"
	"github.com/carefuly/careful-admin-go-gin/internal/model/careful/system"
	repositorySystem "github.com/carefuly/careful-admin-go-gin/internal/repository/repository/careful/system"
	"github.com/carefuly/careful-admin-go-gin/pkg/models"
	"github.com/carefuly/careful-admin-go-gin/pkg/utils/json_format"
	"github.com/go-sql-driver/mysql"
	"sort"
)

var (
	ErrMenuNotFound             = repositorySystem.ErrMenuNotFound
	ErrMenuDuplicate            = repositorySystem.ErrMenuDuplicate
	ErrMenuParentNotFound       = repositorySystem.ErrMenuParentNotFound
	ErrMenuDisabled             = repositorySystem.ErrMenuDisabled
	ErrMenuHasChildren          = repositorySystem.ErrMenuHasChildren
	ErrMenuVersionInconsistency = repositorySystem.ErrMenuVersionInconsistency
)

// MenuRouteTree 菜单路由结构
type MenuRouteTree struct {
	Id        string          `json:"id"`        // 主键ID
	Name      string          `json:"name"`      // 菜单名称
	Path      string          `json:"path"`      // 路由地址
	Component string          `json:"component"` // 组件地址
	Meta      map[string]any  `json:"meta"`      // 元信息
	Sort      int             `json:"sort"`      // 排序
	Status    bool            `json:"status"`    // 状态
	ParentID  string          `json:"parent_id"` // 上级菜单
	Children  []MenuRouteTree `json:"children"`  // 子菜单
}

type MenuService interface {
	Create(ctx context.Context, domain domainSystem.Menu, isCreateButton bool, user domainSystem.User) error
	Delete(ctx context.Context, id string) error
	BatchDelete(ctx context.Context, ids []string) error
	Update(ctx context.Context, domain domainSystem.Menu) error

	GetById(ctx context.Context, id string) (domainSystem.Menu, error)
	GetMenuRouteTree(ctx context.Context, filter domainSystem.MenuFilter) ([]MenuRouteTree, error)
	GetListPage(ctx context.Context, filters domainSystem.MenuFilter) ([]domainSystem.Menu, int64, error)
	GetListAll(ctx context.Context, filter domainSystem.MenuFilter) ([]domainSystem.Menu, error)

	// GetMenuRouteTree(ctx context.Context, filter domainSystem.MenuFilter) ([]MenuRouteTree, error)
	// GetListTree(ctx context.Context, filter domainSystem.MenuFilter) ([]domainSystem.Menu, error)
}

type menuService struct {
	repo           repositorySystem.MenuRepository
	menuButtonRepo repositorySystem.MenuButtonRepository
}

func NewMenuService(repo repositorySystem.MenuRepository, menuButtonRepo repositorySystem.MenuButtonRepository) MenuService {
	return &menuService{
		repo:           repo,
		menuButtonRepo: menuButtonRepo,
	}
}

// Create 创建
func (svc *menuService) Create(ctx context.Context, domain domainSystem.Menu, isCreateButton bool, user domainSystem.User) error {
	root := "root"
	// 判空处理
	if domain.ParentID == nil || *domain.ParentID == "" {
		domain.ParentID = &root
	}
	// 检查name、path、title和parentId是否同时存在
	exists, err := svc.repo.CheckExistByNameAndPathAndTitleAndParentId(ctx,
		domain.Name, domain.Path, domain.Title, *domain.ParentID, "")
	if err != nil {
		return err
	}
	if exists {
		return repositorySystem.ErrMenuDuplicate
	}
	// 查询上级菜单
	parent, err := svc.repo.GetByParentId(ctx, *domain.ParentID)
	if err != nil {
		if errors.Is(err, repositorySystem.ErrMenuParentNotFound) {
			return repositorySystem.ErrMenuParentNotFound
		}
		return err
	}
	if parent.Id == "" {
		return repositorySystem.ErrMenuParentNotFound
	}
	if !parent.Status {
		return repositorySystem.ErrMenuDisabled
	}

	menu, err := svc.repo.Create(ctx, domain)
	if err != nil {
		if svc.IsDuplicateEntryError(err) {
			return repositorySystem.ErrMenuDuplicate
		}
		return err
	}

	// 自动创建菜单默认按钮
	buttons := svc.publicMenuButton(ctx, menu, user)
	for _, button := range buttons {
		json_format.PrintFormattedJSON(button)
	}

	return nil
}

// Delete 删除
func (svc *menuService) Delete(ctx context.Context, id string) error {
	count, err := svc.repo.GetChildCount(ctx, id)
	if err != nil {
		return err
	}
	if count > 0 {
		return repositorySystem.ErrMenuHasChildren
	}
	return svc.repo.Delete(ctx, id)
}

// BatchDelete 批量删除
func (svc *menuService) BatchDelete(ctx context.Context, ids []string) error {
	var failedIds []string

	for _, id := range ids {
		count, _ := svc.repo.GetChildCount(ctx, id)
		if count == 0 {
			failedIds = append(failedIds, id)
			continue
		}
	}

	return svc.repo.BatchDelete(ctx, failedIds)
}

// Update 更新
func (svc *menuService) Update(ctx context.Context, domain domainSystem.Menu) error {
	root := "root"
	// 判空处理
	if domain.ParentID == nil || *domain.ParentID == "" {
		domain.ParentID = &root
	}
	// 检查name、path、title和parentId是否同时存在
	exists, err := svc.repo.CheckExistByNameAndPathAndTitleAndParentId(ctx,
		domain.Name, domain.Path, domain.Title, *domain.ParentID, "")
	if err != nil {
		return err
	}
	if exists {
		return repositorySystem.ErrMenuDuplicate
	}
	// 查询上级菜单
	parent, err := svc.repo.GetByParentId(ctx, *domain.ParentID)
	if err != nil {
		if errors.Is(err, repositorySystem.ErrMenuParentNotFound) {
			return repositorySystem.ErrMenuParentNotFound
		}
		return err
	}
	if parent.Id == "" {
		return repositorySystem.ErrMenuParentNotFound
	}
	if !parent.Status {
		return repositorySystem.ErrMenuDisabled
	}

	if err := svc.repo.Update(ctx, domain); err != nil {
		switch {
		case svc.IsDuplicateEntryError(err):
			return repositorySystem.ErrMenuDuplicate
		case errors.Is(err, repositorySystem.ErrMenuVersionInconsistency):
			return repositorySystem.ErrMenuVersionInconsistency
		default:
			return err
		}
	}

	return nil
}

// GetById 获取详情
func (svc *menuService) GetById(ctx context.Context, id string) (domainSystem.Menu, error) {
	domain, err := svc.repo.GetById(ctx, id)
	if err != nil {
		if errors.Is(err, repositorySystem.ErrMenuNotFound) {
			return domain, repositorySystem.ErrMenuNotFound
		}
		return domain, err
	}
	if domain.Id == "" {
		return domain, repositorySystem.ErrMenuNotFound
	}
	return domain, err
}

// GetMenuRouteTree 获取菜单路由
func (svc *menuService) GetMenuRouteTree(ctx context.Context, filter domainSystem.MenuFilter) ([]MenuRouteTree, error) {
	listAll, err := svc.repo.GetListAll(ctx, filter)
	if err != nil {
		return []MenuRouteTree{}, err
	}

	var listRoute []MenuRouteTree

	for _, l := range listAll {
		route := MenuRouteTree{
			Id:        l.Id,
			Name:      l.Name,
			Path:      l.Path,
			Component: l.Component,
			Meta: map[string]any{
				"title":         l.Title,
				"icon":          l.Icon,
				"showBadge":     l.ShowBadge,
				"showTextBadge": l.ShowTextBadge,
				"isHide":        l.IsHide,
				"isHideTab":     l.IsHideTab,
				"link":          l.Link,
				"isIframe":      l.IsIframe,
				"keepAlive":     l.KeepAlive,
				"isFirstLevel":  l.IsFirstLevel,
				"fixedTab":      l.FixedTab,
				"activePath":    l.ActivePath,
				"isFullPage":    l.IsFullPage,
				"isAuthButton":  l.IsAuthButton,
				"authMark":      l.AuthMark,
			},
			Sort:     l.Sort,
			Status:   l.Status,
			ParentID: "",
			Children: []MenuRouteTree{},
		}
		if *l.ParentID == "root" {
			route.ParentID = ""
		} else {
			route.ParentID = *l.ParentID
		}

		listRoute = append(listRoute, route)
	}

	// 顶级菜单ParentID为root
	return svc.buildMenuRouteTree(listRoute, ""), nil
}

// GetListPage 分页查询列表
func (svc *menuService) GetListPage(ctx context.Context, filters domainSystem.MenuFilter) ([]domainSystem.Menu, int64, error) {
	return svc.repo.GetListPage(ctx, filters)
}

// GetListAll 查询所有列表
func (svc *menuService) GetListAll(ctx context.Context, filter domainSystem.MenuFilter) ([]domainSystem.Menu, error) {
	return svc.repo.GetListAll(ctx, filter)
}

// IsDuplicateEntryError 判断是否是唯一冲突错误
func (svc *menuService) IsDuplicateEntryError(err error) bool {
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) {
		// MySQL 错误码 1062 表示唯一冲突
		return mysqlErr.Number == 1062
	}
	return false
}

// publicMenuButton 生成公共按钮权限
func (svc *menuService) publicMenuButton(ctx context.Context, menu domainSystem.Menu, user domainSystem.User) []domainSystem.MenuButton {
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
				AuthMark: "add",
				Method:   0,
				Api:      "",
				MenuID:   menu.Id,
			},
		},
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
				Title:    "删除",
				AuthMark: "delete",
				Method:   0,
				Api:      "",
				MenuID:   menu.Id,
			},
		},
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
				Title:    "修改",
				AuthMark: "update",
				Method:   0,
				Api:      "",
				MenuID:   menu.Id,
			},
		},
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
				Title:    "查询",
				AuthMark: "query",
				Method:   0,
				Api:      "",
				MenuID:   menu.Id,
			},
		},
	}
}

// buildMenuRouteTree 递归构建菜单路由
func (svc *menuService) buildMenuRouteTree(menus []MenuRouteTree, ParentId string) []MenuRouteTree {
	var tree []MenuRouteTree

	// 筛选当前父ID的子菜单
	for _, menu := range menus {
		if menu.ParentID == ParentId {
			// 递归查询子菜单
			children := svc.buildMenuRouteTree(menus, menu.Id)
			if len(children) == 0 {
				menu.Children = []MenuRouteTree{}
			} else {
				menu.Children = children
			}

			// 按Sort排序
			sort.Slice(children, func(i, j int) bool {
				return children[i].Sort < children[j].Sort
			})

			tree = append(tree, menu)
		}
	}

	// 按Sort排序当前层级
	sort.Slice(tree, func(i, j int) bool {
		return tree[i].Sort < tree[j].Sort
	})

	return tree
}

// ----------------------------------------------

// buildMenuTree 递归构建菜单树
// func (svc *menuService) buildMenuTree(menus []domainSystem.Menu, ParentId string) []domainSystem.Menu {
// var tree []domainSystem.Menu
//
// // 筛选当前父ID的子菜单
// for _, menu := range menus {
// 	if menu.ParentId == ParentId {
// 		// 递归查询子菜单
// 		children := svc.buildMenuTree(menus, menu.Id)
// 		if len(children) == 0 {
// 			menu.Children = []domainSystem.Menu{}
// 		} else {
// 			menu.Children = children
// 		}
//
// 		// 按Sort排序
// 		sort.Slice(children, func(i, j int) bool {
// 			return children[i].Sort < children[j].Sort
// 		})
//
// 		tree = append(tree, menu)
// 	}
// }
//
// // 按Sort排序当前层级
// sort.Slice(tree, func(i, j int) bool {
// 	return tree[i].Sort < tree[j].Sort
// })
//
// return tree
// return nil
// }

// GetMenuRouteTree 获取菜单路由
// func (svc *menuService) GetMenuRouteTree(ctx context.Context, filter domainSystem.MenuFilter) ([]MenuRouteTree, error) {
// 	listAll, err := svc.repo.GetListAll(ctx, filter)
// 	if err != nil {
// 		return []MenuRouteTree{}, err
// 	}
//
// 	var listRoute []MenuRouteTree
//
// 	for _, l := range listAll {
// 		route := MenuRouteTree{
// 			Id:        l.Id,
// 			Name:      l.Name,
// 			Path:      l.Path,
// 			Component: l.Component,
// 			Meta: map[string]any{
// 				"title":         l.Title,
// 				"icon":          l.Icon,
// 				"showBadge":     l.ShowBadge,
// 				"showTextBadge": l.ShowTextBadge,
// 				"isHide":        l.IsHide,
// 				"isHideTab":     l.IsHideTab,
// 				"link":          l.Link,
// 				"isIframe":      l.IsIframe,
// 				"keepAlive":     l.KeepAlive,
// 				"isFirstLevel":  l.IsFirstLevel,
// 				"fixedTab":      l.FixedTab,
// 				"activePath":    l.ActivePath,
// 				"isFullPage":    l.IsFullPage,
// 				"isAuthButton":  l.IsAuthButton,
// 				"authMark":      l.AuthMark,
// 			},
// 			Sort:     l.Sort,
// 			Status:   l.Status,
// 			ParentId: "",
// 			Children: []MenuRouteTree{},
// 		}
// 		if l.ParentId == "root" {
// 			route.ParentId = ""
// 		} else {
// 			route.ParentId = l.ParentId
// 		}
//
// 		listRoute = append(listRoute, route)
// 	}
//
// 	顶级菜单ParentID为root
// 	return svc.buildMenuRouteTree(listRoute, ""), nil
//
// 	return nil, nil
// }

// GetListTree 获取菜单树形结构
// func (svc *menuService) GetListTree(ctx context.Context, filter domainSystem.MenuFilter) ([]domainSystem.Menu, error) {
// 	// listAll, err := svc.repo.GetListAll(ctx, filter)
// 	// if err != nil {
// 	// 	return listAll, err
// 	// }
// 	//
// 	// // 顶级菜单ParentID为root
// 	// return svc.buildMenuTree(listAll, "root"), nil
//
// 	return nil, nil
// }
