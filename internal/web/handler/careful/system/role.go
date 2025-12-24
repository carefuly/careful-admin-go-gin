/**
 * Description：
 * FileName：role.go
 * Author：CJiaの用心
 * Create：2025/12/19 22:22:12
 * Remark：
 */

package system

import (
	"errors"
	"fmt"
	"github.com/carefuly/careful-admin-go-gin/config"
	domainSystem "github.com/carefuly/careful-admin-go-gin/internal/domain/careful/system"
	modelSystem "github.com/carefuly/careful-admin-go-gin/internal/model/careful/system"
	serviceSystem "github.com/carefuly/careful-admin-go-gin/internal/service/careful/system"
	"github.com/carefuly/careful-admin-go-gin/pkg/constants/careful/system/role"
	"github.com/carefuly/careful-admin-go-gin/pkg/ginx/filters"
	"github.com/carefuly/careful-admin-go-gin/pkg/ginx/response"
	"github.com/carefuly/careful-admin-go-gin/pkg/models"
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

// CreateRoleRequest 创建
type CreateRoleRequest struct {
	Status      bool           `json:"status" binding:"omitempty" default:"true"`     // 状态【true-启用 false-停用】
	Name        string         `json:"name" binding:"required,max=64" default:""`     // 角色名称
	Code        string         `json:"code" binding:"required,max=64" default:""`     // 角色编码
	DataScope   role.DataScope `json:"data_scope" binding:"omitempty" default:"1"`    // 数据权限范围
	Description string         `json:"description" binding:"omitempty" default:""`    // 角色描述
	Sort        int            `json:"sort" binding:"omitempty" default:"1"`          // 排序
	Remark      string         `json:"remark" binding:"omitempty,max=255" default:""` // 备注
}

// ImportRoleRequest 导入
type ImportRoleRequest struct {
	File *multipart.FileHeader `form:"file" binding:"required"` // 文件
}

// UpdateRoleRequest 更新
type UpdateRoleRequest struct {
	Id            string         `json:"id" binding:"required" default:""`                 // 主键ID
	Status        bool           `json:"status" binding:"omitempty" default:"true"`        // 状态【true-启用 false-停用】
	Name          string         `json:"name" binding:"required,max=64" default:""`        // 角色名称
	Code          string         `json:"code" binding:"required,max=64" default:""`        // 角色编码
	DataScope     role.DataScope `json:"data_scope" binding:"omitempty" default:"1"`       // 数据权限范围
	Description   string         `json:"description" binding:"omitempty" default:""`       // 角色描述
	DeptIDs       []string       `json:"dept_ids" binding:"omitempty" default:"[]"`        // 部门ID数组
	MenuIDs       []string       `json:"menu_ids" binding:"omitempty" default:"[]"`        // 菜单ID数组
	MenuButtonIDs []string       `json:"menu_button_ids" binding:"omitempty" default:"[]"` // 按钮ID数组
	Sort          int            `json:"sort" binding:"omitempty" default:"1"`             // 排序
	Timestamp     int64          `json:"timestamp" binding:"omitempty"`                    // 版本
	Remark        string         `json:"remark" binding:"omitempty,max=255" default:""`    // 备注
}

// RoleListPageResponse 列表分页响应
type RoleListPageResponse struct {
	List     []domainSystem.Role `json:"list"`     // 列表
	Total    int64               `json:"total"`    // 总数
	Page     int                 `json:"page"`     // 页码
	PageSize int                 `json:"pageSize"` // 每页数量
}

type RoleHandler interface {
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

type roleHandler struct {
	rely    config.RelyConfig
	svc     serviceSystem.RoleService
	userSvc serviceSystem.UserService
}

func NewRoleHandler(rely config.RelyConfig, svc serviceSystem.RoleService, userSvc serviceSystem.UserService) RoleHandler {
	return &roleHandler{
		rely:    rely,
		svc:     svc,
		userSvc: userSvc,
	}
}

func (h *roleHandler) RegisterRoutes(router *gin.RouterGroup) {
	base := router.Group("/role")
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
// @Summary 创建角色
// @Description 创建角色
// @Tags 系统管理/角色管理
// @Accept application/json
// @Produce application/json
// @Security BearerAuth
// @Param CreateRoleRequest body CreateRoleRequest true "请求"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Router /v1/system/role/create [post]
// @Security LoginToken
func (h *roleHandler) Create(ctx *gin.Context) {
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

	var req CreateRoleRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		validate.NewValidatorErrorHandler(h.rely.Trans).Handle(ctx, err)
		return
	}

	// 校验参数
	dataScopeValidValues := []string{"仅本人数据", "本部门数据", "本部门及下级部门数据", "全部数据", "自定义数据"}
	converter := enumconv.NewEnumConverter(role.DataScopeMapping, role.DataScopeImportMapping, dataScopeValidValues, "数据权限范围")
	_, err = converter.FromEnum(req.DataScope)
	if err != nil {
		response.NewResponse().Error(ctx, http.StatusBadRequest, err.Error(), nil)
		return
	}

	// 转换为领域模型
	domain := domainSystem.Role{
		Role: modelSystem.Role{
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
			DataScope:   req.DataScope,
			Description: req.Description,
		},
	}

	if err := h.svc.Create(ctx, domain); err != nil {
		switch {
		case errors.Is(err, serviceSystem.ErrRoleCodeDuplicate):
			response.NewResponse().Error(ctx, http.StatusBadRequest, "角色编码已存在", nil)
			return
		default:
			ctx.Set("internalError", fmt.Sprintf("创建角色异常 >>> %v", err.Error()))
			zap.S().Error("创建角色异常 >>> ", err.Error())
			response.NewResponse().Error(ctx, http.StatusInternalServerError, "服务器异常", nil)
			return
		}
	}

	response.NewResponse().Success(ctx, "新增成功", nil)
}

// Import
// @Summary 导入角色
// @Description 导入角色
// @Tags 系统管理/角色管理
// @Accept multipart/form-data
// @Produce application/json
// @Security BearerAuth
// @Param file formData file true "文件(支持xlsx/csv格式)"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Router /v1/system/role/import [post]
// @Security LoginToken
func (h *roleHandler) Import(ctx *gin.Context) {
	// response.NewResponse().Success(ctx, "导入成功", nil)
}

// Delete
// @Summary 删除角色
// @Description 删除指定id角色
// @Tags 系统管理/角色管理
// @Accept application/json
// @Produce application/json
// @Security BearerAuth
// @Param id path string true "id"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Router /v1/system/role/delete/{id} [delete]
// @Security LoginToken
func (h *roleHandler) Delete(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" || len(id) == 0 {
		response.NewResponse().Error(ctx, http.StatusBadRequest, "ID不能为空", nil)
		return
	}

	if err := h.svc.Delete(ctx, id); err != nil {
		switch {
		default:
			ctx.Set("internalError", fmt.Sprintf("删除角色异常 >>> %v", err.Error()))
			zap.S().Error("删除角色异常 >>> ", zap.Error(err))
			response.NewResponse().Error(ctx, http.StatusInternalServerError, "服务器异常", nil)
			return
		}
	}

	response.NewResponse().Success(ctx, "删除成功", nil)
}

// BatchDelete
// @Summary 批量删除角色
// @Description 批量删除角色
// @Tags 系统管理/角色管理
// @Accept application/json
// @Produce application/json
// @Security BearerAuth
// @Param ids body []string true "id数组"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Router /v1/system/role/batchDelete [post]
// @Security LoginToken
func (h *roleHandler) BatchDelete(ctx *gin.Context) {
	var ids []string
	if err := ctx.ShouldBindJSON(&ids); err != nil {
		validate.NewValidatorErrorHandler(h.rely.Trans).Handle(ctx, err)
		return
	}

	err := h.svc.BatchDelete(ctx, ids)
	if err != nil {
		ctx.Set("internalError", fmt.Sprintf("批量删除角色异常 >>> %v", err.Error()))
		zap.S().Error("批量删除角色异常 >>> ", zap.Error(err))
		response.NewResponse().Error(ctx, http.StatusInternalServerError, "服务器异常", nil)
		return
	}

	response.NewResponse().Success(ctx, "批量删除成功", nil)
}

// Update
// @Summary 更新角色
// @Description 更新角色
// @Tags 系统管理/角色管理
// @Accept application/json
// @Produce application/json
// @Security BearerAuth
// @Param UpdateRoleRequest body UpdateRoleRequest true "请求"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Router /v1/system/role/update [put]
// @Security LoginToken
func (h *roleHandler) Update(ctx *gin.Context) {
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

	var req UpdateRoleRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		validate.NewValidatorErrorHandler(h.rely.Trans).Handle(ctx, err)
		return
	}

	// 校验参数
	dataScopeValidValues := []string{"仅本人数据", "本部门数据", "本部门及下级部门数据", "全部数据", "自定义数据"}
	converter := enumconv.NewEnumConverter(role.DataScopeMapping, role.DataScopeImportMapping, dataScopeValidValues, "数据权限范围")
	_, err = converter.FromEnum(req.DataScope)
	if err != nil {
		response.NewResponse().Error(ctx, http.StatusBadRequest, err.Error(), nil)
		return
	}

	// 转换为领域模型
	domain := domainSystem.Role{
		Role: modelSystem.Role{
			CoreModels: models.CoreModels{
				Id:         req.Id,
				Sort:       req.Sort,
				Timestamp:  req.Timestamp,
				Modifier:   user.Id,
				BelongDept: user.DeptID,
				Remark:     req.Remark,
			},
			Status:      req.Status,
			Name:        req.Name,
			Code:        req.Code,
			DataScope:   req.DataScope,
			Description: req.Description,
		},
		DeptIDs:       req.DeptIDs,
		MenuIDs:       req.MenuIDs,
		MenuButtonIDs: req.MenuButtonIDs,
	}

	if err := h.svc.Update(ctx, domain); err != nil {
		switch {
		case errors.Is(err, serviceSystem.ErrPostCodeDuplicate):
			response.NewResponse().Error(ctx, http.StatusBadRequest, "角色编码已存在", nil)
			return
		case errors.Is(err, serviceSystem.ErrPostVersionInconsistency):
			response.NewResponse().Error(ctx, http.StatusBadRequest, "数据版本不一致，取消修改，请刷新后重试", nil)
			return
		default:
			ctx.Set("internalError", fmt.Sprintf("更新角色异常 >>> %v", err.Error()))
			zap.S().Error("更新角色异常 >>> ", err.Error())
			response.NewResponse().Error(ctx, http.StatusInternalServerError, "服务器异常", nil)
			return
		}
	}

	response.NewResponse().Success(ctx, "更新成功", nil)
}

