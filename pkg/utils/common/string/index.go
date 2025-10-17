/**
 * Description：
 * FileName：index.go
 * Author：CJiaの用心
 * Create：2025/10/10 11:43:50
 * Remark：
 */

package _string

import "strings"

// ContainsAnySubstring 判断是否包含特定子串
func ContainsAnySubstring(s string, subs ...string) bool {
	for _, substr := range subs {
		if strings.Contains(s, substr) {
			return true
		}
	}
	return false
}

// CleanInputString 清理输入中的特殊字符
func CleanInputString(input string) string {
	return strings.TrimSpace(strings.NewReplacer("\n", "", "\r", "").Replace(input))
}
