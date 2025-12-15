/**
 * Description：
 * FileName：dept.go
 * Author：CJiaの用心
 * Create：2025/11/26 09:43:00
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
	"github.com/carefuly/careful-admin-go-gin/pkg/constants/careful/system/dept"
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

// CreateDeptRequest 创建
type CreateDeptRequest struct {
	Status      bool      `json:"status" binding:"omitempty" default:"true"`                 // 状态【true-启用 false-停用】
	Name        string    `json:"name" binding:"required,max=50" default:""`                 // 部门名称
	Code        string    `json:"code" binding:"required,max=50" default:""`                 // 部门编码
	DeptType    dept.Type `json:"dept_type" binding:"omitempty,max=20" default:"department"` // 部门类型
	Owner       string    `json:"owner" binding:"omitempty,max=64" default:""`               // 部门负责人
	Phone       string    `json:"phone" binding:"omitempty,max=64" default:""`               // 部门电话
	Email       string    `json:"email" binding:"omitempty,max=64" default:""`               // 部门邮箱
	Description string    `json:"description" binding:"omitempty" default:""`                // 部门描述
	ParentID    *string   `json:"parent_id" binding:"omitempty,max=110" default:""`          // 父部门ID
	Sort        int       `json:"sort" binding:"omitempty" default:"1"`                      // 排序
	Remark      string    `json:"remark" binding:"omitempty,max=255" default:""`             // 备注
}

// ImportDeptRequest 导入
type ImportDeptRequest struct {
	File *multipart.FileHeader `form:"file" binding:"required"` // 文件
}

// UpdateDeptRequest 更新
type UpdateDeptRequest struct {
	Id          string    `json:"id" binding:"required" default:""`                          // 主键ID
	Status      bool      `json:"status" binding:"omitempty" default:"true"`                 // 状态【true-启用 false-停用】
	Name        string    `json:"name" binding:"required,max=50" default:""`                 // 部门名称
	Code        string    `json:"code" binding:"required,max=50" default:""`                 // 部门编码
	DeptType    dept.Type `json:"dept_type" binding:"omitempty,max=20" default:"department"` // 部门类型
	Owner       string    `json:"owner" binding:"omitempty,max=64" default:""`               // 部门负责人
	Phone       string    `json:"phone" binding:"omitempty,max=64" default:""`               // 部门电话
	Email       string    `json:"email" binding:"omitempty,max=64" default:""`               // 部门邮箱
	Description string    `json:"description" binding:"omitempty" default:""`                // 部门描述
	ParentID    *string   `json:"parent_id" binding:"omitempty,max=110" default:""`          // 父部门ID
	Sort        int       `json:"sort" binding:"omitempty" default:"1"`                      // 排序
	Timestamp   int64     `json:"timestamp" binding:"omitempty"`                             // 版本
	Remark      string    `json:"remark" binding:"omitempty,max=255" default:""`             // 备注
}

type DeptHandler interface {
	RegisterRoutes(router *gin.RouterGroup)
	Create(ctx *gin.Context)
	Import(ctx *gin.Context)
	Delete(ctx *gin.Context)
	BatchDelete(ctx *gin.Context)
	Update(ctx *gin.Context)
	GetById(ctx *gin.Context)
	GetListTree(ctx *gin.Context)
	GetListAll(ctx *gin.Context)
	Export(ctx *gin.Context)
}

type deptHandler struct {
	rely    config.RelyConfig
	svc     serviceSystem.DeptService
	userSvc serviceSystem.UserService
}

func NewDeptHandler(rely config.RelyConfig, svc serviceSystem.DeptService, userSvc serviceSystem.UserService) DeptHandler {
	return &deptHandler{
		rely:    rely,
		svc:     svc,
		userSvc: userSvc,
	}
}

func (h *deptHandler) RegisterRoutes(router *gin.RouterGroup) {
	base := router.Group("/dept")
	base.POST("/create", h.Create)
	base.POST("/import", h.Import)
	base.DELETE("/delete/:id", h.Delete)
	base.POST("/batchDelete", h.BatchDelete)
	base.PUT("/update", h.Update)
	base.GET("/getById/:id", h.GetById)
	base.GET("/listTree", h.GetListTree)
	base.GET("/listAll", h.GetListAll)
	base.GET("/export", h.Export)
}

// Create
// @Summary 创建部门
// @Description 创建部门
// @Tags 系统管理/部门管理
// @Accept application/json
// @Produce application/json
// @Security BearerAuth
// @Param CreateDeptRequest body CreateDeptRequest true "请求"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Router /v1/system/dept/create [post]
// @Security LoginToken
func (h *deptHandler) Create(ctx *gin.Context) {
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

	var req CreateDeptRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		validate.NewValidatorErrorHandler(h.rely.Trans).Handle(ctx, err)
		return
	}

	// 校验参数
	typeValidValues := []string{"company", "department", "team", "other"}
	converter := enumconv.NewEnumConverter(dept.TypeMapping, dept.TypeImportMapping, typeValidValues, "部门类型")
	_, err = converter.FromEnum(req.DeptType)
	if err != nil {
		response.NewResponse().Error(ctx, http.StatusBadRequest, err.Error(), nil)
		return
	}

	// 转换为领域模型
	domain := domainSystem.Dept{
		Dept: modelSystem.Dept{
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
			DeptType:    req.DeptType,
			Owner:       req.Owner,
			Phone:       req.Phone,
			Email:       req.Email,
			Description: req.Description,
			ParentID:    req.ParentID,
		},
	}

	if err := h.svc.Create(ctx, domain); err != nil {
		switch {
		case errors.Is(err, serviceSystem.ErrDeptCodeDuplicate):
			response.NewResponse().Error(ctx, http.StatusBadRequest, "部门编码已存在", nil)
			return
		case errors.Is(err, serviceSystem.ErrDeptNameParentDuplicate):
			response.NewResponse().Error(ctx, http.StatusBadRequest, "同级别下已存在相同的部门信息", nil)
			return
		case errors.Is(err, serviceSystem.ErrDeptParentNotFound):
			response.NewResponse().Error(ctx, http.StatusBadRequest, "父部门不存在", nil)
			return
		case errors.Is(err, serviceSystem.ErrDeptDisabled):
			response.NewResponse().Error(ctx, http.StatusBadRequest, "父部门已被禁用，无法在其下创建子部门", nil)
			return
		default:
			ctx.Set("internalError", fmt.Sprintf("创建部门异常 >>> %v", err.Error()))
			zap.S().Error("创建部门异常 >>> ", err.Error())
			response.NewResponse().Error(ctx, http.StatusInternalServerError, "服务器异常", nil)
			return
		}
	}

	response.NewResponse().Success(ctx, "新增成功", nil)
}

// Import
// @Summary 导入部门
// @Description 导入部门
// @Tags 系统管理/部门管理
// @Accept multipart/form-data
// @Produce application/json
// @Security BearerAuth
// @Param file formData file true "文件(支持xlsx/csv格式)"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Router /v1/system/dept/import [post]
// @Security LoginToken
func (h *deptHandler) Import(ctx *gin.Context) {
	// TODO implement me
	response.NewResponse().Success(ctx, "接口已预留，暂未开放使用", nil)
}

// Delete
// @Summary 删除部门
// @Description 删除指定id部门
// @Tags 系统管理/部门管理
// @Accept application/json
// @Produce application/json
// @Security BearerAuth
// @Param id path string true "id"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Router /v1/system/dept/delete/{id} [delete]
// @Security LoginToken
func (h *deptHandler) Delete(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" || len(id) == 0 {
		response.NewResponse().Error(ctx, http.StatusBadRequest, "ID不能为空", nil)
		return
	}

	if err := h.svc.Delete(ctx, id); err != nil {
		switch {
		case errors.Is(err, serviceSystem.ErrDeptHasChildren):
			response.NewResponse().Error(ctx, http.StatusBadRequest, "部门含有子部门，无法删除", nil)
			return
		case errors.Is(err, serviceSystem.ErrDeptHasUsers):
			response.NewResponse().Error(ctx, http.StatusBadRequest, "部门下仍有用户，无法删除", nil)
			return
		default:
			ctx.Set("internalError", fmt.Sprintf("删除部门异常 >>> %v", err.Error()))
			zap.S().Error("删除部门异常 >>> ", zap.Error(err))
			response.NewResponse().Error(ctx, http.StatusInternalServerError, "服务器异常", nil)
			return
		}
	}

	response.NewResponse().Success(ctx, "删除成功", nil)
}

// BatchDelete
// @Summary 批量删除部门
// @Description 批量删除部门
// @Tags 系统管理/部门管理
// @Accept application/json
// @Produce application/json
// @Security BearerAuth
// @Param ids body []string true "id数组"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Router /v1/system/dept/batchDelete [post]
// @Security LoginToken
func (h *deptHandler) BatchDelete(ctx *gin.Context) {
	var ids []string
	if err := ctx.ShouldBindJSON(&ids); err != nil {
		validate.NewValidatorErrorHandler(h.rely.Trans).Handle(ctx, err)
		return
	}

	err := h.svc.BatchDelete(ctx, ids)
	if err != nil {
		ctx.Set("internalError", fmt.Sprintf("批量删除部门异常 >>> %v", err.Error()))
		zap.S().Error("批量删除部门异常 >>> ", zap.Error(err))
		response.NewResponse().Error(ctx, http.StatusInternalServerError, "服务器异常", nil)
		return
	}

	response.NewResponse().Success(ctx, "批量删除成功", nil)
}

// Update
// @Summary 更新部门
// @Description 更新部门
// @Tags 系统管理/部门管理
// @Accept application/json
// @Produce application/json
// @Security BearerAuth
// @Param UpdateDeptRequest body UpdateDeptRequest true "请求"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Router /v1/system/dept/update [put]
// @Security LoginToken
func (h *deptHandler) Update(ctx *gin.Context) {
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

	var req UpdateDeptRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		validate.NewValidatorErrorHandler(h.rely.Trans).Handle(ctx, err)
		return
	}

	// 校验参数
	typeValidValues := []string{"company", "department", "team", "other"}
	converter := enumconv.NewEnumConverter(dept.TypeMapping, dept.TypeImportMapping, typeValidValues, "部门类型")
	_, err = converter.FromEnum(req.DeptType)
	if err != nil {
		response.NewResponse().Error(ctx, http.StatusBadRequest, err.Error(), nil)
		return
	}

	// 转换为领域模型
	domain := domainSystem.Dept{
		Dept: modelSystem.Dept{
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
			DeptType:    req.DeptType,
			Owner:       req.Owner,
			Phone:       req.Phone,
			Email:       req.Email,
			Description: req.Description,
			ParentID:    req.ParentID,
		},
	}

	if err := h.svc.Update(ctx, domain); err != nil {
		switch {
		case errors.Is(err, serviceSystem.ErrDeptCodeDuplicate):
			response.NewResponse().Error(ctx, http.StatusBadRequest, "部门编码已存在", nil)
			return
		case errors.Is(err, serviceSystem.ErrDeptNameParentDuplicate):
			response.NewResponse().Error(ctx, http.StatusBadRequest, "同级别下已存在相同的部门信息", nil)
			return
		case errors.Is(err, serviceSystem.ErrDeptYourParent):
			response.NewResponse().Error(ctx, http.StatusBadRequest, "不能将自己设置为父部门", nil)
			return
		case errors.Is(err, serviceSystem.ErrDeptParentNotFound):
			response.NewResponse().Error(ctx, http.StatusBadRequest, "父部门信息不存在", nil)
			return
		case errors.Is(err, serviceSystem.ErrDeptDisabled):
			response.NewResponse().Error(ctx, http.StatusBadRequest, "父部门已被禁用，无法在其下创建子部门", nil)
			return
		case errors.Is(err, serviceSystem.ErrDeptCycleReference):
			response.NewResponse().Error(ctx, http.StatusBadRequest, "不能将子部门设置为父部门，会形成循环引用", nil)
			return
		case errors.Is(err, serviceSystem.ErrDeptVersionInconsistency):
			response.NewResponse().Error(ctx, http.StatusBadRequest, "数据版本不一致，取消修改，请刷新后重试", nil)
			return
		default:
			ctx.Set("internalError", fmt.Sprintf("更新部门异常 >>> %v", err.Error()))
			zap.S().Error("更新部门异常 >>> ", err.Error())
			response.NewResponse().Error(ctx, http.StatusInternalServerError, "服务器异常", nil)
			return
		}
	}

	response.NewResponse().Success(ctx, "更新成功", nil)
}

// GetById
// @Summary 获取部门
// @Description 获取指定id部门
// @Tags 系统管理/部门管理
// @Accept application/json
// @Produce application/json
// @Security BearerAuth
// @Param id path string true "id"
// @Success 200 {object} domainSystem.Dept
// @Failure 400 {object} response.Response
// @Router /v1/system/dept/getById/{id} [get]
// @Security LoginToken
func (h *deptHandler) GetById(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" || len(id) == 0 {
		response.NewResponse().Error(ctx, http.StatusBadRequest, "id不能为空", nil)
		return
	}

	detail, err := h.svc.GetById(ctx, id)
	if err != nil {
		if errors.Is(err, serviceSystem.ErrDeptNotFound) {
			response.NewResponse().Error(ctx, http.StatusBadRequest, "部门不存在", nil)
			return
		}
		ctx.Set("internalError", fmt.Sprintf("获取部门异常 >>> %v", err.Error()))
		zap.S().Error("获取部门异常 >>> ", err.Error())
		response.NewResponse().Error(ctx, http.StatusInternalServerError, "服务器异常", nil)
		return
	}

	response.NewResponse().Success(ctx, "获取成功", detail)
}

// GetListTree
// @Summary 获取部门树
// @Description 获取部门树
// @Tags 系统管理/部门管理
// @Accept application/json
// @Produce application/json
// @Security BearerAuth
// @Param creator query string false "创建人"
// @Param modifier query string false "修改人"
// @Param status query bool false "状态" default(true)
// @Param name query string false "部门名称"
// @Param code query string false "部门编码"
// @Param dept_type query string false "部门类型"
// @Param level query int true "层级深度" default(0)
// @Success 200 {object} serviceSystem.DeptTree
// @Failure 400 {object} response.Response
// @Router /v1/system/dept/listTree [get]
// @Security LoginToken
func (h *deptHandler) GetListTree(ctx *gin.Context) {
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
	deptType := ctx.DefaultQuery("dept_type", "")
	level, _ := strconv.Atoi(ctx.DefaultQuery("level", "0"))

	filter := domainSystem.DeptFilter{
		Filters: filters.Filters{
			Creator:    creator,
			Modifier:   modifier,
			BelongDept: *user.DeptID,
		},
		Status:   status,
		Name:     name,
		Code:     code,
		DeptType: dept.Type(deptType),
		Level:    level,
	}

	list, err := h.svc.GetListTree(ctx, filter)
	if err != nil {
		ctx.Set("internalError", fmt.Sprintf("获取部门树异常 >>> %v", err.Error()))
		zap.S().Error("获取部门树异常 >>> ", err.Error())
		response.NewResponse().Error(ctx, http.StatusInternalServerError, "服务器异常", nil)
		return
	}

	response.NewResponse().Success(ctx, "查询成功", list)
}

// GetListAll
// @Summary 获取所有部门列表
// @Description 获取所有部门列表
// @Tags 系统管理/部门管理
// @Accept application/json
// @Produce application/json
// @Security BearerAuth
// @Param creator query string false "创建人"
// @Param modifier query string false "修改人"
// @Param status query bool false "状态" default(true)
// @Param name query string false "部门名称"
// @Param code query string false "部门编码"
// @Param dept_type query string false "部门类型"
// @Param level query int true "层级深度" default(0)
// @Success 200 {array} []domainSystem.Dept
// @Failure 400 {object} response.Response
// @Router /v1/system/dept/listAll [get]
// @Security LoginToken
func (h *deptHandler) GetListAll(ctx *gin.Context) {
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
	deptType := ctx.DefaultQuery("dept_type", "")
	level, _ := strconv.Atoi(ctx.DefaultQuery("level", "0"))

	filter := domainSystem.DeptFilter{
		Filters: filters.Filters{
			Creator:    creator,
			Modifier:   modifier,
			BelongDept: *user.DeptID,
		},
		Status:   status,
		Name:     name,
		Code:     code,
		DeptType: dept.Type(deptType),
		Level:    level,
	}

	list, err := h.svc.GetListAll(ctx, filter)
	if err != nil {
		ctx.Set("internalError", fmt.Sprintf("获取部门列表异常 >>> %v", err.Error()))
		zap.S().Error("获取部门列表异常 >>> ", err.Error())
		response.NewResponse().Error(ctx, http.StatusInternalServerError, "服务器异常", nil)
		return
	}

	response.NewResponse().Success(ctx, "查询成功", list)
}

// Export
// @Summary 导出部门
// @Description 导出部门到Excel文件
// @Tags 系统管理/部门管理
// @Accept application/json
// @Produce application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
// @Security BearerAuth
// @Param creator query string false "创建人"
// @Param modifier query string false "修改人"
// @Param status query bool false "状态" default(true)
// @Param name query string false "部门名称"
// @Param code query string false "部门编码"
// @Param dept_type query string false "部门类型"
// @Param level query int true "层级深度" default(0)
// @Success 200 {file} file "Excel文件"
// @Failure 500 {object} response.Response
// @Router /v1/system/post/export [get]
// @Security LoginToken
func (h *deptHandler) Export(ctx *gin.Context) {
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
	deptType := ctx.DefaultQuery("dept_type", "")
	level, _ := strconv.Atoi(ctx.DefaultQuery("level", "0"))

	filter := domainSystem.DeptFilter{
		Filters: filters.Filters{
			Creator:    creator,
			Modifier:   modifier,
			BelongDept: *user.DeptID,
		},
		Status:   status,
		Name:     name,
		Code:     code,
		DeptType: dept.Type(deptType),
		Level:    level,
	}

	list, err := h.svc.GetListAll(ctx, filter)
	if err != nil {
		ctx.Set("internalError", fmt.Sprintf("获取部门列表异常 >>> %v", err.Error()))
		zap.S().Error("获取部门列表异常 >>> ", err.Error())
		response.NewResponse().Error(ctx, http.StatusInternalServerError, "服务器异常", nil)
		return
	}

	// 准备导出配置
	filename := fmt.Sprintf("部门导出_%s.xlsx", time.Now().Format("20060102150405"))
	cfg := excelutil.ExcelExportConfig{
		SheetName:  "部门",
		FileName:   filename,
		StreamMode: true,
		Columns: []excelutil.ExcelColumn{
			{Title: "部门名称", Field: "Name", Width: 22},
			{Title: "部门编码", Field: "Code", Width: 22},
			{
				Title: "部门类型",
				Field: "DeptType",
				Width: 15,
				Formatter: func(value interface{}) string {
					typeValidValues := []string{"company", "department", "team", "other"}
					converter := enumconv.NewEnumConverter(dept.TypeMapping, dept.TypeImportMapping, typeValidValues, "部门类型")
					str, _ := converter.FromEnum(value.(dept.Type))
					return str
				},
			},
			{Title: "部门负责人", Field: "Owner", Width: 22},
			{Title: "部门电话", Field: "Phone", Width: 22},
			{Title: "部门邮箱", Field: "Email", Width: 22},
			{Title: "部门描述", Field: "Description", Width: 22},
			{Title: "层级深度", Field: "Level", Width: 22},
			{Title: "父部门名称", Field: "Parent.Name", Width: 22},
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
		ctx.Set("internalError", fmt.Sprintf("导出部门异常 >>> %v", err.Error()))
		zap.S().Error("导出部门异常 >>> ", err.Error())
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
