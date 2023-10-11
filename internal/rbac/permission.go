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

		// Just for test
		PermissionCreateOne, PermissionViewOne,
		PermissionUpdateOne, PermissionDeleteOne,
		// Just for test
		PermissionCreateTwo, PermissionViewTwo,
		PermissionUpdateTwo, PermissionDeleteTwo,
		// Just for test
		PermissionCreateThree, PermissionViewThree,
		PermissionUpdateThree, PermissionDeleteThree,
		// Just for test
		PermissionCreateFour, PermissionViewFour,
		PermissionUpdateFour, PermissionDeleteFour,
		// Just for test
		PermissionCreateFive, PermissionViewFive,
		PermissionUpdateFive, PermissionDeleteFive,
		// Just for test
		PermissionCreateSix, PermissionViewSix,
		PermissionUpdateSix, PermissionDeleteSix,
		// Just for test
		PermissionCreateSeven, PermissionViewSeven,
		PermissionUpdateSeven, PermissionDeleteSeven,
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

	// Just For Test Large JWT data
	PermissionCreateOne = "create-one"
	PermissionViewOne   = "view-one"
	PermissionUpdateOne = "update-one"
	PermissionDeleteOne = "delete-one"

	PermissionCreateTwo = "create-two"
	PermissionViewTwo   = "view-two"
	PermissionUpdateTwo = "update-two"
	PermissionDeleteTwo = "delete-two"

	PermissionCreateThree = "create-three"
	PermissionViewThree   = "view-three"
	PermissionUpdateThree = "update-three"
	PermissionDeleteThree = "delete-three"

	PermissionCreateFour = "create-four"
	PermissionViewFour   = "view-four"
	PermissionUpdateFour = "update-four"
	PermissionDeleteFour = "delete-four"

	PermissionCreateFive = "create-five"
	PermissionViewFive   = "view-five"
	PermissionUpdateFive = "update-five"
	PermissionDeleteFive = "delete-five"

	PermissionCreateSix = "create-six"
	PermissionViewSix   = "view-six"
	PermissionUpdateSix = "update-six"
	PermissionDeleteSix = "delete-six"

	PermissionCreateSeven = "create-seven"
	PermissionViewSeven   = "view-seven"
	PermissionUpdateSeven = "update-seven"
	PermissionDeleteSeven = "delete-seven"
)