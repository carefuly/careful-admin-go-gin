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
	Id         string     `gorm:"type:char(40);primaryKey;column:id;comment:主键ID" json:"id"`                // 主键ID(自增)
	Sort       int        `gorm:"type:bigint;default:1;index;column:sort;comment:显示排序" json:"sort"`         // 显示排序
	Timestamp  int64      `gorm:"type:bigint;column:timestamp;comment:版本号(时间戳)" json:"timestamp"`           // 版本号(时间戳)
	Creator    string     `gorm:"type:char(40);index;column:creator;comment:创建人" json:"creator"`             // 创建人
	Modifier   string     `gorm:"type:char(40);index;column:modifier;comment:修改人" json:"modifier"`          // 修改人
	BelongDept *string    `gorm:"type:char(40);index;column:belong_dept;comment:数据归属部门" json:"belong_dept"` // 数据归属部门
	CreateTime *time.Time `gorm:"autoCreateTime;index;column:create_time;comment:创建时间" json:"create_time"`  // 创建时间
	UpdateTime *time.Time `gorm:"autoUpdateTime;index;column:update_time;comment:修改时间" json:"update_time"`  // 修改时间
	Remark     string     `gorm:"size:512;column:remark;comment:备注" json:"remark"`                          // 备注
}

// BeforeCreate 创建前钩子
func (c *CoreModels) BeforeCreate(tx *gorm.DB) (err error) {
	// 增加一个判断：如果 Id 字段已经被手动设置了值，就直接返回，不做任何操作
	if c.Id != "" {
		return nil
	}

	var idStr string
	u7, err := uuid7.NewV7()
	if err != nil {
		idStr = uuid4.New().String() // 小写
	} else {
		idStr = u7.String() // 小写
	}

	c.Id = strings.ToUpper(idStr) // 建议保持小写，除非业务强要求大写
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
