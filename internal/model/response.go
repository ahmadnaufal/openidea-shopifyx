package model

type ResponseMeta struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
	Total  int `json:"total"`
}

type DataResponse struct {
	Message string        `json:"message"`
	Data    any           `json:"data,omitempty"`
	Meta    *ResponseMeta `json:"meta,omitempty"`
}

type ErrorResponse struct {
	Message string `json:"message"`
	Code    string `json:"code"`
}
