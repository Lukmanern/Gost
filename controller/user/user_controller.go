package controller

import (
	"math"
	"strings"
	"sync"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"

	"github.com/Lukmanern/gost/domain/model"
	"github.com/Lukmanern/gost/internal/consts"
	"github.com/Lukmanern/gost/internal/helper"
	"github.com/Lukmanern/gost/internal/middleware"
	"github.com/Lukmanern/gost/internal/response"
	service "github.com/Lukmanern/gost/service/user"
)

// UserController defines methods for handling
// user-related operations.
type UserController interface {
	// NO-AUTH

	// Register handles for user registration for all roles.
	// This handler will send an email for confirmation.
	Register(c *fiber.Ctx) error

	// AccountActivation handles the account activation
	// process after registration.
	AccountActivation(c *fiber.Ctx) error

	// Login handles user login.
	Login(c *fiber.Ctx) error

	// ForgetPassword handles the forget password request.
	// This handler will send an email for reset password.
	ForgetPassword(c *fiber.Ctx) error

	// ResetPassword handles the password reset process.
	ResetPassword(c *fiber.Ctx) error

	// AUTH

	// MyProfile retrieves the user's own profile information.
	MyProfile(c *fiber.Ctx) error

	// Logout handles user logout.
	// Black-listing the user token / JWT.
	Logout(c *fiber.Ctx) error

	// UpdateProfile handles updating user profile information.
	UpdateProfile(c *fiber.Ctx) error

	// UpdatePassword handles updating user password.
	UpdatePassword(c *fiber.Ctx) error

	// DeleteAccount handles the account deletion process.
	DeleteAccount(c *fiber.Ctx) error

	// AUTH and ROLE-ADMIN AREA

	// GetAll retrieves all user accounts (need admin access).
	GetAll(c *fiber.Ctx) error

	// BanAccount handles banning a user account (need admin access).
	BanAccount(c *fiber.Ctx) error
}

// UserControllerImpl is the implementation of
// UserController with a UserService dependency.
type UserControllerImpl struct {
	service service.UserService
}

var (
	userController     *UserControllerImpl
	userControllerOnce sync.Once
)

// NewUserController creates a singleton UserController
// instance with the given UserService.
func NewUserController(service service.UserService) UserController {
	userControllerOnce.Do(func() {
		userController = &UserControllerImpl{
			service: service,
		}
	})

	return userController
}

// Register handles for user registration for all roles.
// This handler will send an email for confirmation.
func (ctr *UserControllerImpl) Register(c *fiber.Ctx) error {
	var user model.UserRegister
	if err := c.BodyParser(&user); err != nil {
		return response.BadRequest(c, consts.InvalidJSONBody+err.Error())
	}
	if len(user.RoleIDs) < 1 {
		return response.BadRequest(c, "please choose one or more role")
	}
	user.Email = strings.ToLower(user.Email)

	validate := validator.New()
	if err := validate.Struct(&user); err != nil {
		return response.BadRequest(c, consts.InvalidJSONBody+err.Error())
	}

	ctx := c.Context()
	id, err := ctr.service.Register(ctx, user)
	if err != nil {
		fiberErr, ok := err.(*fiber.Error)
		if ok {
			return response.CreateResponse(c, fiberErr.Code, response.Response{
				Message: fiberErr.Message, Success: false, Data: nil,
			})
		}
		return response.Error(c, consts.ErrServer+err.Error())
	}

	message := "account successfully created. please check " + user.Email
	message += " inbox; our system has sent a verification code or link."
	data := map[string]any{
		"id": id,
	}
	return response.CreateResponse(c, fiber.StatusCreated, response.Response{
		Message: message,
		Success: true,
		Data:    data,
	})
}

// AccountActivation handles the account activation
// process after registration.
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

	return response.CreateResponse(c, fiber.StatusOK, response.Response{
		Message: "thank you for your confirmation. your account is active now, you can login.",
		Success: true,
		Data:    nil,
	})
}

// Login handles user login.
func (ctr *UserControllerImpl) Login(c *fiber.Ctx) error {
	var user model.UserLogin
	if err := c.BodyParser(&user); err != nil {
		return response.BadRequest(c, consts.InvalidJSONBody+err.Error())
	}
	user.IP = helper.RandomIPAddress() // Todo : update to c.IP()
	validate := validator.New()
	if err := validate.Struct(&user); err != nil {
		return response.BadRequest(c, consts.InvalidJSONBody+err.Error())
	}

	ctx := c.Context()
	token, err := ctr.service.Login(ctx, user)
	if err != nil {
		fiberErr, ok := err.(*fiber.Error)
		if ok {
			return response.CreateResponse(c, fiberErr.Code, response.Response{
				Message: fiberErr.Message, Success: false, Data: nil,
			})
		}
		return response.Error(c, consts.ErrServer+err.Error())
	}

	return response.CreateResponse(c, fiber.StatusOK, response.Response{
		Message: "success login",
		Success: true,
		Data: map[string]any{
			"token":        token,
			"token-length": len(token),
		},
	})
}

