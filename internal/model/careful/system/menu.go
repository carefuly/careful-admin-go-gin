/**
 * Description：
 * FileName：menu.go
 * Author：CJiaの用心
 * Create：2025/12/05 11:34:54
 * Remark：
 */

package system

import (
	"github.com/carefuly/careful-admin-go-gin/pkg/models"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Menu 菜单表
type Menu struct {
	models.CoreModels

	Status        bool   `gorm:"type:boolean;index;column:status;comment:状态【true-启用 false-停用】" json:"status"`                    // 状态
	Name          string `gorm:"size:64;not null;uniqueIndex:uni_name_path_title_parent;column:name;comment:菜单名称" json:"name"`   // 菜单名称
	Path          string `gorm:"size:128;not null;uniqueIndex:uni_name_path_title_parent;column:path;comment:路由地址" json:"path"`  // 路由地址
	Component     string `gorm:"size:128;not null;column:component;comment:组件地址" json:"component"`                               // 组件地址
	Title         string `gorm:"size:64;not null;uniqueIndex:uni_name_path_title_parent;column:title;comment:路由标题" json:"title"` // 路由标题
	Icon          string `gorm:"size:64;column:icon;comment:路由图标" json:"icon"`                                                   // 路由图标
	ShowBadge     bool   `gorm:"type:boolean;default:false;column:show_badge;comment:是否显示徽章" json:"showBadge"`                   // 是否显示徽章
	ShowTextBadge string `gorm:"size:64;column:show_text_badge;comment:文本徽章" json:"showTextBadge"`                               // 文本徽章
	IsHide        bool   `gorm:"type:boolean;default:false;column:is_hide;comment:是否在菜单中隐藏" json:"isHide"`                       // 是否在菜单中隐藏
	IsHideTab     bool   `gorm:"type:boolean;default:false;column:is_hide_tab;comment:是否在标签页中隐藏" json:"isHideTab"`               // 是否在标签页中隐藏
	Link          string `gorm:"size:255;column:link;comment:外部链接【不填写默认没有外链】" json:"link"`                                       // 外部链接【不填写默认没有外链】
	IsIframe      bool   `gorm:"type:boolean;default:false;column:is_iframe;comment:是否为iframe" json:"isIframe"`                  // 是否为iframe
	KeepAlive     bool   `gorm:"type:boolean;default:false;column:keep_alive;comment:是否缓存页面" json:"keepAlive"`                   // 是否缓存页面
	IsFirstLevel  bool   `gorm:"type:boolean;default:false;column:is_first_level;comment:是否为一级菜单" json:"isFirstLevel"`           // 是否为一级菜单
	FixedTab      bool   `gorm:"type:boolean;default:false;column:fixed_tab;comment:是否固定标签页" json:"fixedTab"`                    // 是否固定标签页
	ActivePath    string `gorm:"size:128;column:active_path;comment:激活菜单路径" json:"activePath"`                                   // 激活菜单路径
	IsFullPage    bool   `gorm:"type:boolean;default:false;column:is_full_page;comment:是否为全屏页面" json:"isFullPage"`               // 是否为全屏页面
	IsAuthButton  bool   `gorm:"type:boolean;default:false;column:is_auth_button;comment:是否为权限按钮行" json:"isAuthButton"`          // 是否为权限按钮行
	AuthMark      string `gorm:"size:128;column:auth_mark;comment:权限标识" json:"authMark"`                                         // 权限标识
	// 上级菜单
	ParentID *string `gorm:"size:110;uniqueIndex:uni_name_path_title_parent;column:parent_id;comment:上级菜单ID" json:"parent_id"` // 上级菜单ID
	Parent   *Menu   `gorm:"foreignKey:ParentID" json:"parent"`                                                                // 上级菜单
	Children []*Menu `gorm:"foreignKey:ParentID" json:"children,omitempty"`                                                    // 子菜单列表
}

func NewMenu() *Menu {
	return &Menu{}
}

func (m *Menu) TableName() string {
	return "careful_system_menu"
}

func (m *Menu) AutoMigrate(db *gorm.DB) {
	err := db.Set("gorm:foreignKeyConstraint", true).
		Set("gorm:table_options", "ENGINE=InnoDB,COMMENT='菜单表'").
		AutoMigrate(&Menu{})
	if err != nil {
		zap.L().Error("Menu表模型迁移失败", zap.Error(err))
	}
}
