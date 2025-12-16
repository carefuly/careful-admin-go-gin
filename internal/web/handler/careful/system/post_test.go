/**
 * Description：
 * FileName：post_test.go.go
 * Author：CJiaの用心
 * Create：2025/12/2 00:35:44
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

func Test_postHandler_Create(t *testing.T) {
	c := config.RelyConfig{}

	testCases := []struct {
		name     string
		mock     func(ctrl *gomock.Controller) (serviceSystem.PostService, serviceSystem.UserService)
		reqBody  string
		wantCode int
		wantMsg  string
	}{
		{
			name: "新增成功",
			mock: func(ctrl *gomock.Controller) (serviceSystem.PostService, serviceSystem.UserService) {
				postService := svcmocks.NewMockPostService(ctrl)
				userService := svcmocks.NewMockUserService(ctrl)
				userService.EXPECT().GetById(gomock.Any(), "1").Return(domainSystem.User{
					User: system.User{
						CoreModels: models.CoreModels{Id: "1"},
					},
				}, nil)
				postService.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)
				return postService, userService
			},
			reqBody: `
{
	"status": true,
	"name": "岗位名称",
	"code": "岗位编码",
	"post_type": 1,
	"level": 1,
	"description": "",
	"dept_id": null,
	"sort": 1,
	"remark": ""
}
`,
			wantCode: http.StatusOK,
			wantMsg:  "新增成功",
		},
		{
			name: "获取凭证，但用户不存在",
			mock: func(ctrl *gomock.Controller) (serviceSystem.PostService, serviceSystem.UserService) {
				postService := svcmocks.NewMockPostService(ctrl)
				userService := svcmocks.NewMockUserService(ctrl)
				userService.EXPECT().GetById(gomock.Any(), "1").Return(domainSystem.User{
					User: system.User{
						CoreModels: models.CoreModels{Id: "1"},
					},
				}, errors.New("获取凭证，但用户不存在"))
				return postService, userService
			},
			reqBody: `
{
	"status": true,
	"name": "岗位名称",
	"code": "岗位编码",
	"post_type": 1,
	"level": 1,
	"description": "",
	"dept_id": null,
	"sort": 1,
	"remark": ""
}
`,
			wantCode: http.StatusInternalServerError,
			wantMsg:  "服务器异常",
		},
		{
			name: "post_type参数不匹配",
			mock: func(ctrl *gomock.Controller) (serviceSystem.PostService, serviceSystem.UserService) {
				postService := svcmocks.NewMockPostService(ctrl)
				userService := svcmocks.NewMockUserService(ctrl)
				userService.EXPECT().GetById(gomock.Any(), "1").Return(domainSystem.User{
					User: system.User{
						CoreModels: models.CoreModels{Id: "1"},
					},
				}, nil)
				return postService, userService
			},
			reqBody: `
{
	"status": true,
	"name": "岗位名称",
	"code": "岗位编码",
	"post_type": 6,
	"level": 1,
	"description": "",
	"dept_id": null,
	"sort": 1,
	"remark": ""
}
`,
			wantCode: http.StatusBadRequest,
			wantMsg:  "无效的岗位类型枚举值: 6",
		},
		{
			name: "level参数不匹配",
			mock: func(ctrl *gomock.Controller) (serviceSystem.PostService, serviceSystem.UserService) {
				postService := svcmocks.NewMockPostService(ctrl)
				userService := svcmocks.NewMockUserService(ctrl)
				userService.EXPECT().GetById(gomock.Any(), "1").Return(domainSystem.User{
					User: system.User{
						CoreModels: models.CoreModels{Id: "1"},
					},
				}, nil)
				return postService, userService
			},
			reqBody: `
{
	"status": true,
	"name": "岗位名称",
	"code": "岗位编码",
	"post_type": 5,
	"level": 5,
	"description": "",
	"dept_id": null,
	"sort": 1,
	"remark": ""
}
`,
			wantCode: http.StatusBadRequest,
			wantMsg:  "无效的岗位级别枚举值: 5",
		},
		{
			name: "岗位编码已存在",
			mock: func(ctrl *gomock.Controller) (serviceSystem.PostService, serviceSystem.UserService) {
				postService := svcmocks.NewMockPostService(ctrl)
				userService := svcmocks.NewMockUserService(ctrl)
				userService.EXPECT().GetById(gomock.Any(), "1").Return(domainSystem.User{
					User: system.User{
						CoreModels: models.CoreModels{Id: "1"},
					},
				}, nil)
				postService.EXPECT().Create(gomock.Any(), gomock.Any()).Return(serviceSystem.ErrPostCodeDuplicate)
				return postService, userService
			},
			reqBody: `
{
	"status": true,
	"name": "岗位名称",
	"code": "岗位编码",
	"post_type": 1,
	"level": 1,
	"description": "",
	"dept_id": null,
	"sort": 1,
	"remark": ""
}
`,
			wantCode: http.StatusBadRequest,
			wantMsg:  "岗位编码已存在",
		},
		{
			name: "服务器异常",
			mock: func(ctrl *gomock.Controller) (serviceSystem.PostService, serviceSystem.UserService) {
				postService := svcmocks.NewMockPostService(ctrl)
				userService := svcmocks.NewMockUserService(ctrl)
				userService.EXPECT().GetById(gomock.Any(), "1").Return(domainSystem.User{
					User: system.User{
						CoreModels: models.CoreModels{Id: "1"},
					},
				}, nil)
				postService.EXPECT().Create(gomock.Any(), gomock.Any()).Return(errors.New("服务器异常"))
				return postService, userService
			},
			reqBody: `
{
	"status": true,
	"name": "岗位名称",
	"code": "岗位编码",
	"post_type": 1,
	"level": 1,
	"description": "",
	"dept_id": null,
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
			h := NewPostHandler(c, service, userService)
			h.RegisterRoutes(router)

			req, err := http.NewRequest(http.MethodPost,
				"/api/v1/post/create",
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
			assert.Equal(t, tc.wantCode, res.Code)
			assert.Equal(t, tc.wantMsg, res.Message)
		})
	}
}

func Test_postHandler_Delete(t *testing.T) {
	c := config.RelyConfig{}

	testCases := []struct {
		name     string
		mock     func(ctrl *gomock.Controller) serviceSystem.PostService
		id       string
		wantCode int
		wantMsg  string
	}{
		{
			name: "删除成功",
			mock: func(ctrl *gomock.Controller) serviceSystem.PostService {
				postService := svcmocks.NewMockPostService(ctrl)
				postService.EXPECT().Delete(gomock.Any(), "1").Return(nil)
				return postService
			},
			id:       "1",
			wantCode: http.StatusOK,
			wantMsg:  "删除成功",
		},
		{
			name: "岗位下仍有用户，无法删除",
			mock: func(ctrl *gomock.Controller) serviceSystem.PostService {
				postService := svcmocks.NewMockPostService(ctrl)
				postService.EXPECT().Delete(gomock.Any(), "1").Return(serviceSystem.ErrPostHasUsers)
				return postService
			},
			id:       "1",
			wantCode: http.StatusBadRequest,
			wantMsg:  "岗位下仍有用户，无法删除",
		},
		{
			name: "服务器异常",
			mock: func(ctrl *gomock.Controller) serviceSystem.PostService {
				postService := svcmocks.NewMockPostService(ctrl)
				postService.EXPECT().Delete(gomock.Any(), "1").Return(errors.New("服务器异常"))
				return postService
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
			h := NewPostHandler(c, service, nil)
			h.RegisterRoutes(router)

			req, err := http.NewRequest(http.MethodDelete,
				"/api/v1/post/delete/"+tc.id,
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
			assert.Equal(t, tc.wantCode, res.Code)
			assert.Equal(t, tc.wantMsg, res.Message)
		})
	}
}

func Test_postHandler_Update(t *testing.T) {
	c := config.RelyConfig{}

	testCases := []struct {
		name     string
		mock     func(ctrl *gomock.Controller) (serviceSystem.PostService, serviceSystem.UserService)
		reqBody  string
		wantCode int
		wantMsg  string
	}{
		{
			name: "更新成功",
			mock: func(ctrl *gomock.Controller) (serviceSystem.PostService, serviceSystem.UserService) {
				postService := svcmocks.NewMockPostService(ctrl)
				userService := svcmocks.NewMockUserService(ctrl)
				userService.EXPECT().GetById(gomock.Any(), "1").Return(domainSystem.User{
					User: system.User{
						CoreModels: models.CoreModels{Id: "1"},
					},
				}, nil)
				postService.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)
				return postService, userService
			},
			reqBody: `
{
	"id": "1",
	"status": true,
	"name": "岗位名称",
	"code": "岗位编码",
	"post_type": 1,
	"level": 1,
	"description": "",
	"dept_id": "1",
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
			mock: func(ctrl *gomock.Controller) (serviceSystem.PostService, serviceSystem.UserService) {
				postService := svcmocks.NewMockPostService(ctrl)
				userService := svcmocks.NewMockUserService(ctrl)
				userService.EXPECT().GetById(gomock.Any(), "1").Return(domainSystem.User{
					User: system.User{
						CoreModels: models.CoreModels{Id: "1"},
					},
				}, errors.New("获取凭证，但用户不存在"))
				return postService, userService
			},
			reqBody: `
{
	"id": "1",
	"status": true,
	"name": "岗位名称",
	"code": "岗位编码",
	"post_type": 1,
	"level": 1,
	"description": "",
	"dept_id": "1",
	"sort": 1,
	"timestamp": 1,
	"remark": ""
}
`,
			wantCode: http.StatusInternalServerError,
			wantMsg:  "服务器异常",
		},
		{
			name: "post_type参数不匹配",
			mock: func(ctrl *gomock.Controller) (serviceSystem.PostService, serviceSystem.UserService) {
				postService := svcmocks.NewMockPostService(ctrl)
				userService := svcmocks.NewMockUserService(ctrl)
				userService.EXPECT().GetById(gomock.Any(), "1").Return(domainSystem.User{
					User: system.User{
						CoreModels: models.CoreModels{Id: "1"},
					},
				}, nil)
				return postService, userService
			},
			reqBody: `
{
	"id": "1",
	"status": true,
	"name": "岗位名称",
	"code": "岗位编码",
	"post_type": 6,
	"level": 1,
	"description": "",
	"dept_id": "1",
	"sort": 1,
	"timestamp": 1,
	"remark": ""
}
`,
			wantCode: http.StatusBadRequest,
			wantMsg:  "无效的岗位类型枚举值: 6",
		},
		{
			name: "level参数不匹配",
			mock: func(ctrl *gomock.Controller) (serviceSystem.PostService, serviceSystem.UserService) {
				postService := svcmocks.NewMockPostService(ctrl)
				userService := svcmocks.NewMockUserService(ctrl)
				userService.EXPECT().GetById(gomock.Any(), "1").Return(domainSystem.User{
					User: system.User{
						CoreModels: models.CoreModels{Id: "1"},
					},
				}, nil)
				return postService, userService
			},
			reqBody: `
{
	"id": "1",
	"status": true,
	"name": "岗位名称",
	"code": "岗位编码",
	"post_type": 1,
	"level": 5,
	"description": "",
	"dept_id": "1",
	"sort": 1,
	"timestamp": 1,
	"remark": ""
}
`,
			wantCode: http.StatusBadRequest,
			wantMsg:  "无效的岗位级别枚举值: 5",
		},
		{
			name: "岗位编码已存在",
			mock: func(ctrl *gomock.Controller) (serviceSystem.PostService, serviceSystem.UserService) {
				postService := svcmocks.NewMockPostService(ctrl)
				userService := svcmocks.NewMockUserService(ctrl)
				userService.EXPECT().GetById(gomock.Any(), "1").Return(domainSystem.User{
					User: system.User{
						CoreModels: models.CoreModels{Id: "1"},
					},
				}, nil)
				postService.EXPECT().Update(gomock.Any(), gomock.Any()).Return(serviceSystem.ErrPostCodeDuplicate)
				return postService, userService
			},
			reqBody: `
{
	"id": "1",
	"status": true,
	"name": "岗位名称",
	"code": "岗位编码",
	"post_type": 1,
	"level": 1,
	"description": "",
	"dept_id": "1",
	"sort": 1,
	"timestamp": 1,
	"remark": ""
}
`,
			wantCode: http.StatusBadRequest,
			wantMsg:  "岗位编码已存在",
		},
		{
			name: "数据版本不一致，取消修改，请刷新后重试",
			mock: func(ctrl *gomock.Controller) (serviceSystem.PostService, serviceSystem.UserService) {
				postService := svcmocks.NewMockPostService(ctrl)
				userService := svcmocks.NewMockUserService(ctrl)
				userService.EXPECT().GetById(gomock.Any(), "1").Return(domainSystem.User{
					User: system.User{
						CoreModels: models.CoreModels{Id: "1"},
					},
				}, nil)
				postService.EXPECT().Update(gomock.Any(), gomock.Any()).Return(serviceSystem.ErrPostVersionInconsistency)
				return postService, userService
			},
			reqBody: `
{
	"id": "1",
	"status": true,
	"name": "岗位名称",
	"code": "岗位编码",
	"post_type": 1,
	"level": 1,
	"description": "",
	"dept_id": "1",
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
			mock: func(ctrl *gomock.Controller) (serviceSystem.PostService, serviceSystem.UserService) {
				postService := svcmocks.NewMockPostService(ctrl)
				userService := svcmocks.NewMockUserService(ctrl)
				userService.EXPECT().GetById(gomock.Any(), "1").Return(domainSystem.User{
					User: system.User{
						CoreModels: models.CoreModels{Id: "1"},
					},
				}, nil)
				postService.EXPECT().Update(gomock.Any(), gomock.Any()).Return(errors.New("服务器异常"))
				return postService, userService
			},
			reqBody: `
{
	"id": "1",
	"status": true,
	"name": "岗位名称",
	"code": "岗位编码",
	"post_type": 1,
	"level": 1,
	"description": "",
	"dept_id": "1",
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
			h := NewPostHandler(c, service, userService)
			h.RegisterRoutes(router)

			req, err := http.NewRequest(http.MethodPut,
				"/api/v1/post/update",
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
			assert.Equal(t, tc.wantCode, res.Code)
			assert.Equal(t, tc.wantMsg, res.Message)
		})
	}
}

func Test_postHandler_GetById(t *testing.T) {
	c := config.RelyConfig{}

	testCases := []struct {
		name     string
		mock     func(ctrl *gomock.Controller) serviceSystem.PostService
		id       string
		wantCode int
		wantMsg  string
	}{
		{
			name: "获取成功",
			mock: func(ctrl *gomock.Controller) serviceSystem.PostService {
				postService := svcmocks.NewMockPostService(ctrl)
				postService.EXPECT().GetById(gomock.Any(), "1").
					Return(domainSystem.Post{
						Post: system.Post{
							Name: "岗位名称",
						},
					}, nil)
				return postService
			},
			id:       "1",
			wantCode: http.StatusOK,
			wantMsg:  "获取成功",
		},
		{
			name: "岗位不存在",
			mock: func(ctrl *gomock.Controller) serviceSystem.PostService {
				postService := svcmocks.NewMockPostService(ctrl)
				postService.EXPECT().GetById(gomock.Any(), "1").
					Return(domainSystem.Post{
						Post: system.Post{
							Name: "岗位名称",
						},
					}, serviceSystem.ErrPostNotFound)
				return postService
			},
			id:       "1",
			wantCode: http.StatusBadRequest,
			wantMsg:  "岗位不存在",
		},
		{
			name: "服务器异常",
			mock: func(ctrl *gomock.Controller) serviceSystem.PostService {
				postService := svcmocks.NewMockPostService(ctrl)
				postService.EXPECT().GetById(gomock.Any(), "1").
					Return(domainSystem.Post{
						Post: system.Post{
							Name: "岗位名称",
						},
					}, errors.New("服务器异常"))
				return postService
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
			h := NewPostHandler(c, service, nil)
			h.RegisterRoutes(router)

			req, err := http.NewRequest(http.MethodGet,
				"/api/v1/post/getById/"+tc.id,
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
			assert.Equal(t, tc.wantCode, res.Code)
			assert.Equal(t, tc.wantMsg, res.Message)
		})
	}
}
