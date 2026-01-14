// Package types
// @description
// @author      梁志豪
// @datetime    2025/12/29 11:38
package types

// 数据解析函数

type CommonResponse struct {
	Status int    `json:"status"`
	Msg    string `json:"msg"`
	Data   any    `json:"data"`
}

type Code = int

const (
	NormalCode           Code = 20000
	NoInitedCode         Code = 30000
	AuthCode             Code = 40000
	ErrorCode            Code = 50000
	LocalFileNoFoundCode Code = 50001 // 文件未发现码
	FileReadCode         Code = 50002 // 文件读取失败码
)

func ErrorResponse(msg string) CommonResponse {
	resp := CommonResponse{
		Status: ErrorCode,
		Msg:    msg,
	}
	return resp
}

func NoInitedResponse() CommonResponse {
	resp := CommonResponse{
		Status: NoInitedCode,
	}
	return resp
}

func NormalResponse(data any) CommonResponse {
	resp := CommonResponse{
		Status: NormalCode,
		Msg:    "success",
		Data:   data,
	}
	return resp
}

func GenResponse(code Code, msg string) CommonResponse {
	resp := CommonResponse{
		Status: code,
		Msg:    msg,
	}
	return resp
}
