/**
 * Description：
 * FileName：menu.go
 * Author：CJiaの用心
 * Create：2025/12/05 15:04:45
 * Remark：
 */

package system

import (
	"errors"
	"fmt"
	"github.com/carefuly/careful-admin-go-gin/config"
	domainSystem "github.com/carefuly/careful-admin-go-gin/internal/domain/careful/system"
	"github.com/carefuly/careful-admin-go-gin/internal/model/careful/system"
	serviceSystem "github.com/carefuly/careful-admin-go-gin/internal/service/careful/system"
	"github.com/carefuly/careful-admin-go-gin/pkg/ginx/filters"
	"github.com/carefuly/careful-admin-go-gin/pkg/ginx/response"
	"github.com/carefuly/careful-admin-go-gin/pkg/models"
	"github.com/carefuly/careful-admin-go-gin/pkg/utils/excelutil"
	"github.com/carefuly/careful-admin-go-gin/pkg/utils/jwt"
	"github.com/carefuly/careful-admin-go-gin/pkg/validate"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"net/http"
	"strconv"
	"time"
)

// CreateMenuRequest 创建
type CreateMenuRequest struct {
	Status         bool    `json:"status" binding:"omitempty" default:"true"`             // 状态【true-启用 false-停用】
	Name           string  `json:"name" binding:"required,max=64" default:""`             // 菜单名称
	Path           string  `json:"path" binding:"required,max=128" default:""`            // 路由地址
	Component      string  `json:"component" binding:"required,max=128" default:""`       // 组件地址
	Title          string  `json:"title" binding:"required,max=64" default:""`            // 路由标题
	Icon           string  `json:"icon" binding:"omitempty,max=64" default:""`            // 路由图标
	ShowBadge      bool    `json:"show_badge" binding:"omitempty" default:"false"`        // 是否显示徽章
	ShowTextBadge  string  `json:"show_text_badge" binding:"omitempty,max=64" default:""` // 文本徽章
	IsHide         bool    `json:"is_hide" binding:"omitempty" default:"false"`           // 是否在菜单中隐藏
	IsHideTab      bool    `json:"is_hide_tab" binding:"omitempty" default:"false"`       // 是否在标签页中隐藏
	Link           string  `json:"link" binding:"omitempty,max=255" default:""`           // 外部链接【不填写默认没有外链】
	IsIframe       bool    `json:"is_iframe" binding:"omitempty" default:"false"`         // 是否为iframe
	KeepAlive      bool    `json:"keep_alive" binding:"omitempty" default:"false"`        // 是否缓存页面
	IsFirstLevel   bool    `json:"is_first_level" binding:"omitempty" default:"false"`    // 是否为一级菜单
	FixedTab       bool    `json:"fixed_tab" binding:"omitempty" default:"false"`         // 是否固定标签页
	ActivePath     string  `json:"active_path" binding:"omitempty,max=128" default:""`    // 激活菜单路径
	IsFullPage     bool    `json:"is_full_page" binding:"omitempty" default:"false"`      // 是否为全屏页面
	IsAuthButton   bool    `json:"is_auth_button" binding:"omitempty" default:"false"`    // 是否为权限按钮行
	AuthMark       string  `json:"auth_mark" binding:"omitempty,max=128" default:""`      // 权限标识
	IsCreateButton bool    `json:"is_create_button" binding:"omitempty" default:"false"`  // 是否自动创建按钮
	ParentID       *string `json:"parent_id" binding:"omitempty,max=110" default:""`      // 上级菜单
	Sort           int     `json:"sort" binding:"omitempty" default:"1"`                  // 排序
	Remark         string  `json:"remark" binding:"omitempty,max=255" default:""`         // 备注
}

