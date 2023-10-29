package controller

import (
	"net"
	"strings"
	"sync"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"

	"github.com/Lukmanern/gost/domain/model"
	"github.com/Lukmanern/gost/internal/middleware"
	"github.com/Lukmanern/gost/internal/response"
	service "github.com/Lukmanern/gost/service/user"
)

type UserController interface {
	Register(c *fiber.Ctx) error
	AccountActivation(c *fiber.Ctx) error
	DeleteAccountActivation(c *fiber.Ctx) error
	ForgetPassword(c *fiber.Ctx) error
	ResetPassword(c *fiber.Ctx) error
	Login(c *fiber.Ctx) error
	Logout(c *fiber.Ctx) error
	UpdatePassword(c *fiber.Ctx) error
	UpdateProfile(c *fiber.Ctx) error
	MyProfile(c *fiber.Ctx) error
}

type UserControllerImpl struct {
	service service.UserService
}

var (
	userAuthController     *UserControllerImpl
	userAuthControllerOnce sync.Once
)

func NewUserController(service service.UserService) UserController {
	userAuthControllerOnce.Do(func() {
		userAuthController = &UserControllerImpl{
			service: service,
		}
	})

	return userAuthController
}

func (ctr UserControllerImpl) Register(c *fiber.Ctx) error {
	var user model.UserRegister
	if err := c.BodyParser(&user); err != nil {
		return response.BadRequest(c, "invalid json body: "+err.Error())
	}
	user.Email = strings.ToLower(user.Email)
	validate := validator.New()
	if err := validate.Struct(&user); err != nil {
		return response.BadRequest(c, "invalid json body: "+err.Error())
	}

	ctx := c.Context()
	id, regisErr := ctr.service.Register(ctx, user)
	if regisErr != nil {
		fiberErr, ok := regisErr.(*fiber.Error)
		if ok {
			return response.CreateResponse(c, fiberErr.Code, false, fiberErr.Message, nil)
		}
		return response.Error(c, "internal server error: "+regisErr.Error())
	}

	message := "Account success created. please check " + user.Email +
		" inbox, our system has sended verification code or link."
	data := map[string]any{
		"id": id,
	}
	return response.CreateResponse(c, fiber.StatusCreated, true, message, data)
}

func (ctr UserControllerImpl) AccountActivation(c *fiber.Ctx) error {
	var user model.UserVerificationCode
	if err := c.BodyParser(&user); err != nil {
		return response.BadRequest(c, "invalid json body: "+err.Error())
	}
	validate := validator.New()
	if err := validate.Struct(&user); err != nil {
		return response.BadRequest(c, "invalid json body: "+err.Error())
	}
	ctx := c.Context()
	err := ctr.service.Verification(ctx, user.Code)
	if err != nil {
		fiberErr, ok := err.(*fiber.Error)
		if ok {
			return response.CreateResponse(c, fiberErr.Code, false, fiberErr.Message, nil)
		}
		return response.Error(c, "internal server error: "+err.Error())
	}

	return response.CreateResponse(c, fiber.StatusOK, true,
		"Thank you for your confirmation. Your account is active now.", nil)
}

func (ctr UserControllerImpl) DeleteAccountActivation(c *fiber.Ctx) error {
	var user model.UserVerificationCode
	if err := c.BodyParser(&user); err != nil {
		return response.BadRequest(c, "invalid json body: "+err.Error())
	}
	validate := validator.New()
	if err := validate.Struct(&user); err != nil {
		return response.BadRequest(c, "invalid json body: "+err.Error())
	}
	ctx := c.Context()
	err := ctr.service.DeleteUserByVerification(ctx, user.Code)
	if err != nil {
		fiberErr, ok := err.(*fiber.Error)
		if ok {
			return response.CreateResponse(c, fiberErr.Code, false, fiberErr.Message, nil)
		}
		return response.Error(c, "internal server error: "+err.Error())
	}

	message := "Your data is already deleted, thank you for your confirmation."
	return response.CreateResponse(c, fiber.StatusOK, true, message, nil)
}

func (ctr UserControllerImpl) Login(c *fiber.Ctx) error {
	var user model.UserLogin
	// user.IP = c.IP() // Todo : uncomment this line in production
	if err := c.BodyParser(&user); err != nil {
		return response.BadRequest(c, "invalid json body: "+err.Error())
	}

	userIP := net.ParseIP(user.IP)
	if userIP == nil {
		return response.BadRequest(c, "invalid json body: invalid user ip address")
	}
	counter, _ := ctr.service.FailedLoginCounter(userIP.String(), false)
	ipBlockMsg := "Your IP has been blocked by system. Please try again in 1 or 2 Hour"
	if counter >= 5 {
		return response.CreateResponse(c, fiber.StatusBadRequest, false, ipBlockMsg, nil)
	}

	validate := validator.New()
	if err := validate.Struct(&user); err != nil {
		return response.BadRequest(c, "invalid json body: "+err.Error())
	}

	ctx := c.Context()
	token, loginErr := ctr.service.Login(ctx, user)
	if loginErr != nil {
		counter, _ := ctr.service.FailedLoginCounter(userIP.String(), true)
		if counter >= 5 {
			return response.CreateResponse(c, fiber.StatusBadRequest, false, ipBlockMsg, nil)
		}
		fiberErr, ok := loginErr.(*fiber.Error)
		if ok {
			return response.CreateResponse(c, fiberErr.Code, false, fiberErr.Message, nil)
		}
		return response.Error(c, "internal server error: "+loginErr.Error())
	}

	data := map[string]any{
		"token":        token,
		"token-length": len(token),
	}
	return response.CreateResponse(c, fiber.StatusOK, true, "success login", data)
}

func (ctr UserControllerImpl) Logout(c *fiber.Ctx) error {
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

func (ctr UserControllerImpl) ForgetPassword(c *fiber.Ctx) error {
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

func (ctr UserControllerImpl) ResetPassword(c *fiber.Ctx) error {
	var user model.UserResetPassword
	if err := c.BodyParser(&user); err != nil {
		return response.BadRequest(c, "invalid json body: "+err.Error())
	}
	validate := validator.New()
	if err := validate.Struct(&user); err != nil {
		return response.BadRequest(c, "invalid json body: "+err.Error())
	}
	if user.NewPassword != user.NewPasswordConfirm {
		return response.BadRequest(c, "password confirmation not match")
	}

	ctx := c.Context()
	resetErr := ctr.service.ResetPassword(ctx, user)
	if resetErr != nil {
		fiberErr, ok := resetErr.(*fiber.Error)
		if ok {
			return response.CreateResponse(c, fiberErr.Code, false, fiberErr.Message, nil)
		}
		return response.Error(c, "internal server error: "+resetErr.Error())
	}

	message := "your password already updated, you can login with your new password, thank you"
	return response.CreateResponse(c, fiber.StatusAccepted, true, message, nil)
}

func (ctr UserControllerImpl) UpdatePassword(c *fiber.Ctx) error {
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

func (ctr UserControllerImpl) UpdateProfile(c *fiber.Ctx) error {
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

func (ctr UserControllerImpl) MyProfile(c *fiber.Ctx) error {
	userClaims, ok := c.Locals("claims").(*middleware.Claims)
	if !ok || userClaims == nil {
		return response.Unauthorized(c)
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