// ForgetPassword handles the forget password request.
// This handler will send an email for reset password.
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
	err := ctr.service.ForgetPassword(ctx, user)
	if err != nil {
		fiberErr, ok := err.(*fiber.Error)
		if ok {
			return response.CreateResponse(c, fiberErr.Code, response.Response{
				Message: fiberErr.Message, Success: false, Data: nil,
			})
		}
		return response.Error(c, consts.ErrServer+err.Error())
	}

	return response.CreateResponse(c, fiber.StatusOK, response.Response{
		Message: "success sending link for reset password to email, check your email inbox",
		Success: true,
		Data:    nil,
	})
}

// ResetPassword handles the password reset process.
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
		return response.BadRequest(c, "password confirmation isn't match")
	}

	ctx := c.Context()
	err := ctr.service.ResetPassword(ctx, user)
	if err != nil {
		fiberErr, ok := err.(*fiber.Error)
		if ok {
			return response.CreateResponse(c, fiberErr.Code, response.Response{
				Message: fiberErr.Message, Success: false, Data: nil,
			})
		}
		return response.Error(c, consts.ErrServer+err.Error())
	}

	return response.CreateResponse(c, fiber.StatusOK, response.Response{
		Message: "your password already updated, you can login with the new password",
		Success: true,
		Data:    nil,
	})
}

// MyProfile retrieves the user's own profile information.
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

// Logout handles user logout.
// Black-listing the user token / JWT.
func (ctr *UserControllerImpl) Logout(c *fiber.Ctx) error {
	userClaims, ok := c.Locals("claims").(*middleware.Claims)
	if !ok || userClaims == nil {
		return response.Unauthorized(c)
	}
	err := ctr.service.Logout(c)
	if err != nil {
		return response.Error(c, consts.ErrServer+err.Error())
	}
	return response.SuccessNoContent(c)
}

// UpdateProfile handles updating user profile information.
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
	err := ctr.service.UpdateProfile(ctx, user)
	if err != nil {
		fiberErr, ok := err.(*fiber.Error)
		if ok {
			return response.CreateResponse(c, fiberErr.Code, response.Response{
				Message: fiberErr.Message, Success: false, Data: nil,
			})
		}
		return response.Error(c, consts.ErrServer+err.Error())
	}
	return response.SuccessNoContent(c)
}

// UpdatePassword handles updating user password.
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
	err := ctr.service.UpdatePassword(ctx, user)
	if err != nil {
		fiberErr, ok := err.(*fiber.Error)
		if ok {
			return response.CreateResponse(c, fiberErr.Code, response.Response{
				Message: fiberErr.Message, Success: false, Data: nil,
			})
		}
		return response.Error(c, consts.ErrServer+err.Error())
	}
	return response.SuccessNoContent(c)
}

// DeleteAccount handles the account deletion process.
func (ctr *UserControllerImpl) DeleteAccount(c *fiber.Ctx) error {
	userClaims, ok := c.Locals("claims").(*middleware.Claims)
	if !ok || userClaims == nil {
		return response.Unauthorized(c)
	}

	var user model.UserDeleteAccount
	if err := c.BodyParser(&user); err != nil {
		return response.BadRequest(c, consts.InvalidJSONBody+err.Error())
	}
	user.ID = userClaims.ID
	validate := validator.New()
	if err := validate.Struct(&user); err != nil {
		return response.BadRequest(c, consts.InvalidJSONBody+err.Error())
	}
	if user.Password != user.PasswordConfirm {
		return response.BadRequest(c, "password confirmation isn't match")
	}

	ctx := c.Context()
	err := ctr.service.DeleteAccount(ctx, user)
	if err != nil {
		fiberErr, ok := err.(*fiber.Error)
		if ok {
			return response.CreateResponse(c, fiberErr.Code, response.Response{
				Message: fiberErr.Message, Success: false, Data: nil,
			})
		}
		return response.Error(c, consts.ErrServer+err.Error())
	}
	// invalidate / blacklist the token
	logoutErr := ctr.service.Logout(c)
	if logoutErr != nil {
		return response.Error(c, consts.ErrServer+logoutErr.Error())
	}
	return response.SuccessNoContent(c)
}

// GetAll retrieves all user accounts (need admin access).
func (ctr *UserControllerImpl) GetAll(c *fiber.Ctx) error {
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
	users, total, getErr := ctr.service.GetAll(ctx, request)
	if getErr != nil {
		return response.Error(c, consts.ErrServer+getErr.Error())
	}

	data := make([]interface{}, len(users))
	for i := range users {
		data[i] = users[i]
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

// BanAccount handles banning a user account (need admin access).
func (ctr *UserControllerImpl) BanAccount(c *fiber.Ctx) error {
	userClaims, ok := c.Locals("claims").(*middleware.Claims)
	if !ok || userClaims == nil {
		return response.Unauthorized(c)
	}

	id, err := c.ParamsInt("id")
	if err != nil || id <= 0 || userClaims.ID == id {
		return response.BadRequest(c, consts.InvalidID)
	}

	ctx := c.Context()
	err = ctr.service.SoftDelete(ctx, id)
	if err != nil {
		fiberErr, ok := err.(*fiber.Error)
		if ok {
			return response.CreateResponse(c, fiberErr.Code, response.Response{
				Message: fiberErr.Message, Success: false, Data: nil,
			})
		}
		return response.Error(c, consts.ErrServer+err.Error())
	}
	return response.SuccessNoContent(c)
}
