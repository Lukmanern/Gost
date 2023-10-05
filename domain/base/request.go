package base

type RequestGetAll struct {
	Page    int    `query:"page"`
	Limit   int    `query:"limit"`
	Keyword string `query:"search"`
	Sort    string `query:"sort"`
}
