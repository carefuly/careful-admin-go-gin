/**
 * Description：
 * FileName：validator_test.go.go
 * Author：CJiaの用心
 * Create：2025/10/9 16:00:57
 * Remark：
 */

package validate

import (
	"encoding/json"
	"errors"
	"github.com/carefuly/careful-admin-go-gin/pkg/ginx/response"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"
)

// 测试结构体定义
type validatorErrorTestCase struct {
	name          string
	inputErr      error
	setupMocks    func(trans *MockTranslator)
	wantCode      int
	wantMsg       string
	wantData      interface{}
	withRequestID bool
}

func TestValidatorErrorHandler_Handle(t *testing.T) {
	testCases := []validatorErrorTestCase{
		{
			name:     "处理EOF错误",
			inputErr: io.EOF,
			setupMocks: func(trans *MockTranslator) {
				// 不需要设置翻译器
			},
			wantCode:      http.StatusBadRequest,
			wantMsg:       "请求参数为空",
			wantData:      nil,
			withRequestID: true,
		},
		{
			name: "处理简单验证错误",
			inputErr: validator.ValidationErrors{
				newMockFieldError("Name", "required", "Name is required"),
			},
			setupMocks: func(trans *MockTranslator) {
				trans.On("Translate", mock.Anything).Return("姓名不能为空").Once()
			},
			wantCode:      http.StatusBadRequest,
			wantMsg:       "参数验证失败",
			wantData:      map[string]string{"Name": "姓名不能为空"},
			withRequestID: true,
		},
		{
			name: "处理嵌套字段验证错误",
			inputErr: validator.ValidationErrors{
				newMockFieldError("User.Name", "required", "User.Name is required"),
				newMockFieldError("User.Email", "email", "User.Email must be a valid email"),
			},
			setupMocks: func(trans *MockTranslator) {
				trans.On("Translate", mock.MatchedBy(func(e validator.FieldError) bool {
					return e.Field() == "User.Name"
				})).Return("用户名不能为空").Once()
				trans.On("Translate", mock.MatchedBy(func(e validator.FieldError) bool {
					return e.Field() == "User.Email"
				})).Return("邮箱格式不正确").Once()
			},
			wantCode:      http.StatusBadRequest,
			wantMsg:       "参数验证失败",
			wantData:      map[string]string{"Name": "用户名不能为空", "Email": "邮箱格式不正确"},
			withRequestID: true,
		},
		{
			name:     "处理普通错误",
			inputErr: errors.New("some random error"),
			setupMocks: func(trans *MockTranslator) {
				// 不需要设置翻译器
			},
			wantCode:      http.StatusBadRequest,
			wantMsg:       "some random error",
			wantData:      nil,
			withRequestID: true,
		},
		{
			name: "处理多个验证错误",
			inputErr: validator.ValidationErrors{
				newMockFieldError("Age", "min", "Age must be at least 18"),
				newMockFieldError("Password", "min", "Password must be at least 8 characters"),
			},
			setupMocks: func(trans *MockTranslator) {
				trans.On("Translate", mock.MatchedBy(func(e validator.FieldError) bool {
					return e.Field() == "Age"
				})).Return("年龄必须大于18岁").Once()
				trans.On("Translate", mock.MatchedBy(func(e validator.FieldError) bool {
					return e.Field() == "Password"
				})).Return("密码长度至少为8位").Once()
			},
			wantCode:      http.StatusBadRequest,
			wantMsg:       "参数验证失败",
			wantData:      map[string]string{"Age": "年龄必须大于18岁", "Password": "密码长度至少为8位"},
			withRequestID: true,
		},
		{
			name:     "处理无请求ID的错误",
			inputErr: errors.New("no request ID error"),
			setupMocks: func(trans *MockTranslator) {
				// 不需要设置翻译器
			},
			wantCode:      http.StatusBadRequest,
			wantMsg:       "no request ID error",
			wantData:      nil,
			withRequestID: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 创建模拟对象
			mockTrans := new(MockTranslator)
			tc.setupMocks(mockTrans)

			// 创建处理器
			handler := NewValidatorErrorHandler(mockTrans)

			// 设置路由
			router := setupRouter(tc.withRequestID)
			router.GET("/test", func(c *gin.Context) {
				handler.Handle(c, tc.inputErr)
			})

			// 创建请求
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/test", nil)
			router.ServeHTTP(w, req)

			// 验证HTTP状态码
			assert.Equal(t, tc.wantCode, w.Code, "HTTP状态码不匹配")

			// 解析响应体
			var resp response.Response
			err := json.Unmarshal(w.Body.Bytes(), &resp)
			assert.NoError(t, err, "无法解析JSON响应")

			// 验证响应内容
			assert.Equal(t, response.ErrorStatus, resp.Status, "响应状态应为error")
			assert.Equal(t, tc.wantCode, resp.Code, "状态码不匹配")
			assert.Equal(t, tc.wantMsg, resp.Message, "消息内容不匹配")
			assert.EqualValues(t, tc.wantData, resp.Data, "数据内容不匹配")

			// 验证请求ID
			if tc.withRequestID {
				assert.Equal(t, "test-request-id", resp.RequestID, "请求ID不匹配")
			} else {
				assert.Empty(t, resp.RequestID, "请求ID应为空")
			}

			// 验证所有模拟调用
			mockTrans.AssertExpectations(t)
		})
	}
}

