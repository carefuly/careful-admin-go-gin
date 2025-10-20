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
	ErrDictTypeInvalidDictValueType = errors.New("无效的字典值类型")
	ErrDictTypeUniqueIndex          = errors.New("违反唯一约束")
)

// DictType 字典项表
type DictType struct {
	models.CoreModels

	Status    bool                   `gorm:"type:boolean;index:idx_status;default:true;column:status;comment:状态【true-启用 false-停用】" json:"status"` // 状态
	Name      string                 `gorm:"type:varchar(50);not null;index:idx_name;column:name;comment:字典项名称" json:"name"`                      // 字典项名称
	StrValue  sql.NullString         `gorm:"type:varchar(50);column:strValue;comment:字符串-字典项值" swaggertype:"string" json:"strValue"`              // 字符串-字典项值
	IntValue  sql.NullInt64          `gorm:"type:tinyint;column:intValue;comment:整型-字典项值" swaggertype:"number" json:"intValue"`                   // 整型-字典项值
	BoolValue sql.NullBool           `gorm:"type:boolean;column:boolValue;comment:布尔-字典项值" swaggertype:"boolean" json:"boolValue"`                // 布尔-字典项值
	DictTag   dict_type.DictTagConst `gorm:"type:varchar(10);default:primary;index:idx_dict_tag;column:dictTag;comment:标签类型" json:"dictTag"`      // 标签类型
	DictColor string                 `gorm:"type:varchar(50);column:dictColor;comment:标签颜色" json:"dictColor"`                                     // 标签颜色
	DictName  string                 `gorm:"type:varchar(100);index:idx_dict_name;column:dictName;comment:字典名称" json:"dictName"`                  // 字典名称
	ValueType dict.ValueTypeConst    `gorm:"type:tinyint;default:1;index:idx_value_type;column:valueType;comment:数据类型" json:"valueType"`          // 数据类型
	DictId    string                 `gorm:"type:varchar(100);index:idx_dict_id;column:dict_id;comment:所属字典ID" json:"dict_id"`                    // 所属字典ID
	Dict      *Dict                  `gorm:"foreignKey:DictId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"dict"`                         // 数据字典
}

func NewDictType() *DictType {
	return &DictType{}
}

func (d *DictType) TableName() string {
	return "careful_tools_dict_type"
}

func (d *DictType) AutoMigrate(db *gorm.DB) {
	err := db.Set("gorm:table_options", "ENGINE=InnoDB,COMMENT='字典项表'").AutoMigrate(&DictType{})
	if err != nil {
		zap.L().Error("DictType表模型迁移失败", zap.Error(err))
	}

	// MySQL 特殊唯一索引（支持 NULL 值）
	indexes := []string{
		"CREATE UNIQUE INDEX uni_dict_name ON careful_tools_dict_type(dict_id, name)",
		"CREATE UNIQUE INDEX uni_dict_str_value ON careful_tools_dict_type(dict_id, strValue)",
		"CREATE UNIQUE INDEX uni_dict_int_value ON careful_tools_dict_type(dict_id, intValue)",
		"CREATE UNIQUE INDEX uni_dict_bool_value ON careful_tools_dict_type(dict_id, boolValue)",
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
