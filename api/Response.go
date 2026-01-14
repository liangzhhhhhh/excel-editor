// Package main
// @description
// @author      梁志豪
// @datetime    2025/12/25 19:32
package api

type ListActResponse struct {
	List []BaseActConfig `json:"List"`
}

type TokenResponse struct {
	Token      string `json:"Token"`
	ExpireTime int64  `json:"ExpireTime"`
	SessionID  string `json:"SessionID"`
	Tip        string `json:"Tip"`
	Result     int64  `json:"Result"`
}

type UploadConfigResponse struct {
	IsLogin   bool   `json:"IsLogin"`
	Uploaded  bool   `json:"Uploaded"`
	SessionID string `json:"SessionID"`
	Tip       string `json:"Tip"`
	Result    int64  `json:"Result"`
}
