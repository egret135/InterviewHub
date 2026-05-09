package model

// 公共 API 响应模型，供 HTTP/RPC 共用

type Response struct {
	Data interface{} `json:"data,omitempty"`
	Meta *Meta       `json:"meta,omitempty"`
	Err  *APIError   `json:"error,omitempty"`
}

type Meta struct {
	Page  int   `json:"page"`
	Size  int   `json:"size"`
	Total int64 `json:"total"`
}

type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
