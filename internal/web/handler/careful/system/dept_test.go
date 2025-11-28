/**
 * Description：
 * FileName：dept_test.go.go
 * Author：CJiaの用心
 * Create：2025/11/28 01:05:35
 * Remark：
 */

package system

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/carefuly/careful-admin-go-gin/config"
	domainSystem "github.com/carefuly/careful-admin-go-gin/internal/domain/careful/system"
	"github.com/carefuly/careful-admin-go-gin/internal/model/careful/system"
	svcmocks "github.com/carefuly/careful-admin-go-gin/internal/service/careful/mocks"
	serviceSystem "github.com/carefuly/careful-admin-go-gin/internal/service/careful/system"
	"github.com/carefuly/careful-admin-go-gin/pkg/ginx/response"
	"github.com/carefuly/careful-admin-go-gin/pkg/models"
	ijwt "github.com/carefuly/careful-admin-go-gin/pkg/utils/jwt"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test_deptHandler_Create(t *testing.T) {
	c := config.RelyConfig{}

	testCases := []struct {
		name     string
		mock     func(ctrl *gomock.Controller) (serviceSystem.DeptService, serviceSystem.UserService)
		reqBody  string
		wantCode int
		wantMsg  string
	}{
		{
			name: "新增成功",
			mock: func(ctrl *gomock.Controller) (serviceSystem.DeptService, serviceSystem.UserService) {
				deptService := svcmocks.NewMockDeptService(ctrl)
				userService := svcmocks.NewMockUserService(ctrl)
				userService.EXPECT().GetById(gomock.Any(), "1").Return(domainSystem.User{
					User: system.User{
						CoreModels: models.CoreModels{Id: "1"},
					},
				}, nil)
				deptService.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)
				return deptService, userService
			},
			reqBody: `
{
	"status": true,
	"name": "部门名称",
	"code": "部门编码",
	"dept_type": "department",
	"owner": "",
	"phone": "",
	"email": "",
	"description": "",
	"parent_id": "1",
	"sort": 1,
	"remark": ""
}
`,
			wantCode: http.StatusOK,
			wantMsg:  "新增成功",
		},
		{
			name: "获取凭证，但用户不存在",
			mock: func(ctrl *gomock.Controller) (serviceSystem.DeptService, serviceSystem.UserService) {
				deptService := svcmocks.NewMockDeptService(ctrl)
				userService := svcmocks.NewMockUserService(ctrl)
				userService.EXPECT().GetById(gomock.Any(), "1").Return(domainSystem.User{
					User: system.User{},
				}, errors.New("获取凭证，但用户不存在"))
				return deptService, userService
			},
			reqBody: `
{
	"status": true,
	"name": "部门名称",
	"code": "部门编码",
	"dept_type": "department",
	"owner": "",
	"phone": "",
	"email": "",
	"description": "",
	"parent_id": "1",
	"sort": 1,
	"remark": ""
}
`,
			wantCode: http.StatusInternalServerError,
			wantMsg:  "服务器异常",
		},
		{
			name: "dept_type参数不匹配",
			mock: func(ctrl *gomock.Controller) (serviceSystem.DeptService, serviceSystem.UserService) {
				deptService := svcmocks.NewMockDeptService(ctrl)
				userService := svcmocks.NewMockUserService(ctrl)
				userService.EXPECT().GetById(gomock.Any(), "1").Return(domainSystem.User{
					User: system.User{
						CoreModels: models.CoreModels{Id: "1"},
					},
				}, nil)
				return deptService, userService
			},
			reqBody: `
{
	"status": true,
	"name": "部门名称",
	"code": "部门编码",
	"dept_type": "部门类型",
	"owner": "",
	"phone": "",
	"email": "",
	"description": "",
	"parent_id": "1",
	"sort": 1,
	"remark": ""
}
`,
			wantCode: http.StatusBadRequest,
			wantMsg:  "无效的部门类型枚举值: 部门类型",
		},
		{
			name: "部门编码已存在",
			mock: func(ctrl *gomock.Controller) (serviceSystem.DeptService, serviceSystem.UserService) {
				deptService := svcmocks.NewMockDeptService(ctrl)
				userService := svcmocks.NewMockUserService(ctrl)
				userService.EXPECT().GetById(gomock.Any(), "1").Return(domainSystem.User{
					User: system.User{
						CoreModels: models.CoreModels{Id: "1"},
					},
				}, nil)
				deptService.EXPECT().Create(gomock.Any(), gomock.Any()).Return(serviceSystem.ErrDeptCodeDuplicate)
				return deptService, userService
			},
			reqBody: `
{
	"status": true,
	"name": "部门名称",
	"code": "部门编码",
	"dept_type": "department",
	"owner": "",
	"phone": "",
	"email": "",
	"description": "",
	"parent_id": "1",
	"sort": 1,
	"remark": ""
}
`,
			wantCode: http.StatusBadRequest,
			wantMsg:  "部门编码已存在",
		},
		{
			name: "同级别下已存在相同的部门信息",
			mock: func(ctrl *gomock.Controller) (serviceSystem.DeptService, serviceSystem.UserService) {
				deptService := svcmocks.NewMockDeptService(ctrl)
				userService := svcmocks.NewMockUserService(ctrl)
				userService.EXPECT().GetById(gomock.Any(), "1").Return(domainSystem.User{
					User: system.User{
						CoreModels: models.CoreModels{Id: "1"},
					},
				}, nil)
				deptService.EXPECT().Create(gomock.Any(), gomock.Any()).Return(serviceSystem.ErrDeptNameParentDuplicate)
				return deptService, userService
			},
			reqBody: `
{
	"status": true,
	"name": "部门名称",
	"code": "部门编码",
	"dept_type": "department",
	"owner": "",
	"phone": "",
	"email": "",
	"description": "",
	"parent_id": "1",
	"sort": 1,
	"remark": ""
}
`,
			wantCode: http.StatusBadRequest,
			wantMsg:  "同级别下已存在相同的部门信息",
		},
		{
			name: "父部门信息不存在",
			mock: func(ctrl *gomock.Controller) (serviceSystem.DeptService, serviceSystem.UserService) {
				deptService := svcmocks.NewMockDeptService(ctrl)
				userService := svcmocks.NewMockUserService(ctrl)
				userService.EXPECT().GetById(gomock.Any(), "1").Return(domainSystem.User{
					User: system.User{
						CoreModels: models.CoreModels{Id: "1"},
					},
				}, nil)
				deptService.EXPECT().Create(gomock.Any(), gomock.Any()).Return(serviceSystem.ErrDeptParentNotFound)
				return deptService, userService
			},
			reqBody: `
{
	"status": true,
	"name": "部门名称",
	"code": "部门编码",
	"dept_type": "department",
	"owner": "",
	"phone": "",
	"email": "",
	"description": "",
	"parent_id": "1",
	"sort": 1,
	"remark": ""
}
`,
			wantCode: http.StatusBadRequest,
			wantMsg:  "父部门信息不存在",
		},
		{
			name: "父部门已被禁用，无法在其下创建子部门",
			mock: func(ctrl *gomock.Controller) (serviceSystem.DeptService, serviceSystem.UserService) {
				deptService := svcmocks.NewMockDeptService(ctrl)
				userService := svcmocks.NewMockUserService(ctrl)
				userService.EXPECT().GetById(gomock.Any(), "1").Return(domainSystem.User{
					User: system.User{
						CoreModels: models.CoreModels{Id: "1"},
					},
				}, nil)
				deptService.EXPECT().Create(gomock.Any(), gomock.Any()).Return(serviceSystem.ErrDeptDisabled)
				return deptService, userService
			},
			reqBody: `
{
	"status": true,
	"name": "部门名称",
	"code": "部门编码",
	"dept_type": "department",
	"owner": "",
	"phone": "",
	"email": "",
	"description": "",
	"parent_id": "1",
	"sort": 1,
	"remark": ""
}
`,
			wantCode: http.StatusBadRequest,
			wantMsg:  "父部门已被禁用，无法在其下创建子部门",
		},
		{
			name: "服务器异常",
			mock: func(ctrl *gomock.Controller) (serviceSystem.DeptService, serviceSystem.UserService) {
				deptService := svcmocks.NewMockDeptService(ctrl)
				userService := svcmocks.NewMockUserService(ctrl)
				userService.EXPECT().GetById(gomock.Any(), "1").Return(domainSystem.User{
					User: system.User{
						CoreModels: models.CoreModels{Id: "1"},
					},
				}, nil)
				deptService.EXPECT().Create(gomock.Any(), gomock.Any()).Return(errors.New("服务器异常"))
				return deptService, userService
			},
			reqBody: `
{
	"status": true,
	"name": "部门名称",
	"code": "部门编码",
	"dept_type": "department",
	"owner": "",
	"phone": "",
	"email": "",
	"description": "",
	"parent_id": "1",
	"sort": 1,
	"remark": ""
}
`,
			wantCode: http.StatusInternalServerError,
			wantMsg:  "服务器异常",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			server := gin.Default()
			// 设置登录凭证
			server.Use(func(ctx *gin.Context) {
				ctx.Set("claims", &ijwt.Claims{
					UserID: "1", // 避免uuid开销过大
				})
			})
			router := server.Group("/api/v1")
			service, userService := tc.mock(ctrl)
			h := NewDeptHandler(c, service, userService)
			h.RegisterRoutes(router)

			req, err := http.NewRequest(http.MethodPost,
				"/api/v1/dept/create",
				bytes.NewBuffer([]byte(tc.reqBody)))
			require.NoError(t, err)
			// 数据是 JSON 格式
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()
			// 响应写回到 resp 里
			server.ServeHTTP(resp, req)

			var res response.Response
			err = json.Unmarshal(resp.Body.Bytes(), &res)
			require.NoError(t, err)
			assert.Equal(t, tc.wantCode, resp.Code)
			assert.Equal(t, tc.wantMsg, res.Message)
		})
	}
}

