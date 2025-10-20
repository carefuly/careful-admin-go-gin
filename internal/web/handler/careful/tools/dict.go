/**
 * Description：
 * FileName：dict.go
 * Author：CJiaの用心
 * Create：2025/10/11 14:03:05
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
	"github.com/carefuly/careful-admin-go-gin/pkg/utils/enumconv"
	"github.com/carefuly/careful-admin-go-gin/pkg/utils/excelutil"
	"github.com/carefuly/careful-admin-go-gin/pkg/utils/jwt"
	"github.com/carefuly/careful-admin-go-gin/pkg/utils/xlsx"
	"github.com/carefuly/careful-admin-go-gin/pkg/validate"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"mime/multipart"
	"net/http"
	"strconv"
	"time"
)

// CreateDictRequest 创建
type CreateDictRequest struct {
	Name      string              `json:"name" binding:"required,max=100"`           // 字典名称
	Code      string              `json:"code" binding:"required,max=100"`           // 字典编码
	Type      dict.TypeConst      `json:"type" binding:"omitempty" default:"1"`      // 字典分类
	ValueType dict.ValueTypeConst `json:"valueType" binding:"omitempty" default:"1"` // 字典值类型
	Sort      int                 `json:"sort" binding:"omitempty" default:"1"`      // 排序
	Status    bool                `json:"status" binding:"omitempty" default:"true"` // 状态【true-启用 false-停用】
	Remark    string              `json:"remark" binding:"omitempty,max=255"`        // 备注
}

// ImportDictRequest 导入
type ImportDictRequest struct {
	File *multipart.FileHeader `form:"file" binding:"required"`
}

// UpdateDictRequest 更新
type UpdateDictRequest struct {
	Id        string `json:"id" binding:"required"`                     // 主键ID
	Code      string `json:"code" binding:"required,max=100"`           // 字典编码
	Sort      int    `json:"sort" binding:"omitempty" default:"1"`      // 排序
	Status    bool   `json:"status" binding:"omitempty" default:"true"` // 状态【true-启用 false-停用】
	Timestamp int64  `json:"timestamp" binding:"omitempty"`             // 版本
	Remark    string `json:"remark" binding:"omitempty,max=255"`        // 备注
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
	base.POST("/delete/batchDelete", h.BatchDelete)
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
// @Param CreateDictRequest body CreateDictRequest true "请求"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Router /v1/tools/dict/create [post]
// @Security LoginToken
func (h *dictHandler) Create(ctx *gin.Context) {
	// 从上下文中获取登录信息
	claims, ok := ctx.MustGet("claims").(*jwt.Claims)
	if !ok {
		zap.S().Error("未找到用户认证信息 >>> ", zap.Error(errors.New(claims.UserId)))
		response.NewResponse().Error(ctx, http.StatusInternalServerError, "服务器异常", nil)
		return
	}

	user, err := h.userSvc.GetById(ctx, claims.UserId)
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
	converter := enumconv.NewEnumConverter(dict.TypeMapping, dict.TypeImportMapping, typeValidValues, "字典分类")
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
				BelongDept: user.DeptId,
				Remark:     req.Remark,
			},
			Status:    req.Status,
			Name:      req.Name,
			Code:      req.Code,
			Type:      req.Type,
			ValueType: req.ValueType,
		},
	}

	if err := h.svc.Create(ctx, domain); err != nil {
		switch {
		case errors.Is(err, serviceTools.ErrDictNameDuplicate):
			response.NewResponse().Error(ctx, http.StatusBadRequest, "字典名称已存在", nil)
			return
		case errors.Is(err, serviceTools.ErrDictCodeDuplicate):
			response.NewResponse().Error(ctx, http.StatusBadRequest, "字典编码已存在", nil)
			return
		case errors.Is(err, serviceTools.ErrDictDuplicate):
			response.NewResponse().Error(ctx, http.StatusBadRequest, "数据字典已存在", nil)
			return
		default:
			ctx.Set("internalError", fmt.Sprintf("创建数据字典异常 >>> %v", err.Error()))
			zap.S().Error("创建数据字典异常 >>> ", zap.Error(err))
			response.NewResponse().Error(ctx, http.StatusInternalServerError, "服务器异常", nil)
			return
		}
	}

	response.NewResponse().Success(ctx, "新增成功", nil)
}

// Import
// @Summary 导入字典
// @Description 导入字典
// @Tags 系统工具/字典管理
// @Accept multipart/form-data
// @Produce application/json
// @Param file formData file true "文件(支持xlsx/csv格式)"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Router /v1/tools/dict/import [post]
// @Security LoginToken
func (h *dictHandler) Import(ctx *gin.Context) {
	// 从上下文中获取登录信息
	claims, ok := ctx.MustGet("claims").(*jwt.Claims)
	if !ok {
		zap.S().Error("未找到用户认证信息 >>> ", zap.Error(errors.New(claims.UserId)))
		response.NewResponse().Error(ctx, http.StatusInternalServerError, "服务器异常", nil)
		return
	}

	user, err := h.userSvc.GetById(ctx, claims.UserId)
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

	// 保存导入的文件信息
	format := time.Now().Format("2006-01-02")
	filePath := "./uploads/" + format + "/" + req.File.Filename
	if err := ctx.SaveUploadedFile(req.File, filePath); err != nil {
		response.NewResponse().Error(ctx, http.StatusBadRequest, "保存文件失败", nil)
		return
	}

	// 读取Excel文件
	read, err := xlsx.NewXlsxFile(filePath).ReadSheetByName("字典模板")
	if err != nil {
		response.NewResponse().Error(ctx, http.StatusBadRequest, err.Error(), nil)
		return
	}

	result := h.svc.Import(ctx, user, read)
	msg := fmt.Sprintf("导入成功【成功导入【%d】条数据, 失败【%d】条数据】", result.SuccessCount, result.FailCount)

	response.NewResponse().Success(ctx, msg, result)
}

// Delete
// @Summary 删除字典
// @Description 删除指定id字典
// @Tags 系统工具/字典管理
// @Accept application/json
// @Produce application/json
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
		if errors.Is(err, serviceTools.ErrDictNotFound) {
			response.NewResponse().Error(ctx, http.StatusBadRequest, "数据字典不存在", nil)
			return
		}
		ctx.Set("internalError", fmt.Sprintf("删除数据字典失败 >>> %v", err.Error()))
		zap.S().Error("删除数据字典失败 >>> ", zap.Error(err))
		response.NewResponse().Error(ctx, http.StatusInternalServerError, "服务器异常", nil)
		return
	}

	response.NewResponse().Success(ctx, "删除成功", nil)
}

// BatchDelete
// @Summary 批量删除字典
// @Description 批量删除字典
// @Tags 系统工具/字典管理
// @Accept application/json
// @Produce application/json
// @Param ids body []string true "id数组"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Router /v1/tools/dict/delete/batchDelete [post]
// @Security LoginToken
func (h *dictHandler) BatchDelete(ctx *gin.Context) {
	var ids []string
	if err := ctx.ShouldBindJSON(&ids); err != nil {
		validate.NewValidatorErrorHandler(h.rely.Trans).Handle(ctx, err)
		return
	}

	err := h.svc.BatchDelete(ctx, ids)
	if err != nil {
		ctx.Set("internalError", fmt.Sprintf("批量删除字典异常 >>> %v", err.Error()))
		zap.S().Error("批量删除字典异常 >>> ", zap.Error(err))
		response.NewResponse().Error(ctx, http.StatusInternalServerError, "服务器异常", nil)
		return
	}

	response.NewResponse().Success(ctx, "批量删除成功", nil)
}

// Update
// @Summary 更新字典
// @Description 更新字典信息
// @Tags 系统工具/字典管理
// @Accept application/json
// @Produce application/json
// @Param UpdateDictRequest body UpdateDictRequest true "请求"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Router /v1/tools/dict/update [put]
// @Security LoginToken
func (h *dictHandler) Update(ctx *gin.Context) {
	// 从上下文中获取登录信息
	claims, ok := ctx.MustGet("claims").(*jwt.Claims)
	if !ok {
		zap.S().Error("未找到用户认证信息 >>> ", zap.Error(errors.New(claims.UserId)))
		response.NewResponse().Error(ctx, http.StatusInternalServerError, "服务器异常", nil)
		return
	}

	user, err := h.userSvc.GetById(ctx, claims.UserId)
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
				Modifier:   claims.UserId,
				BelongDept: user.DeptId,
				Remark:     req.Remark,
			},
			Status: req.Status,
			Code:   req.Code,
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
		case errors.Is(err, serviceTools.ErrDictDuplicate):
			response.NewResponse().Error(ctx, http.StatusBadRequest, "数据字典已存在", nil)
			return
		case errors.Is(err, serviceTools.ErrDictVersionInconsistency):
			response.NewResponse().Error(ctx, http.StatusBadRequest, "数据版本不一致，取消修改，请刷新后重试", nil)
			return
		default:
			ctx.Set("internalError", fmt.Sprintf("更新数据字典失败 >>> %v", err.Error()))
			zap.S().Error("更新数据字典失败 >>> ", err.Error())
			response.NewResponse().Error(ctx, http.StatusInternalServerError, "服务器异常", nil)
			return
		}
	}

	response.NewResponse().Success(ctx, "更新成功", nil)
}

// GetById
// @Summary 获取字典
// @Description 获取指定id字典信息
// @Tags 系统工具/字典管理
// @Accept application/json
// @Produce application/json
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
		ctx.Set("internalError", fmt.Sprintf("获取字典失败 >>> %v", err.Error()))
		zap.S().Error("获取字典失败 >>> ", err.Error())
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
// @Param page query int true "页码" default(1)
// @Param pageSize query int true "每页数量" default(10)
// @Param creator query string false "创建人"
// @Param modifier query string false "修改人"
// @Param status query bool false "状态" default(true)
// @Param name query string false "字典名称"
// @Param code query string false "字典编码"
// @Param type query int true "字典分类" default(0)
// @Param valueType query int true "字典值类型" default(0)
// @Success 200 {object} DictListPageResponse
// @Failure 400 {object} response.Response
// @Router /v1/tools/dict/listPage [get]
// @Security LoginToken
func (h *dictHandler) GetListPage(ctx *gin.Context) {
	// 从上下文中获取登录信息
	claims, ok := ctx.MustGet("claims").(*jwt.Claims)
	if !ok {
		zap.S().Error("未找到用户认证信息 >>> ", zap.Error(errors.New(claims.UserId)))
		response.NewResponse().Error(ctx, http.StatusInternalServerError, "服务器异常", nil)
		return
	}

	user, err := h.userSvc.GetById(ctx, claims.UserId)
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
	valueType, _ := strconv.Atoi(ctx.DefaultQuery("valueType", "0"))

	filter := domainTools.DictFilter{
		Pagination: filters.Pagination{
			Page:     page,
			PageSize: pageSize,
		},
		Filters: filters.Filters{
			Creator:    creator,
			Modifier:   modifier,
			BelongDept: user.DeptId,
		},
		Status:    status,
		Name:      name,
		Code:      code,
		Type:      dict.TypeConst(dictType),
		ValueType: dict.ValueTypeConst(valueType),
	}

	list, total, err := h.svc.GetListPage(ctx, filter)
	if err != nil {
		ctx.Set("internalError", fmt.Sprintf("获取数据字典分页列表异常 >>> %v", err.Error()))
		zap.S().Error("获取数据字典分页列表异常 >>> ", err.Error())
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
// @Summary 获取所有字典
// @Description 获取所有字典列表
// @Tags 系统工具/字典管理
// @Accept application/json
// @Produce application/json
// @Param creator query string false "创建人"
// @Param modifier query string false "修改人"
// @Param status query bool false "状态" default(true)
// @Param name query string false "字典名称"
// @Param code query string false "字典编码"
// @Param type query int true "字典分类" default(0)
// @Param valueType query int true "字典值类型" default(0)
// @Success 200 {array} []domainTools.Dict
// @Failure 400 {object} response.Response
// @Router /v1/tools/dict/listAll [get]
// @Security LoginToken
func (h *dictHandler) GetListAll(ctx *gin.Context) {
	// 从上下文中获取登录信息
	claims, ok := ctx.MustGet("claims").(*jwt.Claims)
	if !ok {
		zap.S().Error("未找到用户认证信息 >>> ", zap.Error(errors.New(claims.UserId)))
		response.NewResponse().Error(ctx, http.StatusInternalServerError, "服务器异常", nil)
		return
	}

	user, err := h.userSvc.GetById(ctx, claims.UserId)
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
	valueType, _ := strconv.Atoi(ctx.DefaultQuery("valueType", "0"))

	filter := domainTools.DictFilter{
		Filters: filters.Filters{
			Creator:    creator,
			Modifier:   modifier,
			BelongDept: user.DeptId,
		},
		Status:    status,
		Name:      name,
		Code:      code,
		Type:      dict.TypeConst(dictType),
		ValueType: dict.ValueTypeConst(valueType),
	}

	list, err := h.svc.GetListAll(ctx, filter)
	if err != nil {
		ctx.Set("internalError", fmt.Sprintf("获取数据字典列表异常 >>> %v", err.Error()))
		zap.S().Error("获取数据字典列表异常 >>> ", err.Error())
		response.NewResponse().Error(ctx, http.StatusInternalServerError, "服务器异常", nil)
		return
	}

	response.NewResponse().Success(ctx, "查询成功", list)
}

// Export
// @Summary 导出字典数据
// @Description 导出字典数据到Excel文件
// @Tags 系统工具/字典管理
// @Accept application/json
// @Produce application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
// @Param creator query string false "创建人"
// @Param modifier query string false "修改人"
// @Param status query bool false "状态" default(true)
// @Param name query string false "字典名称"
// @Param code query string false "字典编码"
// @Param type query int true "字典分类" default(1)
// @Param valueType query int true "字典值类型" default(1)
// @Success 200 {file} file "Excel文件"
// @Failure 500 {object} response.Response
// @Router /v1/tools/dict/export [get]
// @Security LoginToken
func (h *dictHandler) Export(ctx *gin.Context) {
	// 从上下文中获取登录信息
	claims, ok := ctx.MustGet("claims").(*jwt.Claims)
	if !ok {
		zap.S().Error("未找到用户认证信息 >>> ", zap.Error(errors.New(claims.UserId)))
		response.NewResponse().Error(ctx, http.StatusInternalServerError, "服务器异常", nil)
		return
	}

	user, err := h.userSvc.GetById(ctx, claims.UserId)
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
	valueType, _ := strconv.Atoi(ctx.DefaultQuery("valueType", "0"))

	filter := domainTools.DictFilter{
		Filters: filters.Filters{
			Creator:    creator,
			Modifier:   modifier,
			BelongDept: user.DeptId,
		},
		Status:    status,
		Name:      name,
		Code:      code,
		Type:      dict.TypeConst(dictType),
		ValueType: dict.ValueTypeConst(valueType),
	}

	list, err := h.svc.GetListAll(ctx, filter)
	if err != nil {
		ctx.Set("internalError", fmt.Sprintf("获取数据字典列表异常 >>> %v", err.Error()))
		zap.S().Error("获取数据字典列表异常 >>> ", err.Error())
		response.NewResponse().Error(ctx, http.StatusInternalServerError, "服务器异常", nil)
		return
	}

	// 准备导出配置
	filename := fmt.Sprintf("数据字典导出_%s.xlsx", time.Now().Format("20060102150405"))
	cfg := excelutil.ExcelExportConfig{
		SheetName:  "数据字典",
		FileName:   filename,
		StreamMode: true,
		Columns: []excelutil.ExcelColumn{
			{Title: "字典名称", Field: "Name", Width: 22},
			{Title: "字典编码", Field: "Code", Width: 17},
			{
				Title: "字典编码",
				Field: "Type",
				Width: 15,
				Formatter: func(value interface{}) string {
					typeValidValues := []string{"普通字典", "系统字典", "枚举字典"}
					converter := enumconv.NewEnumConverter(dict.TypeMapping, dict.TypeImportMapping, typeValidValues, "字典分类")
					str, _ := converter.FromEnum(value.(dict.TypeConst))
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
					str, _ := typeValueConverter.FromEnum(value.(dict.ValueTypeConst))
					return str
				},
			},
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
		ctx.Set("internalError", fmt.Sprintf("导出数据字典异常 >>> %v", err.Error()))
		zap.S().Error("导出数据字典异常 >>> ", err.Error())
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
