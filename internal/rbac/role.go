package rbac

import "github.com/Lukmanern/gost/domain/entity"

// for migration and seeder
func AllRoles() []entity.Role {
	roleNames := []string{
		"user",
		"admin",
	}

	roles := []entity.Role{}
	for _, name := range roleNames {
		newRoleEntity := entity.Role{
			Name: name,
		}
		newRoleEntity.SetTimes()
		roles = append(roles, newRoleEntity)
	}

	return roles
}
