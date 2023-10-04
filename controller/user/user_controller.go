package service

import (
	"github.com/Lukmanern/gost/domain/base"
	"github.com/Lukmanern/gost/domain/model"
	service "github.com/Lukmanern/gost/service/user"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

type UserController interface {
	Create(c *fiber.Ctx) error
	Get(c *fiber.Ctx) error
	GetAll(c *fiber.Ctx) error
	Update(c *fiber.Ctx) error
	Delete(c *fiber.Ctx) error
}

type UserControllerImpl struct {
	service service.UserService
}

func (ctr UserControllerImpl) Create(c *fiber.Ctx) error {
	var user model.UserCreate
	if err := c.BodyParser(&user); err != nil {
		return base.ResponseBadRequest(c, "invalid json body: "+err.Error())
	}
	validate := validator.New()
	if err := validate.Struct(user); err != nil {
		return base.ResponseBadRequest(c, "invalid json body: "+err.Error())
	}

	ctx := c.Context()
	id, createErr := ctr.service.Create(ctx, user)
	if createErr != nil {
		return base.ResponseInternalServerError(c, "internal server error: "+createErr.Error())
	}

	return base.ResponseCreated(c, "success create data", id)
}

func (ctr UserControllerImpl) Get(c *fiber.Ctx) error {
	return base.ResponseLoaded(c, nil)
}

func (ctr UserControllerImpl) GetAll(c *fiber.Ctx) error {
	return base.ResponseLoaded(c, nil)
}

func (ctr UserControllerImpl) Update(c *fiber.Ctx) error {
	return base.ResponseNoContent(c, "success update data")
}

func (ctr UserControllerImpl) Delete(c *fiber.Ctx) error {
	return base.ResponseNoContent(c, "success delete data")
}
