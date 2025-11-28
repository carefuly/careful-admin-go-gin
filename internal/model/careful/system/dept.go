/**
 * Description：
 * FileName：dept.go
 * Author：CJiaの用心
 * Create：2025/11/25 02:05:29
 * Remark：
 */

package system

import (
	"github.com/carefuly/careful-admin-go-gin/pkg/constants/careful/system/dept"
	"github.com/carefuly/careful-admin-go-gin/pkg/models"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Dept 部门表
type Dept struct {
	models.CoreModels

	Status      bool      `gorm:"type:boolean;index;column:status;comment:状态【true-启用 false-停用】" json:"status"`              // 状态
	Name        string    `gorm:"size:50;not null;uniqueIndex:uni_name_parent;column:name;comment:部门名称" json:"name"`        // 部门名称
	Code        string    `gorm:"size:50;not null;uniqueIndex;column:code;comment:部门编码" json:"code"`                        // 部门编码
	DeptType    dept.Type `gorm:"size:20;not null;index;default:department;column:dept_type;comment:部门类型" json:"dept_type"` // 部门类型
	Owner       string    `gorm:"size:64;column:owner;comment:部门负责人" json:"owner"`                                          // 部门负责人
	Phone       string    `gorm:"size:64;column:phone;comment:部门电话" json:"phone"`                                           // 部门电话
	Email       string    `gorm:"size:64;column:email;comment:部门邮箱" json:"email"`                                           // 部门邮箱
	Description string    `gorm:"type:text;column:description;comment:部门描述" json:"description"`                             // 部门描述
	// 上级部门
	ParentID *string `gorm:"size:110;uniqueIndex:uni_name_parent;column:parent_id;comment:父部门ID" json:"parent_id"` // 父部门ID
	Parent   *Dept   `gorm:"foreignKey:ParentID" json:"parent"`                                                    // 父部门
	// 关联查询字段
	Level int    `gorm:"type:int;not null;index;default:0;column:level;comment:层级深度" json:"level"` // 层级深度，根节点为0
	Path  string `gorm:"size:512;index;column:path;comment:部门路径，格式：/1/2/3/" json:"path"`           // 部门路径，格式：/1/2/3/

	// UserCount  int     `gorm:"type:int;default:0;column:user_count;comment:用户数量" json:"user_count"`      // 用户数量
	// ChildCount int     `gorm:"type:int;default:0;column:child_count;comment:子部门数量" json:"child_count"`   // 子部门数量
	// Children   []*Dept `gorm:"foreignKey:ParentID" json:"children"`                                      // 子部门列表
}

func NewDept() *Dept {
	return &Dept{}
}

func (d *Dept) TableName() string {
	return "careful_system_dept"
}

func (d *Dept) AutoMigrate(db *gorm.DB) {
	err := db.Set("gorm:foreignKeyConstraint", true).
		Set("gorm:table_options", "ENGINE=InnoDB,COMMENT='部门表'").
		AutoMigrate(&Dept{})
	if err != nil {
		zap.L().Error("Dept表模型迁移失败", zap.Error(err))
	}
}
