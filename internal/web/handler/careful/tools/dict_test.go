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
	"fmt"
	"github.com/carefuly/careful-admin-go-gin/config"
	domainTools "github.com/carefuly/careful-admin-go-gin/internal/domain/careful/tools"
	modelsTools "github.com/carefuly/careful-admin-go-gin/internal/model/careful/tools"
	svcmocks "github.com/carefuly/careful-admin-go-gin/internal/service/careful/mocks"
	serviceSystem "github.com/carefuly/careful-admin-go-gin/internal/service/careful/system"
	serviceTools "github.com/carefuly/careful-admin-go-gin/internal/service/careful/tools"
	"github.com/carefuly/careful-admin-go-gin/ioc"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test_dictHandler_Create(t *testing.T) {
	c := config.RelyConfig{}
	server := ioc.NewServer(c, "zh")
	// 初始化翻译器
	if err := server.InitTranslator(); err != nil {
		zap.L().Fatal("翻译器初始化失败", zap.Error(err))
	}
	c.Trans = server.Translator

	testCases := []struct {
		name     string
		mock     func(ctrl *gomock.Controller) (serviceTools.DictService, serviceSystem.UserService)
		reqBody  string
		wantCode int
		wantBody string
	}{
		{
			name: "新增成功",
			mock: func(ctrl *gomock.Controller) (serviceTools.DictService, serviceSystem.UserService) {
				dictService := svcmocks.NewMockDictService(ctrl)
				userService := svcmocks.NewMockUserService(ctrl)
				dictService.EXPECT().Create(gomock.Any(), domainTools.Dict{
					Dict: modelsTools.Dict{
						Status:    true,
						Name:      "字典名称",
						Code:      "字典编码",
						Type:      1,
						ValueType: 2,
					},
				}).Return(nil)
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
			wantBody: "新增成功",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			server := gin.Default()
			router := server.Group("/dev-api/v1")
			service, userService := tc.mock(ctrl)
			h := NewDictHandler(c, service, userService)
			h.RegisterRoutes(router)
			req, err := http.NewRequest(http.MethodPost,
				"/dict/create", bytes.NewBuffer([]byte(tc.reqBody)))
			require.NoError(t, err)
			// 数据是 JSON 格式
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()
			// 响应写回到 resp 里
			server.ServeHTTP(resp, req)
			fmt.Println(resp)
			assert.Equal(t, tc.wantCode, resp.Code)
			assert.Equal(t, tc.wantBody, resp.Body.String())
		})
	}
}
