/**
 * Description：
 * FileName：dict.go
 * Author：CJiaの用心
 * Create：2025/12/3 11:37:34
 * Remark：
 */

package tools

import (
	"errors"
	"fmt"
	"github.com/carefuly/careful-admin-go-gin/config"
	domainTools "github.com/carefuly/careful-admin-go-gin/internal/domain/careful/tools"
	modelTools "github.com/carefuly/careful-admin-go-gin/internal/model/careful/tools"
	serviceSystem "github.com/carefuly/careful-admin-go-gin/internal/service/careful/system"
	serviceTools "github.com/carefuly/careful-admin-go-gin/internal/service/careful/tools"
	"github.com/carefuly/careful-admin-go-gin/pkg/constants/careful/tools/dict"
	"github.com/carefuly/careful-admin-go-gin/pkg/ginx/filters"
	"github.com/carefuly/careful-admin-go-gin/pkg/ginx/response"
	"github.com/carefuly/careful-admin-go-gin/pkg/models"
	"github.com/carefuly/careful-admin-go-gin/pkg/utils/common/xlsx"
	"github.com/carefuly/careful-admin-go-gin/pkg/utils/enumconv"
	"github.com/carefuly/careful-admin-go-gin/pkg/utils/excelutil"
	"github.com/carefuly/careful-admin-go-gin/pkg/utils/jwt"
	"github.com/carefuly/careful-admin-go-gin/pkg/validate"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// CreateDictRequest 创建
type CreateDictRequest struct {
	Status      bool           `json:"status" binding:"omitempty" default:"true"`          // 状态【true-启用 false-停用】
	Name        string         `json:"name" binding:"required,max=64" default:""`          // 字典名称
	Code        string         `json:"code" binding:"required,max=64" default:""`          // 字典编码
	Type        dict.Type      `json:"type" binding:"omitempty" default:"1"`               // 字典分类
	ValueType   dict.ValueType `json:"value_type" binding:"omitempty" default:"1"`         // 数据类型
	Description string         `json:"description" binding:"omitempty,max=256" default:""` // 字典描述
	Sort        int            `json:"sort" binding:"omitempty" default:"1"`               // 排序
	Remark      string         `json:"remark" binding:"omitempty,max=512" default:""`      // 备注
}

// ImportDictRequest 导入
type ImportDictRequest struct {
	File *multipart.FileHeader `form:"file" binding:"required"`
}

// UpdateDictRequest 更新
type UpdateDictRequest struct {
	Id          string `json:"id" binding:"required" default:""`                   // 主键ID
	Status      bool   `json:"status" binding:"omitempty" default:"true"`          // 状态【true-启用 false-停用】
	Code        string `json:"code" binding:"required,max=64" default:""`          // 字典编码
	Description string `json:"description" binding:"omitempty,max=256" default:""` // 字典描述
	Sort        int    `json:"sort" binding:"omitempty" default:"1"`               // 排序
	Timestamp   int64  `json:"timestamp" binding:"omitempty"`                      // 版本
	Remark      string `json:"remark" binding:"omitempty,max=512" default:""`      // 备注
}

// DictListPageResponse 列表分页响应
type DictListPageResponse struct {
	List     []domainTools.Dict `json:"list"`     // 列表
	Total    int64              `json:"total"`    // 总数
	Page     int                `json:"page"`     // 页码
	PageSize int                `json:"pageSize"` // 每页数量
}

type DictHandler interface {
	RegisterRoutes(router *gin.RouterGroup)
	Create(ctx *gin.Context)
	Import(ctx *gin.Context)
	Delete(ctx *gin.Context)
	BatchDelete(ctx *gin.Context)
	Update(ctx *gin.Context)
	GetById(ctx *gin.Context)
	GetListPage(ctx *gin.Context)
	GetListAll(ctx *gin.Context)
	Export(ctx *gin.Context)
}

type dictHandler struct {
	rely    config.RelyConfig
	svc     serviceTools.DictService
	userSvc serviceSystem.UserService
}

func NewDictHandler(rely config.RelyConfig, svc serviceTools.DictService, userSvc serviceSystem.UserService) DictHandler {
	return &dictHandler{
		rely:    rely,
		svc:     svc,
		userSvc: userSvc,
	}
}

// RegisterRoutes 注册路由
func (h *dictHandler) RegisterRoutes(router *gin.RouterGroup) {
	base := router.Group("/dict")
	base.POST("/create", h.Create)
	base.POST("/import", h.Import)
	base.DELETE("/delete/:id", h.Delete)
	base.POST("/batchDelete", h.BatchDelete)
	base.PUT("/update", h.Update)
	base.GET("/getById/:id", h.GetById)
	base.GET("/listPage", h.GetListPage)
	base.GET("/listAll", h.GetListAll)
	base.GET("/export", h.Export)
}

// Create
// @Summary 创建字典
// @Description 创建字典
// @Tags 系统工具/字典管理
// @Accept application/json
// @Produce application/json
// @Security BearerAuth
// @Param CreateDictRequest body CreateDictRequest true "请求"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Router /v1/tools/dict/create [post]
// @Security LoginToken
func (h *dictHandler) Create(ctx *gin.Context) {
	// 从上下文中获取登录信息
	claims, ok := ctx.MustGet("claims").(*jwt.Claims)
	if !ok {
		zap.S().Error("未找到用户认证信息 >>> ", zap.Error(errors.New(claims.UserID)))
		response.NewResponse().Error(ctx, http.StatusInternalServerError, "服务器异常", nil)
		return
	}

	user, err := h.userSvc.GetById(ctx, claims.UserID)
	if err != nil {
		ctx.Set("internalError", fmt.Sprintf("获取用户信息异常 >>> %v", err.Error()))
		zap.S().Error("获取用户信息异常 >>> ", err.Error())
		response.NewResponse().Error(ctx, http.StatusInternalServerError, "服务器异常", nil)
		return
	}

	var req CreateDictRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		validate.NewValidatorErrorHandler(h.rely.Trans).Handle(ctx, err)
		return
	}

	// 校验参数
	typeValidValues := []string{"普通字典", "系统字典", "枚举字典"}
	converter := enumconv.NewEnumConverter(dict.TypeMapping, dict.TypeImportMapping, typeValidValues, "字典类型")
	_, err = converter.FromEnum(req.Type)
	if err != nil {
		response.NewResponse().Error(ctx, http.StatusBadRequest, err.Error(), nil)
		return
	}
	valueTypeValidValues := []string{"字符串", "整型", "布尔"}
	valueTypeConverter := enumconv.NewEnumConverter(dict.TypeValueMapping, dict.TypeValueImportMapping, valueTypeValidValues, "数据类型")
	_, err = valueTypeConverter.FromEnum(req.ValueType)
	if err != nil {
		response.NewResponse().Error(ctx, http.StatusBadRequest, err.Error(), nil)
		return
	}

	// 转换为领域模型
	domain := domainTools.Dict{
		Dict: modelTools.Dict{
			CoreModels: models.CoreModels{
				Sort:       req.Sort,
				Creator:    user.Id,
				Modifier:   user.Id,
				BelongDept: user.DeptID,
				Remark:     req.Remark,
			},
			Status:      req.Status,
			Name:        req.Name,
			Code:        req.Code,
			Type:        req.Type,
			ValueType:   req.ValueType,
			Description: req.Description,
		},
	}

	domain, err = h.svc.Create(ctx, domain)
	if err != nil {
		switch {
		case errors.Is(err, serviceTools.ErrDictNameDuplicate):
			response.NewResponse().Error(ctx, http.StatusBadRequest, "字典名称已存在", nil)
			return
		case errors.Is(err, serviceTools.ErrDictCodeDuplicate):
			response.NewResponse().Error(ctx, http.StatusBadRequest, "字典编码已存在", nil)
			return
		default:
			ctx.Set("internalError", fmt.Sprintf("创建字典异常 >>> %v", err.Error()))
			zap.S().Error("创建字典异常 >>> ", zap.Error(err))
			response.NewResponse().Error(ctx, http.StatusInternalServerError, "服务器异常", nil)
			return
		}
	}

	response.NewResponse().Success(ctx, "新增成功", domain)
}

