/**
 * Description：
 * FileName：export.go
 * Author：CJiaの用心
 * Create：2025/10/17 12:04:36
 * Remark：
 */

package excelutil

import (
	"fmt"
	"github.com/xuri/excelize/v2"
	"go.uber.org/zap"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"
)

// ExcelColumn 列配置项
type ExcelColumn struct {
	Title     string                         // 表头标题
	Field     string                         // 结构体字段名（支持点号分隔的嵌套字段）
	Width     float64                        // 列宽
	Formatter func(value interface{}) string // 值格式化函数
	Style     *excelize.Style                // 列样式
}

// ExcelExportConfig Excel导出配置
type ExcelExportConfig struct {
	SheetName    string          // 工作表名称
	FileName     string          // 导出的文件名（不含后缀）
	StreamMode   bool            // 是否以流的方式导出
	Columns      []ExcelColumn   // 列配置
	Data         interface{}     // 要导出的数据（必须是切片）
	TimeFormat   string          // 时间格式化模板
	DefaultStyle *excelize.Style // 默认单元格样式
	HeaderStyle  *excelize.Style // 表头样式
	BasePath     string          // 文件保存的基础路径
}

// ExcelExporter Excel导出工具
type ExcelExporter struct {
	config *ExcelExportConfig
}

// NewExcelExporter 创建Excel导出器
func NewExcelExporter(cfg *ExcelExportConfig) *ExcelExporter {
	// 设置默认值
	if cfg.SheetName == "" {
		cfg.SheetName = "Sheet1"
	}
	if cfg.FileName == "" {
		cfg.FileName = fmt.Sprintf("导出数据_%s", time.Now().Format("20060102150405"))
	}
	if cfg.TimeFormat == "" {
		cfg.TimeFormat = "2006-01-02 15:04:05"
	}
	if cfg.DefaultStyle == nil {
		cfg.DefaultStyle = &excelize.Style{
			// 边框设置
			Border: []excelize.Border{
				{Type: "top", Color: "#000000", Style: 1},    // 顶部黑色实线
				{Type: "bottom", Color: "#000000", Style: 1}, // 底部黑色实线
				{Type: "left", Color: "#000000", Style: 1},   // 左侧黑色实线
				{Type: "right", Color: "#000000", Style: 1},  // 右侧黑色实线
			},
			// 对齐设置
			Alignment: &excelize.Alignment{
				Horizontal: "center", // 水平居中
				Vertical:   "center", // 垂直居中
			},
		}
	}
	if cfg.HeaderStyle == nil {
		cfg.HeaderStyle = &excelize.Style{
			// 字体设置
			Font: &excelize.Font{
				Family: "Noto Sans SC Medium", // 字体名称
				Bold:   true,                  // 加粗
				Size:   12,                    // 字体大小（可选设置）
			},
			// 背景填充
			Fill: excelize.Fill{
				Type:    "pattern",           // 填充类型：纯色填充
				Color:   []string{"#D3D3D3"}, // 灰色背景
				Pattern: 1,                   // 实心填充模式
			},
			// 边框设置
			Border: []excelize.Border{
				{Type: "top", Color: "#000000", Style: 1},    // 顶部黑色实线
				{Type: "bottom", Color: "#000000", Style: 1}, // 底部黑色实线
				{Type: "left", Color: "#000000", Style: 1},   // 左侧黑色实线
				{Type: "right", Color: "#000000", Style: 1},  // 右侧黑色实线
			},
			// 对齐设置
			Alignment: &excelize.Alignment{
				Horizontal: "center", // 水平居中
				Vertical:   "center", // 垂直居中
			},
		}
	}
	// 设置默认文件保存路径
	if cfg.BasePath == "" {
		// 默认为当前目录下的export目录
		cfg.BasePath = filepath.Join(".", "static/export")
	}

	return &ExcelExporter{config: cfg}
}

