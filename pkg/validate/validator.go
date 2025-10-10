/**
 * Description：
 * FileName：validator.go
 * Author：CJiaの用心
 * Create：2025/10/9 16:00:18
 * Remark：
 */

package validate

import (
	"errors"
	"github.com/carefuly/careful-admin-go-gin/pkg/ginx/response"
	"github.com/gin-gonic/gin"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	"io"
	"net/http"
	"strings"
)

// ValidatorErrorHandler 处理验证错误
type ValidatorErrorHandler struct {
	trans ut.Translator
}

// NewValidatorErrorHandler 创建验证错误处理器
func NewValidatorErrorHandler(trans ut.Translator) *ValidatorErrorHandler {
	return &ValidatorErrorHandler{trans: trans}
}

// Handle 处理验证错误并发送响应
func (h *ValidatorErrorHandler) Handle(ctx *gin.Context, err error) {
	switch {
	case errors.Is(err, io.EOF):
		// 请求参数为空
		response.NewResponse().Error(ctx, http.StatusBadRequest, "请求参数为空", nil)

	case isValidationError(err):
		// 参数验证失败
		var validationErrs validator.ValidationErrors
		errors.As(err, &validationErrs)
		translatedErrs := translateValidationErrors(validationErrs, h.trans)
		sanitizedErrs := sanitizeFieldNames(translatedErrs)

		response.NewResponse().Error(
			ctx,
			http.StatusBadRequest,
			"参数验证失败",
			sanitizedErrs,
		)

	default:
		// 其他错误
		response.NewResponse().Error(ctx, http.StatusBadRequest, err.Error(), nil)
	}
}

// translateValidationErrors 转换验证错误信息
func translateValidationErrors(errs validator.ValidationErrors, trans ut.Translator) map[string]string {
	errorsMap := make(map[string]string)
	for _, e := range errs {
		errorsMap[e.Field()] = e.Translate(trans)
	}
	return errorsMap
}

// sanitizeFieldNames 清理字段名称
func sanitizeFieldNames(errorsMap map[string]string) map[string]string {
	sanitized := make(map[string]string)
	for field, msg := range errorsMap {
		// 移除嵌套结构体名称（如 "User.Name" -> "Name"）
		if idx := strings.LastIndex(field, "."); idx != -1 {
			field = field[idx+1:]
		}
		sanitized[field] = msg
	}
	return sanitized
}

// isValidationError 检查是否为验证错误
func isValidationError(err error) bool {
	var ver validator.ValidationErrors
	return errors.As(err, &ver)
}