// setupRouter 设置测试路由器
func setupRouter(withRequestID bool) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	if withRequestID {
		r.Use(func(c *gin.Context) {
			c.Set("X-Request-ID", "test-request-id")
		})
	}

	return r
}

// MockTranslator 模拟翻译器
type MockTranslator struct {
	mock.Mock
}

func (m *MockTranslator) Locale() string {
	// TODO implement me
	panic("implement me")
}

func (m *MockTranslator) PluralsCardinal() []locales.PluralRule {
	// TODO implement me
	panic("implement me")
}

func (m *MockTranslator) PluralsOrdinal() []locales.PluralRule {
	// TODO implement me
	panic("implement me")
}

func (m *MockTranslator) PluralsRange() []locales.PluralRule {
	// TODO implement me
	panic("implement me")
}

func (m *MockTranslator) CardinalPluralRule(num float64, v uint64) locales.PluralRule {
	// TODO implement me
	panic("implement me")
}

func (m *MockTranslator) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	// TODO implement me
	panic("implement me")
}

func (m *MockTranslator) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {
	// TODO implement me
	panic("implement me")
}

func (m *MockTranslator) MonthAbbreviated(month time.Month) string {
	// TODO implement me
	panic("implement me")
}

func (m *MockTranslator) MonthsAbbreviated() []string {
	// TODO implement me
	panic("implement me")
}

func (m *MockTranslator) MonthNarrow(month time.Month) string {
	// TODO implement me
	panic("implement me")
}

func (m *MockTranslator) MonthsNarrow() []string {
	// TODO implement me
	panic("implement me")
}

func (m *MockTranslator) MonthWide(month time.Month) string {
	// TODO implement me
	panic("implement me")
}

func (m *MockTranslator) MonthsWide() []string {
	// TODO implement me
	panic("implement me")
}

func (m *MockTranslator) WeekdayAbbreviated(weekday time.Weekday) string {
	// TODO implement me
	panic("implement me")
}

func (m *MockTranslator) WeekdaysAbbreviated() []string {
	// TODO implement me
	panic("implement me")
}

func (m *MockTranslator) WeekdayNarrow(weekday time.Weekday) string {
	// TODO implement me
	panic("implement me")
}

func (m *MockTranslator) WeekdaysNarrow() []string {
	// TODO implement me
	panic("implement me")
}

func (m *MockTranslator) WeekdayShort(weekday time.Weekday) string {
	// TODO implement me
	panic("implement me")
}

func (m *MockTranslator) WeekdaysShort() []string {
	// TODO implement me
	panic("implement me")
}

func (m *MockTranslator) WeekdayWide(weekday time.Weekday) string {
	// TODO implement me
	panic("implement me")
}

func (m *MockTranslator) WeekdaysWide() []string {
	// TODO implement me
	panic("implement me")
}

func (m *MockTranslator) FmtNumber(num float64, v uint64) string {
	// TODO implement me
	panic("implement me")
}

func (m *MockTranslator) FmtPercent(num float64, v uint64) string {
	// TODO implement me
	panic("implement me")
}

func (m *MockTranslator) FmtCurrency(num float64, v uint64, currency currency.Type) string {
	// TODO implement me
	panic("implement me")
}

func (m *MockTranslator) FmtAccounting(num float64, v uint64, currency currency.Type) string {
	// TODO implement me
	panic("implement me")
}

func (m *MockTranslator) FmtDateShort(t time.Time) string {
	// TODO implement me
	panic("implement me")
}

func (m *MockTranslator) FmtDateMedium(t time.Time) string {
	// TODO implement me
	panic("implement me")
}

func (m *MockTranslator) FmtDateLong(t time.Time) string {
	// TODO implement me
	panic("implement me")
}

func (m *MockTranslator) FmtDateFull(t time.Time) string {
	// TODO implement me
	panic("implement me")
}

