/**
 * Description：
 * FileName：post.go
 * Author：CJiaの用心
 * Create：2025/11/30 00:57:00
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
	"github.com/carefuly/careful-admin-go-gin/pkg/constants/careful/system/post"
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

// CreatePostRequest 创建
type CreatePostRequest struct {
	Status      bool       `json:"status" binding:"omitempty" default:"true"`      // 状态【true-启用 false-停用】
	Name        string     `json:"name" binding:"required,max=50" default:""`      // 岗位名称
	Code        string     `json:"code" binding:"required,max=50" default:""`      // 岗位编码
	PostType    post.Type  `json:"post_type" binding:"omitempty" default:"5"`      // 岗位类型
	Level       post.Level `json:"level" binding:"omitempty" default:"4"`          // 岗位级别
	Description string     `json:"description" binding:"omitempty" default:""`     // 岗位描述
	DeptID      string     `json:"dept_id" binding:"omitempty,max=110" default:""` // 所属部门ID
	Sort        int        `json:"sort" binding:"omitempty" default:"1"`           // 排序
	Remark      string     `json:"remark" binding:"omitempty,max=255" default:""`  // 备注
}

// ImportPostRequest 导入
type ImportPostRequest struct {
	File *multipart.FileHeader `form:"file" binding:"required"` // 文件
}

// UpdatePostRequest 更新
type UpdatePostRequest struct {
	Id          string     `json:"id" binding:"required" default:""`               // 主键ID
	Status      bool       `json:"status" binding:"omitempty" default:"true"`      // 状态【true-启用 false-停用】
	Name        string     `json:"name" binding:"required,max=50" default:""`      // 岗位名称
	Code        string     `json:"code" binding:"required,max=50" default:""`      // 岗位编码
	PostType    post.Type  `json:"post_type" binding:"omitempty" default:"5"`      // 岗位类型
	Level       post.Level `json:"level" binding:"omitempty" default:"4"`          // 岗位级别
	Description string     `json:"description" binding:"omitempty" default:""`     // 岗位描述
	DeptID      *string    `json:"dept_id" binding:"omitempty,max=110" default:""` // 所属部门ID
	Sort        int        `json:"sort" binding:"omitempty" default:"1"`           // 排序
	Timestamp   int64      `json:"timestamp" binding:"omitempty"`                  // 版本
	Remark      string     `json:"remark" binding:"omitempty,max=255" default:""`  // 备注
}

// PostListPageResponse 列表分页响应
type PostListPageResponse struct {
	List     []domainSystem.Post `json:"list"`     // 列表
	Total    int64               `json:"total"`    // 总数
	Page     int                 `json:"page"`     // 页码
	PageSize int                 `json:"pageSize"` // 每页数量
}

type PostHandler interface {
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

type postHandler struct {
	rely    config.RelyConfig
	svc     serviceSystem.PostService
	userSvc serviceSystem.UserService
}

func NewPostHandler(rely config.RelyConfig, svc serviceSystem.PostService, userSvc serviceSystem.UserService) PostHandler {
	return &postHandler{
		rely:    rely,
		svc:     svc,
		userSvc: userSvc,
	}
}

func (h *postHandler) RegisterRoutes(router *gin.RouterGroup) {
	base := router.Group("/post")
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
// @Summary 创建岗位
// @Description 创建岗位
// @Tags 系统管理/岗位管理
// @Accept application/json
// @Produce application/json
// @Security BearerAuth
// @Param CreatePostRequest body CreatePostRequest true "请求"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Router /v1/system/post/create [post]
// @Security LoginToken
func (h *postHandler) Create(ctx *gin.Context) {
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

	var req CreatePostRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		validate.NewValidatorErrorHandler(h.rely.Trans).Handle(ctx, err)
		return
	}

	// 校验参数
	typeValidValues := []string{"管理岗", "技术岗", "业务岗", "职能岗", "其他"}
	converter := enumconv.NewEnumConverter(post.TypeMapping, post.TypeImportMapping, typeValidValues, "岗位类型")
	_, err = converter.FromEnum(req.PostType)
	if err != nil {
		response.NewResponse().Error(ctx, http.StatusBadRequest, err.Error(), nil)
		return
	}

	levelValidValues := []string{"高层", "中层", "基层", "一般员工"}
	levelConverter := enumconv.NewEnumConverter(post.LevelMapping, post.LevelImportMapping, levelValidValues, "岗位级别")
	_, err = levelConverter.FromEnum(req.Level)
	if err != nil {
		response.NewResponse().Error(ctx, http.StatusBadRequest, err.Error(), nil)
		return
	}

	var deptId *string
	if req.DeptID == "" {
		deptId = nil
	} else {
		deptId = &req.DeptID
	}

	// 转换为领域模型
	domain := domainSystem.Post{
		Post: modelSystem.Post{
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
			PostType:    req.PostType,
			Level:       req.Level,
			Description: req.Description,
			DeptID:      deptId,
		},
	}

	if err := h.svc.Create(ctx, domain); err != nil {
		switch {
		case errors.Is(err, serviceSystem.ErrPostCodeDuplicate):
			response.NewResponse().Error(ctx, http.StatusBadRequest, "岗位编码已存在", nil)
			return
		default:
			ctx.Set("internalError", fmt.Sprintf("创建岗位失败 >>> %v", err.Error()))
			zap.S().Error("创建岗位失败 >>> ", err.Error())
			response.NewResponse().Error(ctx, http.StatusInternalServerError, "服务器异常", nil)
			return
		}
	}

	response.NewResponse().Success(ctx, "新增成功", nil)
}

// Import
// @Summary 导入岗位
// @Description 导入岗位信息
// @Tags 系统管理/岗位管理
// @Accept multipart/form-data
// @Produce application/json
// @Security BearerAuth
// @Param file formData file true "文件(支持xlsx/csv格式)"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Router /v1/system/post/import [post]
// @Security LoginToken
func (h *postHandler) Import(ctx *gin.Context) {
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

	var req ImportPostRequest
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
	read, err := xlsx.NewXlsxFile(filePath).ReadSheetByName("岗位模板")
	if err != nil {
		response.NewResponse().Error(ctx, http.StatusBadRequest, err.Error(), nil)
		return
	}

	result := h.svc.Import(ctx, user, read)

	response.NewResponse().Success(ctx, "导入成功", result)
}

// Delete
// @Summary 删除岗位
// @Description 删除指定id岗位信息
// @Tags 系统管理/岗位管理
// @Accept application/json
// @Produce application/json
// @Security BearerAuth
// @Param id path string true "id"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Router /v1/system/post/delete/{id} [delete]
// @Security LoginToken
func (h *postHandler) Delete(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" || len(id) == 0 {
		response.NewResponse().Error(ctx, http.StatusBadRequest, "ID不能为空", nil)
		return
	}

	if err := h.svc.Delete(ctx, id); err != nil {
		switch {
		case errors.Is(err, serviceSystem.ErrPostNotFound):
			response.NewResponse().Error(ctx, http.StatusBadRequest, "岗位信息不存在", nil)
			return
		case errors.Is(err, serviceSystem.ErrPostHasUsers):
			response.NewResponse().Error(ctx, http.StatusBadRequest, "岗位下仍有用户，无法删除", nil)
			return
		default:
			ctx.Set("internalError", fmt.Sprintf("删除岗位信息异常 >>> %v", err.Error()))
			zap.S().Error("删除岗位信息异常 >>> ", zap.Error(err))
			response.NewResponse().Error(ctx, http.StatusInternalServerError, "服务器异常", nil)
			return
		}
	}

	response.NewResponse().Success(ctx, "删除成功", nil)
}

// BatchDelete
// @Summary 批量删除岗位
// @Description 批量删除岗位信息
// @Tags 系统管理/岗位管理
// @Accept application/json
// @Produce application/json
// @Security BearerAuth
// @Param ids body []string true "id数组"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Router /v1/system/post/batchDelete [post]
// @Security LoginToken
func (h *postHandler) BatchDelete(ctx *gin.Context) {
	var ids []string
	if err := ctx.ShouldBindJSON(&ids); err != nil {
		validate.NewValidatorErrorHandler(h.rely.Trans).Handle(ctx, err)
		return
	}

	err := h.svc.BatchDelete(ctx, ids)
	if err != nil {
		ctx.Set("internal", err.Error())
		zap.L().Error("批量删除岗位异常", zap.Error(err))
		response.NewResponse().Error(ctx, http.StatusInternalServerError, "服务器异常", nil)
		return
	}

	response.NewResponse().Success(ctx, "批量删除成功", nil)
}

// Update
// @Summary 更新岗位
// @Description 更新岗位信息
// @Tags 系统管理/岗位管理
// @Accept application/json
// @Produce application/json
// @Security BearerAuth
// @Param UpdatePostRequest body UpdatePostRequest true "请求"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Router /v1/system/post/update [put]
// @Security LoginToken
func (h *postHandler) Update(ctx *gin.Context) {
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

	var req UpdatePostRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		validate.NewValidatorErrorHandler(h.rely.Trans).Handle(ctx, err)
		return
	}

	// 校验参数
	typeValidValues := []string{"管理岗", "技术岗", "业务岗", "职能岗", "其他"}
	converter := enumconv.NewEnumConverter(post.TypeMapping, post.TypeImportMapping, typeValidValues, "岗位类型")
	_, err = converter.FromEnum(req.PostType)
	if err != nil {
		response.NewResponse().Error(ctx, http.StatusBadRequest, err.Error(), nil)
		return
	}

	levelValidValues := []string{"高层", "中层", "基层", "一般员工"}
	levelConverter := enumconv.NewEnumConverter(post.LevelMapping, post.LevelImportMapping, levelValidValues, "岗位级别")
	_, err = levelConverter.FromEnum(req.Level)
	if err != nil {
		response.NewResponse().Error(ctx, http.StatusBadRequest, err.Error(), nil)
		return
	}

	// 转换为领域模型
	domain := domainSystem.Post{
		Post: modelSystem.Post{
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
			PostType:    req.PostType,
			Level:       req.Level,
			Description: req.Description,
			DeptID:      req.DeptID,
		},
	}

	if err := h.svc.Update(ctx, domain); err != nil {
		switch {
		case errors.Is(err, serviceSystem.ErrPostCodeDuplicate):
			response.NewResponse().Error(ctx, http.StatusBadRequest, "岗位编码已存在", nil)
			return
		case errors.Is(err, serviceSystem.ErrPostVersionInconsistency):
			response.NewResponse().Error(ctx, http.StatusBadRequest, "数据版本不一致，取消修改，请刷新后重试", nil)
			return
		default:
			ctx.Set("internalError", fmt.Sprintf("更新岗位信息失败 >>> %v", err.Error()))
			zap.S().Error("更新岗位信息失败 >>> ", err.Error())
			response.NewResponse().Error(ctx, http.StatusInternalServerError, "服务器异常", nil)
			return
		}
	}

	response.NewResponse().Success(ctx, "更新成功", nil)
}

// GetById
// @Summary 获取岗位
// @Description 获取指定id岗位信息
// @Tags 系统管理/岗位管理
// @Accept application/json
// @Produce application/json
// @Security BearerAuth
// @Param id path string true "id"
// @Success 200 {object} domainSystem.Post
// @Failure 400 {object} response.Response
// @Router /v1/system/post/getById/{id} [get]
// @Security LoginToken
func (h *postHandler) GetById(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" || len(id) == 0 {
		response.NewResponse().Error(ctx, http.StatusBadRequest, "id不能为空", nil)
		return
	}

	detail, err := h.svc.GetById(ctx, id)
	if err != nil {
		if errors.Is(err, serviceSystem.ErrPostNotFound) {
			response.NewResponse().Error(ctx, http.StatusBadRequest, "岗位信息不存在", nil)
			return
		}
		ctx.Set("internalError", fmt.Sprintf("获取岗位信息失败 >>> %v", err.Error()))
		zap.S().Error("获取岗位信息失败 >>> ", err.Error())
		response.NewResponse().Error(ctx, http.StatusInternalServerError, "服务器异常", nil)
		return
	}

	response.NewResponse().Success(ctx, "获取成功", detail)
}

// GetListPage
// @Summary 获取岗位分页列表
// @Description 获取岗位分页列表信息
// @Tags 系统管理/岗位管理
// @Accept application/json
// @Produce application/json
// @Security BearerAuth
// @Param page query int true "页码" default(1)
// @Param pageSize query int true "每页数量" default(10)
// @Param creator query string false "创建人"
// @Param modifier query string false "修改人"
// @Param status query bool false "状态" default(true)
// @Param name query string false "岗位名称"
// @Param code query string false "岗位编码"
// @Param post_type query int false "岗位类型" default(0)
// @Param level query int true "岗位级别" default(0)
// @Param dept_id query string true "所属部门ID"
// @Success 200 {array} []domainSystem.Post
// @Success 200 {object} PostListPageResponse
// @Failure 400 {object} response.Response
// @Router /v1/system/post/listPage [get]
// @Security LoginToken
func (h *postHandler) GetListPage(ctx *gin.Context) {
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
	postType, _ := strconv.Atoi(ctx.DefaultQuery("post_type", "0"))
	level, _ := strconv.Atoi(ctx.DefaultQuery("level", "0"))
	deptID := ctx.DefaultQuery("dept_id", "")

	filter := domainSystem.PostFilter{
		Pagination: filters.Pagination{
			Page:     page,
			PageSize: pageSize,
		},
		Filters: filters.Filters{
			Creator:    creator,
			Modifier:   modifier,
			BelongDept: *user.DeptID,
		},
		Status:   status,
		Name:     name,
		Code:     code,
		PostType: post.Type(postType),
		Level:    post.Level(level),
		DeptID:   deptID,
	}

	list, total, err := h.svc.GetListPage(ctx, filter)
	if err != nil {
		ctx.Set("internalError", fmt.Sprintf("获取岗位分页列表异常 >>> %v", err.Error()))
		zap.S().Error("获取岗位分页列表异常 >>> ", err.Error())
		response.NewResponse().Error(ctx, http.StatusInternalServerError, "服务器异常", nil)
		return
	}

	response.NewResponse().Success(ctx, "查询成功", PostListPageResponse{
		List:     list,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}

// GetListAll
// @Summary 获取所有岗位
// @Description 获取所有岗位列表信息
// @Tags 系统管理/岗位管理
// @Accept application/json
// @Produce application/json
// @Security BearerAuth
// @Param creator query string false "创建人"
// @Param modifier query string false "修改人"
// @Param status query bool false "状态" default(true)
// @Param name query string false "岗位名称"
// @Param code query string false "岗位编码"
// @Param post_type query int false "岗位类型" default(0)
// @Param level query int true "岗位级别" default(0)
// @Param dept_id query string true "所属部门ID"
// @Success 200 {array} []domainSystem.Post
// @Failure 400 {object} response.Response
// @Router /v1/system/post/listAll [get]
// @Security LoginToken
func (h *postHandler) GetListAll(ctx *gin.Context) {
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
	postType, _ := strconv.Atoi(ctx.DefaultQuery("post_type", "0"))
	level, _ := strconv.Atoi(ctx.DefaultQuery("level", "0"))
	deptID := ctx.DefaultQuery("dept_id", "")

	filter := domainSystem.PostFilter{
		Filters: filters.Filters{
			Creator:    creator,
			Modifier:   modifier,
			BelongDept: *user.DeptID,
		},
		Status:   status,
		Name:     name,
		Code:     code,
		PostType: post.Type(postType),
		Level:    post.Level(level),
		DeptID:   deptID,
	}

	list, err := h.svc.GetListAll(ctx, filter)
	if err != nil {
		ctx.Set("internalError", fmt.Sprintf("获取岗位列表异常 >>> %v", err.Error()))
		zap.S().Error("获取岗位列表异常 >>> ", err.Error())
		response.NewResponse().Error(ctx, http.StatusInternalServerError, "服务器异常", nil)
		return
	}

	response.NewResponse().Success(ctx, "查询成功", list)
}

// Export
// @Summary 导出岗位数据
// @Description 导出岗位数据到Excel文件
// @Tags 系统管理/岗位管理
// @Accept application/json
// @Produce application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
// @Security BearerAuth
// @Param creator query string false "创建人"
// @Param modifier query string false "修改人"
// @Param status query bool false "状态" default(true)
// @Param name query string false "岗位名称"
// @Param code query string false "岗位编码"
// @Param post_type query int false "岗位类型" default(0)
// @Param level query int true "岗位级别" default(0)
// @Param dept_id query string true "所属部门ID"
// @Success 200 {file} file "Excel文件"
// @Failure 500 {object} response.Response
// @Router /v1/system/post/export [get]
// @Security LoginToken
func (h *postHandler) Export(ctx *gin.Context) {
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
	postType, _ := strconv.Atoi(ctx.DefaultQuery("post_type", "0"))
	level, _ := strconv.Atoi(ctx.DefaultQuery("level", "0"))
	deptID := ctx.DefaultQuery("dept_id", "")

	filter := domainSystem.PostFilter{
		Filters: filters.Filters{
			Creator:    creator,
			Modifier:   modifier,
			BelongDept: *user.DeptID,
		},
		Status:   status,
		Name:     name,
		Code:     code,
		PostType: post.Type(postType),
		Level:    post.Level(level),
		DeptID:   deptID,
	}

	list, err := h.svc.GetListAll(ctx, filter)
	if err != nil {
		ctx.Set("internalError", fmt.Sprintf("获取岗位列表异常 >>> %v", err.Error()))
		zap.S().Error("获取岗位列表异常 >>> ", err.Error())
		response.NewResponse().Error(ctx, http.StatusInternalServerError, "服务器异常", nil)
		return
	}

	// 准备导出配置
	filename := fmt.Sprintf("岗位信息导出_%s.xlsx", time.Now().Format("20060102150405"))
	cfg := excelutil.ExcelExportConfig{
		SheetName:  "岗位信息",
		FileName:   filename,
		StreamMode: true,
		Columns: []excelutil.ExcelColumn{
			{Title: "岗位名称", Field: "Name", Width: 20},
			{Title: "岗位编码", Field: "Code", Width: 20},
			{
				Title: "岗位类型",
				Field: "PostType",
				Width: 15,
				Formatter: func(value interface{}) string {
					typeValidValues := []string{"管理岗", "技术岗", "业务岗", "职能岗", "其他"}
					converter := enumconv.NewEnumConverter(post.TypeMapping, post.TypeImportMapping, typeValidValues, "岗位类型")
					str, _ := converter.FromEnum(value.(post.Type))
					return str
				},
			},
			{
				Title: "岗位级别",
				Field: "Level",
				Width: 15,
				Formatter: func(value interface{}) string {
					levelValidValues := []string{"高层", "中层", "基层", "一般员工"}
					levelConverter := enumconv.NewEnumConverter(post.LevelMapping, post.LevelImportMapping, levelValidValues, "岗位级别")
					str, _ := levelConverter.FromEnum(value.(post.Level))
					return str
				},
			},
			{Title: "岗位描述", Field: "Description", Width: 25},
			{Title: "部门名称", Field: "Dept.Name", Width: 20},
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
		ctx.Set("internalError", fmt.Sprintf("导出岗位信息异常 >>> %v", err.Error()))
		zap.S().Error("导出岗位信息异常 >>> ", err.Error())
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
