package entity

// Don't forget to Add Your new
// Table Here : must sorted .
func AllTables() []interface{} {
	return []any{
		&User{},
		&UserHasRoles{},
		&Role{},
		&RoleHasPermission{},
		&Permission{},

		// Add more tables/structs here
	}
}
