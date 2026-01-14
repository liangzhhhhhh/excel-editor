// Package common
// @description
// @author      梁志豪
// @datetime    2025/12/29 11:49
package common

import (
	"os"
	"path/filepath"
)

// 初始化目录信息
func init() {
	homeDir, _ := os.UserHomeDir()
	tempDir := os.TempDir()
	ExportDataDir = filepath.Join(homeDir, "Downloads")
	TempDataDir = filepath.Join(tempDir, "excel_editor")
}
