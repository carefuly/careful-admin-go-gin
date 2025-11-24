/**
 * Description：
 * FileName：index.go
 * Author：CJiaの用心
 * Create：2025/10/8 14:29:18
 * Remark：
 */

package response

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

/**
// 简单成功响应
response.NewResponse().Success(ctx, "操作成功", data)

// 复杂错误响应
response.NewResponse().
	WithStatus("custom_error").
	WithCode(http.StatusBadRequest).
	WithMessage(errors.New("自定义错误")).
	WithData(details).
	WithRequestID(ctx).
	ToJSON(ctx)

// 标准错误响应
response.NewResponse().Error(ctx, http.StatusBadRequest, "参数错误", nil)
*/

const (
	SuccessStatus = "success"
	ErrorStatus   = "error"
)

// Response 响应处理器结构体
type Response struct {
	Code      int         `json:"code"`       // HTTP状态码
	Message   interface{} `json:"msg"`        // 提示信息
	Data      interface{} `json:"data"`       // 数据
	Status    string      `json:"status"`     // 响应状态: success|error
	Timestamp string      `json:"timestamp"`  // 时间戳
	RequestID string      `json:"request_id"` // 请求ID
}

// NewResponse 创建新的响应处理器
func NewResponse() *Response {
	return &Response{
		Timestamp: time.Now().Format(time.RFC3339),
	}
}

// WithStatus 设置响应状态
func (r *Response) WithStatus(status string) *Response {
	r.Status = status
	return r
}

// WithCode 设置HTTP状态码
func (r *Response) WithCode(code int) *Response {
	r.Code = code
	return r
}

// WithMessage 设置消息内容
func (r *Response) WithMessage(message interface{}) *Response {
	r.Message = message
	return r
}

// WithData 设置响应数据
func (r *Response) WithData(data interface{}) *Response {
	r.Data = data
	return r
}

// WithRequestID 从上下文中设置请求ID
func (r *Response) WithRequestID(ctx *gin.Context) *Response {
	if requestID, exists := ctx.Get("X-Request-ID"); exists {
		r.RequestID = requestID.(string)
	}
	return r
}

// ToJSON 生成JSON响应
func (r *Response) ToJSON(ctx *gin.Context) {
	ctx.JSON(r.Code, gin.H{
		"status":     r.Status,
		"code":       r.Code,
		"msg":        r.Message,
		"data":       r.Data,
		"timestamp":  r.Timestamp,
		"request_id": r.RequestID,
	})
}

// Success 成功响应
func (r *Response) Success(ctx *gin.Context, msg string, data interface{}) {
	r.WithStatus(SuccessStatus).
		WithCode(http.StatusOK).
		WithMessage(msg).
		WithData(data).
		WithRequestID(ctx).
		ToJSON(ctx)
}

// Error 错误响应
func (r *Response) Error(ctx *gin.Context, code int, msg interface{}, data interface{}) {
	r.WithStatus(ErrorStatus).
		WithCode(code).
		WithMessage(msg).
		WithData(data).
		WithRequestID(ctx).
		ToJSON(ctx)
}