// Import
// @Summary 导入字典
// @Description 导入字典
// @Tags 系统工具/字典管理
// @Accept multipart/form-data
// @Produce application/json
// @Security BearerAuth
// @Param file formData file true "文件(支持xlsx/csv格式)"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Router /v1/tools/dict/import [post]
// @Security LoginToken
func (h *dictHandler) Import(ctx *gin.Context) {
	// 从上下文中获取登录信息
	claims, ok := ctx.MustGet("claims").(*jwt.Claims)
	if !ok {
		zap.S().Error("未找到用户认证信息 >>> ", zap.Error(errors.New(claims.UserID)))
		response.NewResponse().Error(ctx, http.StatusInternalServerError, "服务器异常", nil)
		return
	}

	user, err := h.userSvc.GetById(ctx, claims.UserID)
	if err != nil {
		ctx.Set("internalError", fmt.Sprintf("获取用户信息异常 >>> %v", err.Error()))
		zap.S().Error("获取用户信息异常 >>> ", err.Error())
		response.NewResponse().Error(ctx, http.StatusInternalServerError, "服务器异常", nil)
		return
	}

	var req ImportDictRequest
	if err := ctx.ShouldBind(&req); err != nil {
		validate.NewValidatorErrorHandler(h.rely.Trans).Handle(ctx, err)
		return
	}

	// 验证文件类型
	ext := strings.ToLower(filepath.Ext(req.File.Filename))
	if ext != ".xls" && ext != ".xlsx" {
		response.NewResponse().Error(ctx, http.StatusBadRequest, "不支持的文件格式，仅支持xls/xlsx", nil)
		return
	}

	// 验证文件大小
	const maxFileSize = 5 << 20 // 5MB
	if req.File.Size > maxFileSize {
		response.NewResponse().Error(ctx, http.StatusBadRequest,
			fmt.Sprintf("文件大小超过限制(%dMB)", maxFileSize/(1<<20)), nil)
		return
	}

	// 创建安全的文件名
	safeFilename := req.File.Filename

	// 创建目录结构
	format := time.Now().Format("2006-01-02")
	uploadDir := filepath.Join("./media/uploads", format)

	// 确保目录存在
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		ctx.Set("internalError", fmt.Sprintf("创建目录失败 >>> %v", err.Error()))
		zap.S().Error("创建目录失败 >>> ", err.Error())
		response.NewResponse().Error(ctx, http.StatusInternalServerError, "服务器异常", nil)
		return
	}

	// 正确的文件保存路径
	filePath := filepath.Join(uploadDir, safeFilename)

	// 保存导入的文件信息
	if err := ctx.SaveUploadedFile(req.File, filePath); err != nil {
		response.NewResponse().Error(ctx, http.StatusBadRequest, "保存文件失败: "+err.Error(), nil)
		return
	}

	// 读取Excel文件
	read, err := xlsx.NewXlsxFile(filePath).ReadSheetByName("字典模板")
	if err != nil {
		response.NewResponse().Error(ctx, http.StatusBadRequest, err.Error(), nil)
		return
	}

	result := h.svc.Import(ctx, user, read)

	response.NewResponse().Success(ctx, "导入成功", result)
}