// Export 执行导出，返回Excel文件内容
func (e *ExcelExporter) Export() (*excelize.File, error) {
	f := excelize.NewFile()
	sheet := e.config.SheetName

	// 延迟关闭文件（仅释放资源，不影响已保存的文件）
	defer func() {
		if err := f.Close(); err != nil {
			zap.S().Errorf("关闭Excel文件失败: %v", err)
		}
	}()

	// 创建一个工作表
	index, err := f.NewSheet(sheet)
	if err != nil {
		zap.S().Errorf("[ExcelExport] 创建工作表[%s]失败: %v", sheet, err)
		return nil, err
	}
	// 设置当前活动工作表
	f.SetActiveSheet(index)
	// 安全删除默认Sheet1（避免删除当前工作表）
	if sheet != "Sheet1" {
		if err = f.DeleteSheet("Sheet1"); err != nil {
			zap.S().Warnf("[ExcelExport] 删除默认工作表[Sheet1]失败: %v", err)
		}
	}

	// ======== 关键修复：创建并应用表头样式 ========
	var headerStyleID int
	if e.config.HeaderStyle != nil {
		// 1. 在循环外只创建一次样式
		if headerStyleID, err = f.NewStyle(e.config.HeaderStyle); err != nil {
			zap.S().Errorf("创建表头样式失败: %v", err)
		} else {
			// 2. 设置表头行高度
			if err = f.SetRowHeight(sheet, 1, 24); err != nil {
				zap.S().Warnf("设置表头行高失败: %v", err)
			}

			// 3. 设置全行样式
			firstCell, _ := excelize.CoordinatesToCellName(1, 1)
			lastCell, _ := excelize.CoordinatesToCellName(len(e.config.Columns), 1)
			if err = f.SetCellStyle(sheet, firstCell, lastCell, headerStyleID); err != nil {
				zap.S().Errorf("设置表头样式失败: %v", err)
			}
		}
	}

	// 创建表头
	for colIdx, col := range e.config.Columns {
		cell, _ := excelize.CoordinatesToCellName(colIdx+1, 1)

		// 设置表头值
		if err := f.SetCellValue(sheet, cell, col.Title); err != nil {
			zap.S().Errorf("设置表头值失败: %v", err)
		}

		// 设置列宽
		if col.Width > 0 {
			colName := getColName(colIdx + 1)
			if err = f.SetColWidth(sheet, colName, colName, col.Width); err != nil {
				zap.S().Warnf("设置列宽失败: %v", err)
			}
		}
	}

	// 准备数据源
	dataSlice := reflect.ValueOf(e.config.Data)
	if dataSlice.Kind() != reflect.Slice {
		zap.S().Warnf("导出数据必须是切片类型, 实际是: %s", dataSlice.Kind().String())
		return nil, fmt.Errorf("导出数据必须是切片类型, 实际是: %s", dataSlice.Kind().String())
	}

	startRow := 2 // 表头占用第1行，数据从第2行开始

	// 填充数据
	for rowIdx := 0; rowIdx < dataSlice.Len(); rowIdx++ {
		rowValue := dataSlice.Index(rowIdx)
		currentRow := startRow + rowIdx

		// 处理指针值
		if rowValue.Kind() == reflect.Ptr {
			rowValue = rowValue.Elem()
		}

		// 遍历列并设置单元格值
		for colIdx, col := range e.config.Columns {
			cell, _ := excelize.CoordinatesToCellName(colIdx+1, currentRow)

			// 获取字段值
			fieldValue, err := getFieldValue(rowValue, col.Field)
			if err != nil {
				// 处理获取字段值错误
				zap.S().Warnf("获取字段值失败: %s, %v", col.Field, err)
				fieldValue = "" // 使用空值代替
				continue
			}

			// 格式化值
			formattedValue := formatValue(fieldValue, col.Formatter, e.config.TimeFormat)

			// 设置单元格样式
			if col.Style != nil {
				// 强制设置居中
				col.Style.Alignment = &excelize.Alignment{
					Horizontal: "center", // 水平居中
					Vertical:   "center", // 垂直居中
				}
				if styleID, err := f.NewStyle(col.Style); err == nil {
					if err := f.SetCellStyle(sheet, cell, cell, styleID); err != nil {
						zap.S().Warnf("设置单元格样式失败: %s, %v", cell, err)
						return nil, err
					}
				}
			} else if e.config.DefaultStyle != nil {
				if styleID, err := f.NewStyle(e.config.DefaultStyle); err == nil {
					if err := f.SetCellStyle(sheet, cell, cell, styleID); err != nil {
						zap.S().Warnf("设置单元格样式失败: %s, %v", cell, err)
						return nil, err
					}
				}
			}

			// 设置单元格值
			if err := f.SetCellValue(sheet, cell, formattedValue); err != nil {
				zap.S().Warnf("设置单元格值失败: %s, %v", cell, err)
				return nil, err
			}
		}
	}

	// 根据指定模式导出文件
	if e.config.StreamMode {
		// 导出文件流
		return f, err
	} else {
		// 创建日期目录（格式：年月日）
		dateDir := time.Now().Format("20060102")
		// 构建完整路径：基础路径 + 日期目录
		fullPath := filepath.Join(e.config.BasePath, dateDir)
		// 确保目录存在
		if err := os.MkdirAll(fullPath, 0755); err != nil {
			zap.S().Errorf("创建目录失败: %s, %v", fullPath, err)
			return nil, fmt.Errorf("创建目录失败: %w", err)
		}
		// 构建完整文件名：路径 + 文件名 + 后缀
		fileName := e.config.FileName + ".xlsx"
		outputPath := filepath.Join(fullPath, fileName)
		// 保存文件到指定路径
		if err := f.SaveAs(outputPath); err != nil {
			zap.S().Errorf("保存文件失败: %v", err)
			return nil, err
		}
	}

	return nil, nil
}

