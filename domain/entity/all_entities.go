// ⚠️ Don't forget to Add Your new
// Table Here : must sorted by developer.
package entity

// Table interface is contract that make developer
// not forget to add TableName method for struct
type Table interface {
	TableName() string
}

// AllTables func serve/ return all structs that
// developer has been created. This func used in
// database migration scripts.
func AllTables() []Table {
	return []Table{
		&User{},
		&UserHasRoles{},
		&Role{},
		&RoleHasPermission{},
		&Permission{},

		// ...
		// Add more tables/structs
	}
}