// UpdateMenuRequest 更新
type UpdateMenuRequest struct {
	Id             string  `json:"id" binding:"required" default:""`                      // 主键ID
	Status         bool    `json:"status" binding:"omitempty" default:"true"`             // 状态【true-启用 false-停用】
	Name           string  `json:"name" binding:"required,max=64" default:""`             // 菜单名称
	Path           string  `json:"path" binding:"required,max=128" default:""`            // 路由地址
	Component      string  `json:"component" binding:"required,max=128" default:""`       // 组件地址
	Title          string  `json:"title" binding:"required,max=64" default:""`            // 路由标题
	Icon           string  `json:"icon" binding:"omitempty,max=64" default:""`            // 路由图标
	ShowBadge      bool    `json:"show_badge" binding:"omitempty" default:"false"`        // 是否显示徽章
	ShowTextBadge  string  `json:"show_text_badge" binding:"omitempty,max=64" default:""` // 文本徽章
	IsHide         bool    `json:"is_hide" binding:"omitempty" default:"false"`           // 是否在菜单中隐藏
	IsHideTab      bool    `json:"is_hide_tab" binding:"omitempty" default:"false"`       // 是否在标签页中隐藏
	Link           string  `json:"link" binding:"omitempty,max=255" default:""`           // 外部链接【不填写默认没有外链】
	IsIframe       bool    `json:"is_iframe" binding:"omitempty" default:"false"`         // 是否为iframe
	KeepAlive      bool    `json:"keep_alive" binding:"omitempty" default:"false"`        // 是否缓存页面
	IsFirstLevel   bool    `json:"is_first_level" binding:"omitempty" default:"false"`    // 是否为一级菜单
	FixedTab       bool    `json:"fixed_tab" binding:"omitempty" default:"false"`         // 是否固定标签页
	ActivePath     string  `json:"active_path" binding:"omitempty,max=128" default:""`    // 激活菜单路径
	IsFullPage     bool    `json:"is_full_page" binding:"omitempty" default:"false"`      // 是否为全屏页面
	IsAuthButton   bool    `json:"is_auth_button" binding:"omitempty" default:"false"`    // 是否为权限按钮行
	AuthMark       string  `json:"auth_mark" binding:"omitempty,max=128" default:""`      // 权限标识
	IsCreateButton bool    `json:"is_create_button" binding:"omitempty" default:"false"`  // 是否自动创建按钮
	ParentID       *string `json:"parent_id" binding:"omitempty,max=110" default:""`      // 上级菜单
	Sort           int     `json:"sort" binding:"omitempty" default:"1"`                  // 排序
	Timestamp      int64   `json:"timestamp" binding:"omitempty"`                         // 版本
	Remark         string  `json:"remark" binding:"omitempty,max=255"`                    // 备注
}

// MenuListPageResponse 列表分页响应
type MenuListPageResponse struct {
	List     []domainSystem.Menu `json:"list"`     // 列表
	Total    int64               `json:"total"`    // 总数
	Page     int                 `json:"page"`     // 页码
	PageSize int                 `json:"pageSize"` // 每页数量
}

type MenuHandler interface {
	RegisterRoutes(router *gin.RouterGroup)
	Create(ctx *gin.Context)
	Delete(ctx *gin.Context)
	BatchDelete(ctx *gin.Context)
	Update(ctx *gin.Context)
	GetById(ctx *gin.Context)
	GetMenuRouteTree(ctx *gin.Context)
	GetListAll(ctx *gin.Context)
	Export(ctx *gin.Context)
	// GetMenuTree(ctx *gin.Context)
}

type menuHandler struct {
	rely    config.RelyConfig
	svc     serviceSystem.MenuService
	userSvc serviceSystem.UserService
}

func NewMenuHandler(rely config.RelyConfig, svc serviceSystem.MenuService, userSvc serviceSystem.UserService) MenuHandler {
	return &menuHandler{
		rely:    rely,
		svc:     svc,
		userSvc: userSvc,
	}
}