// Delete
// @Summary 删除字典
// @Description 删除指定id字典
// @Tags 系统工具/字典管理
// @Accept application/json
// @Produce application/json
// @Security BearerAuth
// @Param id path string true "id"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Router /v1/tools/dict/delete/{id} [delete]
// @Security LoginToken
func (h *dictHandler) Delete(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" || len(id) == 0 {
		response.NewResponse().Error(ctx, http.StatusBadRequest, "ID不能为空", nil)
		return
	}

	if err := h.svc.Delete(ctx, id); err != nil {
		switch {
		case errors.Is(err, serviceTools.ErrDictHasType):
			response.NewResponse().Error(ctx, http.StatusBadRequest, "字典下仍有字典项，无法删除", nil)
			return
		default:
			ctx.Set("internalError", fmt.Sprintf("删除字典异常 >>> %v", err.Error()))
			zap.S().Error("删除字典异常 >>> ", zap.Error(err))
			response.NewResponse().Error(ctx, http.StatusInternalServerError, "服务器异常", nil)
			return
		}
	}

	response.NewResponse().Success(ctx, "删除成功", nil)
}

// BatchDelete
// @Summary 批量删除字典
// @Description 批量删除字典
// @Tags 系统工具/字典管理
// @Accept application/json
// @Produce application/json
// @Security BearerAuth
// @Param ids body []string true "id数组"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Router /v1/tools/dict/batchDelete [post]
// @Security LoginToken
func (h *dictHandler) BatchDelete(ctx *gin.Context) {
	var ids []string
	if err := ctx.ShouldBindJSON(&ids); err != nil {
		validate.NewValidatorErrorHandler(h.rely.Trans).Handle(ctx, err)
		return
	}

	if err := h.svc.BatchDelete(ctx, ids); err != nil {
		ctx.Set("internalError", fmt.Sprintf("批量删除字典异常 >>> %v", err.Error()))
		zap.S().Error("批量删除字典异常 >>> ", zap.Error(err))
		response.NewResponse().Error(ctx, http.StatusInternalServerError, "服务器异常", nil)
		return
	}

	response.NewResponse().Success(ctx, "批量删除成功", nil)
}

