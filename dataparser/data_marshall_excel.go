// Package main
// @description
// @author      梁志豪
// @datetime    2025/12/15 11:10
package dataparser

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"github.com/labstack/gommon/log"
	"github.com/tealeg/xlsx/v3"
	"io/ioutil"
)

type ExcelDataMarshall struct {
	*Workbook
	*xlsx.File
}

//go:embed template.xlsx
var tpl []byte

func (dp *ExcelDataMarshall) ReadFile(filepath string) []byte {
	fileData, err := ioutil.ReadFile(filepath)
	if err != nil {
		log.Errorf("读取文件失败:%v", err)
	}
	return fileData
}

func (dp *ExcelDataMarshall) LoadWorkbook(data []byte) (err error) {
	var wb *Workbook
	err = json.Unmarshal(data, &wb)
	if err != nil {
		return
	}
	dp.Workbook = wb
	return
}

func (dp *ExcelDataMarshall) Marshall() error {
	// 新建 XLSX 文件
	file := xlsx.NewFile()

	// 循环按顺序创建并填充 sheet
	for _, orderName := range dp.Workbook.SheetOrder {
		sheetInfo, ok := dp.Workbook.Sheets[orderName]
		if !ok {
			return fmt.Errorf("不存在的sheet: %s", orderName)
		}

		// 添加 sheet
		sheet, err := file.AddSheet(sheetInfo.Name)
		if err != nil {
			return fmt.Errorf("新建 sheet %s 失败: %w", sheetInfo.Name, err)
		}

		if sheetInfo.RowCount < 1 {
			return fmt.Errorf("不完整的配置数据: %s", sheetInfo.Name)
		}

		// 填充数据
		for r := 0; r < sheetInfo.RowCount; r++ {
			row := sheet.AddRow()
			for c := 0; c < sheetInfo.ColumnCount; c++ {
				cell := row.AddCell()
				// 避免空/nil 展示成字符串 "nil"
				if r < len(sheetInfo.CellData) && c < len(sheetInfo.CellData[r]) {
					val := sheetInfo.CellData[r][c].V
					if val == nil {
						cell.SetString("")
						continue
					}
					switch v := val.(type) {
					case string:
						cell.SetString(v)
					case int:
						cell.SetInt(v)
					case int64:
						cell.SetInt(int(v))
					case float64:
						cell.SetFloat(v)
					case bool:
						cell.SetBool(v)
					default:
						cell.SetString(fmt.Sprintf("%v", v))
					}
				} else {
					cell.SetString("") // 空白
				}
			}
		}
	}

	dp.File = file
	return nil
}
