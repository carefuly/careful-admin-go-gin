/**
 * Description：
 * FileName：user.go
 * Author：CJiaの用心
 * Create：2025/10/9 16:42:27
 * Remark：
 */

package system

import (
	"context"
	"github.com/carefuly/careful-admin-go-gin/internal/model/careful/system"
	"gorm.io/gorm"
)

var (
	ErrUserNotFound = gorm.ErrRecordNotFound
)

type UserDAO interface {
	Insert(ctx context.Context, model system.User) (*system.User, error)

	FindById(ctx context.Context, id string) (*system.User, error)
	FindByUsername(ctx context.Context, username string) (*system.User, error)
}

type GORMUserDAO struct {
	db *gorm.DB
}

func NewGORMUserDAO(db *gorm.DB) UserDAO {
	return &GORMUserDAO{
		db: db,
	}
}

// Insert 新增
func (dao *GORMUserDAO) Insert(ctx context.Context, model system.User) (*system.User, error) {
	return &model, dao.db.WithContext(ctx).Create(&model).Error
}

// FindById 根据id获取详情
func (dao *GORMUserDAO) FindById(ctx context.Context, id string) (*system.User, error) {
	var model system.User
	err := dao.db.WithContext(ctx).
		Preload("Dept").
		Where("id = ?", id).
		First(&model).Error
	return &model, err
}

// FindByUsername 根据用户名获取详情
func (dao *GORMUserDAO) FindByUsername(ctx context.Context, username string) (*system.User, error) {
	var model system.User
	err := dao.db.WithContext(ctx).
		Preload("Dept").
		Where("username = ?", username).
		First(&model).Error
	return &model, err
}