// GetById
// @Summary 获取角色
// @Description 获取指定id角色
// @Tags 系统管理/角色管理
// @Accept application/json
// @Produce application/json
// @Security BearerAuth
// @Param id path string true "id"
// @Success 200 {object} domainSystem.Role
// @Failure 400 {object} response.Response
// @Router /v1/system/role/getById/{id} [get]
// @Security LoginToken
func (h *roleHandler) GetById(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" || len(id) == 0 {
		response.NewResponse().Error(ctx, http.StatusBadRequest, "id不能为空", nil)
		return
	}

	detail, err := h.svc.GetById(ctx, id)
	if err != nil {
		if errors.Is(err, serviceSystem.ErrPostNotFound) {
			response.NewResponse().Error(ctx, http.StatusBadRequest, "角色不存在", nil)
			return
		}
		ctx.Set("internalError", fmt.Sprintf("获取角色异常 >>> %v", err.Error()))
		zap.S().Error("获取角色异常 >>> ", err.Error())
		response.NewResponse().Error(ctx, http.StatusInternalServerError, "服务器异常", nil)
		return
	}

	response.NewResponse().Success(ctx, "获取成功", detail)
}

// GetListPage
// @Summary 获取角色分页列表
// @Description 获取角色分页列表
// @Tags 系统管理/角色管理
// @Accept application/json
// @Produce application/json
// @Security BearerAuth
// @Param page query int true "页码" default(1)
// @Param pageSize query int true "每页数量" default(10)
// @Param creator query string false "创建人"
// @Param modifier query string false "修改人"
// @Param status query bool false "状态" default(true)
// @Param name query string false "角色名称"
// @Param code query string false "角色编码"
// @Param data_scope query int true "数据权限范围" default(0)
// @Success 200 {array} []domainSystem.Role
// @Success 200 {object} RoleListPageResponse
// @Failure 400 {object} response.Response
// @Router /v1/system/role/listPage [get]
// @Security LoginToken
func (h *roleHandler) GetListPage(ctx *gin.Context) {
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
	code := ctx.DefaultQuery("code", "")
	dataScope, _ := strconv.Atoi(ctx.DefaultQuery("data_scope", "0"))

	filter := domainSystem.RoleFilter{
		Pagination: filters.Pagination{
			Page:     page,
			PageSize: pageSize,
		},
		Filters: filters.Filters{
			Creator:    creator,
			Modifier:   modifier,
			BelongDept: *user.DeptID,
		},
		Status:    status,
		Name:      name,
		Code:      code,
		DataScope: role.DataScope(dataScope),
	}

	list, total, err := h.svc.GetListPage(ctx, filter)
	if err != nil {
		ctx.Set("internalError", fmt.Sprintf("获取角色分页列表异常 >>> %v", err.Error()))
		zap.S().Error("获取角色分页列表异常 >>> ", err.Error())
		response.NewResponse().Error(ctx, http.StatusInternalServerError, "服务器异常", nil)
		return
	}

	response.NewResponse().Success(ctx, "查询成功", RoleListPageResponse{
		List:     list,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}

// GetListAll
// @Summary 获取所有角色
// @Description 获取所有角色列表
// @Tags 系统管理/角色管理
// @Accept application/json
// @Produce application/json
// @Security BearerAuth
// @Param creator query string false "创建人"
// @Param modifier query string false "修改人"
// @Param status query bool false "状态" default(true)
// @Param name query string false "角色名称"
// @Param code query string false "角色编码"
// @Param data_scope query int true "数据权限范围" default(0)
// @Success 200 {array} []domainSystem.Role
// @Failure 400 {object} response.Response
// @Router /v1/system/role/listAll [get]
// @Security LoginToken
func (h *roleHandler) GetListAll(ctx *gin.Context) {
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
	code := ctx.DefaultQuery("code", "")
	dataScope, _ := strconv.Atoi(ctx.DefaultQuery("data_scope", "0"))

	filter := domainSystem.RoleFilter{
		Filters: filters.Filters{
			Creator:    creator,
			Modifier:   modifier,
			BelongDept: *user.DeptID,
		},
		Status:    status,
		Name:      name,
		Code:      code,
		DataScope: role.DataScope(dataScope),
	}

	list, err := h.svc.GetListAll(ctx, filter)
	if err != nil {
		ctx.Set("internalError", fmt.Sprintf("获取角色列表异常 >>> %v", err.Error()))
		zap.S().Error("获取角色列表异常 >>> ", err.Error())
		response.NewResponse().Error(ctx, http.StatusInternalServerError, "服务器异常", nil)
		return
	}

	response.NewResponse().Success(ctx, "查询成功", list)
}

// Export
// @Summary 导出角色
// @Description 导出角色到Excel文件
// @Tags 系统管理/角色管理
// @Accept application/json
// @Produce application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
// @Security BearerAuth
// @Param creator query string false "创建人"
// @Param modifier query string false "修改人"
// @Param status query bool false "状态" default(true)
// @Param name query string false "角色名称"
// @Param code query string false "角色编码"
// @Param data_scope query int true "数据权限范围" default(0)
// @Success 200 {file} file "Excel文件"
// @Failure 500 {object} response.Response
// @Router /v1/system/role/export [get]
// @Security LoginToken
func (h *roleHandler) Export(ctx *gin.Context) {
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
	code := ctx.DefaultQuery("code", "")
	dataScope, _ := strconv.Atoi(ctx.DefaultQuery("data_scope", "0"))

	filter := domainSystem.RoleFilter{
		Filters: filters.Filters{
			Creator:    creator,
			Modifier:   modifier,
			BelongDept: *user.DeptID,
		},
		Status:    status,
		Name:      name,
		Code:      code,
		DataScope: role.DataScope(dataScope),
	}

	list, err := h.svc.GetListAll(ctx, filter)
	if err != nil {
		ctx.Set("internalError", fmt.Sprintf("获取角色列表异常 >>> %v", err.Error()))
		zap.S().Error("获取角色列表异常 >>> ", err.Error())
		response.NewResponse().Error(ctx, http.StatusInternalServerError, "服务器异常", nil)
		return
	}

	// 准备导出配置
	filename := fmt.Sprintf("角色导出_%s.xlsx", time.Now().Format("20060102150405"))
	cfg := excelutil.ExcelExportConfig{
		SheetName:  "角色",
		FileName:   filename,
		StreamMode: true,
		Columns: []excelutil.ExcelColumn{
			{Title: "角色名称", Field: "Name", Width: 22},
			{Title: "角色编码", Field: "Code", Width: 22},
			{
				Title: "数据权限范围",
				Field: "DataScope",
				Width: 15,
				Formatter: func(value interface{}) string {
					dataScopeValidValues := []string{"仅本人数据", "本部门数据", "本部门及下级部门数据", "全部数据", "自定义数据"}
					converter := enumconv.NewEnumConverter(role.DataScopeMapping, role.DataScopeImportMapping, dataScopeValidValues, "数据权限范围")
					str, _ := converter.FromEnum(value.(role.DataScope))
					return str
				},
			},
			{Title: "角色描述", Field: "Description", Width: 22},
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
		ctx.Set("internalError", fmt.Sprintf("导出角色异常 >>> %v", err.Error()))
		zap.S().Error("导出角色异常 >>> ", err.Error())
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
