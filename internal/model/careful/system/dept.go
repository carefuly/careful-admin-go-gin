/**
 * Description：
 * FileName：dept.go
 * Author：CJiaの用心
 * Create：2025/11/25 02:05:29
 * Remark：
 */

package system

import (
	"errors"
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

// GetUserCount 获取部门下的用户数量
func (d *Dept) GetUserCount(tx *gorm.DB) (int64, error) {
	var userCount int64
	err := tx.Model(&User{}).Where("dept_id = ?", d.Id).Count(&userCount).Error
	return userCount, err
}

// GetChildCount 获取直接子部门数量
func (d *Dept) GetChildCount(tx *gorm.DB) (int64, error) {
	var count int64
	err := tx.Model(&Dept{}).Where("parent_id = ?", d.Id).Count(&count).Error
	return count, err
}

// IsLeaf 判断是否为叶子节点（没有子部门）
func (d *Dept) IsLeaf(tx *gorm.DB) (bool, error) {
	count, err := d.GetChildCount(tx)
	return count == 0, err
}

// CanDelete 判断是否可以删除（没有子部门和用户）
func (d *Dept) CanDelete(tx *gorm.DB) (bool, error) {
	isLeaf, err := d.IsLeaf(tx)
	if err != nil || !isLeaf {
		return false, err
	}

	userCount, err := d.GetUserCount(tx)
	if err != nil || userCount > 0 {
		return false, err
	}

	return true, nil
}

// GetAncestors 获取所有祖先部门
func (d *Dept) GetAncestors(tx *gorm.DB) ([]Dept, error) {
	var ancestors []Dept
	currentID := d.ParentID
	for currentID != nil {
		var parent Dept
		if err := tx.First(&parent, currentID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				break // 防止因数据不一致导致死循环
			}
			return nil, err
		}
		ancestors = append([]Dept{parent}, ancestors...) //  prepend
		currentID = parent.ParentID
	}
	return ancestors, nil
}

// // GetDeptTypeDisplayName 获取部门类型的显示名称
// func (d *Dept) GetDeptTypeDisplayName() string {
// 	switch d.DeptType {
// 	case dept.TypeCompany:
// 		return "公司"
// 	case dept.TypeDepartment:
// 		return "部门"
// 	case dept.TypeTeam:
// 		return "小组"
// 	case dept.TypeOther:
// 		return "其他"
// 	default:
// 		return "未知"
// 	}
// }
//
// // GetFullName 获取部门全名（包含父部门）
// // 这个方法会触发数据库查询，建议在 Service 层通过 Preload 加载完整路径后再调用
// func (d *Dept) GetFullName(tx *gorm.DB) (string, error) {
// 	if d.ParentID == nil {
// 		return d.Name, nil
// 	}
// 	var parent Dept
// 	if err := tx.First(&parent, d.ParentID).Error; err != nil {
// 		return "", err
// 	}
// 	parentFullName, err := parent.GetFullName(tx)
// 	if err != nil {
// 		return "", err
// 	}
// 	return fmt.Sprintf("%s / %s", parentFullName, d.Name), nil
// }
//
// // GetDescendants 获取所有后代部门 (递归)
// // 注意：这是一个简单实现，对于层级很深的树可能效率不高。可以考虑使用 GORM 的 Preload("Children.*") 或数据库层面的递归查询。
// func (d *Dept) GetDescendants(tx *gorm.DB) ([]Dept, error) {
// 	var descendants []Dept
// 	var children []Dept
// 	if err := tx.Where("parent_id = ?", d.Id).Find(&children).Error; err != nil {
// 		return nil, err
// 	}
// 	descendants = append(descendants, children...)
// 	for _, child := range children {
// 		grandChildren, err := child.GetDescendants(tx)
// 		if err != nil {
// 			return nil, err
// 		}
// 		descendants = append(descendants, grandChildren...)
// 	}
// 	return descendants, nil
// }
//

//
// // IsRoot 判断是否为根节点（没有父部门）
// func (d *Dept) IsRoot() bool {
// 	return d.ParentID == nil
// }
