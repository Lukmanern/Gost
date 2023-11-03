package application

import (
	"github.com/Lukmanern/gost/internal/middleware"
	"github.com/Lukmanern/gost/internal/rbac"
	"github.com/Lukmanern/gost/internal/response"
	"github.com/gofiber/fiber/v2"
)

func getMiddlewareTestingRoute(router fiber.Router) {
	jwtHandler := middleware.NewJWTHandler()
	handler := FakeHandler
	middlewareTesting := router.Group("middleware").Use(jwtHandler.IsAuthenticated)

	// role admin has these permissions
	// and role user hasn't, so they (user with role user)
	// can't see the endpoint
	middlewareTesting.Get("create-rhp",
		jwtHandler.CheckHasPermission(rbac.PermCreateRoleHasPermissions.ID), handler)
	middlewareTesting.Get("view-rhp",
		jwtHandler.CheckHasPermission(rbac.PermViewRoleHasPermissions.ID), handler)
	middlewareTesting.Get("update-rhp",
		jwtHandler.CheckHasPermission(rbac.PermUpdateRoleHasPermissions.ID), handler)
	middlewareTesting.Get("delete-rhp",
		jwtHandler.CheckHasPermission(rbac.PermDeleteRoleHasPermissions.ID), handler)

	// role user has these permissions
	// and role admin hasn't, so they (user with role admin)
	// can't see the endpoint
	// middlewareTesting.Get("create-exmpl",
	// 	jwtHandler.CheckHasPermission(rbac.PermCreateOne.ID), handler)
	// middlewareTesting.Get("view-exmpl",
	// 	jwtHandler.CheckHasPermission(rbac.PermViewOne.ID), handler)
	// middlewareTesting.Get("update-exmpl",
	// 	jwtHandler.CheckHasPermission(rbac.PermUpdateOne.ID), handler)
	// middlewareTesting.Get("delete-exmpl",
	// 	jwtHandler.CheckHasPermission(rbac.PermDeleteOne.ID), handler)
}

func FakeHandler(c *fiber.Ctx) error {
	return response.CreateResponse(c, fiber.StatusOK, true, "success-view-endpoint", nil)
}