// RegisterRoutes 注册路由
func (h *menuHandler) RegisterRoutes(router *gin.RouterGroup) {
	base := router.Group("/menu")
	base.POST("/create", h.Create)
	base.DELETE("/delete/:id", h.Delete)
	base.POST("/batchDelete", h.BatchDelete)
	base.PUT("/update", h.Update)
	base.GET("/getById/:id", h.GetById)
	base.GET("/listRouteTree", h.GetMenuRouteTree)
	base.GET("/listAll", h.GetListAll)
	base.GET("/export", h.Export)
	// base.GET("/listTree", h.GetMenuTree)
}

// Create
// @Summary 创建菜单
// @Description 创建菜单信息
// @Tags 系统管理/菜单管理
// @Accept application/json
// @Produce application/json
// @Security BearerAuth
// @Param CreateMenuRequest body CreateMenuRequest true "请求"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Router /v1/system/menu/create [post]
// @Security LoginToken
func (h *menuHandler) Create(ctx *gin.Context) {
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

	var req CreateMenuRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		validate.NewValidatorErrorHandler(h.rely.Trans).Handle(ctx, err)
		return
	}

	// 转换为领域模型
	domain := domainSystem.Menu{
		Menu: system.Menu{
			CoreModels: models.CoreModels{
				Sort:       req.Sort,
				Creator:    user.Id,
				Modifier:   user.Id,
				BelongDept: user.DeptID,
				Remark:     req.Remark,
			},
			Status:        req.Status,
			Name:          req.Name,
			Path:          req.Path,
			Component:     req.Component,
			Title:         req.Title,
			Icon:          req.Icon,
			ShowBadge:     req.ShowBadge,
			ShowTextBadge: req.ShowTextBadge,
			IsHide:        req.IsHide,
			IsHideTab:     req.IsHideTab,
			Link:          req.Link,
			IsIframe:      req.IsIframe,
			KeepAlive:     req.KeepAlive,
			IsFirstLevel:  req.IsFirstLevel,
			FixedTab:      req.FixedTab,
			ActivePath:    req.ActivePath,
			IsFullPage:    req.IsFullPage,
			IsAuthButton:  req.IsAuthButton,
			AuthMark:      req.AuthMark,
			ParentID:      req.ParentID,
		},
	}

	if err := h.svc.Create(ctx, domain, req.IsCreateButton, user); err != nil {
		switch {
		case errors.Is(err, serviceSystem.ErrMenuDuplicate):
			response.NewResponse().Error(ctx, http.StatusBadRequest, "同级别下已存在相同的菜单信息", nil)
			return
		case errors.Is(err, serviceSystem.ErrMenuParentNotFound):
			response.NewResponse().Error(ctx, http.StatusBadRequest, "上级菜单不存在", nil)
			return
		case errors.Is(err, serviceSystem.ErrMenuDisabled):
			response.NewResponse().Error(ctx, http.StatusBadRequest, "上级菜单已被禁用，无法在其下创建子菜单", nil)
			return
		default:
			ctx.Set("internalError", fmt.Sprintf("创建菜单异常 >>> %v", err.Error()))
			zap.S().Error("创建菜单异常 >>> ", zap.Error(err))
			response.NewResponse().Error(ctx, http.StatusInternalServerError, "服务器异常", nil)
			return
		}
	}

	response.NewResponse().Success(ctx, "新增成功", nil)
}

// Delete
// @Summary 删除菜单
// @Description 删除指定id菜单信息
// @Tags 系统管理/菜单管理
// @Accept application/json
// @Produce application/json
// @Security BearerAuth
// @Param id path string true "id"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Router /v1/system/menu/delete/{id} [delete]
// @Security LoginToken
func (h *menuHandler) Delete(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" || len(id) == 0 {
		response.NewResponse().Error(ctx, http.StatusBadRequest, "ID不能为空", nil)
		return
	}

	if err := h.svc.Delete(ctx, id); err != nil {
		if errors.Is(err, serviceSystem.ErrMenuHasChildren) {
			response.NewResponse().Error(ctx, http.StatusBadRequest, "菜单含有子菜单，无法删除", nil)
			return
		}
		ctx.Set("internalError", fmt.Sprintf("删除菜单异常 >>> %v", err.Error()))
		zap.S().Error("删除菜单异常 >>> ", zap.Error(err))
		response.NewResponse().Error(ctx, http.StatusInternalServerError, "服务器异常", nil)
		return
	}

	response.NewResponse().Success(ctx, "删除成功", nil)
}

