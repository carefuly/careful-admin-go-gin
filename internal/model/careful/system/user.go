/**
 * Description：
 * FileName：user.go
 * Author：CJiaの用心
 * Create：2025/11/24 15:08:04
 * Remark：
 */

package system

import (
	"fmt"
	"github.com/carefuly/careful-admin-go-gin/pkg/constants/careful/system/user"
	"github.com/carefuly/careful-admin-go-gin/pkg/models"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"strings"
	"time"
)

// User 用户表
type User struct {
	models.CoreModels

	Status      user.Status      `gorm:"type:tinyint;index;default:1;column:status;comment:状态" json:"status"`                       // 状态
	Username    string           `gorm:"size:50;not null;uniqueIndex;column:username;comment:用户名" json:"username"`                  // 用户名
	Password    string           `gorm:"size:512;not null;column:password;comment:密码" json:"-"`                                     // 密码（加密存储）, 返回给前端时隐藏
	Gender      user.GenderConst `gorm:"type:tinyint;default:1;column:gender;comment:性别" json:"gender"`                             // 性别
	Email       string           `gorm:"size:128;index;column:email;comment:邮箱" json:"email"`                                       // 邮箱
	Mobile      string           `gorm:"size:11;index;column:mobile;comment:手机号" json:"mobile"`                                     // 手机号
	Name        string           `gorm:"size:64;index;column:name;comment:真实姓名" json:"name"`                                        // 真实姓名
	Avatar      string           `gorm:"type:mediumtext;column:avatar;comment:头像（url地址）" json:"avatar"`                             // 头像
	Birthday    *time.Time       `gorm:"column:birthday;comment:生日" json:"birthday"`                                                // 生日
	City        string           `gorm:"size:100;column:city;comment:所在城市" json:"city"`                                             // 所在城市
	Address     string           `gorm:"size:200;column:address;comment:详细地址" json:"address"`                                       // 详细地址
	Bio         string           `gorm:"size:512;column:bio;comment:个人简介" json:"bio"`                                               // 个人简介
	IsSuperuser bool             `gorm:"type:boolean;index;default:false;column:is_superuser;comment:是否为超级管理员" json:"is_superuser"` // 是否为超级管理员
	LastLogin   *time.Time       `gorm:"column:last_login;comment:最后登录时间" json:"last_login"`                                        // 最后登录时间
	LastLoginIp string           `gorm:"size:64;column:last_login_ip;comment:最后登录IP" json:"last_login_ip"`                          // 最后登录IP

	DeptID *string `gorm:"size:110;index;column:dept_id;comment:部门ID" json:"dept_id"`                   // 部门ID
	Dept   *Dept   `gorm:"foreignKey:DeptID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL" json:"dept"` // 部门
	// User -> Post
	Posts []*Post `gorm:"many2many:careful_system_user_post;constraint:OnDelete:CASCADE;"` // 关联岗位
}

func NewUser() *User {
	return &User{}
}

func (u *User) TableName() string {
	return "careful_system_user"
}

func (u *User) AutoMigrate(db *gorm.DB) {
	err := db.Set("gorm:foreignKeyConstraint", true).
		Set("gorm:table_options", "ENGINE=InnoDB,COMMENT='用户表'").
		AutoMigrate(&User{})
	if err != nil {
		zap.L().Error("User表模型迁移失败", zap.Error(err))
	}

	// 迁移中间表并设置备注
	u.migrateManyToManyTable(db, "careful_system_user_post", "用户-关联岗位表")
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

// 迁移many2many中间表并设置表备注
func (u *User) migrateManyToManyTable(db *gorm.DB, tableName string, comment string) {
	err := db.Exec(fmt.Sprintf(
		"ALTER TABLE %s COMMENT = '%s'",
		tableName,
		comment,
	)).Error

	if err != nil {
		zap.L().Error(fmt.Sprintf("%s表备注设置失败", tableName), zap.Error(err))
	}
}
