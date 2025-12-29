/**
 * Description：
 * FileName：operate_logger.go
 * Author：CJiaの用心
 * Create：2025/11/25 01:51:40
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

// OperateLogger 操作日志表
type OperateLogger struct {
	models.CoreModels

	Status          bool   `gorm:"type:boolean;index;column:status;comment:状态【true-启用 false-停用】" json:"status"`            // 状态
	RequestUsername string `gorm:"size:32;index:idx_search;column:request_username;comment:请求用户名" json:"request_username"` // 请求用户名
	RequestTime     string `gorm:"size:32;column:request_time;comment:请求耗时" json:"request_time"`                           // 请求耗时
	RequestStatus   int    `gorm:"type:int;column:request_status;comment:响应状态码" json:"request_status"`                     // 响应状态码
	RequestMethod   string `gorm:"size:32;index:idx_search;column:request_method;comment:请求方式" json:"request_method"`      // 请求方式
	RequestIp       string `gorm:"size:32;column:request_ip;comment:请求IP地址" json:"request_ip"`                             // 请求IP地址
	RequestPath     string `gorm:"size:256;column:request_path;comment:请求地址" json:"request_path"`                          // 请求地址
	RequestQuery    string `gorm:"type:text;column:request_query;comment:请求查询参数" json:"request_query"`                     // 请求查询参数
	RequestBody     any    `gorm:"type:mediumtext;column:request_body;comment:请求体(大文本)" json:"request_body"`               // 请求体(大文本)
	RequestOs       string `gorm:"size:32;column:request_os;comment:操作系统" json:"request_os"`                               // 操作系统
	RequestBrowser  string `gorm:"size:64;column:request_browser;comment:操作浏览器" json:"request_browser"`                    // 操作浏览器
	UserAgent       string `gorm:"size:256;column:user_agent;comment:用户代理" json:"user_agent"`                              // 用户代理
	RequestCode     int    `gorm:"type:int;column:request_code;comment:自定义响应状态码" json:"request_code"`                      // 自定义响应状态码
	RequestResult   string `gorm:"type:text;column:request_result;comment:响应信息" json:"request_result"`                     // 响应信息
	RequestInternal string `gorm:"type:text;column:request_internal;comment:系统错误" json:"request_internal"`                 // 系统错误
}

func NewOperateLogger() *OperateLogger {
	return &OperateLogger{}
}

func (l *OperateLogger) TableName() string {
	return "careful_logger_operate_log"
}

func (l *OperateLogger) AutoMigrate(db *gorm.DB) {
	err := db.Set("gorm:table_options", "ENGINE=InnoDB,COMMENT='操作日志表'").AutoMigrate(&OperateLogger{})
	if err != nil {
		zap.L().Error("OperateLogger表模型迁移失败", zap.Error(err))
	}
}

func (l *OperateLogger) Insert(ctx context.Context, db *gorm.DB, model OperateLogger) {
	currentLogger := db.Config.Logger
	// 临时禁用日志
	db.Config.Logger = logger.Default.LogMode(logger.Silent)

	err := db.WithContext(ctx).Create(&model).Error
	if err != nil {
		zap.L().Error("日志记录异常", zap.String("err", err.Error()))
	}

	// 恢复日志级别
	db.Config.Logger = currentLogger
}
