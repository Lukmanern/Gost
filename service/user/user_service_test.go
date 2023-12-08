package service

import (
	"log"
	"testing"
	"time"

	"github.com/Lukmanern/gost/domain/entity"
	"github.com/Lukmanern/gost/domain/model"
	"github.com/Lukmanern/gost/internal/env"
	"github.com/Lukmanern/gost/internal/errors"
	"github.com/Lukmanern/gost/internal/hash"
	"github.com/Lukmanern/gost/internal/helper"
	repository "github.com/Lukmanern/gost/repository/user"
	permService "github.com/Lukmanern/gost/service/permission"
	roleService "github.com/Lukmanern/gost/service/role"
	"github.com/stretchr/testify/assert"
)

const (
	fileTestName string = "at UserRepoTest"
)

var (
	timeNow        time.Time
	userRepository repository.UserRepository
)

// Register
// Verification
// DeleteUserByVerification
// Login
// ForgetPassword
// ResetPassword
// UpdatePassword
// UpdateProfile
// MyProfile Done

func init() {
	envFilePath := "./../../.env"
	env.ReadConfig(envFilePath)
	timeNow = time.Now()
	userRepository = repository.NewUserRepository()
}

func TestRegisterAndDelete(t *testing.T) {
	permService := permService.NewPermissionService()
	assert.NotNil(t, permService, errors.ShouldNotNil, fileTestName)
	roleService := roleService.NewRoleService(permService)
	assert.NotNil(t, roleService, errors.ShouldNotNil, fileTestName)
	service := NewUserService(roleService)
	assert.NotNil(t, service, errors.ShouldNotNil, fileTestName)
	ctx := helper.NewFiberCtx().Context()
	assert.NotNil(t, ctx, errors.ShouldNotNil, fileTestName)

	validUser := createUser(2)
	defer userRepository.Delete(ctx, validUser.ID)

	type testCase struct {
		Name    string
		Payload model.UserRegister
		WantErr bool
	}

	testCases := []testCase{
		{
			Name: "Failed Register -1: email already used",
			Payload: model.UserRegister{
				Name:     helper.ToTitle(helper.RandomString(12)),
				Email:    validUser.Email,
				Password: helper.RandomString(13),
				RoleID:   2,
			},
			WantErr: true,
		},
		{
			Name: "Failed Register -2: invalid role id",
			Payload: model.UserRegister{
				Name:     helper.ToTitle(helper.RandomString(12)),
				Email:    helper.RandomEmail(),
				Password: helper.RandomString(13),
				RoleID:   -2,
			},
			WantErr: true,
		},
		{
			Name: "Success Register -1",
			Payload: model.UserRegister{
				Name:     helper.ToTitle(helper.RandomString(12)),
				Email:    helper.RandomEmail(),
				Password: helper.RandomString(13),
				RoleID:   2,
			},
		},
		{
			Name: "Success Register -2",
			Payload: model.UserRegister{
				Name:     helper.ToTitle(helper.RandomString(12)),
				Email:    helper.RandomEmail(),
				Password: helper.RandomString(13),
				RoleID:   2,
			},
		},
	}

	for _, tc := range testCases {
		log.Println(tc.Name, fileTestName)

		id, createErr := service.Register(ctx, tc.Payload)
		if tc.WantErr {
			assert.Error(t, createErr, errors.ShouldErr, tc.Name, fileTestName)
			continue
		}
		assert.NoError(t, createErr, errors.ShouldNotErr, tc.Name, fileTestName)

		userByID, getErr := service.MyProfile(ctx, id)
		assert.NoError(t, getErr, errors.ShouldNotErr, tc.Name, fileTestName)
		assert.Equal(t, userByID.Name, tc.Payload.Name, errors.ShouldEqual, tc.Name, fileTestName)
		assert.Equal(t, userByID.Email, tc.Payload.Email, errors.ShouldEqual, tc.Name, fileTestName)

		_, getErr = service.MyProfile(ctx, id*99)
		assert.Error(t, getErr, errors.ShouldErr, tc.Name, fileTestName)

		userByID2, getErr2 := userRepository.GetByID(ctx, id)
		assert.NoError(t, getErr2, errors.ShouldNotErr, tc.Name, fileTestName)

		deleteErr := service.DeleteUserByVerification(ctx, model.UserVerificationCode{
			Email: tc.Payload.Email,
			Code:  *userByID2.VerificationCode,
		})
		assert.NoError(t, deleteErr, errors.ShouldNotErr, tc.Name, fileTestName)

		deleteErr = service.DeleteUserByVerification(ctx, model.UserVerificationCode{
			Email: tc.Payload.Email,
			Code:  *userByID2.VerificationCode,
		})
		assert.Error(t, deleteErr, errors.ShouldErr, tc.Name, fileTestName)

		_, getErr = service.MyProfile(ctx, id)
		assert.Error(t, getErr, errors.ShouldErr, tc.Name, fileTestName)
	}
}

