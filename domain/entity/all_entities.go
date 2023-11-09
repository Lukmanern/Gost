package entity

// ⚠️ Don't forget to Add Your new
// Table Here : must sorted by developer.
func AllTables() []interface{} {
	return []any{
		&User{},
		&UserHasRoles{},
		&Role{},
		&RoleHasPermission{},
		&Permission{},

		// ...
		// Add more tables/structs
	}
}
