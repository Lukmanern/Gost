package main

import (
	"fmt"
	"log"

	"github.com/Lukmanern/gost/database/connection"
	"github.com/Lukmanern/gost/domain/entity"
)

// Becareful using this
// This will delete entire DB Tables,
// and recreate from beginning
func main() {
	db := connection.LoadDatabase()
	fmt.Print("\n\nStart Migration\n\n")
	defer fmt.Print("\n\nFinish Migration\n\n")

	// do in development
	// Becoreful, delete entire Tables of Your Database...
	deleteErr := db.Migrator().DropTable(AllTables()...)
	if deleteErr != nil {
		log.Panicf("Error while deleting tables DB : %s", deleteErr)
	}

	migrateErr := db.AutoMigrate(
		AllTables()...,
	)
	if migrateErr != nil {
		log.Panicf("Error while migration DB : %s", migrateErr)
		db.Rollback()
	}
}

// Add Your Table Here : must sorted right.
func AllTables() []interface{} {
	return []any{
		&entity.User{},
		&entity.Role{},
		&entity.UserHasRoles{},
		&entity.Permission{},
		&entity.RoleHasPermissions{},
	}
}