func TestRegisterAndVerification(t *testing.T) {
	permService := permService.NewPermissionService()
	assert.NotNil(t, permService, errors.ShouldNotNil, fileTestName)
	roleService := roleService.NewRoleService(permService)
	assert.NotNil(t, roleService, errors.ShouldNotNil, fileTestName)
	service := NewUserService(roleService)
	assert.NotNil(t, service, errors.ShouldNotNil, fileTestName)
	ctx := helper.NewFiberCtx().Context()
	assert.NotNil(t, ctx, errors.ShouldNotNil, fileTestName)

	validUser := createUser(2)
	defer userRepository.Delete(ctx, validUser.ID)

	type testCase struct {
		Name    string
		Payload model.UserRegister
		WantErr bool
	}

	testCases := []testCase{
		{
			Name: "Failed Register -1: email already used",
			Payload: model.UserRegister{
				Name:     helper.ToTitle(helper.RandomString(12)),
				Email:    validUser.Email,
				Password: helper.RandomString(13),
				RoleID:   2,
			},
			WantErr: true,
		},
		{
			Name: "Success Register -1",
			Payload: model.UserRegister{
				Name:     helper.ToTitle(helper.RandomString(12)),
				Email:    helper.RandomEmail(),
				Password: helper.RandomString(13),
				RoleID:   2,
			},
		},
		{
			Name: "Success Register -2",
			Payload: model.UserRegister{
				Name:     helper.ToTitle(helper.RandomString(12)),
				Email:    helper.RandomEmail(),
				Password: helper.RandomString(13),
				RoleID:   2,
			},
		},
	}

	for _, tc := range testCases {
		log.Println(tc.Name, fileTestName)

		id, registerErr := service.Register(ctx, tc.Payload)
		if tc.WantErr {
			assert.Error(t, registerErr, errors.ShouldErr, tc.Name, fileTestName)
			continue
		}
		assert.NoError(t, registerErr, errors.ShouldNotErr, tc.Name, fileTestName)

		userByID, getErr := service.MyProfile(ctx, id)
		assert.NoError(t, getErr, errors.ShouldNotErr, tc.Name, fileTestName)
		assert.Equal(t, userByID.Name, tc.Payload.Name, errors.ShouldEqual, tc.Name, fileTestName)
		assert.Equal(t, userByID.Email, tc.Payload.Email, errors.ShouldEqual, tc.Name, fileTestName)

		userByID2, getErr2 := userRepository.GetByID(ctx, id)
		assert.NoError(t, getErr2, errors.ShouldNotErr, tc.Name, fileTestName)

		verErr := service.Verification(ctx, model.UserVerificationCode{
			Email: tc.Payload.Email,
			Code:  *userByID2.VerificationCode,
		})
		assert.NoError(t, verErr, errors.ShouldNotErr, tc.Name, fileTestName)

		token, loginErr := service.Login(ctx, model.UserLogin{
			Email:    tc.Payload.Email,
			Password: tc.Payload.Password,
		})
		assert.NoError(t, loginErr, errors.ShouldNotErr, tc.Name, fileTestName)
		assert.True(t, len(token) > 100, "Should more than 100", tc.Name, fileTestName)

		token, loginErr = service.Login(ctx, model.UserLogin{
			Email:    tc.Payload.Email,
			Password: tc.Payload.Password + "+wrongPass",
		})
		assert.Error(t, loginErr, errors.ShouldErr, tc.Name, fileTestName)
		assert.True(t, len(token) == 0, tc.Name, fileTestName)

		token, loginErr = service.Login(ctx, model.UserLogin{
			Email:    tc.Payload.Email + "wrongEmail",
			Password: tc.Payload.Password,
		})
		assert.Error(t, loginErr, errors.ShouldErr, tc.Name, fileTestName)
		assert.True(t, len(token) == 0, tc.Name, fileTestName)

		verErr = service.Verification(ctx, model.UserVerificationCode{
			Email: tc.Payload.Email,
			Code:  *userByID2.VerificationCode,
		})
		assert.Error(t, verErr, errors.ShouldErr, tc.Name, fileTestName)

		forgetPassErr := service.ForgetPassword(ctx, model.UserForgetPassword{
			Email: tc.Payload.Email,
		})
		assert.NoError(t, forgetPassErr, errors.ShouldNotErr, tc.Name, fileTestName)

		userByID2, getErr2 = userRepository.GetByID(ctx, id)
		assert.NoError(t, getErr2, errors.ShouldNotErr, tc.Name, fileTestName)

		deleteErr := service.DeleteUserByVerification(ctx, model.UserVerificationCode{
			Email: tc.Payload.Email,
			Code:  *userByID2.VerificationCode,
		})
		assert.Error(t, deleteErr, errors.ShouldErr, tc.Name, fileTestName)

		deleteErr = userRepository.Delete(ctx, id)
		assert.NoError(t, deleteErr, errors.ShouldErr, tc.Name, fileTestName)
	}
}

