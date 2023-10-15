package service

import (
	"math"
	"sync"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"

	"github.com/Lukmanern/gost/domain/base"
	"github.com/Lukmanern/gost/domain/model"
	"github.com/Lukmanern/gost/internal/response"
	service "github.com/Lukmanern/gost/service/rbac"
)

type RoleController interface {
	Create(c *fiber.Ctx) error
	Connect(c *fiber.Ctx) error // Add/connect more permissions to role. Table : role_has_permissions
	Get(c *fiber.Ctx) error
	GetAll(c *fiber.Ctx) error
	Update(c *fiber.Ctx) error
	Delete(c *fiber.Ctx) error
}

type RoleControllerImpl struct {
	service service.RoleService
}

var (
	roleControllerImpl     *RoleControllerImpl
	roleControllerImplOnce sync.Once
)

func NewRoleController(service service.RoleService) RoleController {
	roleControllerImplOnce.Do(func() {
		roleControllerImpl = &RoleControllerImpl{
			service: service,
		}
	})
	return roleControllerImpl
}

func (ctr RoleControllerImpl) Create(c *fiber.Ctx) error {
	var role model.RoleCreate
	if err := c.BodyParser(&role); err != nil {
		return response.BadRequest(c, "invalid json body: "+err.Error())
	}
	// hashmap
	idCheckers := make(map[int]bool)
	for _, id := range role.PermissionsID {
		if id < 1 {
			return response.BadRequest(c, "One of the permission IDs is invalid")
		}
		if idCheckers[id] {
			return response.BadRequest(c, "Permission IDs contain the same value")
		}
		idCheckers[id] = true
	}
	validate := validator.New()
	if err := validate.Struct(&role); err != nil {
		return response.BadRequest(c, "invalid json body: "+err.Error())
	}

	ctx := c.Context()
	id, createErr := ctr.service.Create(ctx, role)
	if createErr != nil {
		fiberErr, ok := createErr.(*fiber.Error)
		if ok {
			return response.CreateResponse(c, fiberErr.Code, false, fiberErr.Message, nil)
		}
		return response.Error(c, "internal server error: "+createErr.Error())
	}
	data := map[string]any{
		"id": id,
	}
	return response.SuccessCreated(c, data)
}

func (ctr RoleControllerImpl) Connect(c *fiber.Ctx) error {
	var role model.RoleConnectToPermissions
	if err := c.BodyParser(&role); err != nil {
		return response.BadRequest(c, "invalid json body: "+err.Error())
	}
	// hashmap
	idCheckers := make(map[int]bool)
	for _, id := range role.PermissionsID {
		if id < 1 {
			return response.BadRequest(c, "One of the permission IDs is invalid")
		}
		if idCheckers[id] {
			return response.BadRequest(c, "Permission IDs contain the same value")
		}
		idCheckers[id] = true
	}
	validate := validator.New()
	if err := validate.Struct(&role); err != nil {
		return response.BadRequest(c, "invalid json body: "+err.Error())
	}

	ctx := c.Context()
	connectErr := ctr.service.ConnectPermissions(ctx, role)
	if connectErr != nil {
		fiberErr, ok := connectErr.(*fiber.Error)
		if ok {
			return response.CreateResponse(c, fiberErr.Code, false, fiberErr.Message, nil)
		}
		return response.Error(c, "internal server error: "+connectErr.Error())
	}

	return response.SuccessCreated(c, "role and permissions success connected")
}

func (ctr RoleControllerImpl) Get(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil || id <= 0 {
		return response.BadRequest(c, "invalid id")
	}

	ctx := c.Context()
	role, getErr := ctr.service.GetByID(ctx, id)
	if getErr != nil {
		fiberErr, ok := getErr.(*fiber.Error)
		if ok {
			return response.CreateResponse(c, fiberErr.Code, false, fiberErr.Message, nil)
		}
		return response.Error(c, "internal server error: "+getErr.Error())
	}

	return response.SuccessLoaded(c, role)
}

func (ctr RoleControllerImpl) GetAll(c *fiber.Ctx) error {
	request := base.RequestGetAll{
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
		return response.Error(c, "internal server error: "+getErr.Error())
	}

	data := make([]interface{}, len(roles))
	for i := range roles {
		data[i] = roles[i]
	}

	responseData := base.GetAllResponse{
		Meta: base.PageMeta{
			Total: total,
			Pages: int(math.Ceil(float64(total) / float64(request.Limit))),
			Page:  request.Page,
		},
		Data: data,
	}

	return response.SuccessLoaded(c, responseData)
}

func (ctr RoleControllerImpl) Update(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil || id <= 0 {
		return response.BadRequest(c, "invalid id")
	}
	var role model.RoleUpdate
	if err := c.BodyParser(&role); err != nil {
		return response.BadRequest(c, "invalid json body: "+err.Error())
	}
	role.ID = id
	validate := validator.New()
	if err := validate.Struct(&role); err != nil {
		return response.BadRequest(c, "invalid json body: "+err.Error())
	}

	ctx := c.Context()
	updateErr := ctr.service.Update(ctx, role)
	if updateErr != nil {
		fiberErr, ok := updateErr.(*fiber.Error)
		if ok {
			return response.CreateResponse(c, fiberErr.Code, false, fiberErr.Message, nil)
		}
		return response.Error(c, "internal server error: "+updateErr.Error())
	}

	return response.SuccessNoContent(c)
}

func (ctr RoleControllerImpl) Delete(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil || id <= 0 {
		return response.BadRequest(c, "invalid id")
	}

	ctx := c.Context()
	deleteErr := ctr.service.Delete(ctx, id)
	if deleteErr != nil {
		fiberErr, ok := deleteErr.(*fiber.Error)
		if ok {
			return response.CreateResponse(c, fiberErr.Code, false, fiberErr.Message, nil)
		}
		return response.Error(c, "internal server error: "+deleteErr.Error())
	}

	return response.SuccessNoContent(c)
}
