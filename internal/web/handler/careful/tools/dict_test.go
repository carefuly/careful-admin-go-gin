/**
 * Description：
 * FileName：dict_test.go.go
 * Author：CJiaの用心
 * Create：2025/10/11 15:40:25
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
	"github.com/carefuly/careful-admin-go-gin/internal/model/careful/system"
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

func Test_dictHandler_Create(t *testing.T) {
	c := config.RelyConfig{}

	testCases := []struct {
		name     string
		mock     func(ctrl *gomock.Controller) (serviceTools.DictService, serviceSystem.UserService)
		reqBody  string
		wantCode int
		wantMsg  string
	}{
		{
			name: "新增成功",
			mock: func(ctrl *gomock.Controller) (serviceTools.DictService, serviceSystem.UserService) {
				dictService := svcmocks.NewMockDictService(ctrl)
				userService := svcmocks.NewMockUserService(ctrl)
				userService.EXPECT().GetById(gomock.Any(), "1").Return(domainSystem.User{
					User: system.User{
						CoreModels: models.CoreModels{Id: "1"},
					},
				}, nil)
				dictService.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)
				return dictService, userService
			},
			reqBody: `
{
	"status": true,
	"name": "字典名称",
	"code": "字典编码",
	"type": 1,
	"valueType": 2
}
`,
			wantCode: http.StatusOK,
			wantMsg:  "新增成功",
		},
		{
			name: "获取凭证，但用户不存在",
			mock: func(ctrl *gomock.Controller) (serviceTools.DictService, serviceSystem.UserService) {
				dictService := svcmocks.NewMockDictService(ctrl)
				userService := svcmocks.NewMockUserService(ctrl)
				userService.EXPECT().GetById(gomock.Any(), "1").Return(domainSystem.User{
					User: system.User{},
				}, errors.New("获取凭证，但用户不存在"))
				return dictService, userService
			},
			reqBody: `
{
	"status": true,
	"name": "字典名称",
	"code": "字典编码",
	"type": 1,
	"valueType": 2
}
`,
			wantCode: http.StatusInternalServerError,
			wantMsg:  "服务器异常",
		},
		{
			name: "type参数不匹配",
			mock: func(ctrl *gomock.Controller) (serviceTools.DictService, serviceSystem.UserService) {
				dictService := svcmocks.NewMockDictService(ctrl)
				userService := svcmocks.NewMockUserService(ctrl)
				userService.EXPECT().GetById(gomock.Any(), "1").Return(domainSystem.User{
					User: system.User{
						CoreModels: models.CoreModels{Id: "1"},
					},
				}, nil)
				return dictService, userService
			},
			reqBody: `
{
	"status": true,
	"name": "字典名称",
	"code": "字典编码",
	"type": 4,
	"valueType": 2
}
`,
			wantCode: http.StatusBadRequest,
			wantMsg:  "无效的字典分类枚举值: 4",
		},
		{
			name: "valueType参数不匹配",
			mock: func(ctrl *gomock.Controller) (serviceTools.DictService, serviceSystem.UserService) {
				dictService := svcmocks.NewMockDictService(ctrl)
				userService := svcmocks.NewMockUserService(ctrl)
				userService.EXPECT().GetById(gomock.Any(), "1").Return(domainSystem.User{
					User: system.User{
						CoreModels: models.CoreModels{Id: "1"},
					},
				}, nil)
				return dictService, userService
			},
			reqBody: `
{
	"status": true,
	"name": "字典名称",
	"code": "字典编码",
	"type": 2,
	"valueType": 4
}
`,
			wantCode: http.StatusBadRequest,
			wantMsg:  "无效的数据类型枚举值: 4",
		},
		{
			name: "字典名称已存在",
			mock: func(ctrl *gomock.Controller) (serviceTools.DictService, serviceSystem.UserService) {
				dictService := svcmocks.NewMockDictService(ctrl)
				userService := svcmocks.NewMockUserService(ctrl)
				userService.EXPECT().GetById(gomock.Any(), "1").Return(domainSystem.User{
					User: system.User{
						CoreModels: models.CoreModels{Id: "1"},
					},
				}, nil)
				dictService.EXPECT().Create(gomock.Any(), gomock.Any()).Return(serviceTools.ErrDictNameDuplicate)
				return dictService, userService
			},
			reqBody: `
{
	"status": true,
	"name": "字典名称",
	"code": "字典编码",
	"type": 1,
	"valueType": 2
}
`,
			wantCode: http.StatusBadRequest,
			wantMsg:  "字典名称已存在",
		},
		{
			name: "字典编码已存在",
			mock: func(ctrl *gomock.Controller) (serviceTools.DictService, serviceSystem.UserService) {
				dictService := svcmocks.NewMockDictService(ctrl)
				userService := svcmocks.NewMockUserService(ctrl)
				userService.EXPECT().GetById(gomock.Any(), "1").Return(domainSystem.User{
					User: system.User{
						CoreModels: models.CoreModels{Id: "1"},
					},
				}, nil)
				dictService.EXPECT().Create(gomock.Any(), gomock.Any()).Return(serviceTools.ErrDictCodeDuplicate)
				return dictService, userService
			},
			reqBody: `
{
	"status": true,
	"name": "字典名称",
	"code": "字典编码",
	"type": 1,
	"valueType": 2
}
`,
			wantCode: http.StatusBadRequest,
			wantMsg:  "字典编码已存在",
		},
		{
			name: "服务器异常",
			mock: func(ctrl *gomock.Controller) (serviceTools.DictService, serviceSystem.UserService) {
				dictService := svcmocks.NewMockDictService(ctrl)
				userService := svcmocks.NewMockUserService(ctrl)
				userService.EXPECT().GetById(gomock.Any(), "1").Return(domainSystem.User{
					User: system.User{
						CoreModels: models.CoreModels{Id: "1"},
					},
				}, nil)
				dictService.EXPECT().Create(gomock.Any(), gomock.Any()).Return(errors.New("服务器异常"))
				return dictService, userService
			},
			reqBody: `
{
	"status": true,
	"name": "字典名称",
	"code": "字典编码",
	"type": 1,
	"valueType": 2
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
			h := NewDictHandler(c, service, userService)
			h.RegisterRoutes(router)

			req, err := http.NewRequest(http.MethodPost,
				"/dev-api/v1/dict/create",
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

func Test_dictHandler_Delete(t *testing.T) {
	c := config.RelyConfig{}

	testCases := []struct {
		name     string
		mock     func(ctrl *gomock.Controller) serviceTools.DictService
		id       string
		wantCode int
		wantMsg  string
	}{
		{
			name: "删除成功",
			mock: func(ctrl *gomock.Controller) serviceTools.DictService {
				dictService := svcmocks.NewMockDictService(ctrl)
				dictService.EXPECT().Delete(gomock.Any(), "1").Return(nil)
				return dictService
			},
			id:       "1",
			wantCode: http.StatusOK,
			wantMsg:  "删除成功",
		},
		{
			name: "数据字典不存在",
			mock: func(ctrl *gomock.Controller) serviceTools.DictService {
				dictService := svcmocks.NewMockDictService(ctrl)
				dictService.EXPECT().Delete(gomock.Any(), "1").
					Return(serviceTools.ErrDictNotFound)
				return dictService
			},
			id:       "1",
			wantCode: http.StatusBadRequest,
			wantMsg:  "数据字典不存在",
		},
		{
			name: "服务器异常",
			mock: func(ctrl *gomock.Controller) serviceTools.DictService {
				dictService := svcmocks.NewMockDictService(ctrl)
				dictService.EXPECT().Delete(gomock.Any(), "1").Return(errors.New("服务器异常"))
				return dictService
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
			h := NewDictHandler(c, service, nil)
			h.RegisterRoutes(router)

			req, err := http.NewRequest(http.MethodDelete,
				"/dev-api/v1/dict/delete/"+tc.id,
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

func Test_dictHandler_Update(t *testing.T) {
	c := config.RelyConfig{}

	testCases := []struct {
		name     string
		mock     func(ctrl *gomock.Controller) (serviceTools.DictService, serviceSystem.UserService)
		reqBody  string
		wantCode int
		wantMsg  string
	}{
		{
			name: "更新成功",
			mock: func(ctrl *gomock.Controller) (serviceTools.DictService, serviceSystem.UserService) {
				dictService := svcmocks.NewMockDictService(ctrl)
				userService := svcmocks.NewMockUserService(ctrl)
				userService.EXPECT().GetById(gomock.Any(), "1").Return(domainSystem.User{
					User: system.User{
						CoreModels: models.CoreModels{Id: "1"},
					},
				}, nil)
				dictService.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)
				return dictService, userService
			},
			reqBody: `
{
	"id": "1",
	"code": "字典编码",
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
			mock: func(ctrl *gomock.Controller) (serviceTools.DictService, serviceSystem.UserService) {
				dictService := svcmocks.NewMockDictService(ctrl)
				userService := svcmocks.NewMockUserService(ctrl)
				userService.EXPECT().GetById(gomock.Any(), "1").Return(domainSystem.User{
					User: system.User{},
				}, errors.New("获取凭证，但用户不存在"))
				return dictService, userService
			},
			reqBody: `
{
	"status": true,
	"name": "字典名称",
	"code": "字典编码",
	"type": 1,
	"valueType": 2
}
`,
			wantCode: http.StatusInternalServerError,
			wantMsg:  "服务器异常",
		},
		{
			name: "字典名称已存在",
			mock: func(ctrl *gomock.Controller) (serviceTools.DictService, serviceSystem.UserService) {
				dictService := svcmocks.NewMockDictService(ctrl)
				userService := svcmocks.NewMockUserService(ctrl)
				userService.EXPECT().GetById(gomock.Any(), "1").Return(domainSystem.User{
					User: system.User{
						CoreModels: models.CoreModels{Id: "1"},
					},
				}, nil)
				dictService.EXPECT().Update(gomock.Any(), gomock.Any()).Return(serviceTools.ErrDictNameDuplicate)
				return dictService, userService
			},
			reqBody: `
{
	"id": "1",
	"code": "字典编码",
	"sort": 1,
	"status": true,
	"timestamp": 1,
	"remark": ""
}
`,
			wantCode: http.StatusBadRequest,
			wantMsg:  "字典名称已存在",
		},
		{
			name: "字典编码已存在",
			mock: func(ctrl *gomock.Controller) (serviceTools.DictService, serviceSystem.UserService) {
				dictService := svcmocks.NewMockDictService(ctrl)
				userService := svcmocks.NewMockUserService(ctrl)
				userService.EXPECT().GetById(gomock.Any(), "1").Return(domainSystem.User{
					User: system.User{
						CoreModels: models.CoreModels{Id: "1"},
					},
				}, nil)
				dictService.EXPECT().Update(gomock.Any(), gomock.Any()).Return(serviceTools.ErrDictCodeDuplicate)
				return dictService, userService
			},
			reqBody: `
{
	"id": "1",
	"code": "字典编码",
	"sort": 1,
	"status": true,
	"timestamp": 1,
	"remark": ""
}
`,
			wantCode: http.StatusBadRequest,
			wantMsg:  "字典编码已存在",
		},
		{
			name: "数据版本不一致，取消修改，请刷新后重试",
			mock: func(ctrl *gomock.Controller) (serviceTools.DictService, serviceSystem.UserService) {
				dictService := svcmocks.NewMockDictService(ctrl)
				userService := svcmocks.NewMockUserService(ctrl)
				userService.EXPECT().GetById(gomock.Any(), "1").Return(domainSystem.User{
					User: system.User{
						CoreModels: models.CoreModels{Id: "1"},
					},
				}, nil)
				dictService.EXPECT().Update(gomock.Any(), gomock.Any()).
					Return(serviceTools.ErrDictVersionInconsistency)
				return dictService, userService
			},
			reqBody: `
{
	"id": "1",
	"code": "字典编码",
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
			mock: func(ctrl *gomock.Controller) (serviceTools.DictService, serviceSystem.UserService) {
				dictService := svcmocks.NewMockDictService(ctrl)
				userService := svcmocks.NewMockUserService(ctrl)
				userService.EXPECT().GetById(gomock.Any(), "1").Return(domainSystem.User{
					User: system.User{
						CoreModels: models.CoreModels{Id: "1"},
					},
				}, nil)
				dictService.EXPECT().Update(gomock.Any(), gomock.Any()).Return(errors.New("服务器异常"))
				return dictService, userService
			},
			reqBody: `
{
	"id": "1",
	"code": "字典编码",
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
			h := NewDictHandler(c, service, userService)
			h.RegisterRoutes(router)

			req, err := http.NewRequest(http.MethodPut,
				"/dev-api/v1/dict/update",
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

func Test_dictHandler_GetById(t *testing.T) {
	c := config.RelyConfig{}

	testCases := []struct {
		name     string
		mock     func(ctrl *gomock.Controller) serviceTools.DictService
		id       string
		wantCode int
		wantMsg  string
	}{
		{
			name: "获取成功",
			mock: func(ctrl *gomock.Controller) serviceTools.DictService {
				dictService := svcmocks.NewMockDictService(ctrl)
				dictService.EXPECT().GetById(gomock.Any(), "1").
					Return(domainTools.Dict{
						Dict: tools.Dict{
							Code: "字典编码",
						},
					}, nil)
				return dictService
			},
			id:       "1",
			wantCode: http.StatusOK,
			wantMsg:  "获取成功",
		},
		{
			name: "字典不存在",
			mock: func(ctrl *gomock.Controller) serviceTools.DictService {
				dictService := svcmocks.NewMockDictService(ctrl)
				dictService.EXPECT().GetById(gomock.Any(), "1").
					Return(domainTools.Dict{}, serviceTools.ErrDictNotFound)
				return dictService
			},
			id:       "1",
			wantCode: http.StatusBadRequest,
			wantMsg:  "字典不存在",
		},
		{
			name: "服务器异常",
			mock: func(ctrl *gomock.Controller) serviceTools.DictService {
				dictService := svcmocks.NewMockDictService(ctrl)
				dictService.EXPECT().GetById(gomock.Any(), "1").
					Return(domainTools.Dict{}, errors.New("服务器异常"))
				return dictService
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
			h := NewDictHandler(c, service, nil)
			h.RegisterRoutes(router)

			req, err := http.NewRequest(http.MethodGet,
				"/dev-api/v1/dict/getById/"+tc.id,
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
