package entity

import "github.com/Lukmanern/gost/domain/base"

type Role struct {
	ID          int          `gorm:"type:bigint(20) unsigned not null;autoIncrement;primaryKey" json:"id"`
	Name        string       `gorm:"type:varchar(255) not null unique" json:"name"`
	Description string       `gorm:"type:varchar(255) not null" json:"description"`
	Permissions []Permission `gorm:"many2many:role_has_permissions" json:"permissions"`
	base.TimeFields
}

func (r *Role) TableName() string {
	return "roles"
}

type RoleHasPermission struct {
	/*automated created by gorm*/
	RoleID       int
	PermissionID int
}

func (r *RoleHasPermission) TableName() string {
	return "role_has_permissions"
}

type Permission struct {
	ID          int    `gorm:"type:bigint(20) unsigned not null;autoIncrement;primaryKey" json:"id"`
	Name        string `gorm:"type:varchar(255) not null unique" json:"name"`
	Description string `gorm:"type:varchar(255) not null" json:"description"`
	base.TimeFields
}

func (r *Permission) TableName() string {
	return "permissions"
}