func Test_deptHandler_Delete(t *testing.T) {
	c := config.RelyConfig{}

	testCases := []struct {
		name     string
		mock     func(ctrl *gomock.Controller) serviceSystem.DeptService
		id       string
		wantCode int
		wantMsg  string
	}{
		{
			name: "删除成功",
			mock: func(ctrl *gomock.Controller) serviceSystem.DeptService {
				deptService := svcmocks.NewMockDeptService(ctrl)
				deptService.EXPECT().Delete(gomock.Any(), "1").Return(nil)
				return deptService
			},
			id:       "1",
			wantCode: http.StatusOK,
			wantMsg:  "删除成功",
		},
		{
			name: "部门信息不存在",
			mock: func(ctrl *gomock.Controller) serviceSystem.DeptService {
				deptService := svcmocks.NewMockDeptService(ctrl)
				deptService.EXPECT().Delete(gomock.Any(), "1").Return(serviceSystem.ErrDeptNotFound)
				return deptService
			},
			id:       "1",
			wantCode: http.StatusBadRequest,
			wantMsg:  "部门信息不存在",
		},
		{
			name: "部门含有子部门，无法删除",
			mock: func(ctrl *gomock.Controller) serviceSystem.DeptService {
				deptService := svcmocks.NewMockDeptService(ctrl)
				deptService.EXPECT().Delete(gomock.Any(), "1").Return(serviceSystem.ErrDeptHasChildren)
				return deptService
			},
			id:       "1",
			wantCode: http.StatusBadRequest,
			wantMsg:  "部门含有子部门，无法删除",
		},
		{
			name: "部门下仍有用户，无法删除",
			mock: func(ctrl *gomock.Controller) serviceSystem.DeptService {
				deptService := svcmocks.NewMockDeptService(ctrl)
				deptService.EXPECT().Delete(gomock.Any(), "1").Return(serviceSystem.ErrDeptHasUsers)
				return deptService
			},
			id:       "1",
			wantCode: http.StatusBadRequest,
			wantMsg:  "部门下仍有用户，无法删除",
		},
		{
			name: "服务器异常",
			mock: func(ctrl *gomock.Controller) serviceSystem.DeptService {
				deptService := svcmocks.NewMockDeptService(ctrl)
				deptService.EXPECT().Delete(gomock.Any(), "1").Return(errors.New("服务器异常"))
				return deptService
			},
			id:       "1",
			wantCode: http.StatusInternalServerError,
			wantMsg:  "服务器异常",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			server := gin.Default()
			// 设置登录凭证
			server.Use(func(ctx *gin.Context) {
				ctx.Set("claims", &ijwt.Claims{
					UserID: "1", // 避免uuid开销过大
				})
			})
			router := server.Group("/api/v1")
			service := tc.mock(ctrl)
			h := NewDeptHandler(c, service, nil)
			h.RegisterRoutes(router)

			req, err := http.NewRequest(http.MethodDelete,
				"/api/v1/dept/delete/"+tc.id,
				bytes.NewBuffer([]byte("")))
			require.NoError(t, err)
			// 数据是 JSON 格式
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()
			// 响应写回到 resp 里
			server.ServeHTTP(resp, req)

			var res response.Response
			err = json.Unmarshal(resp.Body.Bytes(), &res)
			require.NoError(t, err)
			assert.Equal(t, tc.wantCode, resp.Code)
			assert.Equal(t, tc.wantMsg, res.Message)
		})
	}
}

