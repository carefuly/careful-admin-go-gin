/**
 * Description：
 * FileName：xlsx_test.go.go
 * Author：CJiaの用心
 * Create：2025/10/17 10:14:33
 * Remark：
 */

package xlsx

import (
	"testing"
)

func TestXlsx_ReadFirstSheet(t *testing.T) {
	testCases := []struct {
		name      string
		path      string
		wantFirst map[string]string
		wantRows  int
		wantCols  int
		wantErr   bool
	}{
		{
			name: "正常读取第一个sheet",
			path: "xlsx_test.xlsx",
			wantFirst: map[string]string{
				"Name": "Alice",
				"Age":  "30",
			},
			wantRows: 2,
			wantCols: 2,
			wantErr:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			xlsxFile := NewXlsxFile(tc.path)
			data, err := xlsxFile.ReadFirstSheet()

			// 错误校验
			if (err != nil) != tc.wantErr {
				t.Fatalf("预期错误: %v, 实际错误: %v", tc.wantErr, err)
			}
			if tc.wantErr {
				return
			}

			// 行数校验
			if len(data) != tc.wantRows {
				t.Errorf("预期行数: %d, 实际行数: %d", tc.wantRows, len(data))
			}

			// 列数校验
			if len(data[0]) != tc.wantCols {
				t.Errorf("预期列数: %d, 实际列数: %d", tc.wantCols, len(data[0]))
			}

			// 第一行数据校验
			firstRow := data[0]
			for k, v := range tc.wantFirst {
				if firstRow[k] != v {
					t.Errorf("字段[%s]预期值: %s, 实际值: %s", k, v, firstRow[k])
				}
			}
		})
	}
}

func TestXlsx_ReadSheetByName(t *testing.T) {
	testCases := []struct {
		name      string
		path      string
		sheetName string
		wantFirst map[string]string
		wantRows  int
		wantCols  int
		wantErr   string
	}{
		{
			name:      "读取存在的Sheet2",
			path:      "xlsx_test.xlsx",
			sheetName: "Sheet2",
			wantRows:  2, // Sheet2有1行数据
			wantCols:  2,
			wantFirst: map[string]string{
				"ID":    "1001",
				"Score": "95",
			},
			wantErr: "",
		},
		{
			name:      "读取存在的Sheet3（含特殊表头）",
			path:      "xlsx_test.xlsx",
			sheetName: "Sheet3",
			wantRows:  2, // Sheet3有2行数据
			wantCols:  2,
			wantFirst: map[string]string{ // 注意重复表头和空表头的处理结果
				"Value": "x",
			},
			wantErr: "",
		},
		{
			name:      "读取不存在的Sheet",
			path:      "xlsx_test.xlsx",
			sheetName: "Sheet4",
			wantRows:  0,
			wantCols:  0,
			wantFirst: nil,
			wantErr:   "工作表[Sheet4]不存在", // 预期错误信息
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			xlsxFile := NewXlsxFile(tc.path)
			data, err := xlsxFile.ReadSheetByName(tc.sheetName)

			// 错误校验
			if tc.wantErr != "" {
				if err == nil {
					t.Fatal("预期错误但未发生错误")
				}
				if err.Error() != tc.wantErr {
					t.Errorf("预期错误: %s, 实际错误: %s", tc.wantErr, err.Error())
				}
				return
			}
			if err != nil {
				t.Fatalf("未预期错误: %v", err)
			}

			// 行数校验
			if len(data) != tc.wantRows {
				t.Errorf("预期行数: %d, 实际行数: %d", tc.wantRows, len(data))
			}

			// 列数校验
			if len(data[0]) != tc.wantCols {
				t.Errorf("预期列数: %d, 实际列数: %d", tc.wantCols, len(data[0]))
			}

			// 第一行数据校验
			firstRow := data[0]
			for k, v := range tc.wantFirst {
				if firstRow[k] != v {
					t.Errorf("字段[%s]预期值: %s, 实际值: %s", k, v, firstRow[k])
				}
			}
		})
	}
}

func TestXlsx_ReadAllSheets(t *testing.T) {
	testCases := []struct {
		name       string
		path       string
		wantSheets int
		wantSheet1 map[string]string
		wantSheet3 map[string]string
		wantErr    bool
	}{
		{
			name:       "读取所有3个sheet",
			path:      "xlsx_test.xlsx",
			wantSheets: 3,
			wantSheet1: map[string]string{ // Sheet1第一行
				"Name": "Alice",
				"Age":  "30",
			},
			wantSheet3: map[string]string{ // Sheet3第一行（特殊表头处理）
				"Value": "x",
			},
			wantErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			xlsxFile := NewXlsxFile(tc.path)
			allData, err := xlsxFile.ReadAllSheets()

			// 错误校验
			if (err != nil) != tc.wantErr {
				t.Fatalf("预期错误: %v, 实际错误: %v", tc.wantErr, err)
			}
			if tc.wantErr {
				return
			}

			// Sheet数量校验
			if len(allData) != tc.wantSheets {
				t.Errorf("预期Sheet数量: %d, 实际数量: %d", tc.wantSheets, len(allData))
			}

			// 校验Sheet1数据
			sheet1Data, ok := allData["Sheet1"]
			if !ok {
				t.Fatal("未找到Sheet1数据")
			}
			firstRow1 := sheet1Data[0]
			for k, v := range tc.wantSheet1 {
				if firstRow1[k] != v {
					t.Errorf("Sheet1字段[%s]预期值: %s, 实际值: %s", k, v, firstRow1[k])
				}
			}

			// 校验Sheet3数据（含特殊表头处理）
			sheet3Data, ok := allData["Sheet3"]
			if !ok {
				t.Fatal("未找到Sheet3数据")
			}
			firstRow3 := sheet3Data[0]
			for k, v := range tc.wantSheet3 {
				if firstRow3[k] != v {
					t.Errorf("Sheet3字段[%s]预期值: %s, 实际值: %s", k, v, firstRow3[k])
				}
			}
		})
	}
}
