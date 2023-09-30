package entity

import "gost/domain/base"

type Role struct {
	ID   int    `gorm:"type:bigint(20) unsigned not null;autoIncrement;primaryKey" json:"id"`
	Name string `gorm:"type:varchar(255) not null" json:"name"`

	base.TimeAttributes
}

func (r *Role) TableName() string {
	return "roles"
}

// User Has Many Roles === user <-many to many-> role
// Role/s Has Many Permissions === role <-many to many-> permission

type RoleHasPermissions struct {
	RoleID       int        `json:"role_id"`
	Role         Role       `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"role"`
	PermissionID int        `json:"permission_id"`
	Permission   Permission `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"permission"`
}

func (r *RoleHasPermissions) TableName() string {
	return "role_has_permissions"
}

type Permission struct {
	ID   int    `gorm:"type:bigint(20) unsigned not null;autoIncrement;primaryKey" json:"id"`
	Name string `gorm:"type:varchar(255) not null" json:"name"`

	base.TimeAttributes
}

func (r *Permission) TableName() string {
	return "permissions"
}
