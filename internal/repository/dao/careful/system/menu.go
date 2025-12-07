/**
 * Description：
 * FileName：menu.go
 * Author：CJiaの用心
 * Create：2025/12/05 12:03:03
 * Remark：
 */

package system

import (
	"context"
	"errors"
	domainSystem "github.com/carefuly/careful-admin-go-gin/internal/domain/careful/system"
	"github.com/carefuly/careful-admin-go-gin/internal/model/careful/system"
	"github.com/carefuly/careful-admin-go-gin/pkg/ginx/filters"
	"gorm.io/gorm"
	"time"
)

var (
	ErrMenuNotFound             = gorm.ErrRecordNotFound
	ErrMenuDuplicate            = errors.New("同级别下已存在相同的菜单信息")
	ErrMenuParentNotFound       = gorm.ErrRecordNotFound
	ErrMenuDisabled             = errors.New("上级菜单已被禁用，无法在其下创建子菜单")
	ErrMenuHasChildren          = errors.New("菜单含有子菜单，无法删除")
	ErrMenuVersionInconsistency = errors.New("数据已被修改，请刷新后重试")
)

type MenuDAO interface {
	Insert(ctx context.Context, model system.Menu) (*system.Menu, error)
	Delete(ctx context.Context, id string) error
	BatchDelete(ctx context.Context, ids []string) error
	Update(ctx context.Context, model system.Menu) error

	FindById(ctx context.Context, id string) (*system.Menu, error)
	FindByParentId(ctx context.Context, parentId string) (*system.Menu, error)
	FindChildCount(ctx context.Context, id string) (int64, error)
	FindListPage(ctx context.Context, filter domainSystem.MenuFilter) ([]*system.Menu, int64, error)
	FindListAll(ctx context.Context, filter domainSystem.MenuFilter) ([]*system.Menu, error)

	CheckExistByNameAndPathAndTitleAndParentId(ctx context.Context, name, path, title, parentId, excludeId string) (bool, error)
}

type GORMMenuDAO struct {
	db *gorm.DB
}

func NewGORMMenuDAO(db *gorm.DB) MenuDAO {
	return &GORMMenuDAO{
		db: db,
	}
}

// Insert 新增
func (dao *GORMMenuDAO) Insert(ctx context.Context, model system.Menu) (*system.Menu, error) {
	return &model, dao.db.WithContext(ctx).Create(&model).Error
}

// Delete 删除
func (dao *GORMMenuDAO) Delete(ctx context.Context, id string) error {
	return dao.db.WithContext(ctx).Where("id = ?", id).Delete(&system.Menu{}).Error
}

// BatchDelete 批量删除
func (dao *GORMMenuDAO) BatchDelete(ctx context.Context, ids []string) error {
	return dao.db.WithContext(ctx).Where("id IN ?", ids).Delete(&system.Menu{}).Error
}

// Update 更新
func (dao *GORMMenuDAO) Update(ctx context.Context, model system.Menu) error {
	result := dao.db.WithContext(ctx).Model(&model).
		Where("id = ? AND timestamp = ?", model.Id, model.Timestamp).
		Updates(map[string]any{
			"status":          model.Status,
			"title":           model.Title,
			"icon":            model.Icon,
			"show_badge":      model.ShowBadge,
			"show_text_badge": model.ShowTextBadge,
			"is_hide":         model.IsHide,
			"is_hide_tab":     model.IsHideTab,
			"link":            model.Link,
			"is_iframe":       model.IsIframe,
			"keep_alive":      model.KeepAlive,
			"is_first_level":  model.IsFirstLevel,
			"fixed_tab":       model.FixedTab,
			"active_path":     model.ActivePath,
			"is_full_page":    model.IsFullPage,
			"is_auth_button":  model.IsAuthButton,
			"auth_mark":       model.AuthMark,
			"parent_id":       model.ParentID,
			"sort":            model.Sort,
			"timestamp":       time.Now().UnixMicro(),
			"modifier":        model.Modifier,
			"remark":          model.Remark,
		})
	// 处理行影响数为0的情况
	if result.RowsAffected == 0 {
		// 先检查记录是否存在
		var exists bool
		dao.db.WithContext(ctx).
			Model(&system.Menu{}).
			Select("1").
			Where("id = ?", model.Id).
			Limit(1).
			Find(&exists)

		if !exists {
			return ErrMenuNotFound
		}
		return ErrMenuVersionInconsistency
	}
	return result.Error
}

// FindById 根据id获取详情
func (dao *GORMMenuDAO) FindById(ctx context.Context, id string) (*system.Menu, error) {
	var model system.Menu
	err := dao.db.WithContext(ctx).Where("id = ?", id).First(&model).Error
	return &model, err
}

// FindByParentId 根据parent_id获取详情
func (dao *GORMMenuDAO) FindByParentId(ctx context.Context, parentId string) (*system.Menu, error) {
	var model system.Menu
	err := dao.db.WithContext(ctx).Where("id = ?", parentId).First(&model).Error
	return &model, err
}

// FindChildCount 获取子菜单数量
func (dao *GORMMenuDAO) FindChildCount(ctx context.Context, id string) (int64, error) {
	var count int64
	err := dao.db.WithContext(ctx).
		Model(&system.Menu{}).
		Where("parent_id = ?", id).
		Count(&count).
		Error
	return count, err
}

// FindListPage 分页查询
func (dao *GORMMenuDAO) FindListPage(ctx context.Context, filter domainSystem.MenuFilter) ([]*system.Menu, int64, error) {
	var total int64
	var models []*system.Menu

	query := dao.buildQuery(ctx, filter)

	err := query.Count(&total).
		Offset((filter.Page - 1) * filter.PageSize).
		Limit(filter.PageSize).
		Find(&models).Error

	return models, total, err
}

// FindListAll 获取所有列表
func (dao *GORMMenuDAO) FindListAll(ctx context.Context, filter domainSystem.MenuFilter) ([]*system.Menu, error) {
	var models []*system.Menu

	query := dao.buildQuery(ctx, filter)

	// 查询
	if err := query.Find(&models).Error; err != nil {
		return nil, err
	}

	return models, nil
}

// buildQuery 构建查询条件
func (dao *GORMMenuDAO) buildQuery(ctx context.Context, filter domainSystem.MenuFilter) *gorm.DB {
	builder := &domainSystem.MenuFilter{
		Filters: filters.Filters{
			Creator:    filter.Creator,
			Modifier:   filter.Modifier,
			BelongDept: filter.BelongDept,
		},
		Status: filter.Status,
		Title:  filter.Title,
	}
	return builder.QueryFilter(ctx, dao.db.WithContext(ctx).Model(&system.Menu{}))
}

// CheckExistByNameAndPathAndTitleAndParentId 检查name、path、title和parentId是否同时存在
func (dao *GORMMenuDAO) CheckExistByNameAndPathAndTitleAndParentId(ctx context.Context, name, path, title, parentId, excludeId string) (bool, error) {
	var model system.Menu
	query := dao.db.WithContext(ctx).Model(&system.Menu{}).
		Select("id"). // 只查询必要的字段
		Where("name = ? AND path = ? AND title = ? AND parent_id = ?", name, path, title, parentId)

	if excludeId != "" {
		query = query.Where("id != ?", excludeId)
	}

	// 使用 LIMIT 1 快速判断是否存在
	err := query.Limit(1).First(&model).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil // 不存在
	}
	return err == nil, err // 存在或查询出错
}
