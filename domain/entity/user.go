package entity

import "github.com/Lukmanern/gost/domain/base"

type User struct {
	ID       int    `gorm:"type:bigint(20) unsigned not null;autoIncrement;primaryKey" json:"id"`
	Name     string `gorm:"type:varchar(255) not null" json:"name"`
	Email    string `gorm:"type:varchar(100) not null unique" json:"email"`
	Password string `gorm:"type:varchar(255) not null" json:"password"`

	base.TimeFieds

	Roles []Role `gorm:"many2many:user_has_roles" json:"roles"`
}

func (u *User) TableName() string {
	return "users"
}

type UserHasRoles struct {
	/*automated created by gorm*/
	UserID int
	RoleID int
}

func (u *UserHasRoles) TableName() string {
	return "user_has_roles"
}
