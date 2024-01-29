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

type RoleController interface {
	// auth + admin
	Create(c *fiber.Ctx) error
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

func (ctr *RoleControllerImpl) Create(c *fiber.Ctx) error {
	userClaims, ok := c.Locals("claims").(*middleware.Claims)
	if !ok || userClaims == nil {
		return response.Unauthorized(c)
	}

	var role model.RoleCreate
	if err := c.BodyParser(&role); err != nil {
		return response.BadRequest(c, consts.InvalidJSONBody)
	}
	validate := validator.New()
	if err := validate.Struct(&role); err != nil {
		return response.BadRequest(c, consts.InvalidJSONBody)
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
	validate := validator.New()
	if err := validate.Struct(&role); err != nil {
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
