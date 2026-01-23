package main

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
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

	httpResp, resp := a.getActInfoResponse(actId)
	if resp.Status != NormalCode {
		return resp
	}

	err := dp.UnmarshallHeader([]byte(httpResp.Data.Header))
	if err != nil {
		return ErrorResponse(err.Error())
	}

	workbook, err := dp.UnmarshallContent(ab, []byte(httpResp.Data.Content))
	if err != nil {
		return ErrorResponse(err.Error())
	}

	md5StrNew := md5Hex(httpResp.Data.Content)
	if err := UpdateActMD5(actId, md5StrNew); err != nil {
		return ErrorResponse(err.Error())
	}

	return NormalResponse(workbook)
}

// FetchActList
// @Description: 获取活动列表
// @author liangzh
// @update 2025-12-15 17:47:54
func (a *App) FetchActList(_ string) CommonResponse {
	httpResp, err := api.GetActList()
	if err != nil {
		return ErrorResponse(err.Error())
	}

	if httpResp.Status != NormalCode {
		return ErrorResponse(httpResp.Msg)
	}

	return NormalResponse(httpResp.Data)
}

// KeepActionConfig
// @Description: 保存配置至线上
// @author liangzh
// @update 2025-12-29 11:39:03
func (a *App) KeepActionConfig(wb dataparser.Workbook, token string) CommonResponse {
	fileName := wb.Id + "_配置表.xlsx"
	dp := new(dataparser.ExcelDataMarshall)
	dp.Workbook = &wb

	if err := dp.Marshall(); err != nil {
		return ErrorResponse(err.Error())
	}

	var buf bytes.Buffer
	if err := dp.File.Write(&buf); err != nil {
		return ErrorResponse(err.Error())
	}

	fileInfo := api.UploadFile{
		Reader:   &buf,
		Filename: fileName,
	}

	httpResp, err := api.UploadConfig(fileInfo, token)
	if err != nil {
		return ErrorResponse(err.Error())
	}

	if httpResp.Result != 0 {
		if !httpResp.IsLogin {
			return GenResponse(AuthCode, "")
		}
		return ErrorResponse(httpResp.Tip)
	}

	return NormalResponse("更新成功")
}

// TempActKeep
// @Description: 临时存储活动信息
// @author liangzh
// @update 2025-12-29 13:31:22
func (a *App) TempActKeep(wb dataparser.Workbook) CommonResponse {
	baseDir := path.Join(common.TempDataDir, common.TmpExcelEditorDir)
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return ErrorResponse("create temp dir failed: " + err.Error())
	}

	// 清理相关的旧文件
	a.cleanupOldTempFiles(baseDir, wb.Id)

	// 原子写入新文件
	tempFilepath := path.Join(baseDir, wb.Id)
	data, err := json.Marshal(wb)
	if err != nil {
		return ErrorResponse("marshal failed: " + err.Error())
	}

	if err := a.writeFileAtomically(tempFilepath, data); err != nil {
		return ErrorResponse("write temp file failed: " + err.Error())
	}

	return NormalResponse(nil)
}

// GetTempAct
// @Description: 获取临时存储的活动信息
// @author liangzh
// @update 2025-12-29 13:31:22
func (a *App) GetTempAct(actId, ab string) CommonResponse {
	dir := path.Join(common.TempDataDir, common.TmpExcelEditorDir)
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return GenResponse(LocalFileNoFoundCode, "temp act dir not found")
		}
		return ErrorResponse("read temp act dir failed: " + err.Error())
	}

	// actId 是必须的，如果提供了 actId，尝试查找指定文件
	if actId != "" {
		targetPrefix, err := dataparser.GetWorkbookName(actId, ab)
		if err != nil {
			return ErrorResponse("invalid actId or ab: " + err.Error())
		}

		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			if strings.HasPrefix(entry.Name(), targetPrefix) {
				return a.readTempActFile(dir, entry.Name())
			}
		}
		// 指定文件未找到
		return GenResponse(LocalFileNoFoundCode, "no temp act file found for specified actId and ab")
	}

	// 未提供 actId，查找最近修改的文件
	latestFile := a.findLatestFile(entries)
	if latestFile == nil {
		return GenResponse(LocalFileNoFoundCode, "no temp act file found")
	}

	return a.readTempActFile(dir, latestFile.Name())
}

// readTempActFile 读取并解析临时活动文件
func (a *App) readTempActFile(dir, filename string) CommonResponse {
	filePath := path.Join(dir, filename)
	data, err := os.ReadFile(filePath)
	if err != nil {
		return GenResponse(FileReadCode, "read temp act file failed: "+err.Error())
	}

	wb := new(dataparser.Workbook)
	if err := json.Unmarshal(data, wb); err != nil {
		return ErrorResponse("unmarshal temp act file failed: " + err.Error())
	}

	return NormalResponse(wb)
}

// Login
// @Description: 登录
// @author liangzh
// @update 2026-01-20 15:00:00
func (a *App) Login(username, password string) (resp CommonResponse) {
	httpResp, err := api.Login(username, password)
	if err != nil {
		resp = ErrorResponse(err.Error())
		return
	}
	resp = NormalResponse(httpResp.Token)
	return
}

// LogError
// @Description: 记录错误日志
// @author liangzh
// @update 2026-01-20 15:00:00
func (a *App) LogError(message string, stack string) {
	log.Log.Error(
		"frontend error",
		zap.String("message", message),
		zap.String("stack", stack),
	)
}