func Test_deptHandler_Update(t *testing.T) {
	c := config.RelyConfig{}

	testCases := []struct {
		name     string
		mock     func(ctrl *gomock.Controller) (serviceSystem.DeptService, serviceSystem.UserService)
		reqBody  string
		wantCode int
		wantMsg  string
	}{
		{
			name: "更新成功",
			mock: func(ctrl *gomock.Controller) (serviceSystem.DeptService, serviceSystem.UserService) {
				deptService := svcmocks.NewMockDeptService(ctrl)
				userService := svcmocks.NewMockUserService(ctrl)
				userService.EXPECT().GetById(gomock.Any(), "1").Return(domainSystem.User{
					User: system.User{
						CoreModels: models.CoreModels{Id: "1"},
					},
				}, nil)
				deptService.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)
				return deptService, userService
			},
			reqBody: `
{
	"id": "1",
	"status": true,
	"name": "部门名称",
	"code": "部门编码",
	"dept_type": "department",
	"owner": "",
	"phone": "",
	"email": "",
	"description": "",
	"parent_id": "1",
	"sort": 1,
	"timestamp": 1,
	"remark": ""
}
`,
			wantCode: http.StatusOK,
			wantMsg:  "更新成功",
		},
		{
			name: "获取凭证，但用户不存在",
			mock: func(ctrl *gomock.Controller) (serviceSystem.DeptService, serviceSystem.UserService) {
				deptService := svcmocks.NewMockDeptService(ctrl)
				userService := svcmocks.NewMockUserService(ctrl)
				userService.EXPECT().GetById(gomock.Any(), "1").Return(domainSystem.User{
					User: system.User{},
				}, errors.New("获取凭证，但用户不存在"))
				return deptService, userService
			},
			reqBody: `
{
	"id": "1",
	"status": true,
	"name": "部门名称",
	"code": "部门编码",
	"dept_type": "department",
	"owner": "",
	"phone": "",
	"email": "",
	"description": "",
	"parent_id": "1",
	"sort": 1,
	"timestamp": 1,
	"remark": ""
}
`,
			wantCode: http.StatusInternalServerError,
			wantMsg:  "服务器异常",
		},
		{
			name: "部门编码已存在",
			mock: func(ctrl *gomock.Controller) (serviceSystem.DeptService, serviceSystem.UserService) {
				deptService := svcmocks.NewMockDeptService(ctrl)
				userService := svcmocks.NewMockUserService(ctrl)
				userService.EXPECT().GetById(gomock.Any(), "1").Return(domainSystem.User{
					User: system.User{
						CoreModels: models.CoreModels{Id: "1"},
					},
				}, nil)
				deptService.EXPECT().Update(gomock.Any(), gomock.Any()).Return(serviceSystem.ErrDeptCodeDuplicate)
				return deptService, userService
			},
			reqBody: `
{
	"id": "1",
	"status": true,
	"name": "部门名称",
	"code": "部门编码",
	"dept_type": "department",
	"owner": "",
	"phone": "",
	"email": "",
	"description": "",
	"parent_id": "1",
	"sort": 1,
	"timestamp": 1,
	"remark": ""
}
`,
			wantCode: http.StatusBadRequest,
			wantMsg:  "部门编码已存在",
		},
		{
			name: "同级别下已存在相同的部门信息",
			mock: func(ctrl *gomock.Controller) (serviceSystem.DeptService, serviceSystem.UserService) {
				deptService := svcmocks.NewMockDeptService(ctrl)
				userService := svcmocks.NewMockUserService(ctrl)
				userService.EXPECT().GetById(gomock.Any(), "1").Return(domainSystem.User{
					User: system.User{
						CoreModels: models.CoreModels{Id: "1"},
					},
				}, nil)
				deptService.EXPECT().Update(gomock.Any(), gomock.Any()).Return(serviceSystem.ErrDeptNameParentDuplicate)
				return deptService, userService
			},
			reqBody: `
{
	"id": "1",
	"status": true,
	"name": "部门名称",
	"code": "部门编码",
	"dept_type": "department",
	"owner": "",
	"phone": "",
	"email": "",
	"description": "",
	"parent_id": "1",
	"sort": 1,
	"timestamp": 1,
	"remark": ""
}
`,
			wantCode: http.StatusBadRequest,
			wantMsg:  "同级别下已存在相同的部门信息",
		},
		{
			name: "不能将自己设置为父部门",
			mock: func(ctrl *gomock.Controller) (serviceSystem.DeptService, serviceSystem.UserService) {
				deptService := svcmocks.NewMockDeptService(ctrl)
				userService := svcmocks.NewMockUserService(ctrl)
				userService.EXPECT().GetById(gomock.Any(), "1").Return(domainSystem.User{
					User: system.User{
						CoreModels: models.CoreModels{Id: "1"},
					},
				}, nil)
				deptService.EXPECT().Update(gomock.Any(), gomock.Any()).Return(serviceSystem.ErrDeptYourParent)
				return deptService, userService
			},
			reqBody: `
{
	"id": "1",
	"status": true,
	"name": "部门名称",
	"code": "部门编码",
	"dept_type": "department",
	"owner": "",
	"phone": "",
	"email": "",
	"description": "",
	"parent_id": "1",
	"sort": 1,
	"timestamp": 1,
	"remark": ""
}
`,
			wantCode: http.StatusBadRequest,
			wantMsg:  "不能将自己设置为父部门",
		},
		{
			name: "父部门信息不存在",
			mock: func(ctrl *gomock.Controller) (serviceSystem.DeptService, serviceSystem.UserService) {
				deptService := svcmocks.NewMockDeptService(ctrl)
				userService := svcmocks.NewMockUserService(ctrl)
				userService.EXPECT().GetById(gomock.Any(), "1").Return(domainSystem.User{
					User: system.User{
						CoreModels: models.CoreModels{Id: "1"},
					},
				}, nil)
				deptService.EXPECT().Update(gomock.Any(), gomock.Any()).Return(serviceSystem.ErrDeptParentNotFound)
				return deptService, userService
			},
			reqBody: `
{
	"id": "1",
	"status": true,
	"name": "部门名称",
	"code": "部门编码",
	"dept_type": "department",
	"owner": "",
	"phone": "",
	"email": "",
	"description": "",
	"parent_id": "1",
	"sort": 1,
	"timestamp": 1,
	"remark": ""
}
`,
			wantCode: http.StatusBadRequest,
			wantMsg:  "父部门信息不存在",
		},
		{
			name: "父部门已被禁用，无法在其下创建子部门",
			mock: func(ctrl *gomock.Controller) (serviceSystem.DeptService, serviceSystem.UserService) {
				deptService := svcmocks.NewMockDeptService(ctrl)
				userService := svcmocks.NewMockUserService(ctrl)
				userService.EXPECT().GetById(gomock.Any(), "1").Return(domainSystem.User{
					User: system.User{
						CoreModels: models.CoreModels{Id: "1"},
					},
				}, nil)
				deptService.EXPECT().Update(gomock.Any(), gomock.Any()).Return(serviceSystem.ErrDeptDisabled)
				return deptService, userService
			},
			reqBody: `
{
	"id": "1",
	"status": true,
	"name": "部门名称",
	"code": "部门编码",
	"dept_type": "department",
	"owner": "",
	"phone": "",
	"email": "",
	"description": "",
	"parent_id": "1",
	"sort": 1,
	"timestamp": 1,
	"remark": ""
}
`,
			wantCode: http.StatusBadRequest,
			wantMsg:  "父部门已被禁用，无法在其下创建子部门",
		},
		{
			name: "不能将子部门设置为父部门，会形成循环引用",
			mock: func(ctrl *gomock.Controller) (serviceSystem.DeptService, serviceSystem.UserService) {
				deptService := svcmocks.NewMockDeptService(ctrl)
				userService := svcmocks.NewMockUserService(ctrl)
				userService.EXPECT().GetById(gomock.Any(), "1").Return(domainSystem.User{
					User: system.User{
						CoreModels: models.CoreModels{Id: "1"},
					},
				}, nil)
				deptService.EXPECT().Update(gomock.Any(), gomock.Any()).Return(serviceSystem.ErrDeptCycleReference)
				return deptService, userService
			},
			reqBody: `
{
	"id": "1",
	"status": true,
	"name": "部门名称",
	"code": "部门编码",
	"dept_type": "department",
	"owner": "",
	"phone": "",
	"email": "",
	"description": "",
	"parent_id": "1",
	"sort": 1,
	"timestamp": 1,
	"remark": ""
}
`,
			wantCode: http.StatusBadRequest,
			wantMsg:  "不能将子部门设置为父部门，会形成循环引用",
		},
		{
			name: "数据版本不一致，取消修改，请刷新后重试",
			mock: func(ctrl *gomock.Controller) (serviceSystem.DeptService, serviceSystem.UserService) {
				deptService := svcmocks.NewMockDeptService(ctrl)
				userService := svcmocks.NewMockUserService(ctrl)
				userService.EXPECT().GetById(gomock.Any(), "1").Return(domainSystem.User{
					User: system.User{
						CoreModels: models.CoreModels{Id: "1"},
					},
				}, nil)
				deptService.EXPECT().Update(gomock.Any(), gomock.Any()).Return(serviceSystem.ErrDeptVersionInconsistency)
				return deptService, userService
			},
			reqBody: `
{
	"id": "1",
	"status": true,
	"name": "部门名称",
	"code": "部门编码",
	"dept_type": "department",
	"owner": "",
	"phone": "",
	"email": "",
	"description": "",
	"parent_id": "1",
	"sort": 1,
	"timestamp": 1,
	"remark": ""
}
`,
			wantCode: http.StatusBadRequest,
			wantMsg:  "数据版本不一致，取消修改，请刷新后重试",
		},
		{
			name: "服务器异常",
			mock: func(ctrl *gomock.Controller) (serviceSystem.DeptService, serviceSystem.UserService) {
				deptService := svcmocks.NewMockDeptService(ctrl)
				userService := svcmocks.NewMockUserService(ctrl)
				userService.EXPECT().GetById(gomock.Any(), "1").Return(domainSystem.User{
					User: system.User{
						CoreModels: models.CoreModels{Id: "1"},
					},
				}, nil)
				deptService.EXPECT().Update(gomock.Any(), gomock.Any()).Return(errors.New("服务器异常"))
				return deptService, userService
			},
			reqBody: `
{
	"id": "1",
	"status": true,
	"name": "部门名称",
	"code": "部门编码",
	"dept_type": "department",
	"owner": "",
	"phone": "",
	"email": "",
	"description": "",
	"parent_id": "1",
	"sort": 1,
	"timestamp": 1,
	"remark": ""
}
`,
			wantCode: http.StatusInternalServerError,
			wantMsg:  "服务器异常",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			server := gin.Default()
			// 设置登录凭证
			server.Use(func(ctx *gin.Context) {
				ctx.Set("claims", &ijwt.Claims{
					UserID: "1", // 避免uuid开销过大
				})
			})
			router := server.Group("/api/v1")
			service, userService := tc.mock(ctrl)
			h := NewDeptHandler(c, service, userService)
			h.RegisterRoutes(router)

			req, err := http.NewRequest(http.MethodPut,
				"/api/v1/dept/update",
				bytes.NewBuffer([]byte(tc.reqBody)))
			require.NoError(t, err)
			// 数据是 JSON 格式
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()
			// 响应写回到 resp 里
			server.ServeHTTP(resp, req)

			var res response.Response
			err = json.Unmarshal(resp.Body.Bytes(), &res)
			require.NoError(t, err)
			assert.Equal(t, tc.wantCode, resp.Code)
			assert.Equal(t, tc.wantMsg, res.Message)
		})
	}
}

