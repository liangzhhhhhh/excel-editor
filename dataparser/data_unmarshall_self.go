// Package dataparser
// @description
// @author      梁志豪
// @datetime    2025/12/11 15:12
package dataparser

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"slices"
	"strconv"
	"strings"

	"github.com/labstack/gommon/log"
	"github.com/tidwall/gjson"
)

type SelfDataUnmarshall struct {
	ActId string `json:"-"`
	WorkbookBaseConfig
}

func (dp *SelfDataUnmarshall) ReadFile(filepath string) []byte {
	fileData, err := ioutil.ReadFile(filepath)
	if err != nil {
		log.Errorf("读取文件失败:%v", err)
	}
	return fileData
}

//func (dp *SelfDataUnmarshall) WriteFile(filepath string) (err error) {
//	var val []byte
//	val, err = json.Marshal(dp.Workbook)
//	if err != nil {
//		return fmt.Errorf("json marshal failed: %w", err)
//	}
//	err = os.WriteFile(filepath, val, 0644)
//	if err != nil {
//		return fmt.Errorf("write file failed: %w", err)
//	}
//	return nil
//}

// UnmarshallHeader
// @Description: sheetId组成规则：sheet- 加上 sheetIndex=>用此来排序
// @author liangzh
// @update 2025-12-11 18:08:55
func (dp *SelfDataUnmarshall) UnmarshallHeader(content []byte) (err error) {
	var originConfig map[string]SheetBaseConfig
	cfg := WorkbookBaseConfig{
		SheetConfigMap: make(map[string]SheetBaseConfig),
	}

	err = json.Unmarshal(content, &originConfig)
	if err != nil {
		return
	}
	var sheetOrder []string
	var orderIndexes []int
	// 去除key的_后缀
	for key, value := range originConfig {
		if len(value.Header.Cn) != len(value.Header.En) || len(value.Header.En) != len(value.Header.Type) {
			err = errors.New("表头数据长度设定不一致")
			return
		}
		mainKeys := strings.Split(key, "_")
		// key可能由英文名+_+中文名组成;当前需要分离
		sheetKey := mainKeys[0]

		orderIndexes = append(orderIndexes, value.OrderIndex)
		if _, convErr := strconv.Atoi(mainKeys[len(mainKeys)-1]); convErr == nil {
			sheetKey = mainKeys[0] + "-" + mainKeys[len(mainKeys)-1]
			// cfg.ConfigLevel.Open = true
			if !slices.Contains(cfg.ConfigLevel.SheetKeys, sheetKey) {
				cfg.ConfigLevel.SheetKeys = append(cfg.ConfigLevel.SheetKeys, sheetKey)
			}
		} else {
			// 此处表达的是 Contant-1 这种形式
			sheetKey = mainKeys[0] + "-1"
		}
		cfg.SheetConfigMap[sheetKey] = value
	}
	// 数据装载
	for key, value := range cfg.SheetConfigMap {
		if len(value.Header.Rule) == 0 {
			value.Header.Rule = make([]string, len(value.Header.Cn))
		}
		value.Header.PropertyMap = make(map[string]*Property)
		for i := 0; i < len(value.Header.En); i++ {
			value.Header.PropertyMap[value.Header.En[i]] = &Property{
				En:    value.Header.En[i],
				Cn:    value.Header.Cn[i],
				Type:  value.Header.Type[i],
				Rule:  value.Header.Rule[i],
				Index: i,
			}
		}
		// pk存储的是En,所以判断这个字段是否存在，存在则设置pk为true
		for _, pk := range value.Header.Pk {
			if property, ok := value.Header.PropertyMap[pk]; !ok {
				err = fmt.Errorf("pk：%s在sheet：%s中不存在", pk, key)
				return
			} else {
				property.Pk = true
			}
		}
	}
	slices.Sort(orderIndexes)
	for _, orderIndex := range orderIndexes {
		sheetOrder = append(sheetOrder, fmt.Sprintf("sheet-%d", orderIndex))
	}
	cfg.SheetOrder = sheetOrder
	dp.WorkbookBaseConfig = cfg
	return
}

func (dp *SelfDataUnmarshall) UnmarshallContent(ab string, content []byte) (workbook Workbook, err error) {
	workBooksData := gjson.Parse(string(content))
	if workBooksData.Get("HighValue").Exists() {
		if ab == "" {
			ab = "A"
		}
		workBooksData.ForEach(func(highValueKey, highValueContent gjson.Result) bool {
			if highValueKey.Str == "HighValue" {
				highValueContent.ForEach(func(abKey, abValue gjson.Result) bool {
					if (abKey.Str == "A" || abKey.Str == "B") && ab == abKey.Str {
						workbook = dp.tidyWorkBookUnmarshall(abValue)
						workbookKey, _ := GetWorkbookName(dp.ActId, abKey.Str)
						workbook.Id = workbookKey
						workbook.Name = workbookKey
					}
					return true
				})
			}
			return true
		})
	} else {
		if ab != "" {
			err = errors.New("数据解析失败")
			return
		}
		workbook = dp.tidyWorkBookUnmarshall(workBooksData)
		workbookKey, _ := GetWorkbookName(dp.ActId, "")
		workbook.Id = workbookKey
		workbook.Name = workbookKey
	}
	return
}

