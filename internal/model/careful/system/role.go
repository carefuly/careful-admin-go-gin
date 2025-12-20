/**
 * Description：
 * FileName：role.go
 * Author：CJiaの用心
 * Create：2025/12/19 21:20:01
 * Remark：
 */

package system

import (
	"github.com/carefuly/careful-admin-go-gin/pkg/constants/careful/system/role"
	"github.com/carefuly/careful-admin-go-gin/pkg/models"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Role 角色表
type Role struct {
	models.CoreModels

	Status      bool           `gorm:"t1ype:boolean;index;column:status;comment:状态【true-启用 false-停用】" json:"status"`    // 状态
	Name        string         `gorm:"size:64;not null;index;column:name;comment:角色名称" json:"name"`                     // 角色名称
	Code        string         `gorm:"size:64;not null;uniqueIndex;column:code;comment:角色编码" json:"code"`               // 角色编码
	DataScope   role.DataScope `gorm:"type:tinyint;index;default:1;column:data_scope;comment:数据权限范围" json:"data_scope"` // 数据权限范围
	Description string         `gorm:"type:text;column:description;comment:角色描述" json:"description"`                    // 角色描述

	// 关联部门 Role -> Dept
	// Dept []*Dept `gorm:"many2many:careful_system_role_dept;constraint:OnDelete:CASCADE;"` // 关联部门
	// 关联菜单 Role -> Menu
	// Menu []*Menu `gorm:"many2many:careful_system_role_menu;constraint:OnDelete:CASCADE;"` // 关联菜单
	// 关联菜单按钮 Role -> MenuButton
	// MenuButton []*MenuButton `gorm:"many2many:careful_system_role_menu_button;constraint:OnDelete:CASCADE;"` // 关联菜单按钮
}

func NewRole() *Role {
	return &Role{}
}

func (d *Role) TableName() string {
	return "careful_system_role"
}

func (d *Role) AutoMigrate(db *gorm.DB) {
	err := db.Set("gorm:foreignKeyConstraint", true).
		Set("gorm:table_options", "ENGINE=InnoDB,COMMENT='角色表'").
		AutoMigrate(&Role{})
	if err != nil {
		zap.L().Error("Role表模型迁移失败", zap.Error(err))
	}
}
