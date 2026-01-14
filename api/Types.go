// Package api
// @description
// @author      梁志豪
// @datetime    2025/12/18 15:00
package api

import (
	"encoding/json"
	"io"
)

type Response[T any] struct {
	Status int    `json:"status"`
	Msg    string `json:"msg"`
	Data   T      `json:"data"`
}

// CommonResponse
// @Description: 通用消息返回值
type CommonResponse struct {
	Status int             `json:"status"`
	Msg    string          `json:"msg"`
	Data   json.RawMessage `json:"data"`
}

// CommonResponse
// @Description: SSE返回响应类型
type SSEBody struct {
	Type Type            `json:"type"` // 消息类型
	Data json.RawMessage `json:"data"` // 消息体
}

// 消息类型
type Type = int

const (
	Type_UpdatedActConfig = iota + 1 // 活动配置已被更新消息类型
)

type BaseActConfig struct {
	ActId   int32
	ActName string
	AB      bool
}

type UploadFile struct {
	Filename string
	Reader   io.Reader
}

type ActConfigRequest struct {
	ActId int32 `json:"ActId"`
}

type ActConfigRespBody struct {
	ActId   int32  `json:"ActId"`
	Header  string `json:"Header"`
	Content string `json:"Content"`
	IsAB    bool   `json:"IsAB"`
}

type LoginRespBody struct {
	Token      string `json:"Token"`
	ExpireTime int64  `json:"ExpireTime"`
	SessionID  string `json:"SessionID"`
	Tip        string `json:"Tip"`
	Result     int64  `json:"Result"`
}

type ErrMsg = string

const (
	ServerError ErrMsg = "服务器异常"
	ParamError  ErrMsg = "参数异常"
)
