/**
 * Description：
 * FileName：post.go
 * Author：CJiaの用心
 * Create：2025/11/29 01:06:54
 * Remark：
 */

package system

import (
	"github.com/carefuly/careful-admin-go-gin/pkg/constants/careful/system/post"
	"github.com/carefuly/careful-admin-go-gin/pkg/models"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Post 岗位表
type Post struct {
	models.CoreModels

	Status      bool       `gorm:"type:boolean;index;column:status;comment:状态【true-启用 false-停用】" json:"status"` // 状态
	Name        string     `gorm:"size:50;not null;index;column:name;comment:岗位名称" json:"name"`                 // 岗位名称
	Code        string     `gorm:"size:50;not null;uniqueIndex;column:code;comment:岗位编码" json:"code"`           // 岗位编码
	PostType    post.Type  `gorm:"type:tinyint;index;default:5;column:post_type;comment:岗位类型" json:"post_type"` // 岗位类型
	Level       post.Level `gorm:"type:tinyint;index;default:4;column:level;comment:岗位级别" json:"level"`         // 岗位级别
	Description string     `gorm:"type:text;column:description;comment:岗位描述" json:"description"`                // 岗位描述
	// 所属部门
	DeptID *string `gorm:"size:110;column:dept_id;comment:所属部门ID" json:"dept_id"`                       // 所属部门ID
	Dept   *Dept   `gorm:"foreignKey:DeptID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL" json:"dept"` // 所属部门
	// 关联用户 Post -> User
	Users []*User `gorm:"many2many:careful_system_user_post;constraint:OnDelete:CASCADE;"` // 关联用户
}

func NewPost() *Post {
	return &Post{}
}

func (d *Post) TableName() string {
	return "careful_system_post"
}

func (d *Post) AutoMigrate(db *gorm.DB) {
	err := db.Set("gorm:foreignKeyConstraint", true).
		Set("gorm:table_options", "ENGINE=InnoDB,COMMENT='岗位表'").
		AutoMigrate(&Post{})
	if err != nil {
		zap.L().Error("Post表模型迁移失败", zap.Error(err))
	}
}
