package rbac

import "github.com/Lukmanern/gost/domain/entity"

// for migration and seeder
func AllRoles() []entity.Role {
	roleNames := []string{
		RoleAdmin,
		RoleUser,
		// ...
		// add more here
	}

	roles := []entity.Role{}
	for _, name := range roleNames {
		newRoleEntity := entity.Role{
			Name: name,
		}
		newRoleEntity.SetCreateTimes()
		roles = append(roles, newRoleEntity)
	}

	return roles
}

const (
	RoleUser  = "user"
	RoleAdmin = "admin"
	// ...
	// add more here
)