// Update
// @Summary 更新字典
// @Description 更新字典
// @Tags 系统工具/字典管理
// @Accept application/json
// @Produce application/json
// @Security BearerAuth
// @Param UpdateDictRequest body UpdateDictRequest true "请求"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Router /v1/tools/dict/update [put]
// @Security LoginToken
func (h *dictHandler) Update(ctx *gin.Context) {
	// 从上下文中获取登录信息
	claims, ok := ctx.MustGet("claims").(*jwt.Claims)
	if !ok {
		zap.S().Error("未找到用户认证信息 >>> ", zap.Error(errors.New(claims.UserID)))
		response.NewResponse().Error(ctx, http.StatusInternalServerError, "服务器异常", nil)
		return
	}

	user, err := h.userSvc.GetById(ctx, claims.UserID)
	if err != nil {
		ctx.Set("internalError", fmt.Sprintf("获取用户信息异常 >>> %v", err.Error()))
		zap.S().Error("获取用户信息异常 >>> ", err.Error())
		response.NewResponse().Error(ctx, http.StatusInternalServerError, "服务器异常", nil)
		return
	}

	var req UpdateDictRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		validate.NewValidatorErrorHandler(h.rely.Trans).Handle(ctx, err)
		return
	}

	// 转换为领域模型
	domain := domainTools.Dict{
		Dict: modelTools.Dict{
			CoreModels: models.CoreModels{
				Id:         req.Id,
				Sort:       req.Sort,
				Timestamp:  req.Timestamp,
				Modifier:   user.Id,
				BelongDept: user.DeptID,
				Remark:     req.Remark,
			},
			Status:      req.Status,
			Code:        req.Code,
			Description: req.Description,
		},
	}

	if err := h.svc.Update(ctx, domain); err != nil {
		switch {
		case errors.Is(err, serviceTools.ErrDictNameDuplicate):
			response.NewResponse().Error(ctx, http.StatusBadRequest, "字典名称已存在", nil)
			return
		case errors.Is(err, serviceTools.ErrDictCodeDuplicate):
			response.NewResponse().Error(ctx, http.StatusBadRequest, "字典编码已存在", nil)
			return
		case errors.Is(err, serviceTools.ErrDictVersionInconsistency):
			response.NewResponse().Error(ctx, http.StatusBadRequest, "数据版本不一致，取消修改，请刷新后重试", nil)
			return
		default:
			ctx.Set("internalError", fmt.Sprintf("更新字典异常 >>> %v", err.Error()))
			zap.S().Error("更新字典异常 >>> ", err.Error())
			response.NewResponse().Error(ctx, http.StatusInternalServerError, "服务器异常", nil)
			return
		}
	}

	response.NewResponse().Success(ctx, "更新成功", nil)
}

