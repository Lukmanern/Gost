package controller

import (
	"strings"
	"sync"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"

	"github.com/Lukmanern/gost/domain/model"
	"github.com/Lukmanern/gost/internal/consts"
	"github.com/Lukmanern/gost/internal/middleware"
	"github.com/Lukmanern/gost/internal/response"
	service "github.com/Lukmanern/gost/service/user"
)

type UserController interface {
	// Register function register user account,
	// than send verification-code to email
	Register(c *fiber.Ctx) error

	// AccountActivation function activates user account with
	// verification code that has been sended to the user's email
	AccountActivation(c *fiber.Ctx) error

	// ForgetPassword function send
	// verification code into user's email
	ForgetPassword(c *fiber.Ctx) error

	// ResetPassword func resets password by creating
	// new password by email and verification code
	ResetPassword(c *fiber.Ctx) error

	// Login func gives token and access to user
	Login(c *fiber.Ctx) error

	// Logout func stores user's token into Redis
	Logout(c *fiber.Ctx) error

	// UpdatePassword func updates user's password
	UpdatePassword(c *fiber.Ctx) error

	// UpdateProfile func updates user's profile data
	UpdateProfile(c *fiber.Ctx) error

	// MyProfile func shows user's profile data
	MyProfile(c *fiber.Ctx) error
}

type UserControllerImpl struct {
	service service.UserService
}

var (
	userController     *UserControllerImpl
	userControllerOnce sync.Once
)

func NewUserController(service service.UserService) UserController {
	userControllerOnce.Do(func() {
		userController = &UserControllerImpl{
			service: service,
		}
	})

	return userController
}

func (ctr *UserControllerImpl) Register(c *fiber.Ctx) error {
	var user model.UserRegister
	if err := c.BodyParser(&user); err != nil {
		return response.BadRequest(c, consts.InvalidJSONBody+err.Error())
	}
	user.Email = strings.ToLower(user.Email)
	validate := validator.New()
	if err := validate.Struct(&user); err != nil {
		return response.BadRequest(c, consts.InvalidJSONBody+err.Error())
	}

	ctx := c.Context()
	id, regisErr := ctr.service.Register(ctx, user)
	if regisErr != nil {
		fiberErr, ok := regisErr.(*fiber.Error)
		if ok {
			return response.CreateResponse(c, fiberErr.Code, response.Response{
				Message: fiberErr.Message, Success: false, Data: nil,
			})
		}
		return response.Error(c, consts.ErrServer+regisErr.Error())
	}

	message := "Account success created. please check " + user.Email + " "
	message += "inbox, our system has sended verification code or link."
	data := map[string]any{
		"id": id,
	}
	return response.SuccessCreated(c, data)
}

func (ctr *UserControllerImpl) AccountActivation(c *fiber.Ctx) error {
	var user model.UserActivation
	if err := c.BodyParser(&user); err != nil {
		return response.BadRequest(c, consts.InvalidJSONBody+err.Error())
	}
	validate := validator.New()
	if err := validate.Struct(&user); err != nil {
		return response.BadRequest(c, consts.InvalidJSONBody+err.Error())
	}
	ctx := c.Context()
	err := ctr.service.AccountActivation(ctx, user)
	if err != nil {
		fiberErr, ok := err.(*fiber.Error)
		if ok {
			return response.CreateResponse(c, fiberErr.Code, response.Response{
				Message: fiberErr.Message, Success: false, Data: nil,
			})
		}
		return response.Error(c, consts.ErrServer+err.Error())
	}

	message := "Thank you for your confirmation. Your account is active now, you can login."
	return response.CreateResponse(c, fiber.StatusOK, response.Response{
		Message: message,
		Success: true,
		Data:    nil,
	})
}

func (ctr *UserControllerImpl) Login(c *fiber.Ctx) error {
	var user model.UserLogin
	// user.IP = c.IP() // Note : uncomment this line in production
	if err := c.BodyParser(&user); err != nil {
		return response.BadRequest(c, consts.InvalidJSONBody+err.Error())
	}

	validate := validator.New()
	if err := validate.Struct(&user); err != nil {
		return response.BadRequest(c, consts.InvalidJSONBody+err.Error())
	}

	ctx := c.Context()
	token, loginErr := ctr.service.Login(ctx, user)
	if loginErr != nil {
		fiberErr, ok := loginErr.(*fiber.Error)
		if ok {
			return response.CreateResponse(c, fiberErr.Code, response.Response{
				Message: fiberErr.Message, Success: false, Data: nil,
			})
		}
		return response.Error(c, consts.ErrServer+loginErr.Error())
	}

	data := map[string]any{
		"token":        token,
		"token-length": len(token),
	}
	return response.CreateResponse(c, fiber.StatusOK, response.Response{
		Message: "Success Login",
		Success: true,
		Data:    data,
	})
}

