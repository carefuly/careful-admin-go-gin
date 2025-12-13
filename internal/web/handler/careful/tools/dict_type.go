/**
 * Description：
 * FileName：dict_type.go
 * Author：CJiaの用心
 * Create：2025/10/19 15:28:27
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
	"github.com/carefuly/careful-admin-go-gin/pkg/constants/careful/tools/dict_type"
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
	"strconv"
	"time"
)

// CreateDictTypeRequest 创建
type CreateDictTypeRequest struct {
	Status    bool              `json:"status" binding:"omitempty" default:"true"`             // 状态【true-启用 false-停用】
	Name      string            `json:"name" binding:"required,max=50" default:""`             // 字典项名称
	StrValue  string            `json:"str_value" binding:"omitempty,max=50" default:""`       // 字符串-字典项值
	IntValue  int64             `json:"int_value" binding:"omitempty"`                         // 整型-字典项值
	BoolValue bool              `json:"bool_value" binding:"omitempty"`                        // 布尔-字典项值
	DictTag   dict_type.DictTag `json:"dict_tag" binding:"omitempty,max=10" default:"primary"` // 标签类型
	DictColor string            `json:"dict_color" binding:"omitempty,max=50" default:""`      // 标签颜色
	DictID    string            `json:"dict_id" binding:"required,max=100" default:""`         // 所属字典ID
	Sort      int               `json:"sort" binding:"omitempty" default:"1"`                  // 排序
	Remark    string            `json:"remark" binding:"omitempty,max=255" default:""`         // 备注
}

// ImportDictTypeRequest 导入
type ImportDictTypeRequest struct {
	File *multipart.FileHeader `form:"file" binding:"required"`
}

// UpdateDictTypeRequest 更新
type UpdateDictTypeRequest struct {
	Id        string            `json:"id" binding:"required" default:""`                      // 主键ID
	Status    bool              `json:"status" binding:"omitempty" default:"true"`             // 状态【true-启用 false-停用】
	Name      string            `json:"name" binding:"required,max=50" default:""`             // 字典项名称
	StrValue  string            `json:"str_value" binding:"omitempty,max=50" default:""`       // 字符串-字典项值
	IntValue  int64             `json:"int_value" binding:"omitempty"`                         // 整型-字典项值
	BoolValue bool              `json:"bool_value" binding:"omitempty"`                        // 布尔-字典项值
	DictTag   dict_type.DictTag `json:"dict_tag" binding:"omitempty,max=10" default:"primary"` // 标签类型
	DictColor string            `json:"dict_color" binding:"omitempty,max=50" default:""`      // 标签颜色
	DictID    string            `json:"dict_id" binding:"required,max=100" default:""`         // 所属字典ID
	Sort      int               `json:"sort" binding:"omitempty" default:"1"`                  // 排序
	Timestamp int64             `json:"timestamp" binding:"omitempty"`                         // 版本
	Remark    string            `json:"remark" binding:"omitempty,max=255" default:""`         // 备注
}

type ListByDictNamesRequest struct {
	DictNames []string `json:"dictNames"` // 数组参数格式: ?dictNames=性别&dictNames=计量单位
}

// DictTypeListPageResponse 列表分页响应
type DictTypeListPageResponse struct {
	List     []domainTools.DictType `json:"list"`     // 列表
	Total    int64                  `json:"total"`    // 总数
	Page     int                    `json:"page"`     // 页码
	PageSize int                    `json:"pageSize"` // 每页数量
}

type DictTypeHandler interface {
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

type dictTypeHandler struct {
	rely    config.RelyConfig
	svc     serviceTools.DictTypeService
	userSvc serviceSystem.UserService
}

func NewDictTypeHandler(rely config.RelyConfig, svc serviceTools.DictTypeService, userSvc serviceSystem.UserService) DictTypeHandler {
	return &dictTypeHandler{
		rely:    rely,
		svc:     svc,
		userSvc: userSvc,
	}
}

func (h *dictTypeHandler) RegisterRoutes(router *gin.RouterGroup) {
	base := router.Group("/dictType")
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
// @Summary 创建字典项
// @Description 创建字典项
// @Tags 系统工具/字典项管理
// @Accept application/json
// @Produce application/json
// @Security BearerAuth
// @Param CreateDictTypeRequest body CreateDictTypeRequest true "请求"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Router /v1/tools/dictType/create [post]
// @Security LoginToken
func (h *dictTypeHandler) Create(ctx *gin.Context) {
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

	var req CreateDictTypeRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		validate.NewValidatorErrorHandler(h.rely.Trans).Handle(ctx, err)
		return
	}

	// 校验参数
	dictTagValues := []string{"primary", "success", "warning", "danger", "info"}
	converter := enumconv.NewEnumConverter(dict_type.DictTagMapping, dict_type.DictTagImportMapping, dictTagValues, "标签类型")
	_, err = converter.FromEnum(req.DictTag)
	if err != nil {
		response.NewResponse().Error(ctx, http.StatusBadRequest, err.Error(), nil)
		return
	}

	// 转换为领域模型
	domain := domainTools.DictType{
		DictType: modelTools.DictType{
			CoreModels: models.CoreModels{
				Sort:       req.Sort,
				Creator:    user.Id,
				Modifier:   user.Id,
				BelongDept: user.DeptID,
				Remark:     req.Remark,
			},
			Status:    req.Status,
			Name:      req.Name,
			DictTag:   req.DictTag,
			DictColor: req.DictColor,
			DictID:    req.DictID,
		},
		StrValue:  req.StrValue,
		IntValue:  req.IntValue,
		BoolValue: req.BoolValue,
	}

	if err := h.svc.Create(ctx, domain); err != nil {
		switch {
		case errors.Is(err, serviceTools.ErrDictNotFound):
			response.NewResponse().Error(ctx, http.StatusBadRequest, "数据字典不存在", nil)
			return
		case errors.Is(err, serviceTools.ErrDictDisabled):
			response.NewResponse().Error(ctx, http.StatusBadRequest, "字典已被禁用，无法在其下创建字典项", nil)
			return
		case errors.Is(err, serviceTools.ErrDictTypeDuplicate):
			response.NewResponse().Error(ctx, http.StatusBadRequest, "同一字典下存在相同的字典项/值", nil)
			return
		case errors.Is(err, serviceTools.ErrDictTypeInvalidDictValueType):
			response.NewResponse().Error(ctx, http.StatusBadRequest, "无效的数据类型", nil)
			return
		default:
			ctx.Set("internalError", fmt.Sprintf("创建字典项异常 >>> %v", err.Error()))
			zap.S().Error("创建字典项异常 >>> ", err.Error())
			response.NewResponse().Error(ctx, http.StatusInternalServerError, "服务器异常", nil)
			return
		}
	}

	response.NewResponse().Success(ctx, "新增成功", nil)
}

// Import
// @Summary 导入字典项
// @Description 导入字典项
// @Tags 系统工具/字典项管理
// @Accept multipart/form-data
// @Produce application/json
// @Security BearerAuth
// @Param file formData file true "文件(支持xlsx/csv格式)"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Router /v1/tools/dictType/import [post]
// @Security LoginToken
func (h *dictTypeHandler) Import(ctx *gin.Context) {
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

	// 保存导入的文件信息
	format := time.Now().Format("2006-01-02")
	filePath := "./media/uploads/" + format + "/" + req.File.Filename
	if err := ctx.SaveUploadedFile(req.File, filePath); err != nil {
		response.NewResponse().Error(ctx, http.StatusBadRequest, "保存文件失败", nil)
		return
	}

	// 读取Excel文件
	read, err := xlsx.NewXlsxFile(filePath).ReadSheetByName("字典项模板")
	if err != nil {
		response.NewResponse().Error(ctx, http.StatusBadRequest, err.Error(), nil)
		return
	}

	result := h.svc.Import(ctx, user, read)

	response.NewResponse().Success(ctx, "导入成功", result)
}

// Delete
// @Summary 删除字典项
// @Description 删除指定id字典项
// @Tags 系统工具/字典项管理
// @Accept application/json
// @Produce application/json
// @Security BearerAuth
// @Param id path string true "id"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Router /v1/tools/dictType/delete/{id} [delete]
// @Security LoginToken
func (h *dictTypeHandler) Delete(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" || len(id) == 0 {
		response.NewResponse().Error(ctx, http.StatusBadRequest, "ID不能为空", nil)
		return
	}

	if err := h.svc.Delete(ctx, id); err != nil {
		ctx.Set("internalError", fmt.Sprintf("删除字典项异常 >>> %v", err.Error()))
		zap.S().Error("删除字典项异常 >>> ", zap.Error(err))
		response.NewResponse().Error(ctx, http.StatusInternalServerError, "服务器异常", nil)
		return
	}

	response.NewResponse().Success(ctx, "删除成功", nil)
}

// BatchDelete
// @Summary 批量删除字典项
// @Description 批量删除字典项
// @Tags 系统工具/字典项管理
// @Accept application/json
// @Produce application/json
// @Security BearerAuth
// @Param ids body []string true "id数组"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Router /v1/tools/dictType/batchDelete [post]
// @Security LoginToken
func (h *dictTypeHandler) BatchDelete(ctx *gin.Context) {
	var ids []string
	if err := ctx.ShouldBindJSON(&ids); err != nil {
		validate.NewValidatorErrorHandler(h.rely.Trans).Handle(ctx, err)
		return
	}

	err := h.svc.BatchDelete(ctx, ids)
	if err != nil {
		ctx.Set("internalError", fmt.Sprintf("批量删除字典项异常 >>> %v", err.Error()))
		zap.S().Error("批量删除字典项异常 >>> ", zap.Error(err))
		response.NewResponse().Error(ctx, http.StatusInternalServerError, "服务器异常", nil)
		return
	}

	response.NewResponse().Success(ctx, "批量删除成功", nil)
}

// Update
// @Summary 更新字典项
// @Description 更新字典项
// @Tags 系统工具/字典项管理
// @Accept application/json
// @Produce application/json
// @Security BearerAuth
// @Param UpdateDictTypeRequest body UpdateDictTypeRequest true "请求"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Router /v1/tools/dictType/update [put]
// @Security LoginToken
func (h *dictTypeHandler) Update(ctx *gin.Context) {
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

	var req UpdateDictTypeRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		validate.NewValidatorErrorHandler(h.rely.Trans).Handle(ctx, err)
		return
	}

	// 校验参数
	dictTagValues := []string{"primary", "success", "warning", "danger", "info"}
	converter := enumconv.NewEnumConverter(dict_type.DictTagMapping, dict_type.DictTagImportMapping, dictTagValues, "标签类型")
	_, err = converter.FromEnum(req.DictTag)
	if err != nil {
		response.NewResponse().Error(ctx, http.StatusBadRequest, err.Error(), nil)
		return
	}

	// 转换为领域模型
	domain := domainTools.DictType{
		DictType: modelTools.DictType{
			CoreModels: models.CoreModels{
				Id:         req.Id,
				Sort:       req.Sort,
				Timestamp:  req.Timestamp,
				Modifier:   user.Id,
				BelongDept: user.DeptID,
				Remark:     req.Remark,
			},
			Status:    req.Status,
			Name:      req.Name,
			DictTag:   req.DictTag,
			DictColor: req.DictColor,
			DictID:    req.DictID,
		},
		StrValue:  req.StrValue,
		IntValue:  req.IntValue,
		BoolValue: req.BoolValue,
	}

	if err := h.svc.Update(ctx, domain); err != nil {
		switch {
		case errors.Is(err, serviceTools.ErrDictNotFound):
			response.NewResponse().Error(ctx, http.StatusBadRequest, "数据字典不存在", nil)
			return
		case errors.Is(err, serviceTools.ErrDictDisabled):
			response.NewResponse().Error(ctx, http.StatusBadRequest, "字典已被禁用，无法在其下创建字典项", nil)
			return
		case errors.Is(err, serviceTools.ErrDictTypeDuplicate):
			response.NewResponse().Error(ctx, http.StatusBadRequest, "同一字典下存在相同的字典项/值", nil)
			return
		case errors.Is(err, serviceTools.ErrDictTypeInvalidDictValueType):
			response.NewResponse().Error(ctx, http.StatusBadRequest, "无效的数据类型", nil)
			return
		case errors.Is(err, serviceTools.ErrDictTypeVersionInconsistency):
			response.NewResponse().Error(ctx, http.StatusBadRequest, "数据版本不一致，取消修改，请刷新后重试", nil)
			return
		default:
			ctx.Set("internalError", fmt.Sprintf("更新字典项异常 >>> %v", err.Error()))
			zap.S().Error("更新字典项异常 >>> ", err.Error())
			response.NewResponse().Error(ctx, http.StatusInternalServerError, "服务器异常", nil)
			return
		}
	}

	response.NewResponse().Success(ctx, "更新成功", nil)
}

// GetById
// @Summary 获取字典项
// @Description 获取指定id字典项
// @Tags 系统工具/字典项管理
// @Accept application/json
// @Produce application/json
// @Security BearerAuth
// @Param id path string true "id"
// @Success 200 {object} domainTools.DictType
// @Failure 400 {object} response.Response
// @Router /v1/tools/dictType/getById/{id} [get]
// @Security LoginToken
func (h *dictTypeHandler) GetById(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" || len(id) == 0 {
		response.NewResponse().Error(ctx, http.StatusBadRequest, "id不能为空", nil)
		return
	}

	detail, err := h.svc.GetById(ctx, id)
	if err != nil {
		if errors.Is(err, serviceTools.ErrDictTypeNotFound) {
			response.NewResponse().Error(ctx, http.StatusBadRequest, "字典项不存在", nil)
			return
		}
		ctx.Set("internalError", fmt.Sprintf("获取字典项异常 >>> %v", err.Error()))
		zap.S().Error("获取字典项异常 >>> ", err.Error())
		response.NewResponse().Error(ctx, http.StatusInternalServerError, "服务器异常", nil)
		return
	}

	response.NewResponse().Success(ctx, "获取成功", detail)
}

// GetListPage
// @Summary 获取字典项分页列表
// @Description 获取字典项分页列表
// @Tags 系统工具/字典项管理
// @Accept application/json
// @Produce application/json
// @Security BearerAuth
// @Param page query int true "页码" default(1)
// @Param pageSize query int true "每页数量" default(10)
// @Param creator query string false "创建人"
// @Param modifier query string false "修改人"
// @Param status query bool false "状态" default(true)
// @Param name query string false "字典项名称"
// @Param dict_tag query string false "标签类型" default(primary)
// @Param dict_name query string false "字典名称"
// @Param value_type query int true "数据类型" default(0)
// @Param dict_id query string false "所属字典ID"
// @Success 200 {object} DictTypeListPageResponse
// @Failure 400 {object} response.Response
// @Router /v1/tools/dictType/listPage [get]
// @Security LoginToken
func (h *dictTypeHandler) GetListPage(ctx *gin.Context) {
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
	statusStr := ctx.DefaultQuery("status", "true")
	status, err := strconv.ParseBool(statusStr)
	if err != nil { // 空字符串、非法值都会触发错误，此时用默认值
		status = true
	}

	name := ctx.DefaultQuery("name", "")
	dictTag := ctx.DefaultQuery("dict_tag", "")
	dictName := ctx.DefaultQuery("dict_name", "")
	valueType, _ := strconv.Atoi(ctx.DefaultQuery("value_type", "0"))
	dictId := ctx.DefaultQuery("dict_id", "")

	filter := domainTools.DictTypeFilter{
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
		DictTag:   dict_type.DictTag(dictTag),
		DictName:  dictName,
		ValueType: dict.ValueType(valueType),
		DictID:    dictId,
	}

	list, total, err := h.svc.GetListPage(ctx, filter)
	if err != nil {
		ctx.Set("internalError", fmt.Sprintf("获取字典项分页列表异常 >>> %v", err.Error()))
		zap.S().Error("获取字典项分页列表异常 >>> ", err.Error())
		response.NewResponse().Error(ctx, http.StatusInternalServerError, "服务器异常", nil)
		return
	}

	response.NewResponse().Success(ctx, "查询成功", DictTypeListPageResponse{
		List:     list,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}

// GetListAll
// @Summary 获取所有字典项
// @Description 获取所有字典项列表
// @Tags 系统工具/字典项管理
// @Accept application/json
// @Produce application/json
// @Security BearerAuth
// @Param creator query string false "创建人"
// @Param modifier query string false "修改人"
// @Param status query bool false "状态" default(true)
// @Param name query string false "字典项名称"
// @Param dict_tag query string false "标签类型" default(primary)
// @Param dict_name query string false "字典名称"
// @Param value_type query int true "数据类型" default(0)
// @Param dict_id query string false "所属字典ID"
// @Success 200 {array} []domainTools.DictType
// @Failure 400 {object} response.Response
// @Router /v1/tools/dictType/listAll [get]
// @Security LoginToken
func (h *dictTypeHandler) GetListAll(ctx *gin.Context) {
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
	statusStr := ctx.DefaultQuery("status", "true")
	status, err := strconv.ParseBool(statusStr)
	if err != nil { // 空字符串、非法值都会触发错误，此时用默认值
		status = true
	}

	name := ctx.DefaultQuery("name", "")
	dictTag := ctx.DefaultQuery("dict_tag", "")
	dictName := ctx.DefaultQuery("dict_name", "")
	valueType, _ := strconv.Atoi(ctx.DefaultQuery("value_type", "0"))
	dictId := ctx.DefaultQuery("dict_id", "")

	filter := domainTools.DictTypeFilter{
		Filters: filters.Filters{
			Creator:    creator,
			Modifier:   modifier,
			BelongDept: *user.DeptID,
		},
		Status:    status,
		Name:      name,
		DictTag:   dict_type.DictTag(dictTag),
		DictName:  dictName,
		ValueType: dict.ValueType(valueType),
		DictID:    dictId,
	}

	list, err := h.svc.GetListAll(ctx, filter)
	if err != nil {
		ctx.Set("internalError", fmt.Sprintf("获取字典项列表异常 >>> %v", err.Error()))
		zap.S().Error("获取字典项列表异常 >>> ", err.Error())
		response.NewResponse().Error(ctx, http.StatusInternalServerError, "服务器异常", nil)
		return
	}

	response.NewResponse().Success(ctx, "查询成功", list)
}

// Export
// @Summary 导出字典项
// @Description 导出字典项到Excel文件
// @Tags 系统工具/字典项管理
// @Accept application/json
// @Produce application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
// @Security BearerAuth
// @Param creator query string false "创建人"
// @Param modifier query string false "修改人"
// @Param status query bool false "状态" default(true)
// @Param name query string false "字典项名称"
// @Param dict_tag query string false "标签类型" default(primary)
// @Param dict_name query string false "字典名称"
// @Param value_type query int true "数据类型" default(0)
// @Param dict_id query string false "所属字典ID"
// @Success 200 {file} file "Excel文件"
// @Failure 500 {object} response.Response
// @Router /v1/tools/dictType/export [get]
// @Security LoginToken
func (h *dictTypeHandler) Export(ctx *gin.Context) {
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
	statusStr := ctx.DefaultQuery("status", "true")
	status, err := strconv.ParseBool(statusStr)
	if err != nil { // 空字符串、非法值都会触发错误，此时用默认值
		status = true
	}

	name := ctx.DefaultQuery("name", "")
	dictTag := ctx.DefaultQuery("dict_tag", "")
	dictName := ctx.DefaultQuery("dict_name", "")
	valueType, _ := strconv.Atoi(ctx.DefaultQuery("value_type", "0"))
	dictId := ctx.DefaultQuery("dict_id", "")

	filter := domainTools.DictTypeFilter{
		Filters: filters.Filters{
			Creator:    creator,
			Modifier:   modifier,
			BelongDept: *user.DeptID,
		},
		Status:    status,
		Name:      name,
		DictTag:   dict_type.DictTag(dictTag),
		DictName:  dictName,
		ValueType: dict.ValueType(valueType),
		DictID:    dictId,
	}

	list, err := h.svc.GetListAll(ctx, filter)
	if err != nil {
		ctx.Set("internalError", fmt.Sprintf("获取字典项列表异常 >>> %v", err.Error()))
		zap.S().Error("获取字典项列表异常 >>> ", err.Error())
		response.NewResponse().Error(ctx, http.StatusInternalServerError, "服务器异常", nil)
		return
	}

	// 准备导出配置
	filename := fmt.Sprintf("字典项导出_%s.xlsx", time.Now().Format("20060102150405"))
	cfg := excelutil.ExcelExportConfig{
		SheetName:  "数据字典",
		FileName:   filename,
		StreamMode: true,
		Columns: []excelutil.ExcelColumn{
			{Title: "标签", Field: "Label", Width: 22},
			{Title: "值", Field: "Value", Width: 22},
			{
				Title: "标签类型",
				Field: "DictTag",
				Width: 15,
				Formatter: func(value interface{}) string {
					dictTagValues := []string{"primary", "success", "warning", "danger", "info"}
					converter := enumconv.NewEnumConverter(dict_type.DictTagMapping, dict_type.DictTagImportMapping, dictTagValues, "标签类型")
					str, _ := converter.FromEnum(value.(dict_type.DictTag))
					return str
				},
			},
			{Title: "标签颜色", Field: "DictColor", Width: 22},
			{Title: "字典名称", Field: "DictName", Width: 22},
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
		ctx.Set("internalError", fmt.Sprintf("导出字典项异常 >>> %v", err.Error()))
		zap.S().Error("导出字典项异常 >>> ", err.Error())
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
