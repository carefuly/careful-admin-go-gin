/**
 * Description：
 * FileName：request_utils.go
 * Author：CJiaの用心
 * Create：2025/10/10 11:44:23
 * Remark：
 */

package request_utils

import (
	"encoding/json"
	"fmt"
	"github.com/carefuly/careful-admin-go-gin/internal/domain/careful/system"
	"github.com/carefuly/careful-admin-go-gin/internal/model/careful/logger"
	"github.com/carefuly/careful-admin-go-gin/pkg/models"
	"github.com/gin-gonic/gin"
	"github.com/mssola/user_agent"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
)

// GetRequestUser 获取请求user
func GetRequestUser(c *gin.Context) string {
	username, ok := c.Get("username")
	if !ok {
		return "AnonymousUser"
	}
	return username.(string)
}

// NormalizeIP IPv6 本地地址处理
func NormalizeIP(c *gin.Context) string {
	ip := c.ClientIP()

	// 如果地址是 IPv6 本地回环
	if ip == "::1" {
		return "127.0.0.1"
	}

	// 如果地址是 IPv4 本地回环
	if ip == "0:0:0:0:0:0:0:1" {
		return "127.0.0.1"
	}

	// 处理IPv6地址的方括号
	if strings.HasPrefix(ip, "[") && strings.HasSuffix(ip, "]") {
		return ip[1 : len(ip)-1]
	}

	return ip
}

// IPAnalysisData IP分析结果结构体
type IPAnalysisData struct {
	Continent      string `json:"continent"`
	Country        string `json:"country"`
	Province       string `json:"province"`
	City           string `json:"city"`
	District       string `json:"district"`
	Isp            string `json:"isp"`
	AreaCode       string `json:"area_code"`
	CountryEnglish string `json:"country_english"`
	CountryCode    string `json:"country_code"`
	Longitude      string `json:"longitude"`
	Latitude       string `json:"latitude"`
}

// GetIPAnalysis 获取IP详细分析
func GetIPAnalysis(ip string) *IPAnalysisData {
	data := &IPAnalysisData{
		Continent:      "",
		Country:        "",
		Province:       "",
		City:           "",
		District:       "",
		Isp:            "",
		AreaCode:       "",
		CountryEnglish: "",
		CountryCode:    "",
		Longitude:      "",
		Latitude:       "",
	}

	if ip == "" || ip == "unknown" {
		return data
	}

	// 调用IP分析API
	resp, err := http.Get(fmt.Sprintf("https://ip.django-vue-admin.com/ip/analysis?ip=%s", ip))
	if err != nil {
		return data
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			return
		}
	}(resp.Body)

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return data
	}

	var result struct {
		Code int            `json:"code"`
		Data IPAnalysisData `json:"data"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return data
	}

	if result.Code == 0 {
		return &result.Data
	}

	return data
}

// GetUserAgent 获取用户代理信息
func GetUserAgent(c *gin.Context) string {
	return c.GetHeader("User-Agent")
}

// GetBrowser 获取浏览器信息
func GetBrowser(ua string) string {
	parsed := user_agent.New(ua)
	browser, _ := parsed.Browser()
	return browser
}

// GetOS 获取操作系统信息
func GetOS(ua string) string {
	parsed := user_agent.New(ua)
	return parsed.OS()
}

// SaveLoginLog 保存登录日志
func SaveLoginLog(c *gin.Context, user system.User, db *gorm.DB) {
	ip := NormalizeIP(c)
	ua := GetUserAgent(c)

	analysisData := GetIPAnalysis(ip)

	log := logger.LoginLogger{
		LoginUsername:  user.Username,
		Ip:             ip,
		Agent:          ua,
		Browser:        GetBrowser(ua),
		Os:             GetOS(ua),
		Continent:      analysisData.Continent,
		Country:        analysisData.Country,
		Province:       analysisData.Province,
		City:           analysisData.City,
		District:       analysisData.District,
		Isp:            analysisData.Isp,
		AreaCode:       analysisData.AreaCode,
		CountryEnglish: analysisData.CountryEnglish,
		CountryCode:    analysisData.CountryCode,
		Longitude:      analysisData.Longitude,
		Latitude:       analysisData.Latitude,
		CoreModels: models.CoreModels{
			Creator:    user.Id, // 假设用户ID字段为ID
			Modifier:   user.Id,
			BelongDept: user.DeptId, // 假设用户有部门ID字段
		},
	}

	if err := db.Create(&log).Error; err != nil {
		zap.L().Error("保存登录日志失败", zap.Error(err))
	}
}
