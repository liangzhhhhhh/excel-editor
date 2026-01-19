// Package api
// @description
// @author      梁志豪
// @datetime    2025/12/18 14:31
package api

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

func Auth(username, password string) (any, error) {
	var (
		endpoint = "/auth.php"
		sign     = Signature(username, password, OaURL)
	)
	c := NewAPIClient(AuthURL)
	var resp any
	err := c.JSONRequest(nil, http.MethodPost, endpoint, nil, nil, map[string]interface{}{
		"UserName": username,
		"Password": password,
		"Sign":     sign,
	}, &resp)
	return &resp, err
}

func GetActList() (*Response[[]BaseActConfig], error) {
	c := NewAPIClient(BaseURL)
	var resp Response[[]BaseActConfig]
	err := c.JSONRequest(nil, "GET", "/DebugAction", map[string]string{"Content-Type": "application/json"}, nil, map[string]interface{}{
		"Cmd":    "1019",
		"Action": "actlist",
	}, &resp)
	return &resp, err
}

func GetActInfo(actId int32) (*Response[ActConfigRespBody], error) {
	c := NewAPIClient(BaseURL)
	var resp Response[ActConfigRespBody]
	err := c.JSONRequest(nil, "POST", "/DebugAction", map[string]string{"Content-Type": "application/json"}, map[string]interface{}{
		"ActId": actId,
	}, map[string]interface{}{
		"Cmd":    "1019",
		"Action": "actconfig",
	}, &resp)
	return &resp, err
}

func Login(username, password string) (*TokenResponse, error) {
	var endpoint = "/"
	c := NewAPIClient("https://fatcat-admin-test.54030.com")
	var resp TokenResponse
	err := c.FormDataRequest(nil, http.MethodPost, endpoint, nil, map[string]interface{}{
		"UserName": username,
		"Password": md5Hex(password),
	}, map[string]interface{}{
		"CH":    "Api",
		"Opt":   "AuthToken",
		"Debug": "",
	}, &resp)
	if err != nil {
		return nil, err
	}
	if resp.Result != 0 {
		return nil, errors.New(resp.Tip)
	}
	return &resp, nil
}

func UploadConfig(fileInfo UploadFile, token string) (*UploadConfigResponse, error) {
	var endpoint = "/"
	c := NewAPIClient(OaURL)
	var resp UploadConfigResponse
	err := c.FormDataRequest(nil, http.MethodPost, endpoint, nil, map[string]interface{}{
		"Token":      token,
		"UploadFile": fileInfo,
	}, map[string]interface{}{
		"CH":    "api",
		"Opt":   "UploadValueTable",
		"Debug": "",
	}, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func Signature(userName, password, host string) string {
	const SysPassAPIKey = "SysPassAPIKey"

	// 当前时间戳（秒）
	now := time.Now().Unix()

	// 等价于 substr(time(), 0, -4)
	timePart := strconv.FormatInt(now, 10)
	if len(timePart) > 4 {
		timePart = timePart[:len(timePart)-4]
	}

	tpl := fmt.Sprintf(
		"%s|%s|%s|%s|%s",
		userName,
		password,
		host,
		SysPassAPIKey,
		timePart,
	)

	sum := md5.Sum([]byte(tpl))
	return hex.EncodeToString(sum[:])
}
