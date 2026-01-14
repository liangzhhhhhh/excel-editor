package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
)

const BaseURL = "http://192.168.1.6:3001"
const AuthURL = "http://wechat.aaagame.com"
const OaURL = "https://fatcat-admin-test.54030.com"

type APIClient struct {
	BaseURL    string
	HTTPClient *http.Client
}

func NewAPIClient(baseURL string) *APIClient {
	return &APIClient{
		BaseURL:    baseURL,
		HTTPClient: &http.Client{},
	}
}

func (c *APIClient) JSONRequest(
	method, path string,
	header map[string]string,
	requestBody interface{},
	queryBody map[string]interface{},
	result interface{},
) error {
	var reqBody io.Reader

	// 处理 requestBody
	if requestBody != nil {
		jsonData, err := json.Marshal(requestBody)
		if err != nil {
			return err
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	// 处理 queryBody
	uri := c.BaseURL + path
	if len(queryBody) > 0 {
		q := url.Values{}
		for k, v := range queryBody {
			q.Set(k, fmt.Sprintf("%v", v))
		}
		if strings.Contains(uri, "?") {
			uri += "&" + q.Encode()
		} else {
			uri += "?" + q.Encode()
		}
	}
	// 创建请求
	req, err := http.NewRequest(method, uri, reqBody)
	if err != nil {
		return err
	}

	// 设置 header
	if header != nil {
		for k, v := range header {
			req.Header.Set(k, v)
		}
	}
	// 默认 Content-Type 为 JSON，如果没有自定义
	if _, ok := header["Content-Type"]; !ok {
		req.Header.Set("Content-Type", "application/json")
	}

	// 执行请求
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	// 检查 HTTP 状态码
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(bodyBytes))
	}
	//bodyBytes, err := io.ReadAll(resp.Body)
	//if err != nil {
	//	return err
	//}
	//fmt.Println(string(bodyBytes))
	// 解码 JSON 到 result
	if result != nil {
		return json.NewDecoder(resp.Body).Decode(result)
	}
	return nil
}

func (c *APIClient) FormDataRequest(
	method, path string,
	header map[string]string,
	formBody map[string]interface{},
	queryBody map[string]interface{},
	result interface{},
) error {

	// 创建 multipart writer
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	// 写入 form 字段
	for k, v := range formBody {
		switch val := v.(type) {
		case string:
			_ = writer.WriteField(k, val)
		case []byte:
			part, err := writer.CreateFormField(k)
			if err != nil {
				return err
			}
			_, _ = part.Write(val)
		case UploadFile:
			fileInfo := v.(UploadFile)
			part, err := writer.CreateFormFile(k, fileInfo.Filename)
			if err != nil {
				return err
			}
			_, err = io.Copy(part, fileInfo.Reader)
		default:
			_ = writer.WriteField(k, fmt.Sprintf("%v", val))
		}
	}

	// 一定要 close，boundary 才会写完
	if err := writer.Close(); err != nil {
		return err
	}

	// 构建 URI + query
	uri := c.BaseURL + path
	if len(queryBody) > 0 {
		q := url.Values{}
		for k, v := range queryBody {
			q.Set(k, fmt.Sprintf("%v", v))
		}
		if strings.Contains(uri, "?") {
			uri += "&" + q.Encode()
		} else {
			uri += "?" + q.Encode()
		}
	}

	// 创建请求
	req, err := http.NewRequest(method, uri, &buf)
	if err != nil {
		return err
	}

	// 设置 Header
	if header != nil {
		for k, v := range header {
			req.Header.Set(k, v)
		}
	}
	contentType := writer.FormDataContentType()
	// ⚠️ multipart 的 Content-Type 必须来自 writer
	req.Header.Set("Content-Type", contentType)

	// 发请求
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// 状态码检查
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(bodyBytes))
	}

	fmt.Println(string(bodyBytes))

	// 反序列化
	if result != nil {
		return json.Unmarshal(bodyBytes, result)
	}

	return nil
}
