package controller

import (
	"math"
	"sync"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"

	"github.com/Lukmanern/gost/domain/model"
	"github.com/Lukmanern/gost/internal/consts"
	"github.com/Lukmanern/gost/internal/response"
	service "github.com/Lukmanern/gost/service/role"
)

type RoleController interface {
	// Create func creates a new role
	Create(c *fiber.Ctx) error

	// Get func gets a role
	Get(c *fiber.Ctx) error

	// GetAll func gets some roles
	GetAll(c *fiber.Ctx) error

	// Update func updates a role
	Update(c *fiber.Ctx) error

	// Delete func deletes a role
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

func (ctr *RoleControllerImpl) Create(c *fiber.Ctx) error {
	var role model.RoleCreate
	if err := c.BodyParser(&role); err != nil {
		return response.BadRequest(c, consts.InvalidJSONBody+err.Error())
	}
	validate := validator.New()
	if err := validate.Struct(&role); err != nil {
		return response.BadRequest(c, consts.InvalidJSONBody+err.Error())
	}

	ctx := c.Context()
	id, createErr := ctr.service.Create(ctx, role)
	if createErr != nil {
		fiberErr, ok := createErr.(*fiber.Error)
		if ok {
			return response.CreateResponse(c, fiberErr.Code, response.Response{
				Message: fiberErr.Message, Success: false, Data: nil,
			})
		}
		return response.Error(c, consts.ErrServer+createErr.Error())
	}
	data := map[string]any{
		"id": id,
	}
	return response.SuccessCreated(c, data)
}

func (ctr *RoleControllerImpl) Get(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil || id <= 0 {
		return response.BadRequest(c, consts.InvalidID)
	}

	ctx := c.Context()
	role, getErr := ctr.service.GetByID(ctx, id)
	if getErr != nil {
		fiberErr, ok := getErr.(*fiber.Error)
		if ok {
			return response.CreateResponse(c, fiberErr.Code, response.Response{
				Message: fiberErr.Message, Success: false, Data: nil,
			})
		}
		return response.Error(c, consts.ErrServer+getErr.Error())
	}
	return response.SuccessLoaded(c, role)
}

func (ctr *RoleControllerImpl) GetAll(c *fiber.Ctx) error {
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

func (ctr *RoleControllerImpl) Update(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil || id <= 0 {
		return response.BadRequest(c, consts.InvalidID)
	}
	var role model.RoleUpdate
	role.ID = id
	if err := c.BodyParser(&role); err != nil {
		return response.BadRequest(c, consts.InvalidJSONBody+err.Error())
	}
	validate := validator.New()
	if err := validate.Struct(&role); err != nil {
		return response.BadRequest(c, consts.InvalidJSONBody+err.Error())
	}

	ctx := c.Context()
	updateErr := ctr.service.Update(ctx, role)
	if updateErr != nil {
		fiberErr, ok := updateErr.(*fiber.Error)
		if ok {
			return response.CreateResponse(c, fiberErr.Code, response.Response{
				Message: fiberErr.Message, Success: false, Data: nil,
			})
		}
		return response.Error(c, consts.ErrServer+updateErr.Error())
	}
	return response.SuccessNoContent(c)
}

func (ctr *RoleControllerImpl) Delete(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil || id <= 0 {
		return response.BadRequest(c, consts.InvalidID)
	}

	ctx := c.Context()
	deleteErr := ctr.service.Delete(ctx, id)
	if deleteErr != nil {
		fiberErr, ok := deleteErr.(*fiber.Error)
		if ok {
			return response.CreateResponse(c, fiberErr.Code, response.Response{
				Message: fiberErr.Message, Success: false, Data: nil,
			})
		}
		return response.Error(c, consts.ErrServer+deleteErr.Error())
	}
	return response.SuccessNoContent(c)
}
