package role

import "github.com/Lukmanern/gost/domain/entity"

const (
	RoleSuperAdmin = "super-admin"
	RoleAdmin      = "admin"
	RoleUser       = "user"
)

func AllRoles() []entity.Role {
	return []entity.Role{
		{
			Name:        RoleSuperAdmin,
			Description: "description for super-admin role",
		},
		{
			Name:        RoleAdmin,
			Description: "description for admin role",
		},
		{
			Name:        RoleUser,
			Description: "description for user role",
		},
	}
}
