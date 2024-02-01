package controller

import (
	"math"
	"sync"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"

	"github.com/Lukmanern/gost/domain/model"
	"github.com/Lukmanern/gost/internal/consts"
	"github.com/Lukmanern/gost/internal/middleware"
	"github.com/Lukmanern/gost/internal/response"
	service "github.com/Lukmanern/gost/service/role"
)

// RoleController defines all methods for handling
// role-related operations and logic.
// All these operations should be performed by an admin
// or other roles that defined in route-file.
type RoleController interface {
	// Create handles the creation of a new role.
	Create(c *fiber.Ctx) error

	// Get retrieves information about a specific role.
	Get(c *fiber.Ctx) error

	// GetAll retrieves information about all roles.
	GetAll(c *fiber.Ctx) error

	// Update handles updating role information.
	Update(c *fiber.Ctx) error

	// Delete handles the deletion of a role.
	Delete(c *fiber.Ctx) error
}

// RoleControllerImpl is the implementation of
// RoleController with a RoleService dependency.
type RoleControllerImpl struct {
	service  service.RoleService
	validate *validator.Validate
}

var (
	roleControllerImpl     *RoleControllerImpl
	roleControllerImplOnce sync.Once
)

// NewRoleController creates a singleton RoleController
// instance with the provided RoleService.
func NewRoleController(service service.RoleService) RoleController {
	roleControllerImplOnce.Do(func() {
		roleControllerImpl = &RoleControllerImpl{
			service:  service,
			validate: validator.New(),
		}
	})
	return roleControllerImpl
}

// Create handles the creation of a new role.
func (ctr *RoleControllerImpl) Create(c *fiber.Ctx) error {
	userClaims, ok := c.Locals("claims").(*middleware.Claims)
	if !ok || userClaims == nil {
		return response.Unauthorized(c)
	}

	var role model.RoleCreate
	if err := c.BodyParser(&role); err != nil {
		return response.BadRequest(c, consts.InvalidJSONBody+err.Error())
	}
	if err := ctr.validate.Struct(&role); err != nil {
		return response.BadRequest(c, consts.InvalidJSONBody+err.Error())
	}

	ctx := c.Context()
	id, err := ctr.service.Create(ctx, role)
	if err != nil {
		fiberErr, ok := err.(*fiber.Error)
		if ok {
			return response.CreateResponse(c, fiberErr.Code, response.Response{
				Message: fiberErr.Message, Success: false, Data: nil,
			})
		}
		return response.Error(c, consts.ErrServer)
	}
	data := map[string]any{
		"id": id,
	}
	return response.SuccessCreated(c, data)
}

// Get retrieves information about a specific role.
func (ctr *RoleControllerImpl) Get(c *fiber.Ctx) error {
	userClaims, ok := c.Locals("claims").(*middleware.Claims)
	if !ok || userClaims == nil {
		return response.Unauthorized(c)
	}

	id, err := c.ParamsInt("id")
	if err != nil || id <= 0 {
		return response.BadRequest(c, consts.InvalidID)
	}

	ctx := c.Context()
	role, err := ctr.service.GetByID(ctx, id)
	if err != nil {
		fiberErr, ok := err.(*fiber.Error)
		if ok {
			return response.CreateResponse(c, fiberErr.Code, response.Response{
				Message: fiberErr.Message, Success: false, Data: nil,
			})
		}
		return response.Error(c, consts.ErrServer)
	}
	return response.SuccessLoaded(c, role)
}

// GetAll retrieves information about all roles.
func (ctr *RoleControllerImpl) GetAll(c *fiber.Ctx) error {
	userClaims, ok := c.Locals("claims").(*middleware.Claims)
	if !ok || userClaims == nil {
		return response.Unauthorized(c)
	}

	request := model.RequestGetAll{
		Page:    c.QueryInt("page", 1),
		Limit:   c.QueryInt("limit", 20),
		Keyword: c.Query("search"),
		Sort:    c.Query("sort"),
	}
	if request.Page <= 0 || request.Limit <= 0 {
		return response.BadRequest(c, "invalid page or limit value")
	}

	ctx := c.Context()
	roles, total, getErr := ctr.service.GetAll(ctx, request)
	if getErr != nil {
		return response.Error(c, consts.ErrServer+getErr.Error())
	}

	data := make([]interface{}, len(roles))
	for i := range roles {
		data[i] = roles[i]
	}
	responseData := model.GetAllResponse{
		Meta: model.PageMeta{
			TotalData:  total,
			TotalPages: int(math.Ceil(float64(total) / float64(request.Limit))),
			AtPage:     request.Page,
		},
		Data: data,
	}
	return response.SuccessLoaded(c, responseData)
}

// Update handles updating role information.
func (ctr *RoleControllerImpl) Update(c *fiber.Ctx) error {
	userClaims, ok := c.Locals("claims").(*middleware.Claims)
	if !ok || userClaims == nil {
		return response.Unauthorized(c)
	}
	id, err := c.ParamsInt("id")
	if err != nil || id <= 0 {
		return response.BadRequest(c, consts.InvalidID)
	}

	var role model.RoleUpdate
	if err := c.BodyParser(&role); err != nil {
		return response.BadRequest(c, consts.InvalidJSONBody)
	}
	role.ID = id
	if err := ctr.validate.Struct(&role); err != nil {
		return response.BadRequest(c, consts.InvalidJSONBody)
	}

	ctx := c.Context()
	err = ctr.service.Update(ctx, role)
	if err != nil {
		fiberErr, ok := err.(*fiber.Error)
		if ok {
			return response.CreateResponse(c, fiberErr.Code, response.Response{
				Message: fiberErr.Message, Success: false, Data: nil,
			})
		}
		return response.Error(c, consts.ErrServer)
	}
	return response.SuccessNoContent(c)
}

// Delete handles the deletion of a role.
func (ctr *RoleControllerImpl) Delete(c *fiber.Ctx) error {
	userClaims, ok := c.Locals("claims").(*middleware.Claims)
	if !ok || userClaims == nil {
		return response.Unauthorized(c)
	}

	id, err := c.ParamsInt("id")
	if err != nil || id <= 0 {
		return response.BadRequest(c, consts.InvalidID)
	}

	ctx := c.Context()
	err = ctr.service.Delete(ctx, id)
	if err != nil {
		fiberErr, ok := err.(*fiber.Error)
		if ok {
			return response.CreateResponse(c, fiberErr.Code, response.Response{
				Message: fiberErr.Message, Success: false, Data: nil,
			})
		}
		return response.Error(c, consts.ErrServer)
	}
	return response.SuccessNoContent(c)
}
