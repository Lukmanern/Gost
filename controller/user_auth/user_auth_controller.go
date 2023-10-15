package controller

import (
	"sync"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"

	"github.com/Lukmanern/gost/domain/model"
	"github.com/Lukmanern/gost/internal/middleware"
	"github.com/Lukmanern/gost/internal/response"
	service "github.com/Lukmanern/gost/service/user_auth"
)

type UserAuthController interface {
	Login(c *fiber.Ctx) error
	Logout(c *fiber.Ctx) error
	ForgetPassword(c *fiber.Ctx) error
	UpdatePassword(c *fiber.Ctx) error
	UpdateProfile(c *fiber.Ctx) error
	MyProfile(c *fiber.Ctx) error
}

type UserAuthControllerImpl struct {
	service service.UserAuthService
}

var (
	userAuthController     *UserAuthControllerImpl
	userAuthControllerOnce sync.Once
)

func NewUserAuthController(service service.UserAuthService) UserAuthController {
	userAuthControllerOnce.Do(func() {
		userAuthController = &UserAuthControllerImpl{
			service: service,
		}
	})

	return userAuthController
}

func (ctr UserAuthControllerImpl) Login(c *fiber.Ctx) error {
	var user model.UserLogin
	if err := c.BodyParser(&user); err != nil {
		return response.BadRequest(c, "invalid json body: "+err.Error())
	}
	validate := validator.New()
	if err := validate.Struct(&user); err != nil {
		return response.BadRequest(c, "invalid json body: "+err.Error())
	}

	ctx := c.Context()
	token, loginErr := ctr.service.Login(ctx, user)
	if loginErr != nil {
		fiberErr, ok := loginErr.(*fiber.Error)
		if ok {
			return response.CreateResponse(c, fiberErr.Code, false, fiberErr.Message, nil)
		}
		return response.Error(c, "internal server error: "+loginErr.Error())
	}

	data := map[string]any{
		"token": token,
	}
	return response.CreateResponse(c, fiber.StatusOK, true, "success login", data)
}

func (ctr UserAuthControllerImpl) Logout(c *fiber.Ctx) error {
	userClaims, ok := c.Locals("claims").(*middleware.Claims)
	if !ok || userClaims == nil {
		return response.Unauthorized(c)
	}

	logoutErr := ctr.service.Logout(c)
	if logoutErr != nil {
		return response.Error(c, "internal server error: "+logoutErr.Error())
	}

	return response.CreateResponse(c, fiber.StatusOK, true, "success logout", nil)
}

func (ctr UserAuthControllerImpl) ForgetPassword(c *fiber.Ctx) error {
	var user model.UserForgetPassword
	if err := c.BodyParser(&user); err != nil {
		return response.BadRequest(c, "invalid json body: "+err.Error())
	}
	validate := validator.New()
	if err := validate.Struct(&user); err != nil {
		return response.BadRequest(c, "invalid json body: "+err.Error())
	}

	ctx := c.Context()
	forgetErr := ctr.service.ForgetPassword(ctx, user)
	if forgetErr != nil {
		fiberErr, ok := forgetErr.(*fiber.Error)
		if ok {
			return response.CreateResponse(c, fiberErr.Code, false, fiberErr.Message, nil)
		}
		return response.Error(c, "internal server error: "+forgetErr.Error())
	}

	message := "success sending link for reset password to email, check your email inbox"
	return response.CreateResponse(c, fiber.StatusAccepted, true, message, nil)
}

func (ctr UserAuthControllerImpl) UpdatePassword(c *fiber.Ctx) error {
	userClaims, ok := c.Locals("claims").(*middleware.Claims)
	if !ok || userClaims == nil {
		return response.Unauthorized(c)
	}

	var user model.UserPasswordUpdate
	if err := c.BodyParser(&user); err != nil {
		return response.BadRequest(c, "invalid json body: "+err.Error())
	}
	user.ID = userClaims.ID

	validate := validator.New()
	if err := validate.Struct(&user); err != nil {
		return response.BadRequest(c, "invalid json body: "+err.Error())
	}
	if user.NewPassword != user.NewPasswordConfirm {
		return response.BadRequest(c, "new password confirmation is wrong")
	}
	if user.NewPassword == user.OldPassword {
		return response.BadRequest(c, "no new password, try another new password")
	}

	ctx := c.Context()
	updateErr := ctr.service.UpdatePassword(ctx, user)
	if updateErr != nil {
		fiberErr, ok := updateErr.(*fiber.Error)
		if ok {
			return response.CreateResponse(c, fiberErr.Code, false, fiberErr.Message, nil)
		}
		return response.Error(c, "internal server error: "+updateErr.Error())
	}

	return response.SuccessNoContent(c)
}

func (ctr UserAuthControllerImpl) UpdateProfile(c *fiber.Ctx) error {
	userClaims, ok := c.Locals("claims").(*middleware.Claims)
	if !ok || userClaims == nil {
		return response.Unauthorized(c)
	}

	var user model.UserProfileUpdate
	if err := c.BodyParser(&user); err != nil {
		return response.BadRequest(c, "invalid json body: "+err.Error())
	}
	user.ID = userClaims.ID
	validate := validator.New()
	if err := validate.Struct(&user); err != nil {
		return response.BadRequest(c, "invalid json body: "+err.Error())
	}

	ctx := c.Context()
	updateErr := ctr.service.UpdateProfile(ctx, user)
	if updateErr != nil {
		fiberErr, ok := updateErr.(*fiber.Error)
		if ok {
			return response.CreateResponse(c, fiberErr.Code, false, fiberErr.Message, nil)
		}
		return response.Error(c, "internal server error: "+updateErr.Error())
	}

	return response.SuccessNoContent(c)
}

func (ctr UserAuthControllerImpl) MyProfile(c *fiber.Ctx) error {
	userClaims, ok := c.Locals("claims").(*middleware.Claims)
	if !ok {
		return response.BadRequest(c, "invalid token")
	}

	ctx := c.Context()
	userProfile, getErr := ctr.service.MyProfile(ctx, userClaims.ID)
	if getErr != nil {
		fiberErr, ok := getErr.(*fiber.Error)
		if ok {
			return response.CreateResponse(c, fiberErr.Code, false, fiberErr.Message, nil)
		}
		return response.Error(c, "internal server error: "+getErr.Error())
	}
	return response.SuccessLoaded(c, userProfile)
}
