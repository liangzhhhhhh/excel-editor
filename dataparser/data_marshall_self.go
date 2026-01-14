// Package main
// @description
// @author      梁志豪
// @datetime    2025/12/15 11:10
package dataparser

import (
	"encoding/json"
	"fmt"
	"github.com/labstack/gommon/log"
	"github.com/tidwall/sjson"
	"io/ioutil"
	"strings"
)

type SelfDataMarshall struct {
	*Workbook
}

func (dp *SelfDataMarshall) ReadFile(filepath string) []byte {
	fileData, err := ioutil.ReadFile(filepath)
	if err != nil {
		log.Errorf("读取文件失败:%v", err)
	}
	return fileData
}

func (dp *SelfDataMarshall) WriteFile(filepath string, data []byte) {
	err := ioutil.WriteFile(filepath, data, 0644)
	if err != nil {
		log.Errorf("创建文件失败:%v", err)
	}
}

func (dp *SelfDataMarshall) LoadWorkbook(data []byte) (err error) {
	var wb *Workbook
	err = json.Unmarshal(data, &wb)
	if err != nil {
		return
	}
	dp.Workbook = wb
	return
}

func (dp *SelfDataMarshall) Marshall() (headerConfig, contentConfig []byte, err error) {

	actConfigHeaders := map[string]SheetBaseConfig{}
	// 通过sheetName获取sheet具体信息
	actConfigContent := "{}"
	for sheetIndex, sheetName := range dp.Workbook.SheetOrder {
		sheetInfo, ok := dp.Workbook.Sheets[sheetName]
		if !ok {
			err = fmt.Errorf("不存在的sheet:%s", sheetName)
			return
		}
		if sheetInfo.RowCount < 3 {
			// 数据设置不完整
			err = fmt.Errorf("不完整的配置数据:%s", sheetName)
			return
		}
		// 首先设置表头字段
		Cns, Ens, Types, Rules := make([]string, sheetInfo.ColumnCount), make([]string, sheetInfo.ColumnCount), make([]string, sheetInfo.ColumnCount), make([]string, sheetInfo.ColumnCount)
		Pks := make([]string, 0)
		for index := range sheetInfo.ColumnCount {
			if sheetInfo.CellData[0][index].V == nil {
				sheetInfo.ColumnCount = index
				Cns = Cns[:index]
				Ens = Ens[:index]
				Types = Types[:index]
				break
			}
			cn := sheetInfo.CellData[0][index].V.(string)
			secondRowValue := sheetInfo.CellData[1][index].V.(string)
			thirdRowValue := sheetInfo.CellData[2][index].V.(string)
			var rowValues []string
			if rowValues = strings.Split(secondRowValue, "_"); len(rowValues) > 1 {
				if rowValues[1] != Pk {
					err = fmt.Errorf("sheet:%s中属性:%s格式不规范", sheetName, secondRowValue)
					return
				}
				Pks = append(Pks, rowValues[0])
			}
			Ens[index] = rowValues[0]
			if rowValues = strings.Split(thirdRowValue, "_"); len(rowValues) > 1 {
				Rules[index] = rowValues[1]
			}
			Types[index] = rowValues[0]
			Cns[index] = cn
		}
		baseConfig := SheetBaseConfig{OrderIndex: sheetIndex, OrderName: sheetInfo.Name, Header: &Header{En: Ens, Cn: Cns, Type: Types, Rule: Rules, Pk: Pks}}
		actConfigHeaders[sheetInfo.Name] = baseConfig
		// 设置content
		sheetData := ""
		for rowIndex := range sheetInfo.CellData {
			if rowIndex < 3 {
				continue
			}
			rowData := "{}"
			for colIndex := range sheetInfo.ColumnCount {
				cellValue := sheetInfo.CellData[rowIndex][colIndex].V
				rowData, _ = sjson.Set(rowData, baseConfig.Header.En[colIndex], cellValue)
			}
			if sheetInfo.RowCount <= 4 {
				sheetData = rowData
				break
			}

			if sheetData == "" {
				sheetData = "[]"
			}
			sheetData, _ = sjson.SetRaw(sheetData, "-1", rowData)
		}
		// sheetInfo.Name 表示的是例如CommonMediumActiPrice_中期活动促销这种，实际上我只需要CommonMediumActiPrice
		actConfigContent, err = sjson.SetRaw(actConfigContent, strings.Split(sheetInfo.Name, "_")[0], sheetData)
		if err != nil {
			return
		}
	}
	headerConfig, err = json.Marshal(actConfigHeaders)
	if err != nil {
		return
	}
	contentConfig = []byte(actConfigContent)
	return
}
