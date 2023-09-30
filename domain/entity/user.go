package entity

import "github.com/Lukmanern/gost/domain/base"

type User struct {
	ID       int    `gorm:"type:bigint(20) unsigned not null;autoIncrement;primaryKey" json:"id"`
	Name     string `gorm:"type:varchar(255) not null" json:"name"`
	Email    string `gorm:"type:varchar(100) not null unique" json:"email"`
	Password string `gorm:"type:varchar(255) not null" json:"password"`

	base.TimeAttributes
}

func (u *User) TableName() string {
	return "users"
}

// User Has Many Roles === user <-many to many-> role
// Role/s Has Many Permissions === role <-many to many-> permission

type UserHasRoles struct {
	UserID int  `json:"user_id"`
	User   User `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"user"`
	RoleID int  `json:"role_id"`
	Role   Role `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"role"`
}

func (r *UserHasRoles) TableName() string {
	return "user_has_roles"
}
