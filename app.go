package main

import (
	"bytes"
	"context"
	"encoding/json"
	"excel-editor/api"
	"excel-editor/common"
	"excel-editor/dataparser"
	"excel-editor/log"
	. "excel-editor/types"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"go.uber.org/zap"
)

// App
// @Description: 启动后需要向server建立sse连接，主要用于接收通知消息以及获取活动列表
type App struct {
	ctx context.Context
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// ImportExcel
// @Description: 导入EXCEL
// @author liangzh
// @update 2025-12-29 11:39:32
func (a *App) ImportExcel(actId, ab string) (resp CommonResponse) {
	filepath, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "请选择Excel文件",
		Filters: []runtime.FileFilter{
			{
				DisplayName: "Excel",
				Pattern:     "*.xlsx;*.xls",
			},
		},
	})
	if err != nil {
		resp = ErrorResponse(err.Error())
		return
	}
	dp, err := dataparser.NewExcelDataUnmarshall(filepath, actId, ab)
	if err != nil {
		resp = ErrorResponse(err.Error())
	}
	err = dp.Parse()
	if err != nil {
		resp = ErrorResponse(err.Error())
		return
	}
	resp = NormalResponse(dp.Workbook)
	return
}

// ExportExcel
// @Description: 导出EXCEL
// @author liangzh
// @update 2025-12-29 11:39:25
func (a *App) ExportExcel(wb dataparser.Workbook) (resp CommonResponse) {
	dp := new(dataparser.ExcelDataMarshall)
	dp.Workbook = &wb
	absolutePath := path.Join(common.ExportDataDir, dp.Workbook.Name) + ".xlsx"
	err := dp.Marshall()
	if err != nil {
		resp = ErrorResponse(err.Error())
		return
	}
	if err := dp.File.Save(absolutePath); err != nil {
		resp = ErrorResponse(err.Error())
		return
	}

	resp = NormalResponse(absolutePath)
	return
}

// FetchActConfig
// @Description: 参数为活动ID
// @author liangzh
// @update 2025-12-15 17:47:54
func (a *App) FetchActConfig(actId, ab string) (resp CommonResponse) {
	dp := new(dataparser.SelfDataUnmarshall)
	dp.ActId = actId
	formId, _ := strconv.Atoi(actId)
	httpResp, err := api.GetActInfo(int32(formId))
	if err != nil {
		resp = ErrorResponse(err.Error())
		return
	}
	if httpResp.Status != NormalCode {
		if httpResp.Status == NoInitedCode {
			resp = NoInitedResponse()
			return
		}
		resp = ErrorResponse(httpResp.Msg)
		return
	}

	err = dp.UnmarshallHeader([]byte(httpResp.Data.Header))
	if err != nil {
		resp = ErrorResponse(err.Error())
		return
	}

	workbook, err := dp.UnmarshallContent(ab, []byte(httpResp.Data.Content))
	if err != nil {
		resp = ErrorResponse(err.Error())
		return
	}

	resp = NormalResponse(workbook)
	return
}

// FetchActList
// @Description: 获取活动列表
// @author liangzh
// @update 2025-12-15 17:47:54
func (a *App) FetchActList(_ string) (resp CommonResponse) {
	httpResp, err := api.GetActList()
	if err != nil {
		resp = ErrorResponse(err.Error())
		return
	}
	//fmt.Println(httpResp)
	if httpResp.Status != NormalCode {
		resp = ErrorResponse(httpResp.Msg)
		return
	}
	resp = NormalResponse(httpResp.Data)
	//resp = GenResponse(AuthCode, "")
	return
}

// KeepActionConfig
// @Description: 保存配置至线上
// @author liangzh
// @update 2025-12-29 11:39:03
func (a *App) KeepActionConfig(wb dataparser.Workbook, token string) (resp CommonResponse) {
	fileName := wb.Id + "_配置表.xlsx"
	dp := new(dataparser.ExcelDataMarshall)
	dp.Workbook = &wb
	err := dp.Marshall()
	if err != nil {
		resp = ErrorResponse(err.Error())
		return
	}

	var buf bytes.Buffer
	if err := dp.File.Write(&buf); err != nil {
		resp = ErrorResponse(err.Error())
		return
	}
	//fmt.Println(len(buf.Bytes()))
	fileInfo := api.UploadFile{
		Reader:   &buf,
		Filename: fileName,
	}
	httpResp, err := api.UploadConfig(fileInfo, token)
	if err != nil {
		resp = ErrorResponse(err.Error())
		return
	}

	if httpResp.Result != 0 {
		if !httpResp.IsLogin {
			resp = GenResponse(AuthCode, "")
			return
		}
		resp = ErrorResponse(httpResp.Tip)
		return
	}
	resp = NormalResponse("更新成功")
	return
}