// BatchDelete
// @Summary 批量删除菜单
// @Description 批量删除菜单信息
// @Tags 系统管理/菜单管理
// @Accept application/json
// @Produce application/json
// @Security BearerAuth
// @Param ids body []string true "id数组"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Router /v1/system/menu/batchDelete [post]
// @Security LoginToken
func (h *menuHandler) BatchDelete(ctx *gin.Context) {
	var ids []string
	if err := ctx.ShouldBindJSON(&ids); err != nil {
		validate.NewValidatorErrorHandler(h.rely.Trans).Handle(ctx, err)
		return
	}

	err := h.svc.BatchDelete(ctx, ids)
	if err != nil {
		ctx.Set("internalError", fmt.Sprintf("批量删除菜单异常 >>> %v", err.Error()))
		zap.S().Error("批量删除菜单异常 >>> ", zap.Error(err))
		response.NewResponse().Error(ctx, http.StatusInternalServerError, "服务器异常", nil)
		return
	}

	response.NewResponse().Success(ctx, "批量删除成功", nil)
}

// Update
// @Summary 更新菜单
// @Description 更新菜单信息
// @Tags 系统管理/菜单管理
// @Accept application/json
// @Produce application/json
// @Security BearerAuth
// @Param UpdateMenuRequest body UpdateMenuRequest true "请求"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Router /v1/system/menu/update [put]
// @Security LoginToken
func (h *menuHandler) Update(ctx *gin.Context) {
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

	var req UpdateMenuRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		validate.NewValidatorErrorHandler(h.rely.Trans).Handle(ctx, err)
		return
	}

	// 转换为领域模型
	domain := domainSystem.Menu{
		Menu: system.Menu{
			CoreModels: models.CoreModels{
				Id:         req.Id,
				Sort:       req.Sort,
				Timestamp:  req.Timestamp,
				Modifier:   user.Id,
				BelongDept: user.DeptID,
				Remark:     req.Remark,
			},
			Status:        req.Status,
			Name:          req.Name,
			Path:          req.Path,
			Component:     req.Component,
			Title:         req.Title,
			Icon:          req.Icon,
			ShowBadge:     req.ShowBadge,
			ShowTextBadge: req.ShowTextBadge,
			IsHide:        req.IsHide,
			IsHideTab:     req.IsHideTab,
			Link:          req.Link,
			IsIframe:      req.IsIframe,
			KeepAlive:     req.KeepAlive,
			IsFirstLevel:  req.IsFirstLevel,
			FixedTab:      req.FixedTab,
			ActivePath:    req.ActivePath,
			IsFullPage:    req.IsFullPage,
			IsAuthButton:  req.IsAuthButton,
			AuthMark:      req.AuthMark,
			ParentID:      req.ParentID,
		},
	}

	if err := h.svc.Update(ctx, domain); err != nil {
		switch {
		case errors.Is(err, serviceSystem.ErrMenuDuplicate):
			response.NewResponse().Error(ctx, http.StatusBadRequest, "同级别下已存在相同的菜单信息", nil)
			return
		case errors.Is(err, serviceSystem.ErrMenuParentNotFound):
			response.NewResponse().Error(ctx, http.StatusBadRequest, "上级菜单不存在", nil)
			return
		case errors.Is(err, serviceSystem.ErrMenuDisabled):
			response.NewResponse().Error(ctx, http.StatusBadRequest, "上级菜单已被禁用，无法在其下创建子菜单", nil)
			return
		case errors.Is(err, serviceSystem.ErrMenuVersionInconsistency):
			response.NewResponse().Error(ctx, http.StatusBadRequest, "数据版本不一致，取消修改，请刷新后重试", nil)
			return
		default:
			ctx.Set("internalError", fmt.Sprintf("更新菜单异常 >>> %v", err.Error()))
			zap.S().Error("更新菜单异常 >>> ", err.Error())
			response.NewResponse().Error(ctx, http.StatusInternalServerError, "服务器异常", nil)
			return
		}
	}

	response.NewResponse().Success(ctx, "更新成功", nil)
}

