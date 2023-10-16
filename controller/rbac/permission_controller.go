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

type PermissionController interface {
	Create(c *fiber.Ctx) error
	Get(c *fiber.Ctx) error
	GetAll(c *fiber.Ctx) error
	Update(c *fiber.Ctx) error
	Delete(c *fiber.Ctx) error

	TestBitfieldRBAC(c *fiber.Ctx) error
}

type PermissionControllerImpl struct {
	service service.PermissionService
}

var (
	permissionControllerImpl     *PermissionControllerImpl
	permissionControllerImplOnce sync.Once
)

func NewPermissionController(service service.PermissionService) PermissionController {
	permissionControllerImplOnce.Do(func() {
		permissionControllerImpl = &PermissionControllerImpl{
			service: service,
		}
	})
	return permissionControllerImpl
}

func (ctr PermissionControllerImpl) Create(c *fiber.Ctx) error {
	var permission model.PermissionCreate
	if err := c.BodyParser(&permission); err != nil {
		return response.BadRequest(c, "invalid json body: "+err.Error())
	}
	validate := validator.New()
	if err := validate.Struct(&permission); err != nil {
		return response.BadRequest(c, "invalid json body: "+err.Error())
	}

	ctx := c.Context()
	id, createErr := ctr.service.Create(ctx, permission)
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

func (ctr PermissionControllerImpl) Get(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil || id <= 0 {
		return response.BadRequest(c, "invalid id")
	}

	ctx := c.Context()
	permission, getErr := ctr.service.GetByID(ctx, id)
	if getErr != nil {
		fiberErr, ok := getErr.(*fiber.Error)
		if ok {
			return response.CreateResponse(c, fiberErr.Code, false, fiberErr.Message, nil)
		}
		return response.Error(c, "internal server error: "+getErr.Error())
	}

	return response.SuccessLoaded(c, permission)
}

func (ctr PermissionControllerImpl) GetAll(c *fiber.Ctx) error {
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
	permissions, total, getErr := ctr.service.GetAll(ctx, request)
	if getErr != nil {
		return response.Error(c, "internal server error: "+getErr.Error())
	}

	data := make([]interface{}, len(permissions))
	for i := range permissions {
		data[i] = permissions[i]
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

func (ctr PermissionControllerImpl) Update(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil || id <= 0 {
		return response.BadRequest(c, "invalid id")
	}
	var permission model.PermissionUpdate
	if err := c.BodyParser(&permission); err != nil {
		return response.BadRequest(c, "invalid json body: "+err.Error())
	}
	permission.ID = id
	validate := validator.New()
	if err := validate.Struct(&permission); err != nil {
		return response.BadRequest(c, "invalid json body: "+err.Error())
	}

	ctx := c.Context()
	updateErr := ctr.service.Update(ctx, permission)
	if updateErr != nil {
		fiberErr, ok := updateErr.(*fiber.Error)
		if ok {
			return response.CreateResponse(c, fiberErr.Code, false, fiberErr.Message, nil)
		}
		return response.Error(c, "internal server error: "+updateErr.Error())
	}

	return response.SuccessNoContent(c)
}

func (ctr PermissionControllerImpl) Delete(c *fiber.Ctx) error {
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

func (ctr PermissionControllerImpl) TestBitfieldRBAC(c *fiber.Ctx) error {
	return response.SuccessLoaded(c, nil)
}
