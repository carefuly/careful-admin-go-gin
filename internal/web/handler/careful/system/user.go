/**
 * Description：
 * FileName：user.go
 * Author：CJiaの用心
 * Create：2025/12/22 14:43:12
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
	"github.com/carefuly/careful-admin-go-gin/pkg/constants/careful/system/user"
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

// CreateUserRequest 创建
type CreateUserRequest struct {
	Status      user.Status      `json:"status" binding:"omitempty" default:"1"`                // 状态
	Username    string           `json:"username" binding:"required,max=50" default:""`         // 用户名
	Password    string           `json:"password" binding:"omitempty,max=128" default:"123456"` // 密码
	Gender      user.GenderConst `json:"gender" binding:"omitempty" default:"1"`                // 性别
	Email       string           `json:"email" binding:"omitempty,max=128" default:""`          // 邮箱
	Mobile      string           `json:"mobile" binding:"omitempty,max=11" default:""`          // 手机号
	Name        string           `json:"name" binding:"omitempty,max=64" default:""`            // 真实姓名
	Avatar      string           `json:"avatar" binding:"omitempty" default:""`                 // 头像
	Birthday    string           `json:"birthday" binding:"omitempty" default:""`               // 生日
	City        string           `json:"city" binding:"omitempty,max=200" default:""`           // 所在城市
	Address     string           `json:"address" binding:"omitempty,max=100" default:""`        // 详细地址
	Bio         string           `json:"bio" binding:"omitempty,max=512" default:""`            // 个人简介
	IsSuperuser bool             `json:"is_superuser" binding:"omitempty" default:"false"`      // 是否为超级管理员
	ManagerID   *string          `json:"manager_id" binding:"omitempty" default:""`             // 直属上级ID
	DeptID      *string          `json:"dept_id" binding:"omitempty" default:""`                // 所属部门ID
	PostIDs     []string         `json:"post_ids" binding:"omitempty" default:"[]"`             // 岗位ID数组
	RoleIDs     []string         `json:"role_ids" binding:"omitempty" default:"[]"`             // 角色ID数组
	Sort        int              `json:"sort" binding:"omitempty" default:"1"`                  // 排序
	Remark      string           `json:"remark" binding:"omitempty,max=255" default:""`         // 备注
}

// ImportUserRequest 导入
type ImportUserRequest struct {
	File *multipart.FileHeader `form:"file" binding:"required"` // 文件
}

// UpdateUserRequest 更新
type UpdateUserRequest struct {
	Id          string           `json:"id" binding:"required" default:""`                      // 主键ID
	Status      user.Status      `json:"status" binding:"omitempty" default:"1"`                // 状态
	Username    string           `json:"username" binding:"required,max=50" default:""`         // 用户名
	Password    string           `json:"password" binding:"omitempty,max=128" default:"123456"` // 密码
	Gender      user.GenderConst `json:"gender" binding:"omitempty" default:"1"`                // 性别
	Email       string           `json:"email" binding:"omitempty,max=128" default:""`          // 邮箱
	Mobile      string           `json:"mobile" binding:"omitempty,max=11" default:""`          // 手机号
	Name        string           `json:"name" binding:"omitempty,max=64" default:""`            // 真实姓名
	Avatar      string           `json:"avatar" binding:"omitempty" default:""`                 // 头像
	Birthday    string           `json:"birthday" binding:"omitempty" default:""`               // 生日
	City        string           `json:"city" binding:"omitempty,max=200" default:""`           // 所在城市
	Address     string           `json:"address" binding:"omitempty,max=100" default:""`        // 详细地址
	Bio         string           `json:"bio" binding:"omitempty,max=512" default:""`            // 个人简介
	IsSuperuser bool             `json:"is_superuser" binding:"omitempty" default:"false"`      // 是否为超级管理员
	ManagerID   *string          `json:"manager_id" binding:"omitempty" default:""`             // 直属上级ID
	DeptID      *string          `json:"dept_id" binding:"omitempty" default:""`                // 所属部门ID
	PostIDs     []string         `json:"post_ids" binding:"omitempty" default:"[]"`             // 岗位ID数组
	RoleIDs     []string         `json:"role_ids" binding:"omitempty" default:"[]"`             // 角色ID数组
	Sort        int              `json:"sort" binding:"omitempty" default:"1"`                  // 排序
	Timestamp   int64            `json:"timestamp" binding:"omitempty"`                         // 版本
	Remark      string           `json:"remark" binding:"omitempty,max=255" default:""`         // 备注
}

// UserListPageResponse 列表分页响应
type UserListPageResponse struct {
	List     []domainSystem.User `json:"list"`     // 列表
	Total    int64               `json:"total"`    // 总数
	Page     int                 `json:"page"`     // 页码
	PageSize int                 `json:"pageSize"` // 每页数量
}

type UserHandler interface {
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

type userHandler struct {
	rely    config.RelyConfig
	svc     serviceSystem.UserService
	userSvc serviceSystem.UserService
}

func NewUserHandler(rely config.RelyConfig, svc serviceSystem.UserService, userSvc serviceSystem.UserService) UserHandler {
	return &userHandler{
		rely:    rely,
		svc:     svc,
		userSvc: userSvc,
	}
}

func (h *userHandler) RegisterRoutes(router *gin.RouterGroup) {
	base := router.Group("/user")
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
// @Summary 创建用户
// @Description 创建用户
// @Tags 系统管理/用户管理
// @Accept application/json
// @Produce application/json
// @Security BearerAuth
// @Param CreateUserRequest body CreateUserRequest true "请求"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Router /v1/system/user/create [post]
// @Security LoginToken
func (h *userHandler) Create(ctx *gin.Context) {
	// 从上下文中获取登录信息
	claims, ok := ctx.MustGet("claims").(*jwt.Claims)
	if !ok {
		zap.S().Error("未找到用户认证信息 >>> ", zap.Error(errors.New(claims.UserID)))
		response.NewResponse().Error(ctx, http.StatusInternalServerError, "服务器异常", nil)
		return
	}

	loginUser, err := h.userSvc.GetById(ctx, claims.UserID)
	if err != nil {
		ctx.Set("internalError", fmt.Sprintf("获取用户信息异常 >>> %v", err.Error()))
		zap.S().Error("获取用户信息异常 >>> ", err.Error())
		response.NewResponse().Error(ctx, http.StatusInternalServerError, "服务器异常", nil)
		return
	}

	var req CreateUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		validate.NewValidatorErrorHandler(h.rely.Trans).Handle(ctx, err)
		return
	}

	birthdayTime, err := time.Parse("2006-01-02", req.Birthday)
	if err != nil {
		fmt.Printf("解析生日失败：%v\n", err)
		return
	}

	// 校验参数
	genderValidValues := []string{"男", "女", "保密"}
	converter := enumconv.NewEnumConverter(user.GenderMapping, user.GenderImportMapping, genderValidValues, "用户性别")
	_, err = converter.FromEnum(req.Gender)
	if err != nil {
		response.NewResponse().Error(ctx, http.StatusBadRequest, err.Error(), nil)
		return
	}

	// 转换为领域模型
	domain := domainSystem.User{
		User: modelSystem.User{
			CoreModels: models.CoreModels{
				Sort:       req.Sort,
				Creator:    loginUser.Id,
				Modifier:   loginUser.Id,
				BelongDept: loginUser.DeptID,
				Remark:     req.Remark,
			},
			Status:      req.Status,
			Username:    req.Username,
			Password:    req.Password,
			Gender:      req.Gender,
			Email:       req.Email,
			Mobile:      req.Mobile,
			Name:        req.Name,
			Avatar:      req.Avatar,
			Birthday:    &birthdayTime,
			City:        req.City,
			Address:     req.Address,
			Bio:         req.Bio,
			IsSuperuser: req.IsSuperuser,
			ManagerID:   req.ManagerID,
			DeptID:      req.DeptID,
		},
		PostIDs: req.PostIDs,
		RoleIDs: req.RoleIDs,
	}

	if err := h.svc.Create(ctx, domain); err != nil {
		switch {
		case errors.Is(err, serviceSystem.ErrUserUsernameDuplicate):
			response.NewResponse().Error(ctx, http.StatusBadRequest, "用户名已存在", nil)
			return
		default:
			ctx.Set("internalError", fmt.Sprintf("创建用户异常 >>> %v", err.Error()))
			zap.S().Error("创建用户异常 >>> ", err.Error())
			response.NewResponse().Error(ctx, http.StatusInternalServerError, "服务器异常", nil)
			return
		}
	}

	response.NewResponse().Success(ctx, "新增成功", nil)
}

// Import
// @Summary 导入用户
// @Description 导入用户
// @Tags 系统管理/用户管理
// @Accept multipart/form-data
// @Produce application/json
// @Security BearerAuth
// @Param file formData file true "文件(支持xlsx/csv格式)"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Router /v1/system/user/import [post]
// @Security LoginToken
func (h *userHandler) Import(ctx *gin.Context) {

}

// Delete
// @Summary 删除用户
// @Description 删除指定id用户
// @Tags 系统管理/用户管理
// @Accept application/json
// @Produce application/json
// @Security BearerAuth
// @Param id path string true "id"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Router /v1/system/user/delete/{id} [delete]
// @Security LoginToken
func (h *userHandler) Delete(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" || len(id) == 0 {
		response.NewResponse().Error(ctx, http.StatusBadRequest, "ID不能为空", nil)
		return
	}

	if err := h.svc.Delete(ctx, id); err != nil {
		switch {
		default:
			ctx.Set("internalError", fmt.Sprintf("删除用户异常 >>> %v", err.Error()))
			zap.S().Error("删除用户异常 >>> ", zap.Error(err))
			response.NewResponse().Error(ctx, http.StatusInternalServerError, "服务器异常", nil)
			return
		}
	}

	response.NewResponse().Success(ctx, "删除成功", nil)
}

// BatchDelete
// @Summary 批量删除用户
// @Description 批量删除用户
// @Tags 系统管理/用户管理
// @Accept application/json
// @Produce application/json
// @Security BearerAuth
// @Param ids body []string true "id数组"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Router /v1/system/user/batchDelete [post]
// @Security LoginToken
func (h *userHandler) BatchDelete(ctx *gin.Context) {
	var ids []string
	if err := ctx.ShouldBindJSON(&ids); err != nil {
		validate.NewValidatorErrorHandler(h.rely.Trans).Handle(ctx, err)
		return
	}

	err := h.svc.BatchDelete(ctx, ids)
	if err != nil {
		ctx.Set("internalError", fmt.Sprintf("批量删除用户异常 >>> %v", err.Error()))
		zap.S().Error("批量删除用户异常 >>> ", zap.Error(err))
		response.NewResponse().Error(ctx, http.StatusInternalServerError, "服务器异常", nil)
		return
	}

	response.NewResponse().Success(ctx, "批量删除成功", nil)
}

// Update
// @Summary 更新用户
// @Description 更新用户
// @Tags 系统管理/用户管理
// @Accept application/json
// @Produce application/json
// @Security BearerAuth
// @Param UpdateUserRequest body UpdateUserRequest true "请求"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Router /v1/system/user/update [put]
// @Security LoginToken
func (h *userHandler) Update(ctx *gin.Context) {
	// 从上下文中获取登录信息
	claims, ok := ctx.MustGet("claims").(*jwt.Claims)
	if !ok {
		zap.S().Error("未找到用户认证信息 >>> ", zap.Error(errors.New(claims.UserID)))
		response.NewResponse().Error(ctx, http.StatusInternalServerError, "服务器异常", nil)
		return
	}

	loginUser, err := h.userSvc.GetById(ctx, claims.UserID)
	if err != nil {
		ctx.Set("internalError", fmt.Sprintf("获取用户信息异常 >>> %v", err.Error()))
		zap.S().Error("获取用户信息异常 >>> ", err.Error())
		response.NewResponse().Error(ctx, http.StatusInternalServerError, "服务器异常", nil)
		return
	}

	var req UpdateUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		fmt.Println(err)
		validate.NewValidatorErrorHandler(h.rely.Trans).Handle(ctx, err)
		return
	}

	birthdayTime, err := time.Parse("2006-01-02", req.Birthday)
	if err != nil {
		fmt.Printf("解析生日失败：%v\n", err)
		return
	}

	// 校验参数
	genderValidValues := []string{"男", "女", "保密"}
	converter := enumconv.NewEnumConverter(user.GenderMapping, user.GenderImportMapping, genderValidValues, "用户性别")
	_, err = converter.FromEnum(req.Gender)
	if err != nil {
		response.NewResponse().Error(ctx, http.StatusBadRequest, err.Error(), nil)
		return
	}

	// 转换为领域模型
	domain := domainSystem.User{
		User: modelSystem.User{
			CoreModels: models.CoreModels{
				Id:         req.Id,
				Sort:       req.Sort,
				Timestamp:  req.Timestamp,
				Modifier:   loginUser.Id,
				BelongDept: loginUser.DeptID,
				Remark:     req.Remark,
			},
			Status:      req.Status,
			Username:    req.Username,
			Gender:      req.Gender,
			Email:       req.Email,
			Mobile:      req.Mobile,
			Name:        req.Name,
			Avatar:      req.Avatar,
			Birthday:    &birthdayTime,
			City:        req.City,
			Address:     req.Address,
			Bio:         req.Bio,
			IsSuperuser: req.IsSuperuser,
			ManagerID:   req.ManagerID,
			DeptID:      req.DeptID,
		},
		PostIDs: req.PostIDs,
		RoleIDs: req.RoleIDs,
	}

	if err := h.svc.Update(ctx, domain); err != nil {
		switch {
		case errors.Is(err, serviceSystem.ErrUserUsernameDuplicate):
			response.NewResponse().Error(ctx, http.StatusBadRequest, "用户名已存在", nil)
			return
		case errors.Is(err, serviceSystem.ErrUserVersionInconsistency):
			response.NewResponse().Error(ctx, http.StatusBadRequest, "数据版本不一致，取消修改，请刷新后重试", nil)
			return
		default:
			ctx.Set("internalError", fmt.Sprintf("更新用户异常 >>> %v", err.Error()))
			zap.S().Error("更新用户异常 >>> ", err.Error())
			response.NewResponse().Error(ctx, http.StatusInternalServerError, "服务器异常", nil)
			return
		}
	}

	response.NewResponse().Success(ctx, "更新成功", nil)
}

// GetById
// @Summary 获取用户
// @Description 获取指定id用户
// @Tags 系统管理/用户管理
// @Accept application/json
// @Produce application/json
// @Security BearerAuth
// @Param id path string true "id"
// @Success 200 {object} domainSystem.User
// @Failure 400 {object} response.Response
// @Router /v1/system/user/getById/{id} [get]
// @Security LoginToken
func (h *userHandler) GetById(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" || len(id) == 0 {
		response.NewResponse().Error(ctx, http.StatusBadRequest, "id不能为空", nil)
		return
	}

	detail, err := h.svc.GetById(ctx, id)
	if err != nil {
		if errors.Is(err, serviceSystem.ErrPostNotFound) {
			response.NewResponse().Error(ctx, http.StatusBadRequest, "用户不存在", nil)
			return
		}
		ctx.Set("internalError", fmt.Sprintf("获取用户异常 >>> %v", err.Error()))
		zap.S().Error("获取用户异常 >>> ", err.Error())
		response.NewResponse().Error(ctx, http.StatusInternalServerError, "服务器异常", nil)
		return
	}

	response.NewResponse().Success(ctx, "获取成功", detail)
}

// GetListPage
// @Summary 获取用户分页列表
// @Description 获取用户分页列表
// @Tags 系统管理/用户管理
// @Accept application/json
// @Produce application/json
// @Security BearerAuth
// @Param page query int true "页码" default(1)
// @Param pageSize query int true "每页数量" default(10)
// @Param creator query string false "创建人"
// @Param modifier query string false "修改人"
// @Param status query bool false "状态" default(true)
// @Param username query string false "用户名"
// @Param email query string false "邮箱"
// @Param mobile query string false "手机号"
// @Param dept_id query string false "所属部门ID"
// @Success 200 {object} UserListPageResponse
// @Failure 400 {object} response.Response
// @Router /v1/system/user/listPage [get]
// @Security LoginToken
func (h *userHandler) GetListPage(ctx *gin.Context) {
	// 从上下文中获取登录信息
	claims, ok := ctx.MustGet("claims").(*jwt.Claims)
	if !ok {
		zap.S().Error("未找到用户认证信息 >>> ", zap.Error(errors.New(claims.UserID)))
		response.NewResponse().Error(ctx, http.StatusInternalServerError, "服务器异常", nil)
		return
	}

	loginUser, err := h.userSvc.GetById(ctx, claims.UserID)
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

	username := ctx.DefaultQuery("username", "")
	email := ctx.DefaultQuery("email", "")
	mobile := ctx.DefaultQuery("mobile", "")
	// deptId := ctx.DefaultQuery("dept_id", "")

	filter := domainSystem.UserFilter{
		Pagination: filters.Pagination{
			Page:     page,
			PageSize: pageSize,
		},
		Filters: filters.Filters{
			Creator:    creator,
			Modifier:   modifier,
			BelongDept: *loginUser.DeptID,
		},
		Status:   status,
		Username: username,
		Email:    email,
		Mobile:   mobile,
	}

	list, total, err := h.svc.GetListPage(ctx, filter)
	if err != nil {
		ctx.Set("internalError", fmt.Sprintf("获取用户分页列表异常 >>> %v", err.Error()))
		zap.S().Error("获取用户分页列表异常 >>> ", err.Error())
		response.NewResponse().Error(ctx, http.StatusInternalServerError, "服务器异常", nil)
		return
	}

	response.NewResponse().Success(ctx, "查询成功", UserListPageResponse{
		List:     list,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}

// GetListAll
// @Summary 获取所有用户
// @Description 获取所有用户
// @Tags 系统管理/用户管理
// @Accept application/json
// @Produce application/json
// @Security BearerAuth
// @Param creator query string false "创建人"
// @Param modifier query string false "修改人"
// @Param status query bool false "状态" default(true)
// @Param username query string false "用户名"
// @Param email query string false "邮箱"
// @Param mobile query string false "手机号"
// @Param dept_id query string false "所属部门ID"
// @Success 200 {array} []domainSystem.User
// @Failure 400 {object} response.Response
// @Router /v1/system/user/listAll [get]
// @Security LoginToken
func (h *userHandler) GetListAll(ctx *gin.Context) {
	// 从上下文中获取登录信息
	claims, ok := ctx.MustGet("claims").(*jwt.Claims)
	if !ok {
		zap.S().Error("未找到用户认证信息 >>> ", zap.Error(errors.New(claims.UserID)))
		response.NewResponse().Error(ctx, http.StatusInternalServerError, "服务器异常", nil)
		return
	}

	loginUser, err := h.userSvc.GetById(ctx, claims.UserID)
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

	username := ctx.DefaultQuery("username", "")
	email := ctx.DefaultQuery("email", "")
	mobile := ctx.DefaultQuery("mobile", "")
	// deptId := ctx.DefaultQuery("dept_id", "")

	filter := domainSystem.UserFilter{
		Filters: filters.Filters{
			Creator:    creator,
			Modifier:   modifier,
			BelongDept: *loginUser.DeptID,
		},
		Status:   status,
		Username: username,
		Email:    email,
		Mobile:   mobile,
	}

	list, err := h.svc.GetListAll(ctx, filter)
	if err != nil {
		ctx.Set("internalError", fmt.Sprintf("获取用户列表异常 >>> %v", err.Error()))
		zap.S().Error("获取用户列表异常 >>> ", err.Error())
		response.NewResponse().Error(ctx, http.StatusInternalServerError, "服务器异常", nil)
		return
	}

	response.NewResponse().Success(ctx, "查询成功", list)
}

// Export
// @Summary 导出用户
// @Description 导出用户到Excel文件
// @Tags 系统管理/用户管理
// @Accept application/json
// @Produce application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
// @Security BearerAuth
// @Param creator query string false "创建人"
// @Param modifier query string false "修改人"
// @Param status query bool false "状态" default(true)
// @Param username query string false "用户名"
// @Param email query string false "邮箱"
// @Param mobile query string false "手机号"
// @Param dept_id query string false "所属部门ID"
// @Success 200 {file} file "Excel文件"
// @Failure 500 {object} response.Response
// @Router /v1/system/user/export [get]
// @Security LoginToken
func (h *userHandler) Export(ctx *gin.Context) {
	// 从上下文中获取登录信息
	claims, ok := ctx.MustGet("claims").(*jwt.Claims)
	if !ok {
		zap.S().Error("未找到用户认证信息 >>> ", zap.Error(errors.New(claims.UserID)))
		response.NewResponse().Error(ctx, http.StatusInternalServerError, "服务器异常", nil)
		return
	}

	loginUser, err := h.userSvc.GetById(ctx, claims.UserID)
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

	username := ctx.DefaultQuery("username", "")
	email := ctx.DefaultQuery("email", "")
	mobile := ctx.DefaultQuery("mobile", "")
	// deptId := ctx.DefaultQuery("dept_id", "")

	filter := domainSystem.UserFilter{
		Filters: filters.Filters{
			Creator:    creator,
			Modifier:   modifier,
			BelongDept: *loginUser.DeptID,
		},
		Status:   status,
		Username: username,
		Email:    email,
		Mobile:   mobile,
	}

	list, err := h.svc.GetListAll(ctx, filter)
	if err != nil {
		ctx.Set("internalError", fmt.Sprintf("获取用户列表异常 >>> %v", err.Error()))
		zap.S().Error("获取用户列表异常 >>> ", err.Error())
		response.NewResponse().Error(ctx, http.StatusInternalServerError, "服务器异常", nil)
		return
	}

	// 准备导出配置
	filename := fmt.Sprintf("用户导出_%s.xlsx", time.Now().Format("20060102150405"))
	cfg := excelutil.ExcelExportConfig{
		SheetName:  "用户",
		FileName:   filename,
		StreamMode: true,
		Columns: []excelutil.ExcelColumn{
			{Title: "用户名", Field: "Username", Width: 22},
			{
				Title: "性别",
				Field: "Gender",
				Width: 15,
				Formatter: func(value interface{}) string {
					genderValidValues := []string{"男", "女", "保密"}
					converter := enumconv.NewEnumConverter(user.GenderMapping, user.GenderImportMapping, genderValidValues, "用户性别")
					str, _ := converter.FromEnum(value.(user.GenderConst))
					return str
				},
			},
			{Title: "邮箱", Field: "Email", Width: 22},
			{Title: "手机号", Field: "Mobile", Width: 22},
			{Title: "真实姓名", Field: "Name", Width: 22},
			{Title: "生日", Field: "Birthday", Width: 22},
			{Title: "所在城市", Field: "City", Width: 22},
			{Title: "详细地址", Field: "Address", Width: 22},
			{Title: "个人简介", Field: "Bio", Width: 22},
			// {
			// 	Title: "状态",
			// 	Field: "Status",
			// 	Width: 10,
			// 	Formatter: func(value interface{}) string {
			// 		if status, ok := value.(bool); ok {
			// 			if status {
			// 				return "启用"
			// 			}
			// 			return "停用"
			// 		}
			// 		return fmt.Sprintf("%v", value)
			// 	},
			// },
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
		ctx.Set("internalError", fmt.Sprintf("导出用户异常 >>> %v", err.Error()))
		zap.S().Error("导出用户异常 >>> ", err.Error())
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
