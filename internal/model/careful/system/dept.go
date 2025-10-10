/**
 * Description：
 * FileName：dept.go
 * Author：CJiaの用心
 * Create：2025/10/9 16:03:07
 * Remark：
 */

package system

import (
	"database/sql"
	"fmt"
	"github.com/carefuly/careful-admin-go-gin/pkg/models"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Dept 部门表
type Dept struct {
	models.CoreModels

	Status     bool           `gorm:"type:boolean;index:idx_status;default:true;column:status;comment:状态【true-启用 false-停用】" json:"status"`                           // 状态
	Name       string         `gorm:"type:varchar(50);not null;uniqueIndex:uni_dept_name_code_parent;column:name;comment:部门名称" json:"name"`                          // 部门名称
	Code       string         `gorm:"type:varchar(50);not null;uniqueIndex:uni_dept_name_code_parent;column:code;comment:部门编码" json:"code"`                          // 部门编码
	Owner      string         `gorm:"type:varchar(32);column:owner;comment:负责人" json:"owner"`                                                                        // 负责人
	Phone      string         `gorm:"type:varchar(32);column:phone;comment:联系电话" json:"phone"`                                                                       // 联系电话
	Email      string         `gorm:"type:varchar(32);column:email;comment:邮箱" json:"email"`                                                                         // 邮箱
	Level      int            `gorm:"type:int;index:idx_level;default:0;column:level;comment:层级深度，根节点为0" json:"level"`                                               // 层级深度，根节点为0
	Path       string         `gorm:"type:varchar(512);index:idx_path;column:path;comment:节点路径，格式：/1/2/3/" json:"path"`                                              // 节点路径，格式：/1/2/3/"
	UserCount  int            `gorm:"type:int;default:0;column:user_count;comment:用户数量" json:"user_count"`                                                           // 用户数量
	ChildCount int            `gorm:"type:int;default:0;column:child_count;comment:子部门数量" json:"child_count"`                                                        // 子部门数量
	ParentID   sql.NullString `gorm:"type:varchar(100);uniqueIndex:uni_dept_name_code_parent;column:parent_id;comment:上级部门ID" swaggertype:"string" json:"parent_id"` // 上级部门ID
	// 关联查询字段（不存储到数据库）
	Children []*Dept `gorm:"-" json:"children,omitempty"` // 子部门列表
	Parent   *Dept   `gorm:"-" json:"parent,omitempty"`   // 父部门信息
}

func NewDept() *Dept {
	return &Dept{}
}

func (d *Dept) TableName() string {
	return "careful_system_dept"
}

func (d *Dept) AutoMigrate(db *gorm.DB) {
	err := db.Set("gorm:table_options", "ENGINE=InnoDB,COMMENT='部门表'").AutoMigrate(&Dept{})
	if err != nil {
		zap.L().Error("Dept表模型迁移失败", zap.Error(err))
	}
}

func (d *Dept) BeforeCreate(tx *gorm.DB) error {
	return d.calculateTreeFields(tx)
}

func (d *Dept) BeforeUpdate(tx *gorm.DB) error {
	return d.calculateTreeFields(tx)
}

// calculateTreeFields 如果ParentID发生变化，需要重新计算树形字段
func (d *Dept) calculateTreeFields(tx *gorm.DB) error {
	if d.ParentID.Valid {
		// 子节点
		var parent Dept
		if err := tx.Where("id = ?", d.ParentID.String).First(&parent).Error; err != nil {
			return err
		}
		d.Level = parent.Level + 1
		d.Path = parent.Path + fmt.Sprintf("%s/", d.Id)
	} else {
		// 根节点
		d.Level = 0
		d.Path = fmt.Sprintf("/%s/", d.Id)
	}

	return nil
}

/**
高效查询方法
// 查询部门树（某个部门的所有子部门）
func GetDeptTree(db *gorm.DB, rootID int64) ([]*Dept, error) {
    var root Dept
    if err := db.Where("id = ?", rootID).First(&root).Error; err != nil {
        return nil, err
    }

    var d []*Dept
    err := db.Where("path LIKE ? AND status = ?", root.Path+"%", true).
        Order("level, sort").
        Find(&d).Error

    return buildTree(d), err
}

// 查询某层级的所有部门
func GetDByLevel(db *gorm.DB, level int) ([]*Dept, error) {
    var d []*Dept
    return d, db.Where("level = ? AND status = ?", level, true).
        Order("sort").
        Find(&d).Error
}

func (d *Dept) GetChildren(db *gorm.DB, includeUsers bool) ([]*Dept, error) {
	var children []*Dept
	query := db.Where("parent_id = ? AND status = ?", d.Id, true).Order("sort")

	if includeUsers {
		query = query.Preload("Users", "status = ?", true)
	}

	return children, query.Find(&children).Error
}

func (d *Dept) GetDescendants(db *gorm.DB, includeUsers bool) ([]*Dept, error) {
	var descendants []*Dept
	query := db.Where("path LIKE ? AND id != ? AND status = ?", d.Path+"%", d.Id, true).Order("level, sort")

	if includeUsers {
		query = query.Preload("Users", "status = ?", true)
	}

	return descendants, query.Find(&descendants).Error
}

func (d *Dept) GetAncestors(db *gorm.DB) ([]*Dept, error) {
	if d.ParentID == nil {
		return []*Dept{}, nil
	}

	var ancestors []*Dept
	// 通过path字段快速获取所有祖先节点
	pathParts := strings.Split(strings.Trim(d.Path, "/"), "/")
	if len(pathParts) <= 1 {
		return ancestors, nil
	}

	// 排除自己，只获取祖先
	ancestorIds := pathParts[:len(pathParts)-1]
	return ancestors, db.Where("id IN ?", ancestorIds).Order("level").Find(&ancestors).Error
}
*/