func (m *MockTranslator) FmtTimeShort(t time.Time) string {
	// TODO implement me
	panic("implement me")
}

func (m *MockTranslator) FmtTimeMedium(t time.Time) string {
	// TODO implement me
	panic("implement me")
}

func (m *MockTranslator) FmtTimeLong(t time.Time) string {
	// TODO implement me
	panic("implement me")
}

func (m *MockTranslator) FmtTimeFull(t time.Time) string {
	// TODO implement me
	panic("implement me")
}

func (m *MockTranslator) Add(key interface{}, text string, override bool) error {
	// TODO implement me
	panic("implement me")
}

func (m *MockTranslator) AddCardinal(key interface{}, text string, rule locales.PluralRule, override bool) error {
	// TODO implement me
	panic("implement me")
}

func (m *MockTranslator) AddOrdinal(key interface{}, text string, rule locales.PluralRule, override bool) error {
	// TODO implement me
	panic("implement me")
}

func (m *MockTranslator) AddRange(key interface{}, text string, rule locales.PluralRule, override bool) error {
	// TODO implement me
	panic("implement me")
}

func (m *MockTranslator) T(key interface{}, params ...string) (string, error) {
	// TODO implement me
	panic("implement me")
}

func (m *MockTranslator) C(key interface{}, num float64, digits uint64, param string) (string, error) {
	// TODO implement me
	panic("implement me")
}

func (m *MockTranslator) O(key interface{}, num float64, digits uint64, param string) (string, error) {
	// TODO implement me
	panic("implement me")
}

func (m *MockTranslator) R(key interface{}, num1 float64, digits1 uint64, num2 float64, digits2 uint64, param1, param2 string) (string, error) {
	// TODO implement me
	panic("implement me")
}

func (m *MockTranslator) VerifyTranslations() error {
	// TODO implement me
	panic("implement me")
}

func (m *MockTranslator) Translate(err validator.FieldError) string {
	args := m.Called(err)
	return args.String(0)
}

// mockFieldError 模拟字段错误
type mockFieldError struct {
	field string
	tag   string
	param string
}

func (e *mockFieldError) Kind() reflect.Kind {
	// TODO implement me
	panic("implement me")
}

func (e *mockFieldError) Type() reflect.Type {
	// TODO implement me
	panic("implement me")
}

func (e *mockFieldError) Translate(ut ut.Translator) string {
	// TODO implement me
	panic("implement me")
}

func newMockFieldError(field, tag, param string) validator.FieldError {
	return &mockFieldError{field: field, tag: tag, param: param}
}

func (e *mockFieldError) Tag() string             { return e.tag }
func (e *mockFieldError) ActualTag() string       { return e.tag }
func (e *mockFieldError) Namespace() string       { return "" }
func (e *mockFieldError) StructNamespace() string { return "" }
func (e *mockFieldError) Field() string           { return e.field }
func (e *mockFieldError) StructField() string     { return "" }
func (e *mockFieldError) Value() interface{}      { return nil }
func (e *mockFieldError) Param() string           { return e.param }
func (e *mockFieldError) Error() string           { return e.tag + " error for " + e.field }

func TestSanitizeFieldNames(t *testing.T) {
	testCases := []struct {
		name  string
		input map[string]string
		want  map[string]string
	}{
		{
			name:  "空输入",
			input: map[string]string{},
			want:  map[string]string{},
		},
		{
			name: "简单字段",
			input: map[string]string{
				"Name": "不能为空",
				"Age":  "必须大于0",
			},
			want: map[string]string{
				"Name": "不能为空",
				"Age":  "必须大于0",
			},
		},
		{
			name: "嵌套结构体字段",
			input: map[string]string{
				"User.Name":  "用户名不能为空",
				"User.Email": "邮箱格式不正确",
			},
			want: map[string]string{
				"Name":  "用户名不能为空",
				"Email": "邮箱格式不正确",
			},
		},
		{
			name: "深度嵌套字段",
			input: map[string]string{
				"User.Address.City": "城市必须填写",
				"User.Phone":        "手机号格式不正确",
			},
			want: map[string]string{
				"City":  "城市必须填写",
				"Phone": "手机号格式不正确",
			},
		},
		{
			name: "混合字段",
			input: map[string]string{
				"ID":                "ID必须为整数",
				"User.Name":         "用户名不能为空",
				"Order.Items[0].ID": "商品ID不能为空",
			},
			want: map[string]string{
				"ID":    "ID必须为整数",
				"Name":  "用户名不能为空",
				"Items": "商品ID不能为空",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := sanitizeFieldNames(tc.input)
			assert.Equal(t, tc.want, result)
		})
	}
}