// TempActKeep
// @Description: 临时存储活动信息
// @author liangzh
// @update 2025-12-29 13:31:22
func (a *App) TempActKeep(wb dataparser.Workbook) (resp CommonResponse) {
	baseDir := path.Join(common.TempDataDir, common.TmpExcelEditorDir)
	// 1. 确保目录存在
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return ErrorResponse("create temp dir failed: " + err.Error())
	}
	id := wb.Id
	baseID := id

	// 2. 判断是否 _A / _B 结尾
	if strings.HasSuffix(id, "_A") {
		baseID = strings.TrimSuffix(id, "_A")
		_ = os.Remove(path.Join(baseDir, baseID))
	} else if strings.HasSuffix(id, "_B") {
		baseID = strings.TrimSuffix(id, "_B")
		_ = os.Remove(path.Join(baseDir, baseID))
	} else {
		// 没有后缀，删掉 A / B 分支
		_ = os.Remove(path.Join(baseDir, id+"_A"))
		_ = os.Remove(path.Join(baseDir, id+"_B"))
	}
	// 3. 当前最终写入路径
	tempFilepath := path.Join(baseDir, id)
	data, err := json.Marshal(wb)
	if err != nil {
		return ErrorResponse("marshal failed: " + err.Error())
	}

	// 4. 原子写入
	tmp := tempFilepath + ".tmp"
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		return ErrorResponse("write temp file failed: " + err.Error())
	}

	if err := os.Rename(tmp, tempFilepath); err != nil {
		return ErrorResponse("rename temp file failed: " + err.Error())
	}

	return NormalResponse(nil)
}

// GetTempAct
// @Description: 临时存储活动信息
// @author liangzh
// @update 2025-12-29 13:31:22
func (a *App) GetTempAct(actId, ab string) (resp CommonResponse) {
	dir := path.Join(common.TempDataDir, common.TmpExcelEditorDir)
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return GenResponse(LocalFileNoFoundCode, "temp act dir not found")
		}
		return ErrorResponse("read temp act dir failed: " + err.Error())
	}
	var (
		targetPrefix string
		latestFile   os.DirEntry
		latestTime   time.Time
	)
	// 1. 构造目标前缀
	targetPrefix, _ = dataparser.GetWorkbookName(actId, ab)
	// 2. 遍历目录
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		// 2.1 优先：找指定文件
		if targetPrefix != "" && strings.HasPrefix(entry.Name(), targetPrefix) {
			data, err := os.ReadFile(path.Join(dir, entry.Name()))
			if err != nil {
				return GenResponse(FileReadCode, "read temp act file failed: "+err.Error())
			}
			wb := new(dataparser.Workbook)
			if err := json.Unmarshal(data, wb); err != nil {
				return ErrorResponse(err.Error())
			}
			return NormalResponse(wb)
		}

		// 2.2 记录最近文件
		info, err := entry.Info()
		if err != nil {
			continue
		}
		if latestFile == nil || info.ModTime().After(latestTime) {
			latestFile = entry
			latestTime = info.ModTime()
		}
	}

	// 3. 回退：最近文件
	if latestFile == nil {
		return GenResponse(LocalFileNoFoundCode, "no temp act file found")
	}

	if actId != "" && ab != "" {
		return GenResponse(LocalFileNoFoundCode, "no temp act file found")
	}

	data, err := os.ReadFile(path.Join(dir, latestFile.Name()))
	if err != nil {
		return GenResponse(FileReadCode, "read latest temp act file failed: "+err.Error())
	}

	wb := new(dataparser.Workbook)
	if err := json.Unmarshal(data, wb); err != nil {
		return ErrorResponse(err.Error())
	}
	return NormalResponse(wb)
}

func (a *App) Login(username, password string) (resp CommonResponse) {
	httpResp, err := api.Login(username, password)
	if err != nil {
		resp = ErrorResponse(err.Error())
		return
	}
	resp = NormalResponse(httpResp.Token)
	return
}

func (a *App) LogError(message string, stack string) {
	log.Log.Error(
		"frontend error",
		zap.String("message", message),
		zap.String("stack", stack),
	)
}
