// Package dataparser
// @description
// @author      梁志豪
// @datetime    2025/12/16 14:17
package dataparser

import (
	"encoding/json"
	"github.com/labstack/gommon/log"
	"io/ioutil"
	"strconv"
	"strings"
)

type AcrParser struct {
	Data []ActInfo
}

func (ap *AcrParser) ReadFile(filepath string) []byte {
	fileData, err := ioutil.ReadFile(filepath)
	if err != nil {
		log.Errorf("读取文件失败:%v", err)
	}
	return fileData
}

func (ap *AcrParser) ParseData(dt []byte) (err error) {
	var actInfos []ActInfo
	err = json.Unmarshal(dt, &actInfos)
	if err != nil {
		return
	}
	ap.Data = actInfos
	return
}

func (ap *AcrParser) Find(query string) (actInfos []ActInfo) {
	query = strings.TrimSpace(query)
	if query == "" {
		return ap.Data
	}

	// 1. 尝试按 ID 查询
	if id, err := strconv.ParseInt(query, 10, 64); err == nil {
		for _, actInfo := range ap.Data {
			if actInfo.ActId == id {
				actInfos = append(actInfos, actInfo)
			}
		}
		return
	}

	// 2. 按名称查询（忽略大小写 + 包含）
	q := strings.ToLower(query)
	for _, actInfo := range ap.Data {
		if strings.Contains(strings.ToLower(actInfo.ActName), q) {
			actInfos = append(actInfos, actInfo)
		}
	}
	return
}
