package model

type PageMeta struct {
	TotalData  int `json:"total_data"`
	TotalPages int `json:"total_pages"`
	AtPage     int `json:"at_page"`
}

// GetAllResponse struct used for response getAll controller funcs
type GetAllResponse struct {
	Meta PageMeta    `json:"meta"`
	Data interface{} `json:"data"`
}

// RequestGetAll struct used for request getAll controller funcs
type RequestGetAll struct {
	Page    int    `query:"page"`
	Limit   int    `query:"limit"`
	Keyword string `query:"search"`
	Sort    string `query:"sort"`
}