// ConsistentCheck
// @Description: 检查活动配置是否一致
// @author liangzh
// @update 2026-01-20 15:00:00
func (a *App) ConsistentCheck(actId string) CommonResponse {
	httpResp, resp := a.getActInfoResponse(actId)
	if resp.Status != NormalCode && resp.Status != NoInitedCode {
		return resp
	}

	md5StrNew := common.DefaultMD5
	if httpResp.Status == NormalCode {
		md5StrNew = md5Hex(httpResp.Data.Content)
	}

	md5Map, err := LoadActMD5Map()
	if err != nil {
		return ErrorResponse(err.Error())
	}

	md5StrOld, ok := md5Map[actId]
	if !ok {
		md5Map[actId] = common.DefaultMD5
		md5StrOld = common.DefaultMD5
	}

	if md5StrOld != md5StrNew {
		return GenResponse(ConsistentCheckCode, "活动配置已修改，请选择是否重新加载？")
	}
	// 重新存入
	if err := SaveActMD5Map(md5Map); err != nil {
		return ErrorResponse(err.Error())
	}
	return NormalResponse(nil)
}

// ConsistentSync
// @Description: 进行一致性同步json值
func (a *App) ConsistentSync(actId string) CommonResponse {
	httpResp, resp := a.getActInfoResponse(actId)
	if resp.Status != NormalCode {
		return resp
	}
	md5Map, err := LoadActMD5Map()
	if err != nil {
		return ErrorResponse(err.Error())
	}

	md5StrNew := md5Hex(httpResp.Data.Content)
	md5Map[actId] = md5StrNew

	if err := SaveActMD5Map(md5Map); err != nil {
		return ErrorResponse(err.Error())
	}

	return NormalResponse(nil)
}

func md5Hex(s string) string {
	sum := md5.Sum([]byte(s))
	return hex.EncodeToString(sum[:])
}

type ActMD5Map map[string]string

func actMD5FilePath() string {
	return path.Join(
		common.TempDataDir,
		common.TmpExcelEditorDir,
		"act_md5.json",
	)
}

func LoadActMD5Map() (ActMD5Map, error) {
	filePath := actMD5FilePath()

	data := ActMD5Map{}

	b, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return data, nil
		}
		return nil, err
	}

	if len(b) == 0 {
		return data, nil
	}

	if err := json.Unmarshal(b, &data); err != nil {
		return nil, err
	}

	return data, nil
}

func SaveActMD5Map(data ActMD5Map) error {
	dir := path.Join(common.TempDataDir, common.TmpExcelEditorDir)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	filePath := actMD5FilePath()
	tmpPath := filePath + ".tmp"

	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(tmpPath, b, 0644); err != nil {
		return err
	}

	// Windows 上如果目标文件已存在，需要先删除才能重命名
	if _, err := os.Stat(filePath); err == nil {
		if err := os.Remove(filePath); err != nil {
			// 如果删除失败，尝试清理临时文件
			_ = os.Remove(tmpPath)
			return err
		}
	}

	return os.Rename(tmpPath, filePath)
}

func UpdateActMD5(actId, md5Str string) error {
	data, err := LoadActMD5Map()
	if err != nil {
		return err
	}

	data[actId] = md5Str

	return SaveActMD5Map(data)
}

// ========== 辅助函数 ==========

// getActInfoResponse 获取活动信息并处理响应
func (a *App) getActInfoResponse(actId string) (*api.Response[api.ActConfigRespBody], CommonResponse) {
	formId, err := strconv.Atoi(actId)
	if err != nil {
		return nil, ErrorResponse("invalid actId: " + err.Error())
	}

	httpResp, err := api.GetActInfo(int32(formId))
	if err != nil {
		return nil, ErrorResponse(err.Error())
	}

	if httpResp.Status != NormalCode {
		if httpResp.Status == NoInitedCode {
			return httpResp, NoInitedResponse()
		}
		return httpResp, ErrorResponse(httpResp.Msg)
	}

	return httpResp, NormalResponse(nil)
}

// cleanupOldTempFiles 清理旧的临时文件
func (a *App) cleanupOldTempFiles(baseDir, id string) {
	if strings.HasSuffix(id, "_A") {
		baseID := strings.TrimSuffix(id, "_A")
		_ = os.Remove(path.Join(baseDir, baseID))
	} else if strings.HasSuffix(id, "_B") {
		baseID := strings.TrimSuffix(id, "_B")
		_ = os.Remove(path.Join(baseDir, baseID))
	} else {
		// 没有后缀，删掉 A / B 分支
		_ = os.Remove(path.Join(baseDir, id+"_A"))
		_ = os.Remove(path.Join(baseDir, id+"_B"))
	}
}

// writeFileAtomically 原子写入文件
func (a *App) writeFileAtomically(filepath string, data []byte) error {
	tmpPath := filepath + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return err
	}
	return os.Rename(tmpPath, filepath)
}

// findLatestFile 查找最近修改的文件
func (a *App) findLatestFile(entries []os.DirEntry) os.DirEntry {
	var latestFile os.DirEntry
	var latestTime time.Time

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		if latestFile == nil || info.ModTime().After(latestTime) {
			latestFile = entry
			latestTime = info.ModTime()
		}
	}

	return latestFile
}
