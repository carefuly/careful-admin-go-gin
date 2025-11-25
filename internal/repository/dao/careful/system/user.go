/**
 * Description：
 * FileName：user.go
 * Author：CJiaの用心
 * Create：2025/11/24 17:09:15
 * Remark：
 */

package system

import (
	"context"
	"errors"
	"github.com/carefuly/careful-admin-go-gin/internal/model/careful/system"
	"gorm.io/gorm"
	"time"
)

var (
	ErrUserNotFound             = gorm.ErrRecordNotFound
	ErrUserVersionInconsistency = errors.New("数据已被修改，请刷新后重试")
)

type UserDAO interface {
	UpdateLoginField(ctx context.Context, lastLogin string, lastLoginIp string, model system.User) error

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

// UpdateLoginField 更新登录字段
func (dao *GORMUserDAO) UpdateLoginField(ctx context.Context, lastLogin string, lastLoginIp string, model system.User) error {
	result := dao.db.WithContext(ctx).Model(&model).
		Where("id = ? AND timestamp = ?", model.Id, model.Timestamp).
		Updates(map[string]any{
			"last_login":    lastLogin,
			"last_login_ip": lastLoginIp,
			"timestamp":     time.Now().UnixMicro(),
		})
	// 处理行影响数为0的情况
	if result.RowsAffected == 0 {
		// 先检查记录是否存在
		var exists bool
		dao.db.WithContext(ctx).
			Model(&system.User{}).
			Select("1").
			Where("id = ?", model.Id).
			Limit(1).
			Find(&exists)

		if !exists {
			return ErrUserNotFound
		}
		return ErrUserVersionInconsistency
	}
	return result.Error
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
