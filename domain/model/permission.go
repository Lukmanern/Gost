package model

// Permission
type PermissionCreate struct {
	Name        string `validate:"required,string,min=1,max=60" json:"name"`
	Description string `validate:"required,string,min=1,max=100" json:"description"`
}

type PermissionUpdate struct {
	ID          int    `validate:"required,int,min=1" json:"id"`
	Name        string `validate:"required,string,min=1,max=60" json:"name"`
	Description string `validate:"required,string,min=1,max=100" json:"description"`
}

type PermissionResponse struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}
