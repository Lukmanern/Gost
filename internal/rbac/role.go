package rbac

import "github.com/Lukmanern/gost/domain/entity"

// ⚠️ Do not forget to put new role here.
// AllRoles func return all roles entities that has been created
// by developer. This func run self audit that check for name should be unique value.
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
	RoleAdmin = "admin"
	RoleUser  = "user"

	// ...
	// add more here
)