// GetById
// @Summary 获取菜单
// @Description 获取指定id菜单信息
// @Tags 系统管理/菜单管理
// @Accept application/json
// @Produce application/json
// @Security BearerAuth
// @Param id path string true "id"
// @Success 200 {object} domainSystem.Menu
// @Failure 400 {object} response.Response
// @Router /v1/system/menu/getById/{id} [get]
// @Security LoginToken
func (h *menuHandler) GetById(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" || len(id) == 0 {
		response.NewResponse().Error(ctx, http.StatusBadRequest, "id不能为空", nil)
		return
	}

	detail, err := h.svc.GetById(ctx, id)
	if err != nil {
		if errors.Is(err, serviceSystem.ErrMenuNotFound) {
			response.NewResponse().Error(ctx, http.StatusBadRequest, "菜单不存在", nil)
			return
		}
		ctx.Set("internalError", fmt.Sprintf("获取菜单异常 >>> %v", err.Error()))
		zap.S().Error("获取菜单异常 >>> ", err.Error())
		response.NewResponse().Error(ctx, http.StatusInternalServerError, "服务器异常", nil)
		return
	}

	response.NewResponse().Success(ctx, "获取成功", detail)
}

// GetMenuRouteTree
// @Summary 获取菜单路由树形结构
// @Description 获取菜单路由树形结构
// @Tags 系统管理/菜单管理
// @Accept application/json
// @Produce application/json
// @Security BearerAuth
// @Param creator query string false "创建人"
// @Param modifier query string false "修改人"
// @Param status query bool false "状态" default(true)
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Router /v1/system/menu/listRouteTree [get]
// @Security LoginToken
func (h *menuHandler) GetMenuRouteTree(ctx *gin.Context) {
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

	filter := domainSystem.MenuFilter{
		Filters: filters.Filters{
			Creator:    creator,
			Modifier:   modifier,
			BelongDept: *user.DeptID,
		},
		Status: status,
	}

	list, err := h.svc.GetMenuRouteTree(ctx, filter)
	if err != nil {
		ctx.Set("internalError", fmt.Sprintf("获取菜单路由异常 >>> %v", err.Error()))
		zap.S().Error("获取菜单路由异常 >>> ", err.Error())
		response.NewResponse().Error(ctx, http.StatusInternalServerError, "服务器异常", nil)
		return
	}

	response.NewResponse().Success(ctx, "查询成功", list)
}

// GetListAll
// @Summary 获取所有菜单
// @Description 获取所有菜单列表信息
// @Tags 系统管理/菜单管理
// @Accept application/json
// @Produce application/json
// @Security BearerAuth
// @Param creator query string false "创建人"
// @Param modifier query string false "修改人"
// @Param status query bool false "状态" default(true)
// @Param title query string false "菜单标题"
// @Success 200 {array} []domainSystem.Menu
// @Failure 400 {object} response.Response
// @Router /v1/system/menu/listAll [get]
// @Security LoginToken
func (h *menuHandler) GetListAll(ctx *gin.Context) {
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

	title := ctx.DefaultQuery("title", "")

	filter := domainSystem.MenuFilter{
		Filters: filters.Filters{
			Creator:    creator,
			Modifier:   modifier,
			BelongDept: *user.DeptID,
		},
		Status: status,
		Title:  title,
	}

	list, err := h.svc.GetListAll(ctx, filter)
	if err != nil {
		ctx.Set("internalError", fmt.Sprintf("获取菜单列表异常 >>> %v", err.Error()))
		zap.S().Error("获取菜单列表异常 >>> ", err.Error())
		response.NewResponse().Error(ctx, http.StatusInternalServerError, "服务器异常", nil)
		return
	}

	response.NewResponse().Success(ctx, "查询成功", list)
}

