/**
 * Description：
 * FileName：login_log.go
 * Author：CJiaの用心
 * Create：2025/10/10 10:51:18
 * Remark：
 */

package logger

import (
	"context"
	"github.com/carefuly/careful-admin-go-gin/pkg/models"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// LoginLogger 登录日志表
type LoginLogger struct {
	models.CoreModels

	Status         bool   `gorm:"type:boolean;index:idx_status;default:true;column:status;comment:状态【true-启用 false-停用】" json:"status"` // 状态
	LoginUsername  string `gorm:"type:varchar(40);index:idx_search;column:loginUsername;comment:登录用户名" json:"loginUsername"`           // 登录用户名
	Ip             string `gorm:"type:varchar(32);column:ip;comment:登录ip" json:"ip"`                                                   // 登录ip
	Agent          string `gorm:"type:mediumtext;column:agent;comment:agent信息" json:"agent"`                                           // agent信息
	Browser        string `gorm:"type:varchar(255);column:browser;comment:浏览器名" json:"browser"`                                        // 浏览器名
	Os             string `gorm:"type:varchar(255);column:os;comment:操作系统" json:"os"`                                                  // 操作系统
	Continent      string `gorm:"type:varchar(50);column:continent;comment:州" json:"continent"`                                        // 州
	Country        string `gorm:"type:varchar(50);column:country;comment:国家" json:"country"`                                           // 国家
	Province       string `gorm:"type:varchar(50);column:province;comment:省份" json:"province"`                                         // 省份
	City           string `gorm:"type:varchar(50);column:city;comment:城市" json:"city"`                                                 // 城市
	District       string `gorm:"type:varchar(50);column:district;comment:县区" json:"district"`                                         // 县区
	Isp            string `gorm:"type:varchar(50);column:isp;comment:运营商" json:"isp"`                                                  // 运营商
	AreaCode       string `gorm:"type:varchar(50);column:area_code;comment:区域代码" json:"area_code"`                                     // 区域代码
	CountryEnglish string `gorm:"type:varchar(50);column:country_english;comment:英文全称" json:"country_english"`                         // 英文全称
	CountryCode    string `gorm:"type:varchar(50);column:country_code;comment:简称" json:"country_code"`                                 // 简称
	Longitude      string `gorm:"type:varchar(50);column:longitude;comment:经度" json:"longitude"`                                       // 经度
	Latitude       string `gorm:"type:varchar(50);column:latitude;comment:纬度" json:"latitude"`                                         // 纬度
}

func NewLoginLogger() *LoginLogger {
	return &LoginLogger{}
}

func (l *LoginLogger) TableName() string {
	return "careful_logger_login_log"
}

func (l *LoginLogger) AutoMigrate(db *gorm.DB) {
	err := db.Set("gorm:table_options", "ENGINE=InnoDB,COMMENT='登录日志表'").AutoMigrate(&LoginLogger{})
	if err != nil {
		zap.L().Error("LoginLogger表模型迁移失败", zap.Error(err))
	}
}

func (l *LoginLogger) Insert(ctx context.Context, db *gorm.DB, model LoginLogger) {
	currentLogger := db.Config.Logger
	// 临时禁用日志
	db.Config.Logger = logger.Default.LogMode(logger.Silent)

	err := db.WithContext(ctx).Create(&model).Error
	if err != nil {
		zap.L().Error("登录日志异常", zap.String("err", err.Error()))
	}

	// 恢复日志级别
	db.Config.Logger = currentLogger
}