func (ctr *UserControllerImpl) Logout(c *fiber.Ctx) error {
	userClaims, ok := c.Locals("claims").(*middleware.Claims)
	if !ok || userClaims == nil {
		return response.Unauthorized(c)
	}
	logoutErr := ctr.service.Logout(c)
	if logoutErr != nil {
		return response.Error(c, consts.ErrServer+logoutErr.Error())
	}
	return response.SuccessNoContent(c)
}

func (ctr *UserControllerImpl) ForgetPassword(c *fiber.Ctx) error {
	var user model.UserForgetPassword
	if err := c.BodyParser(&user); err != nil {
		return response.BadRequest(c, consts.InvalidJSONBody+err.Error())
	}
	validate := validator.New()
	if err := validate.Struct(&user); err != nil {
		return response.BadRequest(c, consts.InvalidJSONBody+err.Error())
	}

	ctx := c.Context()
	forgetErr := ctr.service.ForgetPassword(ctx, user)
	if forgetErr != nil {
		fiberErr, ok := forgetErr.(*fiber.Error)
		if ok {
			return response.CreateResponse(c, fiberErr.Code, response.Response{
				Message: fiberErr.Message, Success: false, Data: nil,
			})
		}
		return response.Error(c, consts.ErrServer+forgetErr.Error())
	}

	message := "success sending link for reset password to email, check your email inbox"
	return response.CreateResponse(c, fiber.StatusAccepted, response.Response{
		Message: message,
		Success: true,
		Data:    nil,
	})
}

func (ctr *UserControllerImpl) ResetPassword(c *fiber.Ctx) error {
	var user model.UserResetPassword
	if err := c.BodyParser(&user); err != nil {
		return response.BadRequest(c, consts.InvalidJSONBody+err.Error())
	}
	validate := validator.New()
	if err := validate.Struct(&user); err != nil {
		return response.BadRequest(c, consts.InvalidJSONBody+err.Error())
	}
	if user.NewPassword != user.NewPasswordConfirm {
		return response.BadRequest(c, "password confirmation not match")
	}

	ctx := c.Context()
	resetErr := ctr.service.ResetPassword(ctx, user)
	if resetErr != nil {
		fiberErr, ok := resetErr.(*fiber.Error)
		if ok {
			return response.CreateResponse(c, fiberErr.Code, response.Response{
				Message: fiberErr.Message, Success: false, Data: nil,
			})
		}
		return response.Error(c, consts.ErrServer+resetErr.Error())
	}

	message := "your password already updated, you can login with your new password, thank you"
	return response.CreateResponse(c, fiber.StatusAccepted, response.Response{
		Message: message,
		Success: true,
		Data:    nil,
	})
}

func (ctr *UserControllerImpl) UpdatePassword(c *fiber.Ctx) error {
	userClaims, ok := c.Locals("claims").(*middleware.Claims)
	if !ok || userClaims == nil {
		return response.Unauthorized(c)
	}

	var user model.UserPasswordUpdate
	if err := c.BodyParser(&user); err != nil {
		return response.BadRequest(c, consts.InvalidJSONBody+err.Error())
	}
	user.ID = userClaims.ID

	validate := validator.New()
	if err := validate.Struct(&user); err != nil {
		return response.BadRequest(c, consts.InvalidJSONBody+err.Error())
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
			return response.CreateResponse(c, fiberErr.Code, response.Response{
				Message: fiberErr.Message, Success: false, Data: nil,
			})
		}
		return response.Error(c, consts.ErrServer+updateErr.Error())
	}

	return response.SuccessNoContent(c)
}

func (ctr *UserControllerImpl) UpdateProfile(c *fiber.Ctx) error {
	userClaims, ok := c.Locals("claims").(*middleware.Claims)
	if !ok || userClaims == nil {
		return response.Unauthorized(c)
	}

	var user model.UserUpdate
	if err := c.BodyParser(&user); err != nil {
		return response.BadRequest(c, consts.InvalidJSONBody+err.Error())
	}
	user.ID = userClaims.ID
	validate := validator.New()
	if err := validate.Struct(&user); err != nil {
		return response.BadRequest(c, consts.InvalidJSONBody+err.Error())
	}

	ctx := c.Context()
	updateErr := ctr.service.UpdateProfile(ctx, user)
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

func (ctr *UserControllerImpl) MyProfile(c *fiber.Ctx) error {
	userClaims, ok := c.Locals("claims").(*middleware.Claims)
	if !ok || userClaims == nil {
		return response.Unauthorized(c)
	}

	ctx := c.Context()
	userProfile, getErr := ctr.service.MyProfile(ctx, userClaims.ID)
	if getErr != nil {
		fiberErr, ok := getErr.(*fiber.Error)
		if ok {
			return response.CreateResponse(c, fiberErr.Code, response.Response{
				Message: fiberErr.Message, Success: false, Data: nil,
			})
		}
		return response.Error(c, consts.ErrServer+getErr.Error())
	}
	return response.SuccessLoaded(c, userProfile)
}
