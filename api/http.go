package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"time"
)

// const BaseURL = "http://192.168.1.6:3001"
const BaseURL = "http://127.0.0.1:3000"
const AuthURL = "http://wechat.aaagame.com"
const OaURL = "https://fatcat-admin-test.54030.com"

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

func (c *APIClient) JSONRequest(
	ctx context.Context,
	method, path string,
	header map[string]string,
	requestBody interface{},
	queryBody map[string]interface{},
	result interface{},
) error {

	// 如果外部没传 ctx，就给一个默认超时
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(
			context.Background(),
			c.DefaultTimeout,
		)
		defer cancel()
	}

	var reqBody io.Reader
	if requestBody != nil {
		jsonData, err := json.Marshal(requestBody)
		if err != nil {
			return err
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	uri := c.BaseURL + path
	if len(queryBody) > 0 {
		q := url.Values{}
		for k, v := range queryBody {
			q.Set(k, fmt.Sprintf("%v", v))
		}
		uri += "?" + q.Encode()
	}

	// ⚠️ 用 NewRequestWithContext
	req, err := http.NewRequestWithContext(ctx, method, uri, reqBody)
	if err != nil {
		return err
	}

	for k, v := range header {
		req.Header.Set(k, v)
	}
	if _, ok := header["Content-Type"]; !ok {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		// 👇 关键：区分超时 / 主动取消
		if errors.Is(err, context.DeadlineExceeded) {
			return fmt.Errorf("request timeout")
		}
		if errors.Is(err, context.Canceled) {
			return fmt.Errorf("request canceled")
		}
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, body)
	}

	if result != nil {
		return json.NewDecoder(resp.Body).Decode(result)
	}

	return nil
}

func (c *APIClient) FormDataRequest(
	ctx context.Context,
	method, path string,
	header map[string]string,
	formBody map[string]interface{},
	queryBody map[string]interface{},
	result interface{},
) error {

	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), c.DefaultTimeout)
		defer cancel()
	}

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

	uri := c.BaseURL + path
	if len(queryBody) > 0 {
		q := url.Values{}
		for k, v := range queryBody {
			q.Set(k, fmt.Sprintf("%v", v))
		}
		uri += "?" + q.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, method, uri, &buf)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	for k, v := range header {
		req.Header.Set(k, v)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
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
