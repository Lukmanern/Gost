package model

import "github.com/Lukmanern/gost/domain/entity"

type RoleResponse struct {
	ID          int    `validate:"required,numeric,min=1" json:"id"`
	Name        string `validate:"required" json:"name"`
	Description string `validate:"required" json:"description"`
	entity.TimeFields
}

type RoleCreate struct {
	Name        string `validate:"required,min=5,max=60" json:"name"`
	Description string `validate:"required,max=100" json:"description"`
}

type RoleUpdate struct {
	ID          int    `validate:"required,numeric,min=1"`
	Name        string `validate:"required,min=5,max=60" json:"name"`
	Description string `validate:"required,max=100" json:"description"`
}