// getFieldValue 递归获取结构体字段值（支持嵌套结构体）
func getFieldValue(v reflect.Value, fieldPath string) (interface{}, error) {
	if fieldPath == "" || !v.IsValid() {
		return nil, nil
	}

	// 尝试解析点号分隔的字段路径
	fields := strings.Split(fieldPath, ".")
	currentField := fields[0]

	var fieldValue reflect.Value

	switch v.Kind() {
	case reflect.Struct:
		fieldValue = v.FieldByName(currentField)
		if !fieldValue.IsValid() {
			return nil, fmt.Errorf("结构体字段 %s 不存在", currentField)
		}
	case reflect.Map:
		if v.Type().Key().Kind() != reflect.String {
			return nil, fmt.Errorf("map 键类型必须为 string")
		}
		mapValue := v.MapIndex(reflect.ValueOf(currentField))
		if !mapValue.IsValid() {
			return nil, fmt.Errorf("map 键 %s 不存在", currentField)
		}
		fieldValue = mapValue
	default:
		return nil, fmt.Errorf("不支持的类型: %s", v.Kind().String())
	}

	// 处理指针
	if fieldValue.Kind() == reflect.Ptr {
		if fieldValue.IsNil() {
			return nil, nil
		}
		fieldValue = fieldValue.Elem()
	}

	// 如果还有嵌套字段，递归查找
	if len(fields) > 1 {
		remainingPath := strings.Join(fields[1:], ".")
		return getFieldValue(fieldValue, remainingPath)
	}

	return fieldValue.Interface(), nil
}

// formatValue 格式化值
func formatValue(value interface{}, formatter func(interface{}) string, timeFormat string) interface{} {
	if formatter != nil {
		return formatter(value)
	}

	if timeFormat == "" {
		timeFormat = "2006-01-02 15:04:05"
	}

	// 对时间类型进行默认格式化
	switch v := value.(type) {
	case time.Time:
		return v.Format(timeFormat)
	case *time.Time:
		if v != nil {
			return v.Format(timeFormat)
		}
	}

	return value
}

// getColName 根据列索引获取列名
func getColName(colIdx int) string {
	name, _ := excelize.ColumnNumberToName(colIdx)
	return name
}
