package model

// Role
type RoleCreate struct {
	Name        string `validate:"required,min=1,max=60" json:"name"`
	Description string `validate:"required,min=1" json:"description"`
}

type RoleResponse struct {
	ID          int    `validate:"required,numeric,min=1" json:"id"`
	Name        string `validate:"required,min=1" json:"name"`
	Description string `validate:"required,min=1" json:"description"`
}

type RoleUpdate struct {
	ID          int    `validate:"required,numeric,min=1"`
	Name        string `validate:"required,min=1,max=60" json:"name"`
	Description string `validate:"required,min=1" json:"description"`
}

type RoleConnectToPermissions struct {
	RoleID        int   `validate:"required,numeric" json:"role_id"`
	PermissionsID []int `validate:"required" json:"permissions_id"`
}

// Permission
type PermissionCreate struct {
	Name        string `validate:"required,string,min=1,max=60" json:"name"`
	Description string `validate:"required,string,min=1,max=100" json:"description"`
}

type PermissionUpdate struct {
	ID          int    `validate:"required,numeric,min=1" json:"id"`
	Name        string `validate:"required,string,min=1,max=60" json:"name"`
	Description string `validate:"required,string,min=1,max=100" json:"description"`
}

type PermissionResponse struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}
