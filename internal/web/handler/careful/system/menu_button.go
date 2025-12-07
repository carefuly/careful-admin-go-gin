/**
 * Description：
 * FileName：menu_button.go
 * Author：CJiaの用心
 * Create：2025/12/05 14:55:11
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
	"github.com/carefuly/careful-admin-go-gin/pkg/constants/careful/system/menu"
	"github.com/carefuly/careful-admin-go-gin/pkg/ginx/response"
	"github.com/carefuly/careful-admin-go-gin/pkg/models"
	"github.com/carefuly/careful-admin-go-gin/pkg/utils/enumconv"
	"github.com/carefuly/careful-admin-go-gin/pkg/utils/jwt"
	"github.com/carefuly/careful-admin-go-gin/pkg/validate"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"net/http"
)

// CreateMenuButtonRequest 创建
type CreateMenuButtonRequest struct {
	Status   bool        `json:"status" binding:"omitempty" default:"true"`      // 状态【true-启用 false-停用】
	Title    string      `json:"title" binding:"required,max=64" default:""`     // 名称
	AuthMark string      `json:"auth_mark" binding:"required,max=64" default:""` // 权限值
	Method   menu.Method `json:"method" binding:"required" default:"1"`          // 方法类型
	Api      string      `json:"api" binding:"required,max=255" default:""`      // 接口地址
	MenuID   string      `json:"menu_id" binding:"required,max=110" default:""`  // 菜单ID
	Sort     int         `json:"sort" binding:"omitempty" default:"1"`           // 排序
	Remark   string      `json:"remark" binding:"omitempty,max=255" default:""`  // 备注
}

// UpdateMenuButtonRequest 更新
type UpdateMenuButtonRequest struct {
	Id        string      `json:"id" binding:"required" default:""`               // 主键ID
	Status    bool        `json:"status" binding:"omitempty" default:"true"`      // 状态【true-启用 false-停用】
	Title     string      `json:"title" binding:"required,max=64" default:""`     // 名称
	AuthMark  string      `json:"auth_mark" binding:"required,max=64" default:""` // 权限值
	Method    menu.Method `json:"method" binding:"required" default:"1"`          // 方法类型
	Api       string      `json:"api" binding:"required,max=255" default:""`      // 接口地址
	MenuID    string      `json:"menu_id" binding:"required,max=110" default:""`  // 菜单ID
	Sort      int         `json:"sort" binding:"omitempty" default:"1"`           // 排序
	Timestamp int64       `json:"timestamp" binding:"omitempty"`                  // 版本
	Remark    string      `json:"remark" binding:"omitempty,max=255" default:""`  // 备注
}

// MenuButtonListPageResponse 列表分页响应
type MenuButtonListPageResponse struct {
	List     []domainSystem.MenuButton `json:"list"`     // 列表
	Total    int64                     `json:"total"`    // 总数
	Page     int                       `json:"page"`     // 页码
	PageSize int                       `json:"pageSize"` // 每页数量
}

type MenuButtonHandler interface {
	// RegisterRoutes(router *gin.RouterGroup)
	// Create(ctx *gin.Context)
	// Delete(ctx *gin.Context)
	// BatchDelete(ctx *gin.Context)
	// Update(ctx *gin.Context)
	// GetById(ctx *gin.Context)
	// GetListPage(ctx *gin.Context)
	// GetListByMenuIds(ctx *gin.Context)
	// GetListAll(ctx *gin.Context)
}

type menuButtonHandler struct {
	rely    config.RelyConfig
	svc     serviceSystem.MenuButtonService
	userSvc serviceSystem.UserService
}

func NewMenuButtonHandler(rely config.RelyConfig, svc serviceSystem.MenuButtonService, userSvc serviceSystem.UserService) MenuButtonHandler {
	return &menuButtonHandler{
		rely:    rely,
		svc:     svc,
		userSvc: userSvc,
	}
}

// RegisterRoutes 注册路由
func (h *menuButtonHandler) RegisterRoutes(router *gin.RouterGroup) {
	// base := router.Group("/menuButton")
	// base.POST("/create", h.Create)
	// base.DELETE("/delete/:id", h.Delete)
	// base.POST("/delete/batchDelete", h.BatchDelete)
	// base.PUT("/update", h.Update)
	// base.GET("/getById/:id", h.GetById)
	// base.GET("/listPage", h.GetListPage)
	// base.POST("/listByMenuIds", h.GetListByMenuIds)
	// base.GET("/listAll", h.GetListAll)
}

// Create
// @Summary 创建菜单按钮
// @Description 创建菜单按钮信息
// @Tags 系统管理/菜单按钮管理
// @Accept application/json
// @Produce application/json
// @Security BearerAuth
// @Param CreateMenuButtonRequest body CreateMenuButtonRequest true "请求"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Router /v1/system/menuButton/create [post]
// @Security LoginToken
func (h *menuButtonHandler) Create(ctx *gin.Context) {
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

	var req CreateMenuButtonRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		validate.NewValidatorErrorHandler(h.rely.Trans).Handle(ctx, err)
		return
	}

	// 校验参数
	methodValidValues := []string{"GET", "POST", "PUT", "DELETE"}
	converter := enumconv.NewEnumConverter(menu.MethodMapping, menu.MethodImportMapping, methodValidValues, "方法类型")
	_, err = converter.FromEnum(req.Method)
	if err != nil {
		response.NewResponse().Error(ctx, http.StatusBadRequest, err.Error(), nil)
		return
	}

	// 转换为领域模型
	domain := domainSystem.MenuButton{
		MenuButton: modelSystem.MenuButton{
			CoreModels: models.CoreModels{
				Sort:       req.Sort,
				Creator:    user.Id,
				Modifier:   user.Id,
				BelongDept: user.DeptID,
				Remark:     req.Remark,
			},
			Status:   req.Status,
			Title:    req.Title,
			AuthMark: req.AuthMark,
			Api:      req.Api,
			Method:   req.Method,
			MenuID:   req.MenuID,
		},
	}

	if err := h.svc.Create(ctx, domain); err != nil {
		ctx.Set("internalError", fmt.Sprintf("创建菜单按钮异常 >>> %v", err.Error()))
		zap.S().Error("创建菜单按钮异常 >>> ", zap.Error(err))
		response.NewResponse().Error(ctx, http.StatusInternalServerError, "服务器异常", nil)
		return
	}

	response.NewResponse().Success(ctx, "新增成功", nil)
}

// Delete
// @Summary 删除菜单按钮
// @Description 删除指定id菜单按钮信息
// @Tags 系统管理/菜单按钮管理
// @Accept application/json
// @Produce application/json
// @Security BearerAuth
// @Param id path string true "id"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Router /v1/system/menuButton/delete/{id} [delete]
// @Security LoginToken
func (h *menuButtonHandler) Delete(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" || len(id) == 0 {
		response.NewResponse().Error(ctx, http.StatusBadRequest, "ID不能为空", nil)
		return
	}

	if err := h.svc.Delete(ctx, id); err != nil {
		if errors.Is(err, serviceSystem.ErrMenuButtonNotFound) {
			response.NewResponse().Error(ctx, http.StatusBadRequest, "菜单按钮不存在", nil)
			return
		}
		ctx.Set("internalError", fmt.Sprintf("删除菜单按钮异常 >>> %v", err.Error()))
		zap.S().Error("删除菜单按钮异常 >>> ", zap.Error(err))
		response.NewResponse().Error(ctx, http.StatusInternalServerError, "服务器异常", nil)
		return
	}

	response.NewResponse().Success(ctx, "删除成功", nil)
}

// BatchDelete
// @Summary 批量删除菜单按钮
// @Description 批量删除菜单按钮信息
// @Tags 系统管理/菜单按钮管理
// @Accept application/json
// @Produce application/json
// @Security BearerAuth
// @Param ids body []string true "id数组"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Router /v1/system/menuButton/delete/batchDelete [post]
// @Security LoginToken
func (h *menuButtonHandler) BatchDelete(ctx *gin.Context) {
	var ids []string
	if err := ctx.ShouldBindJSON(&ids); err != nil {
		validate.NewValidatorErrorHandler(h.rely.Trans).Handle(ctx, err)
		return
	}

	err := h.svc.BatchDelete(ctx, ids)
	if err != nil {
		ctx.Set("internalError", fmt.Sprintf("批量删除菜单按钮异常 >>> %v", err.Error()))
		zap.S().Error("批量删除菜单按钮异常 >>> ", zap.Error(err))
		response.NewResponse().Error(ctx, http.StatusInternalServerError, "服务器异常", nil)
		return
	}

	response.NewResponse().Success(ctx, "批量删除成功", nil)
}

// Update
// @Summary 更新菜单按钮
// @Description 更新菜单按钮信息
// @Tags 系统管理/菜单按钮管理
// @Accept application/json
// @Produce application/json
// @Security BearerAuth
// @Param UpdateMenuButtonRequest body UpdateMenuButtonRequest true "请求"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Router /v1/system/menuButton/update [put]
// @Security LoginToken
func (h *menuButtonHandler) Update(ctx *gin.Context) {
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

	var req UpdateMenuButtonRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		validate.NewValidatorErrorHandler(h.rely.Trans).Handle(ctx, err)
		return
	}

	// 校验参数
	methodValidValues := []string{"GET", "POST", "PUT", "DELETE"}
	converter := enumconv.NewEnumConverter(menu.MethodMapping, menu.MethodImportMapping, methodValidValues, "方法类型")
	_, err = converter.FromEnum(req.Method)
	if err != nil {
		response.NewResponse().Error(ctx, http.StatusBadRequest, err.Error(), nil)
		return
	}

	// 转换为领域模型
	domain := domainSystem.MenuButton{
		MenuButton: modelSystem.MenuButton{
			CoreModels: models.CoreModels{
				Id:         req.Id,
				Sort:       req.Sort,
				Timestamp:  req.Timestamp,
				Modifier:   user.Id,
				BelongDept: user.DeptID,
				Remark:     req.Remark,
			},
			Status:   req.Status,
			Title:    req.Title,
			AuthMark: req.AuthMark,
			Api:      req.Api,
			Method:   req.Method,
			MenuID:   req.MenuID,
		},
	}

	if err := h.svc.Update(ctx, domain); err != nil {
		switch {
		case errors.Is(err, serviceSystem.ErrMenuButtonVersionInconsistency):
			response.NewResponse().Error(ctx, http.StatusBadRequest, "数据版本不一致，取消修改，请刷新后重试", nil)
			return
		default:
			ctx.Set("internalError", fmt.Sprintf("更新菜单按钮异常 >>> %v", err.Error()))
			zap.S().Error("更新菜单按钮异常 >>> ", err.Error())
			response.NewResponse().Error(ctx, http.StatusInternalServerError, "服务器异常", nil)
			return
		}
	}

	response.NewResponse().Success(ctx, "更新成功", nil)
}
