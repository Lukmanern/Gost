package base

type RequestGetAll struct {
	Keyword string `json:"search" query:"search"`
	Limit   int    `json:"limit" query:"limit"`
	Page    int    `json:"page" query:"page"`
	Sort    string `json:"sort" query:"sort"`
}
