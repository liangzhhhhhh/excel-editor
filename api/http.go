package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"excel-editor/config"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"time"
)

var (
	BaseURL = config.GetConfig().BaseURL
	AuthURL = config.GetConfig().AuthURL
	OaURL   = config.GetConfig().OaURL
)

type APIClient struct {
	BaseURL        string
	HTTPClient     *http.Client
	DefaultTimeout time.Duration
}

func NewAPIClient(baseURL string) *APIClient {
	return &APIClient{
		BaseURL:        baseURL,
		HTTPClient:     &http.Client{Timeout: 0},
		DefaultTimeout: 10 * time.Second,
	}
}

func (c *APIClient) JSONRequest(ctx context.Context, method, path string, header map[string]string, requestBody interface{}, queryBody map[string]interface{}, result interface{}) error {
	ctx, cancel := c.ensureContext(ctx)
	defer cancel()

	var reqBody io.Reader
	if requestBody != nil {
		jsonData, err := json.Marshal(requestBody)
		if err != nil {
			return err
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.buildURI(path, queryBody), reqBody)
	if err != nil {
		return err
	}

	c.setHeaders(req, header, "application/json")
	return c.doRequest(req, result)
}

func (c *APIClient) FormDataRequest(ctx context.Context, method, path string, header map[string]string, formBody map[string]interface{}, queryBody map[string]interface{}, result interface{}) error {
	ctx, cancel := c.ensureContext(ctx)
	defer cancel()

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	for k, v := range formBody {
		switch val := v.(type) {
		case string:
			_ = writer.WriteField(k, val)
		case UploadFile:
			part, err := writer.CreateFormFile(k, val.Filename)
			if err != nil {
				return err
			}
			_, _ = io.Copy(part, val.Reader)
		default:
			_ = writer.WriteField(k, fmt.Sprintf("%v", val))
		}
	}
	_ = writer.Close()

	req, err := http.NewRequestWithContext(ctx, method, c.buildURI(path, queryBody), &buf)
	if err != nil {
		return err
	}

	c.setHeaders(req, header, writer.FormDataContentType())
	return c.doRequest(req, result)
}

// ensureContext 确保有有效的 context，如果没有则创建带超时的 context
// 注意：返回的 context 的 cancel 函数会在请求完成后自动调用
func (c *APIClient) ensureContext(ctx context.Context) (context.Context, context.CancelFunc) {
	if ctx == nil {
		return context.WithTimeout(context.Background(), c.DefaultTimeout)
	}
	return ctx, func() {} // 空函数，因为外部传入的 context 不需要我们取消
}

// setHeaders 设置请求头
func (c *APIClient) setHeaders(req *http.Request, header map[string]string, defaultContentType string) {
	for k, v := range header {
		req.Header.Set(k, v)
	}
	if _, ok := header["Content-Type"]; !ok && defaultContentType != "" {
		req.Header.Set("Content-Type", defaultContentType)
	}
}

// doRequest 执行 HTTP 请求并处理响应
func (c *APIClient) doRequest(req *http.Request, result interface{}) error {
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return fmt.Errorf("request timeout")
		}
		if errors.Is(err, context.Canceled) {
			return fmt.Errorf("request canceled")
		}
		return err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, body)
	}

	if result != nil {
		return json.Unmarshal(body, result)
	}
	return nil
}

// buildURI 构建完整的请求 URI
func (c *APIClient) buildURI(path string, queryBody map[string]interface{}) string {
	uri := c.BaseURL + path
	if len(queryBody) > 0 {
		q := url.Values{}
		for k, v := range queryBody {
			q.Set(k, fmt.Sprintf("%v", v))
		}
		uri += "?" + q.Encode()
	}
	return uri
}
