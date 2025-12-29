/**
 * Description：
 * FileName：login_logger.go
 * Author：CJiaの用心
 * Create：2025/11/25 01:19:57
 * Remark：
 */

package logger

import (
	"github.com/carefuly/careful-admin-go-gin/pkg/models"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// LoginLogger 登录日志表
type LoginLogger struct {
	models.CoreModels

	Status         bool   `gorm:"type:boolean;index;column:status;comment:状态【true-启用 false-停用】" json:"status"`               // 状态
	LoginUsername  string `gorm:"size:32;index;column:login_username;comment:登录用户名" json:"login_username"`                   // 登录用户名
	Ip             string `gorm:"size:32;column:ip;comment:登录ip" json:"ip"`                                                  // 登录ip
	Agent          string `gorm:"type:text;column:agent;comment:agent信息" json:"agent"`                                       // agent信息
	Browser        string `gorm:"size:256;column:browser;comment:浏览器名" json:"browser"`                                       // 浏览器名
	Os             string `gorm:"size:256;column:os;comment:操作系统" json:"os"`                                                 // 操作系统
	Continent      string `gorm:"size:32;column:continent;comment:州" json:"continent"`                                       // 州
	Country        string `gorm:"size:32;column:country;comment:国家" json:"country"`                                          // 国家
	Province       string `gorm:"size:32;column:province;comment:省份" json:"province"`                                        // 省份
	City           string `gorm:"size:32;column:city;comment:城市" json:"city"`                                                // 城市
	District       string `gorm:"size:32;column:district;comment:县区" json:"district"`                                        // 县区
	Isp            string `gorm:"size:32;column:isp;comment:运营商" json:"isp"`                                                 // 运营商
	AreaCode       string `gorm:"size:32;column:area_code;comment:区域代码" json:"area_code"`                                    // 区域代码
	CountryEnglish string `gorm:"size:32;column:country_english;comment:英文全称" json:"country_english"`                        // 英文全称
	CountryCode    string `gorm:"size:32;column:country_code;comment:简称" json:"country_code"`                                // 简称
	Longitude      string `gorm:"size:32;column:longitude;comment:经度" json:"longitude"`                                      // 经度
	Latitude       string `gorm:"size:32;column:latitude;comment:纬度" json:"latitude"`                                        // 纬度
	LoginResult    bool   `gorm:"type:boolean;index;column:login_result;comment:登录结果【true-成功 false-失败】" json:"login_result"` // 登录结果
	FailureReason  string `gorm:"size:256;column:failure_reason;comment:失败结果" json:"failure_reason"`                         // 失败原因
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
