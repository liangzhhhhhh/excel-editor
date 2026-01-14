// Package main
// @description
// @author      梁志豪
// @datetime    2025/12/12 19:31
package dataparser

import (
	"encoding/json"
	"fmt"
	"github.com/xuri/excelize/v2"
	"os"
	"strconv"
	"strings"
)

type ExcelDataUnmarshall struct {
	FilePath string
	*Workbook
}

func NewExcelDataUnmarshall(filePath, actId, ab string) (*ExcelDataUnmarshall, error) {
	workbookId, err := GetWorkbookName(actId, ab)
	if err != nil {
		return nil, err
	}
	return &ExcelDataUnmarshall{
		FilePath: filePath,
		Workbook: &Workbook{
			Id:     workbookId,
			Name:   workbookId,
			Sheets: make(map[string]Sheet),
		},
	}, nil
}

func (dp *ExcelDataUnmarshall) WriteFile(filepath string) (err error) {
	var val []byte
	val, err = json.Marshal(dp.Workbook)
	if err != nil {
		return fmt.Errorf("json marshal failed: %w", err)
	}
	err = os.WriteFile(filepath, val, 0644)
	if err != nil {
		return fmt.Errorf("write file failed: %w", err)
	}
	return nil
}

// Parse
// @Description: 强调：传入Excel格式必须按照 Activity_7XXX格式传入
// @author liangzh
// @update 2025-12-16 18:47:54
func (dp *ExcelDataUnmarshall) Parse() (err error) {
	file, err := excelize.OpenFile(dp.FilePath)
	if err != nil {
		return fmt.Errorf("open file failed: %w", err)
	}
	defer file.Close()

	// 获取sheet列表
	sheets := file.GetSheetList()
	for sheetIdx, sheet := range sheets {
		var (
			contentConfig = Sheet{
				CellData: make(map[int]map[int]CellData),
			}
			headerConfig SheetBaseConfig
		)
		// sheet > 真实名称
		sheetName := file.GetSheetName(sheetIdx)
		rows, err := file.GetRows(sheet)
		if err != nil {
			return fmt.Errorf("get rows failed: %w", err)
		}
		// "sheet-1" > 用于排序使用的名称
		headerConfig.OrderName = fmt.Sprintf("sheet-%d", sheetIdx)
		headerConfig.OrderIndex = sheetIdx
		if len(rows) < 3 {
			err = fmt.Errorf("sheet格式不正确:%s", sheetName)
			return err
		}
		contentConfig.Id = headerConfig.OrderName
		contentConfig.Name = sheetName
		propertyNum := len(rows[0])
		headerConfig.Header = &Header{
			Cn:          make([]string, propertyNum),
			En:          make([]string, propertyNum),
			Type:        make([]string, propertyNum),
			Rule:        make([]string, propertyNum),
			PropertyMap: make(map[string]*Property),
		}
		contentConfig.CellData[0] = make(map[int]CellData)
		contentConfig.CellData[1] = make(map[int]CellData)
		contentConfig.CellData[2] = make(map[int]CellData)
		for colIndex := range propertyNum {
			if rows[0][colIndex] == "" {
				headerConfig.Header.Cn = headerConfig.Header.Cn[:colIndex]
				headerConfig.Header.En = headerConfig.Header.En[:colIndex]
				headerConfig.Header.Type = headerConfig.Header.Type[:colIndex]
				break
			}
			headerConfig.Header.Cn[colIndex] = rows[0][colIndex]
			secondRowValue := rows[1][colIndex]
			thirdRowValue := rows[2][colIndex]
			var rowValues []string
			var hasPk bool
			if rowValues = strings.Split(secondRowValue, "_"); len(rowValues) > 1 {
				headerConfig.Header.Pk = append(headerConfig.Header.Pk, rowValues[0])
				hasPk = true
			}
			headerConfig.Header.En[colIndex] = rowValues[0]
			if rowValues = strings.Split(thirdRowValue, "_"); len(rowValues) > 1 {
				headerConfig.Header.Rule[colIndex] = rowValues[1]
			}
			headerConfig.Header.Type[colIndex] = rowValues[0]
			headerConfig.Header.PropertyMap[rows[1][colIndex]] = &Property{
				Index: colIndex,
				Cn:    rows[0][colIndex],
				En:    headerConfig.Header.En[colIndex],
				Type:  headerConfig.Header.Type[colIndex],
				Rule:  headerConfig.Header.Rule[colIndex],
				Pk:    hasPk,
			}
			contentConfig.CellData[0][colIndex] = CellData{V: rows[0][colIndex]}
			contentConfig.CellData[1][colIndex] = CellData{V: secondRowValue}
			contentConfig.CellData[2][colIndex] = CellData{V: thirdRowValue}
			contentConfig.ColumnCount++
		}
		contentConfig.RowCount = 3
		// 行数据
		for rowIdx, row := range rows {
			if rowIdx < 3 {
				continue
			}
			// 整行判断，如果都是没有值则表示后续无值了
			isEnd := true
			for colIdx, colData := range row {
				if colIdx >= contentConfig.ColumnCount {
					break
				}
				if colData != "" {
					isEnd = false
					break
				}
			}
			if isEnd {
				break
			}
			contentConfig.CellData[rowIdx], err = dp.UnmarshallRow(headerConfig, contentConfig.ColumnCount, row)
			if err != nil {
				return fmt.Errorf("unmarshall failed: %w", err)
			}
			contentConfig.RowCount++
		}
		dp.Workbook.SheetOrder = append(dp.Workbook.SheetOrder, headerConfig.OrderName)
		dp.Workbook.Sheets[headerConfig.OrderName] = contentConfig
	}
	return
}

// UnmarshallRow
// @Description: 处理行数据
// @author liangzh
// @update 2025-12-11 17:05:24
func (dp *ExcelDataUnmarshall) UnmarshallRow(sheetBaseConfig SheetBaseConfig, colNum int, rowData []string) (res map[int]CellData, err error) {
	res = make(map[int]CellData)
	for colIdx, colData := range rowData {
		if colIdx >= colNum {
			break
		}
		switch sheetBaseConfig.Header.Type[colIdx] {
		case INT:
			var value int
			value, err = strconv.Atoi(colData)
			res[colIdx] = CellData{V: value}
		case FLOAT:
			var value float64
			value, err = strconv.ParseFloat(colData, 64)
			res[colIdx] = CellData{V: value}
		case STRING, JSON:
			res[colIdx] = CellData{V: colData}
		case BOOL:
			var value bool
			value, err = strconv.ParseBool(colData)
			res[colIdx] = CellData{V: value}
		}
		if err != nil {
			return
		}
	}
	return
}
