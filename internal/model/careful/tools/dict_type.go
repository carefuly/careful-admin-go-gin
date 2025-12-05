/**
 * Description：
 * FileName：dict_type.go
 * Author：CJiaの用心
 * Create：2025/10/10 15:45:47
 * Remark：
 */

package tools

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/carefuly/careful-admin-go-gin/pkg/constants/careful/tools/dict"
	"github.com/carefuly/careful-admin-go-gin/pkg/constants/careful/tools/dict_type"
	"github.com/carefuly/careful-admin-go-gin/pkg/models"
	"github.com/go-sql-driver/mysql"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var (
	ErrDictTypeInvalidDictValueType = errors.New("无效的数据类型")
	ErrDictTypeUniqueIndex          = errors.New("违反唯一约束")
)

// DictType 字典项表
type DictType struct {
	models.CoreModels

	Status    bool              `gorm:"type:boolean;index;column:status;comment:状态【true-启用 false-停用】" json:"status"`            // 状态
	Name      string            `gorm:"size:50;not null;index;column:name;comment:字典项名称" json:"name"`                           // 字典项名称
	StrValue  sql.NullString    `gorm:"size:50;column:str_value;comment:字符串-字典项值" swaggertype:"string" json:"str_value"`        // 字符串-字典项值
	IntValue  sql.NullInt64     `gorm:"type:tinyint;column:int_value;comment:整型-字典项值" swaggertype:"number" json:"int_value"`    // 整型-字典项值
	BoolValue sql.NullBool      `gorm:"type:boolean;column:bool_value;comment:布尔-字典项值" swaggertype:"boolean" json:"bool_value"` // 布尔-字典项值
	DictTag   dict_type.DictTag `gorm:"size:10;default:primary;index;column:dict_tag;comment:标签类型" json:"dict_tag"`             // 标签类型
	DictColor string            `gorm:"size:20;column:dict_color;comment:标签颜色" json:"dict_color"`                               // 标签颜色
	DictName  string            `gorm:"size:100;index;column:dict_name;comment:字典名称" json:"dict_name"`                          // 字典名称
	ValueType dict.ValueType    `gorm:"type:tinyint;default:1;index;column:value_type;comment:数据类型" json:"value_type"`          // 数据类型
	DictID    string            `gorm:"size:110;index;column:dict_id;comment:所属字典ID" json:"dict_id"`                            // 所属字典ID
	Dict      *Dict             `gorm:"foreignKey:DictID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"dict"`             // 数据字典
}

func NewDictType() *DictType {
	return &DictType{}
}

func (d *DictType) TableName() string {
	return "careful_tools_dict_type"
}

func (d *DictType) AutoMigrate(db *gorm.DB) {
	err := db.Set("gorm:foreignKeyConstraint", true).
		Set("gorm:table_options", "ENGINE=InnoDB,COMMENT='字典项表'").
		AutoMigrate(&DictType{})
	if err != nil {
		zap.L().Error("DictType表模型迁移失败", zap.Error(err))
	}

	// MySQL 特殊唯一索引（支持 NULL 值）
	indexes := []string{
		"CREATE UNIQUE INDEX uni_dict_name ON careful_tools_dict_type(dict_id, name)",
		"CREATE UNIQUE INDEX uni_dict_str_value ON careful_tools_dict_type(dict_id, str_value)",
		"CREATE UNIQUE INDEX uni_dict_int_value ON careful_tools_dict_type(dict_id, int_value)",
		"CREATE UNIQUE INDEX uni_dict_bool_value ON careful_tools_dict_type(dict_id, bool_value)",
	}

	for _, s := range indexes {
		if err := db.Exec(s).Error; err != nil {
			var mysqlErr *mysql.MySQLError
			if errors.As(err, &mysqlErr) && mysqlErr.Number == 1061 {
				// 索引已存在，忽略错误
				zap.L().Debug("索引已存在", zap.String("sql", s))
				continue
			}
			zap.L().Error("创建字典项索引失败", zap.String("sql", s), zap.Error(err))
		}
	}
}

// BeforeSave 在创建/更新时校验数据一致性
func (d *DictType) BeforeSave(tx *gorm.DB) error {
	// 根据类型清理无关字段
	switch d.ValueType {
	case 1: // 字符串
		d.IntValue = sql.NullInt64{Valid: false}
		d.BoolValue = sql.NullBool{Valid: false}
		// 验证字符串值是否有效
		if !d.StrValue.Valid || d.StrValue.String == "" {
			return fmt.Errorf("字符串值不能为空")
		}
	case 2: // 整型
		d.StrValue = sql.NullString{Valid: false}
		d.BoolValue = sql.NullBool{Valid: false}
		// 验证整数值是否有效
		if !d.IntValue.Valid {
			return fmt.Errorf("整数值不能为空")
		}
	case 3: // 布尔
		d.StrValue = sql.NullString{Valid: false}
		d.IntValue = sql.NullInt64{Valid: false}
		// 验证布尔值是否有效
		if !d.BoolValue.Valid {
			return fmt.Errorf("布尔值不能为空")
		}
	default:
		return ErrDictTypeInvalidDictValueType
	}

	return nil
}
