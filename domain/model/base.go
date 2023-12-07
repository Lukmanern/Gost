package model

// RequestGetAll struct used for request getAll controller funcs
type RequestGetAll struct {
	Page    int    `query:"page"`
	Limit   int    `query:"limit"`
	Keyword string `query:"search"`
	Sort    string `query:"sort"`
}

type PageMeta struct {
	Total int `json:"total"`
	Pages int `json:"pages"`
	Page  int `json:"page"`
}

// GetAllResponse struct used for response getAll controller funcs
type GetAllResponse struct {
	Meta PageMeta    `json:"meta"`
	Data interface{} `json:"data"`
}
