// Package dataparser
// @description
// @author      梁志豪
// @datetime    2025/12/15 17:31
package dataparser

import (
	"fmt"
	"path/filepath"
	"strconv"
)

func getFileNameWithoutExt(filePath string) string {
	base := filepath.Base(filePath)
	ext := filepath.Ext(base)
	return base[:len(base)-len(ext)]
}

// GetWorkbookName
// @Description: actId必须要是 7000~8000
// @author liangzh
// @update 2025-12-17 14:26:18
func GetWorkbookName(actId, ab string) (string, error) {
	actIdInt, err := strconv.Atoi(actId)
	if err != nil {
		return "", err
	}
	if actIdInt < 7000 || actIdInt > 8000 {
		return "", fmt.Errorf("无效的活动ID:%s", actId)
	}
	baseActName := "Activity_" + actId
	if ab != "" {
		baseActName += "_" + ab
	}
	return baseActName, nil
}