// GetById
// @Summary 获取字典
// @Description 获取指定id字典
// @Tags 系统工具/字典管理
// @Accept application/json
// @Produce application/json
// @Security BearerAuth
// @Param id path string true "id"
// @Success 200 {object} domainTools.Dict
// @Failure 400 {object} response.Response
// @Router /v1/tools/dict/getById/{id} [get]
// @Security LoginToken
func (h *dictHandler) GetById(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" || len(id) == 0 {
		response.NewResponse().Error(ctx, http.StatusBadRequest, "id不能为空", nil)
		return
	}

	detail, err := h.svc.GetById(ctx, id)
	if err != nil {
		if errors.Is(err, serviceTools.ErrDictNotFound) {
			response.NewResponse().Error(ctx, http.StatusBadRequest, "字典不存在", nil)
			return
		}
		ctx.Set("internalError", fmt.Sprintf("获取字典异常 >>> %v", err.Error()))
		zap.S().Error("获取字典异常 >>> ", err.Error())
		response.NewResponse().Error(ctx, http.StatusInternalServerError, "服务器异常", nil)
		return
	}

	response.NewResponse().Success(ctx, "获取成功", detail)
}

// GetListPage
// @Summary 获取字典分页列表
// @Description 获取字典分页列表
// @Tags 系统工具/字典管理
// @Accept application/json
// @Produce application/json
// @Security BearerAuth
// @Param page query int true "页码" default(1)
// @Param pageSize query int true "每页数量" default(10)
// @Param creator query string false "创建人"
// @Param modifier query string false "修改人"
// @Param status query bool true "状态" default(true)
// @Param name query string false "字典名称"
// @Param code query string false "字典编码"
// @Param type query int true "字典分类" default(0)
// @Param value_type query int true "数据类型" default(0)
// @Success 200 {object} DictListPageResponse
// @Failure 400 {object} response.Response
// @Router /v1/tools/dict/listPage [get]
// @Security LoginToken
func (h *dictHandler) GetListPage(ctx *gin.Context) {
	// 从上下文中获取登录信息
	claims, ok := ctx.MustGet("claims").(*jwt.Claims)
	if !ok {
		zap.S().Error("未找到用户认证信息 >>> ", zap.Error(errors.New(claims.UserID)))
		response.NewResponse().Error(ctx, http.StatusInternalServerError, "服务器异常", nil)
		return
	}

	user, err := h.userSvc.GetById(ctx, claims.UserID)
	if err != nil {
		ctx.Set("internalError", fmt.Sprintf("获取用户信息异常 >>> %v", err.Error()))
		zap.S().Error("获取用户信息异常 >>> ", err.Error())
		response.NewResponse().Error(ctx, http.StatusInternalServerError, "服务器异常", nil)
		return
	}

	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("pageSize", "10"))
	creator := ctx.DefaultQuery("creator", "")
	modifier := ctx.DefaultQuery("modifier", "")
	status, _ := strconv.ParseBool(ctx.DefaultQuery("status", "true"))

	name := ctx.DefaultQuery("name", "")
	code := ctx.DefaultQuery("code", "")
	dictType, _ := strconv.Atoi(ctx.DefaultQuery("type", "0"))
	valueType, _ := strconv.Atoi(ctx.DefaultQuery("value_type", "0"))

	filter := domainTools.DictFilter{
		Filters: filters.Filters{
			Creator:    creator,
			Modifier:   modifier,
			BelongDept: *user.DeptID,
		},
		Pagination: filters.Pagination{
			Page:     page,
			PageSize: pageSize,
		},
		Status:    status,
		Name:      name,
		Code:      code,
		Type:      dict.Type(dictType),
		ValueType: dict.ValueType(valueType),
	}

	list, total, err := h.svc.GetListPage(ctx, filter)
	if err != nil {
		ctx.Set("internalError", fmt.Sprintf("获取字典分页列表异常 >>> %v", err.Error()))
		zap.S().Error("获取字典分页列表异常 >>> ", err.Error())
		response.NewResponse().Error(ctx, http.StatusInternalServerError, "服务器异常", nil)
		return
	}

	response.NewResponse().Success(ctx, "查询成功", DictListPageResponse{
		List:     list,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}

// GetListAll
// @Summary 获取所有字典列表
// @Description 获取所有字典列表
// @Tags 系统工具/字典管理
// @Accept application/json
// @Produce application/json
// @Security BearerAuth
// @Param creator query string false "创建人"
// @Param modifier query string false "修改人"
// @Param status query bool true "状态" default(true)
// @Param name query string false "字典名称"
// @Param code query string false "字典编码"
// @Param type query int true "字典分类" default(0)
// @Param value_type query int true "数据类型" default(0)
// @Success 200 {array} []domainTools.Dict
// @Failure 400 {object} response.Response
// @Router /v1/tools/dict/listAll [get]
// @Security LoginToken
func (h *dictHandler) GetListAll(ctx *gin.Context) {
	// 从上下文中获取登录信息
	claims, ok := ctx.MustGet("claims").(*jwt.Claims)
	if !ok {
		zap.S().Error("未找到用户认证信息 >>> ", zap.Error(errors.New(claims.UserID)))
		response.NewResponse().Error(ctx, http.StatusInternalServerError, "服务器异常", nil)
		return
	}

	user, err := h.userSvc.GetById(ctx, claims.UserID)
	if err != nil {
		ctx.Set("internalError", fmt.Sprintf("获取用户信息异常 >>> %v", err.Error()))
		zap.S().Error("获取用户信息异常 >>> ", err.Error())
		response.NewResponse().Error(ctx, http.StatusInternalServerError, "服务器异常", nil)
		return
	}

	creator := ctx.DefaultQuery("creator", "")
	modifier := ctx.DefaultQuery("modifier", "")
	status, _ := strconv.ParseBool(ctx.DefaultQuery("status", "true"))

	name := ctx.DefaultQuery("name", "")
	code := ctx.DefaultQuery("code", "")
	dictType, _ := strconv.Atoi(ctx.DefaultQuery("type", "0"))
	valueType, _ := strconv.Atoi(ctx.DefaultQuery("value_type", "0"))

	filter := domainTools.DictFilter{
		Filters: filters.Filters{
			Creator:    creator,
			Modifier:   modifier,
			BelongDept: *user.DeptID,
		},
		Status:    status,
		Name:      name,
		Code:      code,
		Type:      dict.Type(dictType),
		ValueType: dict.ValueType(valueType),
	}

	list, err := h.svc.GetListAll(ctx, filter)
	if err != nil {
		ctx.Set("internalError", fmt.Sprintf("获取字典列表异常 >>> %v", err.Error()))
		zap.S().Error("获取字典列表异常 >>> ", err.Error())
		response.NewResponse().Error(ctx, http.StatusInternalServerError, "服务器异常", nil)
		return
	}

	response.NewResponse().Success(ctx, "查询成功", list)
}

// Export
// @Summary 导出字典
// @Description 导出字典到Excel文件
// @Tags 系统工具/字典管理
// @Accept application/json
// @Produce application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
// @Security BearerAuth
// @Param creator query string false "创建人"
// @Param modifier query string false "修改人"
// @Param status query bool true "状态" default(true)
// @Param name query string false "字典名称"
// @Param code query string false "字典编码"
// @Param type query int true "字典分类" default(0)
// @Param value_type query int true "数据类型" default(0)
// @Success 200 {file} file "Excel文件"
// @Failure 500 {object} response.Response
// @Router /v1/tools/dict/export [get]
// @Security LoginToken
func (h *dictHandler) Export(ctx *gin.Context) {
	// 从上下文中获取登录信息
	claims, ok := ctx.MustGet("claims").(*jwt.Claims)
	if !ok {
		zap.S().Error("未找到用户认证信息 >>> ", zap.Error(errors.New(claims.UserID)))
		response.NewResponse().Error(ctx, http.StatusInternalServerError, "服务器异常", nil)
		return
	}

	user, err := h.userSvc.GetById(ctx, claims.UserID)
	if err != nil {
		ctx.Set("internalError", fmt.Sprintf("获取用户信息异常 >>> %v", err.Error()))
		zap.S().Error("获取用户信息异常 >>> ", err.Error())
		response.NewResponse().Error(ctx, http.StatusInternalServerError, "服务器异常", nil)
		return
	}

	creator := ctx.DefaultQuery("creator", "")
	modifier := ctx.DefaultQuery("modifier", "")
	status, _ := strconv.ParseBool(ctx.DefaultQuery("status", "true"))

	name := ctx.DefaultQuery("name", "")
	code := ctx.DefaultQuery("code", "")
	dictType, _ := strconv.Atoi(ctx.DefaultQuery("type", "0"))
	valueType, _ := strconv.Atoi(ctx.DefaultQuery("value_type", "0"))

	filter := domainTools.DictFilter{
		Filters: filters.Filters{
			Creator:    creator,
			Modifier:   modifier,
			BelongDept: *user.DeptID,
		},
		Status:    status,
		Name:      name,
		Code:      code,
		Type:      dict.Type(dictType),
		ValueType: dict.ValueType(valueType),
	}

	list, err := h.svc.GetListAll(ctx, filter)
	if err != nil {
		ctx.Set("internalError", fmt.Sprintf("获取字典列表异常 >>> %v", err.Error()))
		zap.S().Error("获取字典列表异常 >>> ", err.Error())
		response.NewResponse().Error(ctx, http.StatusInternalServerError, "服务器异常", nil)
		return
	}

	// 准备导出配置
	filename := fmt.Sprintf("字典导出_%s.xlsx", time.Now().Format("20060102150405"))
	cfg := excelutil.ExcelExportConfig{
		SheetName:  "字典",
		FileName:   filename,
		StreamMode: true,
		Columns: []excelutil.ExcelColumn{
			{Title: "字典名称", Field: "Name", Width: 22},
			{Title: "字典编码", Field: "Code", Width: 22},
			{
				Title: "字典类型",
				Field: "Type",
				Width: 15,
				Formatter: func(value interface{}) string {
					typeValidValues := []string{"普通字典", "系统字典", "枚举字典"}
					converter := enumconv.NewEnumConverter(dict.TypeMapping, dict.TypeImportMapping, typeValidValues, "字典类型")
					str, _ := converter.FromEnum(value.(dict.Type))
					return str
				},
			},
			{
				Title: "数据类型",
				Field: "ValueType",
				Width: 15,
				Formatter: func(value interface{}) string {
					typeValueValidValues := []string{"字符串", "整型", "布尔"}
					typeValueConverter := enumconv.NewEnumConverter(dict.TypeValueMapping, dict.TypeValueImportMapping, typeValueValidValues, "数据类型")
					str, _ := typeValueConverter.FromEnum(value.(dict.ValueType))
					return str
				},
			},
			{Title: "字典描述", Field: "Description", Width: 30},
			{
				Title: "状态",
				Field: "Status",
				Width: 10,
				Formatter: func(value interface{}) string {
					if status, ok := value.(bool); ok {
						if status {
							return "启用"
						}
						return "停用"
					}
					return fmt.Sprintf("%v", value)
				},
			},
			{Title: "排序", Field: "Sort", Width: 8},
			{Title: "创建时间", Field: "CreateTime", Width: 22},
			{Title: "更新时间", Field: "UpdateTime", Width: 22},
			{Title: "备注", Field: "Remark", Width: 40},
		},
		Data: list,
	}

	// 创建并执行导出器
	exporter := excelutil.NewExcelExporter(&cfg)
	f, err := exporter.Export()
	if err != nil {
		ctx.Set("internalError", fmt.Sprintf("导出字典异常 >>> %v", err.Error()))
		zap.S().Error("导出字典异常 >>> ", err.Error())
		response.NewResponse().Error(ctx, http.StatusInternalServerError, "服务器异常", nil)
		return
	}

	// 设置响应头
	ctx.Header("Content-Type", "application/octet-stream")
	ctx.Header("Content-Disposition", "attachment; filename=export.xlsx")
	ctx.Header("Pragma", "no-cache")
	ctx.Header("Cache-Control", "no-store")

	// 流式写入响应
	if _, err := f.WriteTo(ctx.Writer); err != nil {
		response.NewResponse().Error(ctx, http.StatusInternalServerError, "生成Excel失败", nil)
	}
}