// Export
// @Summary 导出菜单
// @Description 导出菜单信息到Excel文件
// @Tags 系统管理/菜单管理
// @Accept application/json
// @Produce application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
// @Security BearerAuth
// @Param creator query string false "创建人"
// @Param modifier query string false "修改人"
// @Param status query bool false "状态" default(true)
// @Param title query string false "菜单标题"
// @Success 200 {file} file "Excel文件"
// @Failure 500 {object} response.Response
// @Router /v1/system/menu/export [get]
// @Security LoginToken
func (h *menuHandler) Export(ctx *gin.Context) {
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

	title := ctx.DefaultQuery("title", "")

	filter := domainSystem.MenuFilter{
		Filters: filters.Filters{
			Creator:    creator,
			Modifier:   modifier,
			BelongDept: *user.DeptID,
		},
		Status: status,
		Title:  title,
	}

	list, err := h.svc.GetListAll(ctx, filter)
	if err != nil {
		ctx.Set("internalError", fmt.Sprintf("获取菜单列表异常 >>> %v", err.Error()))
		zap.S().Error("获取菜单列表异常 >>> ", err.Error())
		response.NewResponse().Error(ctx, http.StatusInternalServerError, "服务器异常", nil)
		return
	}

	// 准备导出配置
	filename := fmt.Sprintf("菜单信息导出_%s.xlsx", time.Now().Format("20060102150405"))
	cfg := excelutil.ExcelExportConfig{
		SheetName:  "菜单信息",
		FileName:   filename,
		StreamMode: true,
		Columns: []excelutil.ExcelColumn{
			{Title: "菜单名称", Field: "Name", Width: 22},
			{Title: "路由地址", Field: "Path", Width: 22},
			{Title: "组件地址", Field: "Component", Width: 22},
			{Title: "路由标题", Field: "Title", Width: 22},
			{Title: "路由图标", Field: "Icon", Width: 22},
			{
				Title: "是否显示徽章",
				Field: "ShowBadge",
				Width: 22,
				Formatter: func(value interface{}) string {
					if flag, ok := value.(bool); ok {
						if flag {
							return "是"
						}
						return "否"
					}
					return fmt.Sprintf("%v", value)
				},
			},
			{Title: "文本徽章", Field: "ShowTextBadge", Width: 22},
			{
				Title: "是否在菜单中隐藏",
				Field: "IsHide",
				Width: 22,
				Formatter: func(value interface{}) string {
					if flag, ok := value.(bool); ok {
						if flag {
							return "是"
						}
						return "否"
					}
					return fmt.Sprintf("%v", value)
				},
			},
			{
				Title: "是否在标签页中隐藏",
				Field: "IsHideTab",
				Width: 22,
				Formatter: func(value interface{}) string {
					if flag, ok := value.(bool); ok {
						if flag {
							return "是"
						}
						return "否"
					}
					return fmt.Sprintf("%v", value)
				},
			},
			{Title: "外部链接", Field: "Link", Width: 22},
			{
				Title: "是否为iframe",
				Field: "IsIframe",
				Width: 22,
				Formatter: func(value interface{}) string {
					if flag, ok := value.(bool); ok {
						if flag {
							return "是"
						}
						return "否"
					}
					return fmt.Sprintf("%v", value)
				},
			},
			{
				Title: "是否缓存页面",
				Field: "KeepAlive",
				Width: 22,
				Formatter: func(value interface{}) string {
					if flag, ok := value.(bool); ok {
						if flag {
							return "是"
						}
						return "否"
					}
					return fmt.Sprintf("%v", value)
				},
			},
			{
				Title: "是否为一级菜单",
				Field: "IsFirstLevel",
				Width: 22,
				Formatter: func(value interface{}) string {
					if flag, ok := value.(bool); ok {
						if flag {
							return "是"
						}
						return "否"
					}
					return fmt.Sprintf("%v", value)
				},
			},
			{
				Title: "是否固定标签页",
				Field: "FixedTab",
				Width: 22,
				Formatter: func(value interface{}) string {
					if flag, ok := value.(bool); ok {
						if flag {
							return "是"
						}
						return "否"
					}
					return fmt.Sprintf("%v", value)
				},
			},
			{Title: "激活菜单路径", Field: "ActivePath", Width: 22},
			{
				Title: "是否为全屏页面",
				Field: "IsFullPage",
				Width: 22,
				Formatter: func(value interface{}) string {
					if flag, ok := value.(bool); ok {
						if flag {
							return "是"
						}
						return "否"
					}
					return fmt.Sprintf("%v", value)
				},
			},
			{Title: "上级菜单", Field: "Parent.Name", Width: 20},
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
		ctx.Set("internalError", fmt.Sprintf("导出菜单信息异常 >>> %v", err.Error()))
		zap.S().Error("导出菜单信息异常 >>> ", err.Error())
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

// --------------------------------------------------

// // GetMenuTree 获取菜单树形结构
// // @Summary 获取菜单树形结构
// // @Description 获取菜单树形结构
// // @Tags 系统管理/菜单管理
// // @Accept application/json
// // @Produce application/json
// // @Param creator query string false "创建人"
// // @Param modifier query string false "修改人"
// // @Param status query bool false "状态" default(true)
// // @Param title query string false "菜单标题"
// // @Success 200 {object} response.Response
// // @Failure 400 {object} response.Response
// // @Router /v1/system/menu/listTree [get]
// // @Security LoginToken
// func (h *menuHandler) GetMenuTree(ctx *gin.Context) {
// 	// 从上下文中获取登录信息
// 	claims, ok := ctx.MustGet("claims").(*jwt.Claims)
// 	if !ok {
// 		zap.S().Error("未找到用户认证信息 >>> ", zap.Error(errors.New(claims.UserId)))
// 		response.NewResponse().Error(ctx, http.StatusInternalServerError, "服务器异常", nil)
// 		return
// 	}
//
// 	user, err := h.userSvc.GetById(ctx, claims.UserId)
// 	if err != nil {
// 		ctx.Set("internalError", fmt.Sprintf("获取用户信息异常 >>> %v", err.Error()))
// 		zap.S().Error("获取用户信息异常 >>> ", err.Error())
// 		response.NewResponse().Error(ctx, http.StatusInternalServerError, "服务器异常", nil)
// 		return
// 	}
//
// 	creator := ctx.DefaultQuery("creator", "")
// 	modifier := ctx.DefaultQuery("modifier", "")
// 	status, _ := strconv.ParseBool(ctx.DefaultQuery("status", "true"))
//
// 	title := ctx.DefaultQuery("title", "")
//
// 	filter := domainSystem.MenuFilter{
// 		Filters: filters.Filters{
// 			Creator:    creator,
// 			Modifier:   modifier,
// 			BelongDept: user.DeptId,
// 		},
// 		Status: status,
// 		Title:  title,
// 	}
//
// 	list, err := h.svc.GetListTree(ctx, filter)
// 	if err != nil {
// 		ctx.Set("internalError", fmt.Sprintf("获取菜单列表异常 >>> %v", err.Error()))
// 		zap.S().Error("获取菜单列表异常 >>> ", err.Error())
// 		response.NewResponse().Error(ctx, http.StatusInternalServerError, "服务器异常", nil)
// 		return
// 	}
//
// 	response.NewResponse().Success(ctx, "查询成功", list)
// }
