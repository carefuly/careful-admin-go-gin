/**
 * Description：
 * FileName：export_test.go.go
 * Author：CJiaの用心
 * Create：2025/10/17 13:17:14
 * Remark：
 */

package excelutil

import (
	"fmt"
	"testing"
	"time"
)

func TestExcelExporter_Export(t *testing.T) {
	type fields struct {
		config *ExcelExportConfig
	}

	testCases := []struct {
		name    string
		fields  fields
		want    []byte
		wantErr bool
	}{
		{
			name: "用户数据导出[带嵌套]",
			fields: fields{
				config: &ExcelExportConfig{
					SheetName:  "用户数据",
					FileName:   "用户导出",
					StreamMode: false,
					Columns: []ExcelColumn{
						{Title: "ID", Field: "ID", Width: 8},
						{Title: "用户名", Field: "Username", Width: 15},
						{Title: "邮箱", Field: "Email", Width: 15},
						{
							Title: "状态",
							Field: "Active",
							Width: 8,
							Formatter: func(value interface{}) string {
								if active, ok := value.(bool); ok {
									if active {
										return "激活"
									}
									return "冻结"
								}
								return "未知"
							},
						},
						{Title: "创建时间", Field: "CreatedAt", Width: 20},
						{
							Title: "部门",
							Field: "Department.Name", // 嵌套字段
							Width: 15,
						},
					},
					Data: []struct {
						ID         int
						Username   string
						Email      string
						Active     bool
						CreatedAt  time.Time
						Department struct {
							ID   int
							Name string
						}
					}{
						{
							ID:        1,
							Username:  "zhang",
							Email:     "zhang@qq.com",
							Active:    true,
							CreatedAt: time.Date(2025, 6, 27, 12, 05, 50, 38, time.Local),
							Department: struct {
								ID   int
								Name string
							}{
								ID:   1,
								Name: "技术部",
							},
						},
						{
							ID:        2,
							Username:  "li",
							Email:     "li@qq.com",
							Active:    false,
							CreatedAt: time.Date(2025, 6, 27, 12, 05, 50, 38, time.Local),
							Department: struct {
								ID   int
								Name string
							}{
								ID:   2,
								Name: "市场部",
							},
						},
						{
							ID:        3,
							Username:  "wang",
							Email:     "wang@qq.com",
							Active:    true,
							CreatedAt: time.Date(2025, 6, 27, 12, 05, 50, 38, time.Local),
							Department: struct {
								ID   int
								Name string
							}{
								ID:   3,
								Name: "运营部",
							},
						},
						{
							ID:        4,
							Username:  "ma",
							Email:     "ma@qq.com",
							Active:    false,
							CreatedAt: time.Date(2025, 6, 27, 12, 05, 50, 38, time.Local),
							Department: struct {
								ID   int
								Name string
							}{
								ID:   4,
								Name: "财务部",
							},
						},
					},
				},
			},
		},
		{
			name: "产品库存导出",
			fields: fields{
				config: &ExcelExportConfig{
					SheetName:  "产品库存",
					FileName:   "产品库存",
					StreamMode: false,
					Columns: []ExcelColumn{
						{Title: "产品ID", Field: "ID", Width: 10},
						{Title: "产品名称", Field: "Name", Width: 15},
						{Title: "库存数量", Field: "Stock", Width: 13},
						{
							Title: "上次盘点时间",
							Field: "LastCheckAt",
							Formatter: func(value interface{}) string {
								if t, ok := value.(time.Time); ok {
									if t.IsZero() {
										return "未盘点"
									}
									return t.Format("2006-01-02")
								}
								return "无效日期"
							},
							Width: 20,
						},
						{Title: "创建时间", Field: "CreatedAt", Width: 20},
					},
					Data: []struct {
						ID          int
						Name        string
						Stock       int
						LastCheckAt time.Time
						CreatedAt   time.Time
					}{
						{
							ID:          1,
							Name:        "苹果",
							Stock:       100,
							LastCheckAt: time.Date(2025, 6, 27, 12, 05, 50, 38, time.Local),
							CreatedAt:   time.Date(2025, 6, 27, 13, 06, 45, 45, time.Local),
						},
						{
							ID:          2,
							Name:        "香蕉",
							Stock:       200,
							LastCheckAt: time.Date(2025, 6, 27, 12, 05, 50, 38, time.Local),
							CreatedAt:   time.Date(2025, 6, 27, 13, 06, 45, 45, time.Local),
						},
						{
							ID:          3,
							Name:        "橘子",
							Stock:       300,
							LastCheckAt: time.Date(2025, 6, 27, 12, 05, 50, 38, time.Local),
							CreatedAt:   time.Date(2025, 6, 27, 13, 06, 45, 45, time.Local),
						},
					},
				},
			},
		},
	}
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			e := NewExcelExporter(tt.fields.config)

			bytes, err := e.Export()
			if err != nil {
				t.Errorf("导出文件失败：%v", err)
				return
			}

			fmt.Println("bytes", bytes)
		})
	}
}
