/**
 * Description：
 * FileName：types.go
 * Author：CJiaの用心
 * Create：2025/12/2 04:38:19
 * Remark：
 */

package _import

type ImportResult struct {
	Result []map[string]string // 导入数据信息
}

// type ImportError struct {
// 	Row     int    `json:"row"`     // 数据行号
// 	Message string `json:"message"` // 错误信息
// }
//
// func (r *ImportResult) AddError(row int, message string) {
// 	r.FailCount++
// 	r.Errors = append(r.Errors, ImportError{
// 		Row:     row,
// 		Message: message,
// 	})
// }
