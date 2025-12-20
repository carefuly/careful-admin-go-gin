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
	repositorySystem "github.com/carefuly/careful-admin-go-gin/internal/repository/repository/careful/system"
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

// MenuTree 菜单树形结构
type MenuTree struct {
	domainSystem.Menu            // 上级菜单
	Children          []MenuTree `json:"children"` // 子菜单
}

type MenuService interface {
	Create(ctx context.Context, domain domainSystem.Menu, menuButton []string, user domainSystem.User) error
	Delete(ctx context.Context, id string) error
	BatchDelete(ctx context.Context, ids []string) error
	Update(ctx context.Context, domain domainSystem.Menu) error

	GetById(ctx context.Context, id string) (domainSystem.Menu, error)
	GetMenuRouteTree(ctx context.Context, filter domainSystem.MenuFilter) ([]MenuRouteTree, error)
	GetListTree(ctx context.Context, filters domainSystem.MenuFilter) ([]MenuTree, error)
	GetListPage(ctx context.Context, filters domainSystem.MenuFilter) ([]domainSystem.Menu, int64, error)
	GetListAll(ctx context.Context, filter domainSystem.MenuFilter) ([]domainSystem.Menu, error)
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
func (svc *menuService) Create(ctx context.Context, domain domainSystem.Menu, menuButton []string, user domainSystem.User) error {
	root := "root"
	// 判空处理
	if domain.ParentID == nil || *domain.ParentID == "" {
		domain.ParentID = &root
		domain.IsFirstLevel = true
	} else {
		domain.IsFirstLevel = false
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

	_, err = svc.repo.Create(ctx, domain)
	if err != nil {
		if svc.IsDuplicateEntryError(err) {
			return repositorySystem.ErrMenuDuplicate
		}
		return err
	}

	// 自动创建菜单默认按钮
	// buttons := svc.publicMenuButton(ctx, menu, user)
	// for _, button := range buttons {
	// 	json_format.PrintFormattedJSON(button)
	// }

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
		domain.IsFirstLevel = true
	} else {
		domain.IsFirstLevel = false
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

// GetMenuRouteTree 查询菜单路由
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

// GetListTree 查询菜单树
func (svc *menuService) GetListTree(ctx context.Context, filters domainSystem.MenuFilter) ([]MenuTree, error) {
	list, err := svc.repo.GetListAll(ctx, filters)
	if err != nil {
		return nil, err
	}

	// 构建菜单树
	return svc.buildMenuTree(list), nil
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

// buildMenuRouteTree 递归构建菜单路由
func (svc *menuService) buildMenuRouteTree(menus []MenuRouteTree, ParentId string) []MenuRouteTree {
	// 创建映射表，便于快速查找
	menuMap := make(map[string][]MenuRouteTree)
	allMenus := make(map[string]MenuRouteTree)

	// 首先将所有菜单放入allMenus映射表
	for _, menu := range menus {
		allMenus[menu.Id] = menu
	}

	// 按照ParentID分组
	for _, menu := range menus {
		parentID := menu.ParentID
		menuMap[parentID] = append(menuMap[parentID], menu)
	}

	// 获取顶级菜单（ParentID为空的菜单）
	topLevelMenus := menuMap[""]

	// 对顶级菜单按Sort排序
	sort.Slice(topLevelMenus, func(i, j int) bool {
		return topLevelMenus[i].Sort < topLevelMenus[j].Sort
	})

	// 递归构建子菜单树
	for i := range topLevelMenus {
		topLevelMenus[i] = svc.buildChildrenForMenuRoute(topLevelMenus[i], menuMap)
	}

	return topLevelMenus
}

// buildChildrenForMenuRoute 为单个菜单构建子菜单
func (svc *menuService) buildChildrenForMenuRoute(menu MenuRouteTree, menuMap map[string][]MenuRouteTree) MenuRouteTree {
	// 查找当前菜单的子菜单
	children := menuMap[menu.Id]

	// 对子菜单按Sort排序
	sort.Slice(children, func(i, j int) bool {
		return children[i].Sort < children[j].Sort
	})

	// 递归构建子菜单的子菜单
	for i := range children {
		children[i] = svc.buildChildrenForMenuRoute(children[i], menuMap)
	}

	// 设置当前菜单的子菜单
	menu.Children = children

	return menu
}

// buildMenuTree 构建菜单树
func (svc *menuService) buildMenuTree(menus []domainSystem.Menu) []MenuTree {
	// 创建map用于快速查找菜单
	menuMap := make(map[string]domainSystem.Menu)
	childrenMap := make(map[string][]MenuTree)

	// 首先填充menuMap和初始化childrenMap
	for _, menu := range menus {
		menuMap[menu.Id] = menu
		// 初始化当前菜单的子菜单列表
		if _, exists := childrenMap[menu.Id]; !exists {
			childrenMap[menu.Id] = []MenuTree{}
		}
	}

	// 确定根节点ID
	rootID := "root"

	// 遍历所有菜单，建立父子关系
	for _, menu := range menus {
		var parentID string
		if menu.ParentID != nil && *menu.ParentID != "" {
			parentID = *menu.ParentID
		} else {
			// 如果ParentID为nil或为空字符串，则认为是一级菜单
			parentID = rootID
		}

		// 创建当前菜单的树节点
		node := MenuTree{
			Menu:     menu,
			Children: []MenuTree{}, // 初始为空，后面会填充
		}

		// 添加到父菜单的子菜单列表
		childrenMap[parentID] = append(childrenMap[parentID], node)
	}

	// 获取根节点的所有子菜单
	rootNodes := childrenMap[rootID]

	// 对根节点按sort排序
	sort.Slice(rootNodes, func(i, j int) bool {
		return rootNodes[i].Menu.Sort < rootNodes[j].Menu.Sort
	})

	// 递归处理所有节点，填充子菜单并排序
	for i := range rootNodes {
		svc.populateAndSortChildren(&rootNodes[i], childrenMap)
	}

	return rootNodes
}

// populateAndSortChildren 填充子菜单并递归排序
func (svc *menuService) populateAndSortChildren(node *MenuTree, childrenMap map[string][]MenuTree) {
	// 获取当前节点的子菜单
	if children, exists := childrenMap[node.Menu.Id]; exists {
		node.Children = children
	}

	// 对当前节点的子菜单排序
	sort.Slice(node.Children, func(i, j int) bool {
		return node.Children[i].Menu.Sort < node.Children[j].Menu.Sort
	})

	// 递归处理所有子节点
	for i := range node.Children {
		svc.populateAndSortChildren(&node.Children[i], childrenMap)
	}
}
