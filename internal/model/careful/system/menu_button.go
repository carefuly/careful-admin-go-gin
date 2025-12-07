/**
 * Description：
 * FileName：menu_button.go
 * Author：CJiaの用心
 * Create：2025/12/05 11:28:08
 * Remark：
 */

package system

import (
	"github.com/carefuly/careful-admin-go-gin/pkg/constants/careful/system/menu"
	"github.com/carefuly/careful-admin-go-gin/pkg/models"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// MenuButton 菜单按钮表
type MenuButton struct {
	models.CoreModels

	Status   bool        `gorm:"type:boolean;index;column:status;comment:状态【true-启用 false-停用】" json:"status"` // 状态
	Title    string      `gorm:"size:64;not null;index;column:title;comment:按钮名称" json:"title"`               // 按钮名称
	AuthMark string      `gorm:"size:64;not null;index;column:auth_mark;comment:按钮权限值" json:"authMark"`       // 按钮权限值
	Method   menu.Method `gorm:"type:tinyint;default(1);index;column:method;comment:方法类型" json:"method"`      // 方法类型
	Api      string      `gorm:"size:255;not null;index;column:api;comment:接口地址" json:"api"`                  // 接口地址
	// 所属菜单
	MenuID string `gorm:"size:110;index;column:menu_id;comment:关联菜单" json:"menu_id"`                   // 关联菜单
	Menu   *Menu  `gorm:"foreignKey:MenuID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"menu"` // 菜单
}

func NewMenuButton() *MenuButton {
	return &MenuButton{}
}

func (m *MenuButton) TableName() string {
	return "careful_system_menu_button"
}

func (m *MenuButton) AutoMigrate(db *gorm.DB) {
	err := db.Set("gorm:foreignKeyConstraint", true).
		Set("gorm:table_options", "ENGINE=InnoDB,COMMENT='菜单按钮表'").
		AutoMigrate(&MenuButton{})
	if err != nil {
		zap.L().Error("MenuButton表模型迁移失败", zap.Error(err))
	}
}