// ResetPassword
// UpdatePassword
// UpdateProfile
func TestUpdateAndReset(t *testing.T) {
	permService := permService.NewPermissionService()
	assert.NotNil(t, permService, errors.ShouldNotNil, fileTestName)
	roleService := roleService.NewRoleService(permService)
	assert.NotNil(t, roleService, errors.ShouldNotNil, fileTestName)
	service := NewUserService(roleService)
	assert.NotNil(t, service, errors.ShouldNotNil, fileTestName)
	ctx := helper.NewFiberCtx().Context()
	assert.NotNil(t, ctx, errors.ShouldNotNil, fileTestName)

	validUser := createUser(2)
	defer userRepository.Delete(ctx, validUser.ID)

	type testCase struct {
		Name    string
		Payload model.UserResetPassword
		WantErr bool
	}

	testCases := []testCase{
		{
			Name: "Failed Update -1: email already used",
			Payload: model.UserResetPassword{
				Email: validUser.Email,
				Code:  "wrongCode",
			},
			WantErr: true,
		},
		// {
		// 	Name: "Success Update -1",
		// 	Payload: model.UserResetPassword{
		// 		Email: validUser.Email,
		// 	},
		// },
		// {
		// 	Name: "Success Update -2",
		// 	Payload: model.UserResetPassword{
		// 		Email: validUser.Email,
		// 	},
		// },
	}

	for _, tc := range testCases {
		log.Println(tc.Name, fileTestName)

		resetErr := service.ResetPassword(ctx, tc.Payload)
		if tc.WantErr {
			assert.Error(t, resetErr, errors.ShouldErr, tc.Name, fileTestName)
			continue
		}
		assert.NoError(t, resetErr, errors.ShouldNotErr, tc.Name, fileTestName)

		// userByID, getErr2 := userRepository.GetByID(ctx, id)
		// assert.NoError(t, getErr2, errors.ShouldNotErr, tc.Name, fileTestName)

		// verErr := service.Verification(ctx, model.UserVerificationCode{
		// 	Email: tc.Payload.Email,
		// 	Code:  *userByID.VerificationCode,
		// })
		// assert.NoError(t, verErr, errors.ShouldNotErr, tc.Name, fileTestName)

		// token, loginErr := service.Login(ctx, model.UserLogin{
		// 	Email:    tc.Payload.Email,
		// 	Password: tc.Payload.Password,
		// })
		// assert.NoError(t, loginErr, errors.ShouldNotErr, tc.Name, fileTestName)
		// assert.True(t, len(token) > 100, "Should more than 100", tc.Name, fileTestName)

		// token, loginErr = service.Login(ctx, model.UserLogin{
		// 	Email:    tc.Payload.Email,
		// 	Password: tc.Payload.Password + "+wrongPass",
		// })
		// assert.Error(t, loginErr, errors.ShouldErr, tc.Name, fileTestName)
		// assert.True(t, len(token) == 0, tc.Name, fileTestName)

		// token, loginErr = service.Login(ctx, model.UserLogin{
		// 	Email:    tc.Payload.Email + "wrongEmail",
		// 	Password: tc.Payload.Password,
		// })
		// assert.Error(t, loginErr, errors.ShouldErr, tc.Name, fileTestName)
		// assert.True(t, len(token) == 0, tc.Name, fileTestName)

		// verErr = service.Verification(ctx, model.UserVerificationCode{
		// 	Email: tc.Payload.Email,
		// 	Code:  *userByID2.VerificationCode,
		// })
		// assert.Error(t, verErr, errors.ShouldErr, tc.Name, fileTestName)

		// forgetPassErr := service.ForgetPassword(ctx, model.UserForgetPassword{
		// 	Email: tc.Payload.Email,
		// })
		// assert.NoError(t, forgetPassErr, errors.ShouldNotErr, tc.Name, fileTestName)

		// userByID2, getErr2 = userRepository.GetByID(ctx, id)
		// assert.NoError(t, getErr2, errors.ShouldNotErr, tc.Name, fileTestName)

		// deleteErr := service.DeleteUserByVerification(ctx, model.UserVerificationCode{
		// 	Email: tc.Payload.Email,
		// 	Code:  *userByID2.VerificationCode,
		// })
		// assert.Error(t, deleteErr, errors.ShouldErr, tc.Name, fileTestName)

		// deleteErr = userRepository.Delete(ctx, id)
		// assert.NoError(t, deleteErr, errors.ShouldErr, tc.Name, fileTestName)
	}
}

func createUser(roleID int) entity.User {
	ctx := helper.NewFiberCtx().Context()
	pwHashed, _ := hash.Generate(helper.RandomString(10))
	user := entity.User{
		Name:     helper.RandomString(10),
		Email:    helper.RandomEmail(),
		Password: pwHashed,
	}
	user.SetCreateTime()
	userID, createErr := userRepository.Create(ctx, user, roleID)
	if createErr != nil {
		log.Fatal("error while create new user", fileTestName)
	}
	user.ID = userID
	return user
}
