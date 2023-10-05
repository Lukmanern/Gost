package rbac

import "github.com/Lukmanern/gost/domain/entity"

// for migration and seeder
func AllPermissions() []entity.Permission {
	permissionNames := []string{
		// user
		PermissionCreateUser, PermissionViewUser,
		PermissionUpdateUser, PermissionDeleteUser,
		// role
		PermissionCreateRole, PermissionViewRole,
		PermissionUpdateRole, PermissionDeleteRole,
		// user has roles
		PermissionCreateUserHasRole, PermissionViewUserHasRole,
		PermissionUpdateUserHasRole, PermissionDeleteUserHasRole,
		// permission
		PermissionCreatePermission, PermissionViewPermission,
		PermissionUpdatePermission, PermissionDeletePermission,
		// role has permissions
		PermissionCreateRoleHasPermissions, PermissionViewRoleHasPermissions,
		PermissionUpdateRoleHasPermissions, PermissionDeleteRoleHasPermissions,
	}

	permissions := []entity.Permission{}
	for _, name := range permissionNames {
		newPermissionEntity := entity.Permission{
			Name: name,
		}
		newPermissionEntity.SetTimes()
		permissions = append(permissions, newPermissionEntity)
	}

	return permissions
}

const (
	PermissionCreateUser = "create-user"
	PermissionViewUser   = "view-user"
	PermissionUpdateUser = "update-user"
	PermissionDeleteUser = "delete-user"

	PermissionCreateRole = "create-role"
	PermissionViewRole   = "view-role"
	PermissionUpdateRole = "update-role"
	PermissionDeleteRole = "delete-role"

	PermissionCreateUserHasRole = "create-user-has-role"
	PermissionViewUserHasRole   = "view-user-has-role"
	PermissionUpdateUserHasRole = "update-user-has-role"
	PermissionDeleteUserHasRole = "delete-user-has-role"

	PermissionCreatePermission = "create-permission"
	PermissionViewPermission   = "read-permission"
	PermissionUpdatePermission = "update-permission"
	PermissionDeletePermission = "delete-permission"

	PermissionCreateRoleHasPermissions = "create-role-has-permissions"
	PermissionViewRoleHasPermissions   = "view-role-has-permissions"
	PermissionUpdateRoleHasPermissions = "update-role-has-permissions"
	PermissionDeleteRoleHasPermissions = "delete-role-has-permissions"
)
