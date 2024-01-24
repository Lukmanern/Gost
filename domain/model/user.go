package model

import (
	"time"

	"github.com/Lukmanern/gost/domain/entity"
)

type User struct {
	ID          int        `gorm:"type:bigserial;primaryKey" json:"id"`
	Name        string     `gorm:"type:varchar(100) not null" json:"name"`
	Email       string     `gorm:"type:varchar(100) not null unique" json:"email"`
	Password    string     `gorm:"type:varchar(255) not null" json:"password"`
	ActivatedAt *time.Time `gorm:"type:timestamp null;default:null" json:"activated_at"`
}

type UserRegister struct {
	Name     string `validate:"required,min=2,max=60" json:"name"`
	Email    string `validate:"required,email,min=5,max=60" json:"email"`
	Password string `validate:"required,min=8,max=30" json:"password"`
	RoleIDs  []int  `validate:"required" json:"role_id"`
}

type UserActivation struct {
	Code  string `validate:"required,min=21,max=60" json:"code"`
	Email string `validate:"required,email,min=5,max=60" json:"email"`
}

type UserLogin struct {
	Email    string `validate:"required,email,min=5,max=60" json:"email"`
	Password string `validate:"required,min=8,max=30" json:"password"`
	IP       string `validate:"required,min=4,max=20" json:"ip"`
}

// ID          int        `gorm:"type:bigserial;primaryKey" json:"id"`
// Name        string     `gorm:"type:varchar(100) not null" json:"name"`
// Email       string     `gorm:"type:varchar(100) not null unique" json:"email"`
// Password    string     `gorm:"type:varchar(255) not null" json:"password"`
// ActivatedAt *time.Time `gorm:"type:timestamp null;default:null" json:"activated_at"`
// Roles       []Role     `gorm:"many2many:user_has_roles" json:"roles"`

type UserUpdate struct {
	ID   int    `gorm:"type:bigserial;primaryKey" json:"id"`
	Name string `gorm:"type:varchar(100) not null" json:"name"`
}

type UserUpdateRoles struct {
	ID      int   `gorm:"type:bigserial;primaryKey" json:"id"`
	RoleIDs []int `validate:"required" json:"role_id"`
}

type UserForgetPassword struct {
	Email string `validate:"required,email,min=5,max=60" json:"email"`
}

type UserResetPassword struct {
	Email              string `validate:"required,email,min=5,max=60" json:"email"`
	Code               string `validate:"required,min=21,max=60" json:"code"`
	NewPassword        string `validate:"required,min=8,max=30" json:"new_password"`
	NewPasswordConfirm string `validate:"required,min=8,max=30" json:"new_password_confirm"`
}

type UserPasswordUpdate struct {
	ID                 int    `validate:"required,numeric,min=1"`
	OldPassword        string `validate:"required,min=8,max=30" json:"old_password"`
	NewPassword        string `validate:"required,min=8,max=30" json:"new_password"`
	NewPasswordConfirm string `validate:"required,min=8,max=30" json:"new_password_confirm"`
}

type UserProfile struct {
	Email       string
	Name        string
	ActivatedAt *time.Time
	Roles       []string
}

type UserResponse struct {
	ID          int
	Name        string
	ActivatedAt *time.Time
}

type UserResponseDetail struct {
	ID          int
	Email       string
	Name        string
	ActivatedAt *time.Time
	Roles       []entity.Role
}
