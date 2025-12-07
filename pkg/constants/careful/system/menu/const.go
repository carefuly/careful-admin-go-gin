/**
 * Description：
 * FileName：const.go
 * Author：CJiaの用心
 * Create：2025/12/05 11:51:45
 * Remark：
 */

package menu

type Method int

const (
	MethodGET    Method = iota + 1 // GET
	MethodPOST                     // POST
	MethodPUT                      // PUT
	MethodDELETE                   // DELETE
	MethodPATCH                    // DELETE
)

// MethodMapping 接口请求方法映射
var MethodMapping = map[Method]string{
	MethodGET:    "GET",
	MethodPOST:   "POST",
	MethodPUT:    "PUT",
	MethodDELETE: "DELETE",
	MethodPATCH:  "PATCH",
}

// MethodImportMapping 接口请求方法映射
var MethodImportMapping = map[string]Method{
	"GET":    MethodGET,
	"POST":   MethodPOST,
	"PUT":    MethodPUT,
	"DELETE": MethodDELETE,
	"PATCH":  MethodPATCH,
}