func Test_deptHandler_GetById(t *testing.T) {
	c := config.RelyConfig{}

	testCases := []struct {
		name     string
		mock     func(ctrl *gomock.Controller) serviceSystem.DeptService
		id       string
		wantCode int
		wantMsg  string
	}{
		{
			name: "获取成功",
			mock: func(ctrl *gomock.Controller) serviceSystem.DeptService {
				deptService := svcmocks.NewMockDeptService(ctrl)
				deptService.EXPECT().GetById(gomock.Any(), "1").
					Return(domainSystem.Dept{
						Dept: system.Dept{
							Name: "部门名称",
						},
					}, nil)
				return deptService
			},
			id:       "1",
			wantCode: http.StatusOK,
			wantMsg:  "获取成功",
		},
		{
			name: "部门信息不存在",
			mock: func(ctrl *gomock.Controller) serviceSystem.DeptService {
				deptService := svcmocks.NewMockDeptService(ctrl)
				deptService.EXPECT().GetById(gomock.Any(), "1").
					Return(domainSystem.Dept{
						Dept: system.Dept{
							Name: "部门名称",
						},
					}, serviceSystem.ErrDeptNotFound)
				return deptService
			},
			id:       "1",
			wantCode: http.StatusBadRequest,
			wantMsg:  "部门信息不存在",
		},
		{
			name: "服务器异常",
			mock: func(ctrl *gomock.Controller) serviceSystem.DeptService {
				deptService := svcmocks.NewMockDeptService(ctrl)
				deptService.EXPECT().GetById(gomock.Any(), "1").
					Return(domainSystem.Dept{
						Dept: system.Dept{
							Name: "部门名称",
						},
					}, errors.New("服务器异常"))
				return deptService
			},
			id:       "1",
			wantCode: http.StatusInternalServerError,
			wantMsg:  "服务器异常",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			server := gin.Default()
			// 设置登录凭证
			server.Use(func(ctx *gin.Context) {
				ctx.Set("claims", &ijwt.Claims{
					UserID: "1", // 避免uuid开销过大
				})
			})
			router := server.Group("/api/v1")
			service := tc.mock(ctrl)
			h := NewDeptHandler(c, service, nil)
			h.RegisterRoutes(router)

			req, err := http.NewRequest(http.MethodGet,
				"/api/v1/dept/getById/"+tc.id,
				bytes.NewBuffer([]byte("")))
			require.NoError(t, err)
			// 数据是 JSON 格式
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()
			// 响应写回到 resp 里
			server.ServeHTTP(resp, req)

			var res response.Response
			err = json.Unmarshal(resp.Body.Bytes(), &res)
			require.NoError(t, err)
			assert.Equal(t, tc.wantCode, resp.Code)
			assert.Equal(t, tc.wantMsg, res.Message)
		})
	}
}
