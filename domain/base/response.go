package base

type PageMeta struct {
	Total int `json:"total"`
	Pages int `json:"pages"`
	Page  int `json:"page"`
}

type GetAllResponse struct {
	Meta PageMeta      `json:"meta"`
	Data []interface{} `json:"data"`
}
