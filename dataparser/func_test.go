// Package dataparser
// @description
// @author      梁志豪
// @datetime    2025/12/11 15:24
package dataparser

import (
	"fmt"
	"testing"
)

func TestUnmarshallSelf(t *testing.T) {
	dp := new(SelfDataUnmarshall)
	headerContent := dp.ReadFile("./header.json")
	err := dp.UnmarshallHeader(headerContent)
	if err != nil {
		t.Error(err)
	}
	contentContent := dp.ReadFile("./content.json")
	workbook, err := dp.UnmarshallContent(contentContent)
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("%+v\n", workbook)
}

func TestMarshallSelf(t *testing.T) {
	dp := new(SelfDataMarshall)
	onlineExcelData := dp.ReadFile("./output.json")
	err := dp.LoadWorkbook(onlineExcelData)
	if err != nil {
		t.Error(err)
	}
	headerConfig, contentConfig, err := dp.Marshall()
	if err != nil {
		t.Error(err)
	}
	dp.WriteFile("./marshall_header.json", headerConfig)
	dp.WriteFile("./marshall_content.json", contentConfig)
}

//func TestUnMarshallExcel(t *testing.T) {
//	dp := NewExcelDataUnmarshall("Activity_7090_配置表")
//	err := dp.Parse()
//	if err != nil {
//		t.Error(err)
//	}
//	dp.WriteFile("./output_excel.json")
//}

//func TestMarshallExcel(t *testing.T) {
//	dp := new(ExcelDataMarshall)
//	data := dp.ReadFile("./output_excel.json")
//	err := dp.LoadWorkbook(data)
//	if err != nil {
//		t.Error(err)
//	}
//	err = dp.Marshall()
//	if err != nil {
//		t.Error(err)
//	}
//}
