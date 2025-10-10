/**
 * Description：
 * FileName：operate_log.go
 * Author：CJiaの用心
 * Create：2025/10/10 10:55:57
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

	Status          bool   `gorm:"type:boolean;index:idx_status;default:true;column:status;comment:状态【true-启用 false-停用】" json:"status"` // 状态
	RequestUsername string `gorm:"type:varchar(40);index:idx_search;column:requestUsername;comment:请求用户名" json:"requestUsername"`       // 请求用户名
	RequestTime     string `gorm:"type:varchar(40);column:requestTime;comment:请求耗时" json:"requestTime"`                                 // 请求耗时
	RequestStatus   int    `gorm:"type:int;column:requestStatus;comment:响应状态码" json:"requestStatus"`                                    // 响应状态码
	RequestMethod   string `gorm:"type:varchar(20);index:idx_search;column:requestMethod;comment:请求方式" json:"requestMethod"`            // 请求方式
	RequestIp       string `gorm:"type:varchar(20);column:requestIp;comment:请求IP地址" json:"requestIp"`                                   // 请求IP地址
	RequestPath     string `gorm:"type:varchar(255);column:requestPath;comment:请求地址" json:"requestPath"`                                // 请求地址
	RequestQuery    string `gorm:"type:text;column:requestQuery;comment:请求查询参数" json:"requestQuery"`                                    // 请求查询参数
	RequestBody     any    `gorm:"type:mediumtext;column:requestBody;comment:请求体(大文本)" json:"requestBody"`                              // 请求体(大文本)
	RequestOs       string `gorm:"type:varchar(40);column:requestOs;comment:操作系统" json:"requestOs"`                                     // 操作系统
	RequestBrowser  string `gorm:"type:varchar(64);column:requestBrowser;comment:操作浏览器" json:"requestBrowser"`                          // 操作浏览器
	UserAgent       string `gorm:"type:varchar(255);column:userAgent;comment:用户代理" json:"userAgent"`                                    // 用户代理
	RequestCode     int    `gorm:"type:int;column:requestCode;comment:自定义响应状态码" json:"requestCode"`                                     // 自定义响应状态码
	RequestResult   string `gorm:"type:text;column:requestResult;comment:响应信息" json:"requestResult"`                                    // 响应信息
	RequestInternal string `gorm:"type:text;column:requestInternal;comment:系统错误" json:"requestInternal"`                                // 系统错误
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
