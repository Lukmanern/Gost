package model

type RoleCreate struct {
	Name string `validate:"required,string,min=1,max=60" json:"name"`
}

type RoleResponse struct {
	ID   int    `validate:"required,int,min=1"`
	Name string `validate:"required,string,min=1,max=60" json:"name"`
}

type RoleUpdateDelete struct {
	ID   int    `validate:"required,int,min=1"`
	Name string `validate:"required,string,min=1,max=60" json:"name"`
}
