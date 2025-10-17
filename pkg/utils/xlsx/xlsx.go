/**
 * Description：
 * FileName：xlsx.go
 * Author：CJiaの用心
 * Create：2025/10/17 09:13:12
 * Remark：
 */

package xlsx

import (
	"fmt"
	"github.com/xuri/excelize/v2"
	"strings"
)

// Xlsx 封装Excel文件操作
type Xlsx struct {
	FilePath string
	file     *excelize.File // 内部文件句柄
}

// NewXlsxFile 创建Xlsx实例
func NewXlsxFile(filePath string) *Xlsx {
	return &Xlsx{
		FilePath: filePath,
	}
}

// ReadFirstSheet 读取第一个工作表数据
func (x *Xlsx) ReadFirstSheet() ([]map[string]string, error) {
	if err := x.openFile(); err != nil {
		return nil, err
	}

	sheets, err := x.getSheets()
	if err != nil {
		return nil, err
	}
	if len(sheets) == 0 {
		return nil, fmt.Errorf("文件中没有工作表")
	}

	rows, err := x.getSheetRows(sheets[0])
	if err != nil {
		return nil, err
	}

	return x.processRows(rows), nil
}

// ReadSheetByName 按名称读取指定工作表数据
func (x *Xlsx) ReadSheetByName(sheetName string) ([]map[string]string, error) {
	if err := x.openFile(); err != nil {
		return nil, err
	}

	sheets, err := x.getSheets()
	if err != nil {
		return nil, err
	}

	// 快速检查工作表是否存在
	sheetMap := make(map[string]struct{}, len(sheets))
	for _, s := range sheets {
		sheetMap[s] = struct{}{}
	}
	if _, exists := sheetMap[sheetName]; !exists {
		return nil, fmt.Errorf("工作表[%s]不存在", sheetName)
	}

	rows, err := x.getSheetRows(sheetName)
	if err != nil {
		return nil, err
	}

	return x.processRows(rows), nil
}

// ReadAllSheets 读取所有工作表数据
func (x *Xlsx) ReadAllSheets() (map[string][]map[string]string, error) {
	if err := x.openFile(); err != nil {
		return nil, err
	}

	sheets, err := x.getSheets()
	if err != nil {
		return nil, err
	}

	result := make(map[string][]map[string]string, len(sheets))
	for _, sheet := range sheets {
		rows, err := x.getSheetRows(sheet)
		if err != nil {
			return nil, err
		}
		result[sheet] = x.processRows(rows)
	}

	return result, nil
}

// Close 关闭文件释放资源
func (x *Xlsx) Close() error {
	if x.file == nil {
		return nil
	}
	if err := x.file.Close(); err != nil {
		return fmt.Errorf("关闭文件失败: %w", err)
	}
	x.file = nil // 重置文件句柄
	return nil
}

// openFile 打开文件（内部使用，确保文件只被打开一次）
func (x *Xlsx) openFile() error {
	if x.file != nil {
		return nil
	}

	file, err := excelize.OpenFile(x.FilePath)
	if err != nil {
		return fmt.Errorf("打开文件失败: %w", err)
	}
	x.file = file
	return nil
}

// getSheets 获取所有工作表名称
func (x *Xlsx) getSheets() ([]string, error) {
	sheets := x.file.GetSheetList()
	if len(sheets) == 0 {
		return nil, fmt.Errorf("未找到任何工作表")
	}
	return sheets, nil
}

// getSheetRows 获取指定工作表的所有行数据
func (x *Xlsx) getSheetRows(sheetName string) ([][]string, error) {
	rows, err := x.file.GetRows(sheetName)
	if err != nil {
		return nil, fmt.Errorf("读取工作表[%s]数据失败: %w", sheetName, err)
	}
	return rows, nil
}

// processRows 处理行数据为map列表（键为表头，值为单元格内容）
func (x *Xlsx) processRows(rows [][]string) []map[string]string {
	if len(rows) == 0 {
		return []map[string]string{}
	}

	// 处理表头（去重、处理空表头）
	headers := rows[0]
	uniqueHeaders := x.processHeaders(headers)

	// 处理数据行
	result := make([]map[string]string, 0, len(rows)-1)
	for _, row := range rows[1:] {
		rowMap := make(map[string]string, len(uniqueHeaders))
		for colIdx, header := range uniqueHeaders {
			// 处理列数少于表头的情况
			if colIdx < len(row) {
				rowMap[header] = strings.TrimSpace(row[colIdx])
			} else {
				rowMap[header] = ""
			}
		}
		result = append(result, rowMap)
	}

	return result
}

// processHeaders 处理表头，确保唯一性
func (x *Xlsx) processHeaders(headers []string) []string {
	headerCount := make(map[string]int)
	uniqueHeaders := make([]string, len(headers))

	for i, h := range headers {
		trimmedHeader := strings.TrimSpace(h)
		// 处理空表头
		if trimmedHeader == "" {
			trimmedHeader = fmt.Sprintf("column_%d", i+1)
		}

		// 处理重复表头
		headerCount[trimmedHeader]++
		if headerCount[trimmedHeader] > 1 {
			uniqueHeaders[i] = fmt.Sprintf("%s_%d", trimmedHeader, headerCount[trimmedHeader])
		} else {
			uniqueHeaders[i] = trimmedHeader
		}
	}

	return uniqueHeaders
}
