package controller

import (
	"log"

	"github.com/Lukmanern/gost/database/connector"
	"github.com/Lukmanern/gost/internal/env"
	"github.com/Lukmanern/gost/internal/rbac"
	service "github.com/Lukmanern/gost/service/user_dev"
)

var (
	userDevService    service.UserDevService
	userDevController UserController
)

func init() {
	// controller\user_dev\user_dev_controller_test.go
	// Check env and database
	env.ReadConfig("./../../.env")
	c := env.Configuration()
	dbURI := c.GetDatabaseURI()
	privKey := c.GetPrivateKey()
	pubKey := c.GetPublicKey()
	if dbURI == "" || privKey == nil || pubKey == nil {
		log.Fatal("Database URI or keys aren't valid")
	}

	connector.LoadDatabase()
	r := connector.LoadRedisDatabase()
	r.FlushAll() // clear all key:value in redis

	// dump all permissions into hashMap
	rbac.PermissionNameHashMap = rbac.PermissionNamesHashMap()
	rbac.PermissionHashMap = rbac.PermissionsHashMap()

	userDevService = service.NewUserDevService()
	userDevController = NewUserController(userDevService)
}

// func TestDelete(t *testing.T) {
// 	req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/action/%s", actionEntity.ID), nil)
// 	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", userToken))
// 	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
// 	app := fiber.New()
// 	actionService := service.LoadActionService()
// 	actionController := NewActionController(actionService)
// 	app.Put("/action/:id", actionController.Delete)
// 	resp, err := app.Test(req, -1)
// 	if err != nil {
// 		log.Print(err)
// 		tearDown()
// 		return
// 	}
// 	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
// 	saved, _ := repo.Get(ctx, actionEntity.ID, hospitalID)
// 	assert.Empty(t, saved.ID)
// }