func (dp *SelfDataUnmarshall) tidyWorkBookUnmarshall(workBookData gjson.Result) (workbook Workbook) {
	workbook = Workbook{
		Sheets:     map[string]Sheet{},
		SheetOrder: dp.WorkbookBaseConfig.SheetOrder,
	}
	//  进行了分层
	// if dp.ConfigLevel.Open {
	if workBookData.Get("ConfigIDs").Exists() {
		workBookData.ForEach(func(configKey, configContent gjson.Result) bool {
			configContent.ForEach(func(subConfigKey, subConfigContent gjson.Result) bool {
				sheets := dp.tidyBaseUnmarshall(subConfigKey.Str, subConfigContent)
				if len(sheets) > 0 {
					for _, sheet := range sheets {
						workbook.Sheets[sheet.Id] = sheet
					}
				}
				return true
			})
			return true
		})
	} else {
		sheets := dp.tidyBaseUnmarshall("1", workBookData)
		if len(sheets) > 0 {
			for _, sheet := range sheets {
				workbook.Sheets[sheet.Id] = sheet
			}
		}
	}

	return
}

func (dp *SelfDataUnmarshall) tidyBaseUnmarshall(configID string, workBookData gjson.Result) (sheets []Sheet) {
	workBookData.ForEach(func(key, cellData gjson.Result) bool {
		rowIdx := 0
		sheetConfigKey := key.Str + "-" + configID
		sheetBaseConfig, ok := dp.WorkbookBaseConfig.SheetConfigMap[sheetConfigKey]
		if !ok {
			return true
		}
		sheet := Sheet{
			Id:          fmt.Sprintf("sheet-%d", sheetBaseConfig.OrderIndex),
			Name:        sheetBaseConfig.OrderName,
			CellData:    make(map[int]map[int]CellData),
			ColumnCount: len(sheetBaseConfig.Header.En),
		}
		headerCnData := make(map[int]CellData)
		headerEnData := make(map[int]CellData)
		headerTypeData := make(map[int]CellData)
		for headerIdx, cnVal := range sheetBaseConfig.Header.Cn {
			var secondRowValue, thirdRowValue string
			headerCnData[headerIdx] = CellData{V: cnVal}
			secondRowValue = sheetBaseConfig.Header.En[headerIdx]
			thirdRowValue = sheetBaseConfig.Header.Type[headerIdx]
			if property, _ := sheetBaseConfig.Header.PropertyMap[sheetBaseConfig.Header.En[headerIdx]]; property.Pk {
				secondRowValue += "_pk"
			}
			headerEnData[headerIdx] = CellData{V: secondRowValue}
			if sheetBaseConfig.Header.Rule[headerIdx] != "" {
				thirdRowValue += "_" + sheetBaseConfig.Header.Rule[headerIdx]
			}
			headerTypeData[headerIdx] = CellData{V: thirdRowValue}
		}
		sheet.CellData[rowIdx] = headerCnData
		rowIdx++
		sheet.CellData[rowIdx] = headerEnData
		rowIdx++
		sheet.CellData[rowIdx] = headerTypeData
		rowIdx++
		//  只有一行数据
		if cellData.IsObject() {
			var (
				colData map[int]CellData
				err     error
			)
			sheet.RowCount = rowIdx + 1
			colData, err = dp.UnMarshallRow(sheetBaseConfig, cellData)
			if err != nil {
				return false
			}
			sheet.CellData[rowIdx] = colData
		} else if cellData.IsArray() {
			cellData.ForEach(func(_, colValue gjson.Result) bool {
				var (
					colData map[int]CellData
					err     error
				)
				colData, err = dp.UnMarshallRow(sheetBaseConfig, colValue)
				if err != nil {
					return false
				}
				sheet.CellData[rowIdx] = colData
				rowIdx++
				return true
			})
			sheet.RowCount = rowIdx + 1
		}
		//workbook.Sheets[sheet.Id] = sheet
		sheets = append(sheets, sheet)
		return true
	})
	return
}

// UnMarshallRow
// @Description: 处理行数据
// @author liangzh
// @update 2025-12-11 17:05:24
func (dp *SelfDataUnmarshall) UnMarshallRow(sheetBaseConfig SheetBaseConfig, rowData gjson.Result) (res map[int]CellData, err error) {
	res = make(map[int]CellData)
	rowData.ForEach(func(key, value gjson.Result) bool {
		if property, ok := sheetBaseConfig.Header.PropertyMap[key.Str]; !ok {
			err = fmt.Errorf("key:%s not exist", key.Str)
			return false
		} else {
			//res[property.Index] = CellData{V: value.Value()}
			switch property.Type {
			case INT, LAYERV:
				res[property.Index] = CellData{V: int(value.Int())}
			case FLOAT:
				res[property.Index] = CellData{V: value.Float()}
			case BOOL:
				res[property.Index] = CellData{V: value.Bool()}
			default:
				res[property.Index] = CellData{V: value.String()}
			}
		}
		return true
	})
	return
}
