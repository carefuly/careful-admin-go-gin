/**
 * Description：
 * FileName：index.go
 * Author：CJiaの用心
 * Create：2025/10/9 15:56:22
 * Remark：
 */

package json_format

import (
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
)

// PrintFormattedJSON 格式化输出任意数据为 JSON
// 可选参数 indent 用于自定义缩进（默认两个空格）
func PrintFormattedJSON(data any, indent ...string) {
	prefix := ""
	indentStr := "  "
	if len(indent) > 0 {
		indentStr = indent[0]
	}

	jsonData, err := json.MarshalIndent(data, prefix, indentStr)
	if err != nil {
		zap.L().Error("JSON 序列化失败",
			zap.Error(err),
			zap.Any("raw_data", data),
		)
		return
	}

	output := string(jsonData)
	fmt.Println(output)
}
