package model

type RoleCreate struct {
	Name        string `validate:"required,min=1,max=60" json:"name"`
	Description string `validate:"required,min=1" json:"description"`
}

type RoleResponse struct {
	ID          int    `validate:"required,int,min=1" json:"id"`
	Name        string `validate:"required,min=1" json:"name"`
	Description string `validate:"required,min=1" json:"description"`
}

type RoleUpdateDelete struct {
	ID          int    `validate:"required,int,min=1"`
	Name        string `validate:"required,min=1,max=60" json:"name"`
	Description string `validate:"required,min=1" json:"description"`
}
