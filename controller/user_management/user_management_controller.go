// don't use this for production
// use this file just for testing
// and testing management.

package controller

import (
	"math"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"

	"github.com/Lukmanern/gost/domain/base"
	"github.com/Lukmanern/gost/domain/model"
	"github.com/Lukmanern/gost/internal/response"
	service "github.com/Lukmanern/gost/service/user_management"
)

type UserManagementController interface {
	Create(c *fiber.Ctx) error
	Get(c *fiber.Ctx) error
	GetAll(c *fiber.Ctx) error
	Update(c *fiber.Ctx) error
	Delete(c *fiber.Ctx) error
}

type UserManagementControllerImpl struct {
	service service.UserManagementService
}

func NewUserManagementController(userService service.UserManagementService) UserManagementController {
	return &UserManagementControllerImpl{
		service: userService,
	}
}

func (ctr UserManagementControllerImpl) Create(c *fiber.Ctx) error {
	var user model.UserCreate
	if err := c.BodyParser(&user); err != nil {
		return response.BadRequest(c, "invalid json body: "+err.Error())
	}
	user.Email = strings.ToLower(user.Email)
	validate := validator.New()
	if err := validate.Struct(&user); err != nil {
		return response.BadRequest(c, "invalid json body: "+err.Error())
	}

	ctx := c.Context()
	id, createErr := ctr.service.Create(ctx, user)
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

func (ctr UserManagementControllerImpl) Get(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil || id <= 0 {
		return response.BadRequest(c, "invalid id")
	}

	ctx := c.Context()
	userProfile, getErr := ctr.service.GetByID(ctx, id)
	if getErr != nil {
		fiberErr, ok := getErr.(*fiber.Error)
		if ok {
			return response.CreateResponse(c, fiberErr.Code, false, fiberErr.Message, nil)
		}
		return response.Error(c, "internal server error: "+getErr.Error())
	}

	return response.SuccessLoaded(c, userProfile)
}

func (ctr UserManagementControllerImpl) GetAll(c *fiber.Ctx) error {
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
	users, total, getErr := ctr.service.GetAll(ctx, request)
	if getErr != nil {
		return response.Error(c, "internal server error: "+getErr.Error())
	}

	data := make([]interface{}, len(users))
	for i := range users {
		data[i] = users[i]
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

func (ctr UserManagementControllerImpl) Update(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil || id <= 0 {
		return response.BadRequest(c, "invalid id")
	}
	var user model.UserProfileUpdate
	user.ID = id
	if err := c.BodyParser(&user); err != nil {
		return response.BadRequest(c, "invalid json body: "+err.Error())
	}
	validate := validator.New()
	if err := validate.Struct(&user); err != nil {
		return response.BadRequest(c, "invalid json body: "+err.Error())
	}

	ctx := c.Context()
	updateErr := ctr.service.Update(ctx, user)
	if updateErr != nil {
		fiberErr, ok := updateErr.(*fiber.Error)
		if ok {
			return response.CreateResponse(c, fiberErr.Code, false, fiberErr.Message, nil)
		}
		return response.Error(c, "internal server error: "+updateErr.Error())
	}

	return response.SuccessNoContent(c)
}

func (ctr UserManagementControllerImpl) Delete(c *fiber.Ctx) error {
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
