// Package dataparser
// @description
// @author      梁志豪
// @datetime    2025/12/11 15:27
package dataparser

// 数据解析后存储格式
type Workbook struct {
	Id         string                            `json:"id"` // 任意但保证唯一，不在页面展示，只做主键
	Name       string                            `json:"name,omitempty"`
	SheetOrder []string                          `json:"sheetOrder,omitempty"` // 排序
	Sheets     map[string]Sheet                  `json:"sheets,omitempty"`     // key: 任意但保证唯一，不在页面展示，只做主键
	Styles     map[string]map[string]interface{} `json:"styles,omitempty"`     // 样式
}

type Sheet struct {
	Name        string                   `json:"name,omitempty"`
	Id          string                   `json:"id,omitempty"`          // 可以与map的key保持一致
	RowCount    int                      `json:"rowCount,omitempty"`    // 行数
	ColumnCount int                      `json:"columnCount,omitempty"` // 列数
	CellData    map[int]map[int]CellData `json:"cellData,omitempty"`    // 第一个int：行  第二个int：列
	ColumnData  map[int]ColumnStyle      `json:"columnData,omitempty"`  // 列样式
	RowData     map[int]RowStyle         `json:"rowData,omitempty"`     // 行样式
	ZoomRatio   *float64                 `json:"zoomRatio,omitempty"`   // 缩放比例
	TabColor    *string                  `json:"tabColor,omitempty"`    // 颜色
	ScrollTop   *int                     `json:"scrollTop,omitempty"`   // 滚动顶部距离
	ScrollLeft  *int                     `json:"scrollLeft,omitempty"`  // 滚动距离左侧距离
	RowHeader   *RowHeaderStyle          `json:"rowHeader,omitempty"`   // 头样式
}

type CellData struct {
	V interface{} `json:"v,omitempty"`
	S *string     `json:"s,omitempty"` // styleID
	T *int64      `json:"t,omitempty"` // type: 1:number
}

type ColumnStyle struct {
	W int64 `json:"w,omitempty"` // 宽
}

type RowStyle struct {
	H  int64 `json:"h,omitempty"`  //高
	IA int64 `json:"ia,omitempty"` // 是否自动高度
}

type RowHeaderStyle struct {
	Width  int64 `json:"width,omitempty"`
	Hidden int64 `json:"hidden,omitempty"`
}

// 配置信息
type WorkbookBaseConfig struct {
	SheetConfigMap map[string]SheetBaseConfig // key: sheet的英文名
	SheetOrder     []string
	ConfigLevel
}

type ConfigLevel struct {
	Open      bool     `json:"Open"`
	SheetKeys []string `json:"SheetKeys"`
}

type SheetBaseConfig struct {
	OrderIndex int     `json:"SheetIndex"`
	OrderName  string  `json:"SheetName"`
	Header     *Header `json:"Header"`
}

type Header struct {
	PropertyMap map[string]*Property `json:"-"` // key: En
	Cn          []string             `json:"Cn"`
	En          []string             `json:"En"`
	Type        []string             `json:"Type"`
	Rule        []string             `json:"Rule"`
	Pk          []string             `json:"Pk"`
}

type Property struct {
	Cn    string       `json:"Cn"`
	En    string       `json:"En"`
	Type  PropertyType `json:"Type"`
	Index int          `json:"Index"`
	Rule  string       `json:"Rule"`
	Pk    bool         `json:"Pk"`
}

type PropertyType = string

const (
	INT     PropertyType = "int"
	FLOAT   PropertyType = "float"
	BOOL    PropertyType = "bool"
	STRING  PropertyType = "string"
	JSON    PropertyType = "json"
	LAYERV  PropertyType = "LayerV"
	Dollars PropertyType = "Dollars"
	Numbers PropertyType = "Numbers"
)

type ActInfo struct {
	ActId   int64  `json:"ActId"`
	ActName string `json:"ActName"`
}

const Pk = "pk"
