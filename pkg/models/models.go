/**
 * Description：
 * FileName：models.go
 * Author：CJiaの用心
 * Create：2025/10/9 15:51:06
 * Remark：
 */

package models

import (
	uuid7 "github.com/gofrs/uuid"
	uuid4 "github.com/google/uuid"
	"gorm.io/gorm"
	"strings"
	"time"
)

// CoreModels 公共模型
// 核心标准抽象模型,可直接继承使用
// 增加审计字段, 覆盖字段时, 字段名称请勿修改, 必须统一审计字段名称
type CoreModels struct {
	Id         string     `gorm:"type:varchar(110);primaryKey;column:id;comment:主键ID" json:"id"`               // 主键ID(自增)
	Sort       int        `gorm:"type:bigint;default:1;index;column:sort;comment:显示排序" json:"sort"`            // 显示排序
	Timestamp  int64      `gorm:"type:bigint;column:timestamp;comment:版本号(时间戳)" json:"timestamp"`              // 版本号(时间戳)
	Creator    string     `gorm:"type:varchar(100);index;column:creator;comment:创建人" json:"creator"`           // 创建人
	Modifier   string     `gorm:"type:varchar(100);index;column:modifier;comment:修改人" json:"modifier"`         // 修改人
	BelongDept string     `gorm:"type:varchar(100);index;column:belong_dept;comment:数据归属部门" json:"belongDept"` // 数据归属部门
	CreateTime *time.Time `gorm:"autoCreateTime;index;column:create_time;comment:创建时间" json:"-"`               // 创建时间
	UpdateTime *time.Time `gorm:"autoUpdateTime;index;column:update_time;comment:修改时间" json:"-"`               // 修改时间
	Remark     string     `gorm:"type:varchar(512);column:remark;comment:备注" json:"remark"`                    // 备注
}

// BeforeCreate 创建前钩子
func (c *CoreModels) BeforeCreate(tx *gorm.DB) (err error) {
	// 设置id
	u7, err := uuid7.NewV7()
	if err != nil {
		// 失败就换UUID4
		c.Id = strings.ToUpper(uuid4.New().String())
	} else {
		c.Id = strings.ToUpper(u7.String())
	}
	// 设置版本号为当前时间戳
	c.Timestamp = time.Now().UnixMicro()
	return nil
}

// BeforeUpdate 更新前钩子
func (c *CoreModels) BeforeUpdate(tx *gorm.DB) (err error) {
	// 更新时更新版本号为当前时间戳
	c.Timestamp = time.Now().UnixMicro()
	return nil
}
