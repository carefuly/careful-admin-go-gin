/**
 * Description：
 * FileName：dict_type_test.go.go
 * Author：CJiaの用心
 * Create：2025/10/20 10:26:57
 * Remark：
 */

package tools

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/carefuly/careful-admin-go-gin/config"
	domainSystem "github.com/carefuly/careful-admin-go-gin/internal/domain/careful/system"
	domainTools "github.com/carefuly/careful-admin-go-gin/internal/domain/careful/tools"
	modelSystem "github.com/carefuly/careful-admin-go-gin/internal/model/careful/system"
	"github.com/carefuly/careful-admin-go-gin/internal/model/careful/tools"
	svcmocks "github.com/carefuly/careful-admin-go-gin/internal/service/careful/mocks"
	serviceSystem "github.com/carefuly/careful-admin-go-gin/internal/service/careful/system"
	serviceTools "github.com/carefuly/careful-admin-go-gin/internal/service/careful/tools"
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

func Test_dictTypeHandler_Create(t *testing.T) {
	c := config.RelyConfig{}

	testCases := []struct {
		name     string
		mock     func(ctrl *gomock.Controller) (serviceTools.DictTypeService, serviceSystem.UserService)
		reqBody  string
		wantCode int
		wantMsg  string
	}{
		{
			name: "新增成功",
			mock: func(ctrl *gomock.Controller) (serviceTools.DictTypeService, serviceSystem.UserService) {
				dictTypeService := svcmocks.NewMockDictTypeService(ctrl)
				userService := svcmocks.NewMockUserService(ctrl)
				userService.EXPECT().GetById(gomock.Any(), "1").Return(domainSystem.User{
					User: modelSystem.User{
						CoreModels: models.CoreModels{Id: "1"},
					},
				}, nil)
				dictTypeService.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)
				return dictTypeService, userService
			},
			reqBody: `
{
	"name": "男",
	"strValue": "",
	"intValue": 1,
	"boolValue": false,
	"dictTag": "primary",
	"dictColor": "",
	"dict_id": "1",
	"sort": 1,
	"status": true,
	"remark": ""
}
`,
			wantCode: http.StatusOK,
			wantMsg:  "新增成功",
		},
		{
			name: "获取凭证，但用户不存在",
			mock: func(ctrl *gomock.Controller) (serviceTools.DictTypeService, serviceSystem.UserService) {
				dictTypeService := svcmocks.NewMockDictTypeService(ctrl)
				userService := svcmocks.NewMockUserService(ctrl)
				userService.EXPECT().GetById(gomock.Any(), "1").Return(domainSystem.User{
					User: modelSystem.User{},
				}, errors.New("获取凭证，但用户不存在"))
				return dictTypeService, userService
			},
			reqBody: `
{
	"name": "男",
	"strValue": "",
	"intValue": 1,
	"boolValue": false,
	"dictTag": "primary",
	"dictColor": "",
	"dict_id": "1",
	"sort": 1,
	"status": true,
	"remark": ""
}
`,
			wantCode: http.StatusInternalServerError,
			wantMsg:  "服务器异常",
		},
		{
			name: "dictTag参数不匹配",
			mock: func(ctrl *gomock.Controller) (serviceTools.DictTypeService, serviceSystem.UserService) {
				dictTypeService := svcmocks.NewMockDictTypeService(ctrl)
				userService := svcmocks.NewMockUserService(ctrl)
				userService.EXPECT().GetById(gomock.Any(), "1").Return(domainSystem.User{
					User: modelSystem.User{
						CoreModels: models.CoreModels{Id: "1"},
					},
				}, nil)
				return dictTypeService, userService
			},
			reqBody: `
{
	"name": "男",
	"strValue": "",
	"intValue": 1,
	"boolValue": false,
	"dictTag": "primary_",
	"dictColor": "",
	"dict_id": "1",
	"sort": 1,
	"status": true,
	"remark": ""
}
`,
			wantCode: http.StatusBadRequest,
			wantMsg:  "无效的标签类型枚举值: primary_",
		},
		{
			name: "同一字典下存在相同的字典项/值",
			mock: func(ctrl *gomock.Controller) (serviceTools.DictTypeService, serviceSystem.UserService) {
				dictTypeService := svcmocks.NewMockDictTypeService(ctrl)
				userService := svcmocks.NewMockUserService(ctrl)
				userService.EXPECT().GetById(gomock.Any(), "1").Return(domainSystem.User{
					User: modelSystem.User{
						CoreModels: models.CoreModels{Id: "1"},
					},
				}, nil)
				dictTypeService.EXPECT().Create(gomock.Any(), gomock.Any()).Return(serviceTools.ErrDictTypeDuplicate)
				return dictTypeService, userService
			},
			reqBody: `
{
	"name": "男",
	"strValue": "",
	"intValue": 1,
	"boolValue": false,
	"dictTag": "primary",
	"dictColor": "",
	"dict_id": "1",
	"sort": 1,
	"status": true,
	"remark": ""
}
`,
			wantCode: http.StatusBadRequest,
			wantMsg:  "同一字典下存在相同的字典项/值",
		},
		{
			name: "服务器异常",
			mock: func(ctrl *gomock.Controller) (serviceTools.DictTypeService, serviceSystem.UserService) {
				dictTypeService := svcmocks.NewMockDictTypeService(ctrl)
				userService := svcmocks.NewMockUserService(ctrl)
				userService.EXPECT().GetById(gomock.Any(), "1").Return(domainSystem.User{
					User: modelSystem.User{
						CoreModels: models.CoreModels{Id: "1"},
					},
				}, nil)
				dictTypeService.EXPECT().Create(gomock.Any(), gomock.Any()).Return(errors.New("服务器异常"))
				return dictTypeService, userService
			},
			reqBody: `
{
	"name": "男",
	"strValue": "",
	"intValue": 1,
	"boolValue": false,
	"dictTag": "primary",
	"dictColor": "",
	"dict_id": "1",
	"sort": 1,
	"status": true,
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
					UserId: "1", // 避免uuid开销过大
				})
			})
			router := server.Group("/dev-api/v1")
			service, userService := tc.mock(ctrl)
			h := NewDictTypeHandler(c, service, userService)
			h.RegisterRoutes(router)

			req, err := http.NewRequest(http.MethodPost,
				"/dev-api/v1/dictType/create",
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

func Test_dictTypeHandler_Delete(t *testing.T) {
	c := config.RelyConfig{}

	testCases := []struct {
		name     string
		mock     func(ctrl *gomock.Controller) serviceTools.DictTypeService
		id       string
		wantCode int
		wantMsg  string
	}{
		{
			name: "删除成功",
			mock: func(ctrl *gomock.Controller) serviceTools.DictTypeService {
				dictTypeService := svcmocks.NewMockDictTypeService(ctrl)
				dictTypeService.EXPECT().Delete(gomock.Any(), "1").Return(nil)
				return dictTypeService
			},
			id:       "1",
			wantCode: http.StatusOK,
			wantMsg:  "删除成功",
		},
		{
			name: "字典信息不存在",
			mock: func(ctrl *gomock.Controller) serviceTools.DictTypeService {
				dictTypeService := svcmocks.NewMockDictTypeService(ctrl)
				dictTypeService.EXPECT().Delete(gomock.Any(), "1").
					Return(serviceTools.ErrDictTypeNotFound)
				return dictTypeService
			},
			id:       "1",
			wantCode: http.StatusBadRequest,
			wantMsg:  "字典信息不存在",
		},
		// {
		// 	name: "服务器异常",
		// 	mock: func(ctrl *gomock.Controller) serviceTools.DictTypeService {
		// 		dictTypeService := svcmocks.NewMockDictTypeService(ctrl)
		// 		dictTypeService.EXPECT().Delete(gomock.Any(), "1").
		// 			Return(errors.New("服务器异常"))
		// 		return dictTypeService
		// 	},
		// 	id:       "1",
		// 	wantCode: http.StatusInternalServerError,
		// 	wantMsg:  "服务器异常",
		// },
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			server := gin.Default()
			// 设置登录凭证
			server.Use(func(ctx *gin.Context) {
				ctx.Set("claims", &ijwt.Claims{
					UserId: "1", // 避免uuid开销过大
				})
			})
			router := server.Group("/dev-api/v1")
			service := tc.mock(ctrl)
			h := NewDictTypeHandler(c, service, nil)
			h.RegisterRoutes(router)

			req, err := http.NewRequest(http.MethodDelete,
				"/dev-api/v1/dictType/delete/"+tc.id,
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

func Test_dictTypeHandler_Update(t *testing.T) {
	c := config.RelyConfig{}

	testCases := []struct {
		name     string
		mock     func(ctrl *gomock.Controller) (serviceTools.DictTypeService, serviceSystem.UserService)
		reqBody  string
		wantCode int
		wantMsg  string
	}{
		{
			name: "更新成功",
			mock: func(ctrl *gomock.Controller) (serviceTools.DictTypeService, serviceSystem.UserService) {
				dictTypeService := svcmocks.NewMockDictTypeService(ctrl)
				userService := svcmocks.NewMockUserService(ctrl)
				userService.EXPECT().GetById(gomock.Any(), "1").Return(domainSystem.User{
					User: modelSystem.User{
						CoreModels: models.CoreModels{Id: "1"},
					},
				}, nil)
				dictTypeService.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)
				return dictTypeService, userService
			},
			reqBody: `
{
	"id": "1",
	"name": "男",
	"dictTag": "primary",
	"dictColor": "",
	"dict_id": "1",
	"sort": 1,
	"status": true,
	"timestamp": 1,
	"remark": ""
}
`,
			wantCode: http.StatusOK,
			wantMsg:  "更新成功",
		},
		{
			name: "获取凭证，但用户不存在",
			mock: func(ctrl *gomock.Controller) (serviceTools.DictTypeService, serviceSystem.UserService) {
				dictTypeService := svcmocks.NewMockDictTypeService(ctrl)
				userService := svcmocks.NewMockUserService(ctrl)
				userService.EXPECT().GetById(gomock.Any(), "1").Return(domainSystem.User{
					User: modelSystem.User{},
				}, errors.New("获取凭证，但用户不存在"))
				return dictTypeService, userService
			},
			reqBody: `
{
	"id": "1",
	"name": "男",
	"dictTag": "primary",
	"dictColor": "",
	"dict_id": "1",
	"sort": 1,
	"status": true,
	"timestamp": 1,
	"remark": ""
}
`,
			wantCode: http.StatusInternalServerError,
			wantMsg:  "服务器异常",
		},
		{
			name: "字典信息已存在",
			mock: func(ctrl *gomock.Controller) (serviceTools.DictTypeService, serviceSystem.UserService) {
				dictTypeService := svcmocks.NewMockDictTypeService(ctrl)
				userService := svcmocks.NewMockUserService(ctrl)
				userService.EXPECT().GetById(gomock.Any(), "1").Return(domainSystem.User{
					User: modelSystem.User{
						CoreModels: models.CoreModels{Id: "1"},
					},
				}, nil)
				dictTypeService.EXPECT().Update(gomock.Any(), gomock.Any()).Return(serviceTools.ErrDictTypeDuplicate)
				return dictTypeService, userService
			},
			reqBody: `
{
	"id": "1",
	"name": "男",
	"dictTag": "primary",
	"dictColor": "",
	"dict_id": "1",
	"sort": 1,
	"status": true,
	"timestamp": 1,
	"remark": ""
}
`,
			wantCode: http.StatusBadRequest,
			wantMsg:  "字典信息已存在",
		},
		{
			name: "数据版本不一致，取消修改，请刷新后重试",
			mock: func(ctrl *gomock.Controller) (serviceTools.DictTypeService, serviceSystem.UserService) {
				dictTypeService := svcmocks.NewMockDictTypeService(ctrl)
				userService := svcmocks.NewMockUserService(ctrl)
				userService.EXPECT().GetById(gomock.Any(), "1").Return(domainSystem.User{
					User: modelSystem.User{
						CoreModels: models.CoreModels{Id: "1"},
					},
				}, nil)
				dictTypeService.EXPECT().Update(gomock.Any(), gomock.Any()).
					Return(serviceTools.ErrDictTypeVersionInconsistency)
				return dictTypeService, userService
			},
			reqBody: `
{
	"id": "1",
	"name": "男",
	"dictTag": "primary",
	"dictColor": "",
	"dict_id": "1",
	"sort": 1,
	"status": true,
	"timestamp": 1,
	"remark": ""
}
`,
			wantCode: http.StatusBadRequest,
			wantMsg:  "数据版本不一致，取消修改，请刷新后重试",
		},
		{
			name: "服务器异常",
			mock: func(ctrl *gomock.Controller) (serviceTools.DictTypeService, serviceSystem.UserService) {
				dictTypeService := svcmocks.NewMockDictTypeService(ctrl)
				userService := svcmocks.NewMockUserService(ctrl)
				userService.EXPECT().GetById(gomock.Any(), "1").Return(domainSystem.User{
					User: modelSystem.User{
						CoreModels: models.CoreModels{Id: "1"},
					},
				}, nil)
				dictTypeService.EXPECT().Update(gomock.Any(), gomock.Any()).Return(errors.New("服务器异常"))
				return dictTypeService, userService
			},
			reqBody: `
{
	"id": "1",
	"name": "男",
	"dictTag": "primary",
	"dictColor": "",
	"dict_id": "1",
	"sort": 1,
	"status": true,
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
					UserId: "1", // 避免uuid开销过大
				})
			})
			router := server.Group("/dev-api/v1")
			service, userService := tc.mock(ctrl)
			h := NewDictTypeHandler(c, service, userService)
			h.RegisterRoutes(router)

			req, err := http.NewRequest(http.MethodPut,
				"/dev-api/v1/dictType/update",
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

func Test_dictTypeHandler_GetById(t *testing.T) {
	c := config.RelyConfig{}

	testCases := []struct {
		name     string
		mock     func(ctrl *gomock.Controller) serviceTools.DictTypeService
		id       string
		wantCode int
		wantMsg  string
	}{
		{
			name: "获取成功",
			mock: func(ctrl *gomock.Controller) serviceTools.DictTypeService {
				dictTypeService := svcmocks.NewMockDictTypeService(ctrl)
				dictTypeService.EXPECT().GetById(gomock.Any(), "1").
					Return(domainTools.DictType{
						DictType: tools.DictType{
							Name: "男",
						},
					}, nil)
				return dictTypeService
			},
			id:       "1",
			wantCode: http.StatusOK,
			wantMsg:  "获取成功",
		},
		{
			name: "字典信息不存在",
			mock: func(ctrl *gomock.Controller) serviceTools.DictTypeService {
				dictTypeService := svcmocks.NewMockDictTypeService(ctrl)
				dictTypeService.EXPECT().GetById(gomock.Any(), "1").
					Return(domainTools.DictType{}, serviceTools.ErrDictTypeNotFound)
				return dictTypeService
			},
			id:       "1",
			wantCode: http.StatusBadRequest,
			wantMsg:  "字典信息不存在",
		},
		{
			name: "服务器异常",
			mock: func(ctrl *gomock.Controller) serviceTools.DictTypeService {
				dictTypeService := svcmocks.NewMockDictTypeService(ctrl)
				dictTypeService.EXPECT().GetById(gomock.Any(), "1").
					Return(domainTools.DictType{}, errors.New("服务器异常"))
				return dictTypeService
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
					UserId: "1", // 避免uuid开销过大
				})
			})
			router := server.Group("/dev-api/v1")
			service := tc.mock(ctrl)
			h := NewDictTypeHandler(c, service, nil)
			h.RegisterRoutes(router)

			req, err := http.NewRequest(http.MethodGet,
				"/dev-api/v1/dictType/getById/"+tc.id,
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
