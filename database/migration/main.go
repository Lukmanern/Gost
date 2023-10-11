package main

import (
	"fmt"
	"log"
	"time"

	"github.com/Lukmanern/gost/database/connector"
	"github.com/Lukmanern/gost/domain/entity"
	"github.com/Lukmanern/gost/internal/env"
	"github.com/Lukmanern/gost/internal/rbac"
)

// Don't forget to Add Your new
// Table Here : must sorted .
func AllTables() []interface{} {
	return []any{
		&entity.User{},
		&entity.Role{},
		&entity.Permission{},

		// Add more tables/structs here
	}
}

// Becareful using this
// This will delete entire DB Tables,
// and recreate from beginning
func main() {
	db := connector.LoadDatabase()
	fmt.Print("\n\nStart Migration\n\n")
	defer fmt.Print("\n\nFinish Migration\n\n")

	// do in development
	// Becoreful, delete entire
	// Tables and datas of Your Database.
	env.ReadConfig("./.env")
	appInProduction := env.Configuration().GetAppInProduction()
	if !appInProduction {
		func() {
			fmt.Print("\n\nWarning : DROPING ALL DB-TABLES AND RE-CREATE in 9 seconds (CTRL+C to stop)\n\n")
			time.Sleep(9 * time.Second)
			tables := AllTables()
			deleteErr := db.Migrator().DropTable(tables...)
			if deleteErr != nil {
				log.Panicf("Error while deleting tables DB : %s", deleteErr)
			}
		}()
	}

	migrateErr := db.AutoMigrate(
		AllTables()...,
	)
	if migrateErr != nil {
		log.Panicf("Error while migration DB : %s", migrateErr)
		db.Rollback()
	}

	if !appInProduction {
		// add permission and table
		for _, data := range rbac.AllPermissions() {
			time.Sleep(100 * time.Millisecond)
			if createErr := db.Create(&data).Error; createErr != nil {
				log.Panicf("Error while create Permissions : %s", createErr)
			}
		}
		time.Sleep(500 * time.Millisecond)
		for _, data := range rbac.AllRoles() {
			time.Sleep(100 * time.Millisecond)
			if createErr := db.Create(&data).Error; createErr != nil {
				log.Panicf("Error while create Roles : %s", createErr)
			}
		}
	}
}
