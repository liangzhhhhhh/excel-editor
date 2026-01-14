package main

import (
	"embed"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// Create an instance of the app structure
	app := NewApp()

	// Create application with options
	err := wails.Run(&options.App{
		Title:  "Excel Editor",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}

//package main
//
//import (
//	"crypto/md5"
//	"encoding/hex"
//	"encoding/json"
//	"fmt"
//	"io"
//	"net/http"
//	"net/url"
//	"strings"
//	"time"
//)
//
///************** 工具函数 **************/
//
//func md5Hex(s string) string {
//	sum := md5.Sum([]byte(s))
//	return hex.EncodeToString(sum[:])
//}
//
//// 等价 PHP: substr(time(), 0, -4)
//func timePart() int64 {
//	return time.Now().Unix() / 10000
//}
//
//// Signature 等价 PHP static function Signature
//func Signature(userName, password, host string) string {
//	const SysPassAPIKey = "SysPassAPIKey" // TODO 常量
//	partTime := timePart()
//	tpl := fmt.Sprintf(
//		"%s|%s|%s|%s|%d",
//		userName,
//		password,
//		host,
//		SysPassAPIKey,
//		partTime,
//	)
//	return md5Hex(tpl)
//}
//
//// 简化版 CheckSpecialChar
//func checkSpecialChar(s string) bool {
//	illegal := []string{"'", `"`, `\`, "0x", "-", "+", "="}
//	for _, v := range illegal {
//		if strings.Contains(s, v) {
//			return true
//		}
//	}
//	return false
//}
//
///************** 主逻辑 **************/
//
//func LoginAdmin(loginAPI, adminName, adminPass, host string) error {
//	// 1. 检查非法字符（对应 PHP CheckSpecialChar）
//	if checkSpecialChar(adminName) {
//		return fmt.Errorf("illegal AdminName")
//	}
//
//	// 2. 清洗用户名（严格对齐 PHP）
//	userName := strings.TrimSpace(adminName)
//	userName = strings.NewReplacer(
//		`"`, "",
//		`'`, "",
//		`\`, "",
//		"0x", "",
//	).Replace(userName)
//	if len(userName) > 100 {
//		userName = userName[:100]
//	}
//
//	// 3. 密码 MD5
//	password := md5Hex(strings.TrimSpace(adminPass))
//
//	// 4. 生成签名
//	sign := Signature(userName, password, host)
//
//	// 5. 构造 form 表单（关键）
//	data := url.Values{}
//	data.Set("UserName", userName)
//	data.Set("Password", password)
//	data.Set("Sign", sign)
//
//	req, err := http.NewRequest(
//		http.MethodPost,
//		loginAPI,
//		strings.NewReader(data.Encode()),
//	)
//	if err != nil {
//		return err
//	}
//
//	// 必须设置 Content-Type
//	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
//	req.Header.Set("Referer", host)
//	client := &http.Client{}
//	resp, err := client.Do(req)
//	if err != nil {
//		return err
//	}
//	defer resp.Body.Close()
//
//	// 7. 完整打印响应
//	respBody, err := io.ReadAll(resp.Body)
//	if err != nil {
//		return err
//	}
//
//	fmt.Println("HTTP STATUS:", resp.Status)
//	fmt.Println("RAW RESPONSE:")
//	fmt.Println(string(respBody))
//
//	// 8. 尝试解析 JSON（失败也不影响）
//	var adminInfo map[string]interface{}
//	if err := json.Unmarshal(respBody, &adminInfo); err == nil {
//		fmt.Println("PARSED JSON:", adminInfo)
//	}
//
//	return nil
//}
//
///************** main **************/
//
//func main() {
//	loginAPI := "http://wechat.aaagame.com/auth.php" // 等价 PHP 里最后那行
//	adminName := "梁志豪"
//	adminPass := "lzh5201314."
//
//	host := "fatcat-admin-test.54030.com"
//
//	if err := LoginAdmin(loginAPI, adminName, adminPass, host); err != nil {
//		fmt.Println("LOGIN ERROR:", err)
//	}
//}
