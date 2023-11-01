package main

import (
	"fmt"
	"log"
	"time"

	"github.com/Lukmanern/gost/database/connector"
	"github.com/Lukmanern/gost/domain/entity"
	"github.com/Lukmanern/gost/internal/env"
	"github.com/Lukmanern/gost/internal/rbac"
	"gorm.io/gorm"
)

var (
	db              *gorm.DB
	config          env.Config
	appInProduction bool
)

func setup() {
	db = connector.LoadDatabase()
	env.ReadConfig("./.env")
	config = env.Configuration()
	appInProduction = config.GetAppInProduction()
}

// Becareful using this
// This will delete entire DB Tables,
// and recreate from beginning
func main() {
	setup()
	fmt.Print("\n\nStart Migration\n\n")
	defer fmt.Print("\n\nFinish Migration\n\n")

	// delete all table
	if !appInProduction {
		dropAll()
	}
	// recreate or create new table
	migrateErr := db.AutoMigrate(
		entity.AllTables()...,
	)
	if migrateErr != nil {
		db.Rollback()
		log.Panicf("Error while migration DB : %s", migrateErr)
	}

	// seed master-RBAC data : role, permission
	if !appInProduction {
		seeding()
	}
}

func dropAll() {
	fmt.Print("WARNING : DROPING ALL DB-TABLES AND RE-CREATE in 9 seconds (CTRL+C to stop)\n\n")
	time.Sleep(10 * time.Second)
	tables := entity.AllTables()
	deleteErr := db.Migrator().DropTable(tables...)
	if deleteErr != nil {
		log.Panicf("Error while deleting tables DB : %s", deleteErr)
	}
}

func seeding() {
	// seeding permission and role
	for _, data := range rbac.AllRoles() {
		time.Sleep(100 * time.Millisecond)
		if createErr := db.Create(&data).Error; createErr != nil {
			log.Panicf("Error while create Roles : %s", createErr)
		}
	}
	time.Sleep(500 * time.Millisecond)
	for _, data := range rbac.AllPermissions() {
		time.Sleep(100 * time.Millisecond)
		if createErr := db.Create(&data).Error; createErr != nil {
			log.Panicf("Error while create Permissions : %s", createErr)
		}
		if data.ID <= 20 {
			if createErr := db.Create(&entity.RoleHasPermission{
				RoleID:       1,
				PermissionID: data.ID,
			}).Error; createErr != nil {
				log.Panicf("Error while create Roles : %s", createErr)
			}
		}
		if data.ID > 20 {
			if createErr := db.Create(&entity.RoleHasPermission{
				RoleID:       2,
				PermissionID: data.ID,
			}).Error; createErr != nil {
				log.Panicf("Error while create Roles : %s", createErr)
			}
		}
	}
}
