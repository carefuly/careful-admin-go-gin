/**
 * Description：
 * FileName：user.go
 * Author：CJiaの用心
 * Create：2025/10/9 16:03:12
 * Remark：
 */

package system

import (
	"database/sql"
	"fmt"
	"github.com/carefuly/careful-admin-go-gin/pkg/constants/careful/system/user"
	"github.com/carefuly/careful-admin-go-gin/pkg/models"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"strings"
)

// User 用户表
type User struct {
	models.CoreModels

	Status   bool             `gorm:"type:boolean;index:idx_status;default:true;column:status;comment:状态【true-启用 false-停用】" json:"status"` // 状态
	Username string           `gorm:"type:varchar(50);not null;uniqueIndex;column:username;comment:用户名" json:"username"`                   // 用户名
	Password string           `gorm:"type:varchar(512);not null;column:password;comment:密码" json:"-"`                                      // 密码
	Name     string           `gorm:"type:varchar(50);index:idx_search;column:name;comment:姓名" json:"name"`                                // 姓名
	Gender   user.GenderConst `gorm:"type:tinyint;default:1;column:gender;comment:性别" json:"gender"`                                       // 性别
	Email    string           `gorm:"type:varchar(50);index:idx_search;column:email;comment:邮箱" json:"email"`                              // 邮箱
	Mobile   string           `gorm:"type:varchar(20);index:idx_search;column:mobile;comment:电话" json:"mobile"`                            // 电话
	Avatar   string           `gorm:"type:mediumtext;column:avatar;comment:头像（url地址）" json:"avatar"`                                       // 头像

	DeptId sql.NullString `gorm:"type:varchar(100);index;column:dept_id;comment:部门ID（可为空）" swaggertype:"string" json:"dept_id"` // 部门ID（可为空）
	Dept   *Dept          `gorm:"foreignKey:DeptId;constraint:OnUpdate:CASCADE,OnDelete:SET NULL" json:"dept"`                  // 部门
}

func NewUser() *User {
	return &User{}
}

func (u *User) TableName() string {
	return "careful_system_users"
}

func (u *User) AutoMigrate(db *gorm.DB) {
	err := db.Set("gorm:table_options", "ENGINE=InnoDB,COMMENT='用户表'").AutoMigrate(&User{})
	if err != nil {
		zap.L().Error("User表模型迁移失败", zap.Error(err))
	}
}

func (u *User) AfterCreate(tx *gorm.DB) error {
	return u.updateDeptUserCount(tx, 1)
}

func (u *User) AfterDelete(tx *gorm.DB) error {
	return u.updateDeptUserCount(tx, -1)
}

// updateDeptUserCount 更新部门用户数
func (u *User) updateDeptUserCount(tx *gorm.DB, delta int) error {
	if u.DeptId.Valid {
		return tx.Model(&Dept{}).Where("id = ?", u.DeptId.String).
			UpdateColumn("user_count", gorm.Expr("user_count + ?", delta)).Error
	}
	return nil
}

// Validate 验证用户数据
func (u *User) Validate() error {
	if u.Username == "" {
		return fmt.Errorf("用户名不能为空")
	}

	if len(u.Username) < 4 {
		return fmt.Errorf("用户名长度不能少于3位")
	}

	if u.Password == "" {
		return fmt.Errorf("密码不能为空")
	}

	if u.Email != "" {
		if !strings.Contains(u.Email, "@") {
			return fmt.Errorf("邮箱格式不正确")
		}
	}

	if u.Mobile != "" && len(u.Mobile) < 11 {
		return fmt.Errorf("手机号格式不正确")
	}

	return nil
}
