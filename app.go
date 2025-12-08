package main

import (
	"context"
	"encoding/base64"
	"os"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx context.Context
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// OpenFile 打开文件并返回base64编码的内容
func (a *App) OpenFile() string {
	path, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Filters: []runtime.FileFilter{{DisplayName: "Excel Files (*.xlsx)", Pattern: "*.xlsx"}},
	})
	if err != nil || path == "" {
		return ""
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return base64.StdEncoding.EncodeToString(data)
}

// SaveFile 保存base64编码的数据到文件
func (a *App) SaveFile(base64Data string) string {
	path, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		DefaultFilename: "edited.xlsx",
		Filters:         []runtime.FileFilter{{DisplayName: "Excel Files (*.xlsx)", Pattern: "*.xlsx"}},
	})
	if err != nil || path == "" {
		return "error"
	}
	data, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		return "error"
	}
	err = os.WriteFile(path, data, 0644)
	if err != nil {
		return "error"
	}
	return "success"
}
