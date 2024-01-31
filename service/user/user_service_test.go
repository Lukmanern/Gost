package service

import (
	"log"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/Lukmanern/gost/database/connector"
	"github.com/Lukmanern/gost/domain/entity"
	"github.com/Lukmanern/gost/domain/model"
	"github.com/Lukmanern/gost/internal/consts"
	"github.com/Lukmanern/gost/internal/env"
	"github.com/Lukmanern/gost/internal/hash"
	"github.com/Lukmanern/gost/internal/helper"
	repository "github.com/Lukmanern/gost/repository/user"
	"github.com/go-redis/redis"
	"github.com/stretchr/testify/assert"
)

const (
	headerTestName string = "at User Service Test"
)

var (
	timeNow        time.Time
	userRepository repository.UserRepository
	redisConTest   *redis.Client
)

func init() {
	envFilePath := "./../../.env"
	env.ReadConfig(envFilePath)
	timeNow = time.Now()
	userRepository = repository.NewUserRepository()
	redisConTest = connector.LoadRedisCache()
}

func TestRegister(t *testing.T) {
	service := NewUserService()
	assert.NotNil(t, service, consts.ShouldNotNil, headerTestName)
	repository := userRepository
	assert.NotNil(t, repository, consts.ShouldNotNil, headerTestName)
	ctx := helper.NewFiberCtx().Context()
	assert.NotNil(t, ctx, consts.ShouldNotNil, headerTestName)

	type testCase struct {
		Name    string
		Payload model.UserRegister
		WantErr bool
	}

	validUser := createUser()
	defer repository.Delete(ctx, validUser.ID)

	testCases := []testCase{
		{
			Name: "Failed Create -1: email already used",
			Payload: model.UserRegister{
				Name:     helper.RandomString(15),
				Email:    validUser.Email,
				Password: "password00",
			},
			WantErr: true,
		},
		{
			Name: "Failed Create -2: invalid role ID",
			Payload: model.UserRegister{
				Name:     helper.RandomString(15),
				Email:    helper.RandomEmail(),
				Password: "password00",
				RoleIDs:  []int{-1, 0},
			},
			WantErr: true,
		},
		{
			Name: "Success Create -1",
			Payload: model.UserRegister{
				Name:     helper.RandomString(15),
				Email:    helper.RandomEmail(),
				Password: "password00",
			},
			WantErr: false,
		},
		{
			Name: "Success Create -2",
			Payload: model.UserRegister{
				Name:     helper.RandomString(15),
				Email:    helper.RandomEmail(),
				Password: "password00",
			},
			WantErr: false,
		},
	}
	for _, tc := range testCases {
		log.Println(tc.Name, headerTestName)

		id, createErr := service.Register(ctx, tc.Payload)
		if tc.WantErr {
			assert.Error(t, createErr, consts.ShouldErr, tc.Name, headerTestName)
			continue
		}
		assert.NoError(t, createErr, consts.ShouldNotErr, tc.Name, headerTestName)

		user, getErr := repository.GetByID(ctx, id)
		assert.NoError(t, getErr, consts.ShouldNotErr, tc.Name, headerTestName)
		assert.NotNil(t, user, consts.ShouldNotNil, tc.Name, headerTestName)

		deleteErr := service.DeleteAccount(ctx, model.UserDeleteAccount{
			ID:       id,
			Password: tc.Payload.Password,
		})
		assert.NoError(t, deleteErr, consts.ShouldNotErr, tc.Name, headerTestName)

		// value reset
		user = nil
		getErr = nil
		user, getErr = repository.GetByID(ctx, id)
		assert.Error(t, getErr, consts.ShouldErr, tc.Name, headerTestName)
		assert.Nil(t, user, consts.ShouldNil, tc.Name, headerTestName)
	}
}

func TestAccountActivation(t *testing.T) {
	service := NewUserService()
	assert.NotNil(t, service, consts.ShouldNotNil, headerTestName)
	repository := userRepository
	assert.NotNil(t, repository, consts.ShouldNotNil, headerTestName)
	ctx := helper.NewFiberCtx().Context()
	assert.NotNil(t, ctx, consts.ShouldNotNil, headerTestName)

	validUser := model.UserRegister{
		Name:     strings.ToLower(helper.RandomString(12)),
		Email:    helper.RandomEmail(),
		Password: helper.RandomString(12),
		RoleIDs:  []int{1, 2, 3},
	}
	id, err := service.Register(ctx, validUser)
	assert.Nil(t, err, consts.ShouldNotNil, headerTestName)
	defer repository.Delete(ctx, id)

	key := validUser.Email + KEY_ACCOUNT_ACTIVATION
	validCode := redisConTest.Get(key).Val()
	assert.True(t, len(validCode) > 0, consts.ShouldNotNil, headerTestName)

	type testCase struct {
		Name    string
		Payload model.UserActivation
		WantErr bool
	}

	testCases := []testCase{
		{
			Name: "Failed Activation -1: wrong code",
			Payload: model.UserActivation{
				Code:  "wrongcode",
				Email: validUser.Email,
			},
			WantErr: true,
		},
		{
			Name: "Success Activation -1",
			Payload: model.UserActivation{
				Code:  validCode,
				Email: validUser.Email,
			},
			WantErr: false,
		},
		{
			Name: "Failed Activation -2: code is already used",
			Payload: model.UserActivation{
				Code:  validCode,
				Email: validUser.Email,
			},
			WantErr: true,
		},
		{
			Name: "Failed Activation -3: account not found",
			Payload: model.UserActivation{
				Code:  validCode,
				Email: helper.RandomEmail(),
			},
			WantErr: true,
		},
	}

	for _, tc := range testCases {
		log.Println(tc.Name, headerTestName)

		err := service.AccountActivation(ctx, tc.Payload)
		if tc.WantErr {
			assert.Error(t, err, consts.ShouldErr, tc.Name, headerTestName)
			continue
		}
		assert.NoError(t, err, consts.ShouldNotErr, tc.Name, headerTestName)
	}
}

func TestLogin(t *testing.T) {
	service := NewUserService()
	assert.NotNil(t, service, consts.ShouldNotNil, headerTestName)
	repository := userRepository
	assert.NotNil(t, repository, consts.ShouldNotNil, headerTestName)
	ctx := helper.NewFiberCtx().Context()
	assert.NotNil(t, ctx, consts.ShouldNotNil, headerTestName)

	users := make([]entity.User, 2)
	for i := range users {
		users[i] = createUser()
		defer repository.Delete(ctx, users[i].ID)
	}

	type testCase struct {
		Name    string
		Payload model.UserLogin
		WantErr bool
	}
	testCases := []testCase{
		{
			Name:    "Failed Login -1: void payload",
			WantErr: true,
		},
		{
			Name:    "Failed Login -2: data not found",
			WantErr: true,
			Payload: model.UserLogin{
				Email:    "wrong-email",
				Password: "xx",
			},
		},
		{
			Name:    "Failed Login -3: data not found",
			WantErr: true,
			Payload: model.UserLogin{
				Email:    "",
				Password: "xx",
			},
		},
		{
			Name:    "Failed Login -3: wrong password",
			WantErr: true,
			Payload: model.UserLogin{
				Email:    users[0].Email,
				Password: "wrong-password",
			},
		},
	}
	for i, user := range users {
		testCases = append(testCases, testCase{
			Name: "Success login -" + strconv.Itoa(i+1),
			Payload: model.UserLogin{
				Email:    user.Email,
				Password: user.Password,
			},
			WantErr: false,
		})
	}

	for _, tc := range testCases {
		log.Println(tc.Name, headerTestName)
		_, loginErr := service.Login(ctx, tc.Payload)
		if tc.WantErr {
			assert.Error(t, loginErr, consts.ShouldErr, headerTestName)
			continue
		}
		assert.NoError(t, loginErr, consts.ShouldNotErr)
	}
}

func TestForgetPassword(t *testing.T) {
	service := NewUserService()
	assert.NotNil(t, service, consts.ShouldNotNil, headerTestName)
	repository := userRepository
	assert.NotNil(t, repository, consts.ShouldNotNil, headerTestName)
	ctx := helper.NewFiberCtx().Context()
	assert.NotNil(t, ctx, consts.ShouldNotNil, headerTestName)

	validUser := createUser()
	defer repository.Delete(ctx, validUser.ID)

	type testCase struct {
		Name    string
		WantErr bool
		Payload model.UserForgetPassword
	}

	testCases := []testCase{
		{
			Name:    "Success Forget Password -1",
			Payload: model.UserForgetPassword{Email: validUser.Email},
			WantErr: false,
		},
		{
			Name:    "Success Forget Password -2",
			Payload: model.UserForgetPassword{Email: validUser.Email},
			WantErr: false,
		},
		{
			Name:    "Failed Forget Password -1: user not found",
			Payload: model.UserForgetPassword{Email: helper.RandomEmail()},
			WantErr: true,
		},
		{
			Name:    "Failed Forget Password -2: user not found",
			Payload: model.UserForgetPassword{Email: helper.RandomEmail()},
			WantErr: true,
		},
	}
	for _, tc := range testCases {
		log.Println(tc.Name, headerTestName)

		err := service.ForgetPassword(ctx, tc.Payload)
		if tc.WantErr {
			assert.Error(t, err, consts.ShouldErr, tc.Name, headerTestName)
			continue
		}
		assert.NoError(t, err, consts.ShouldNotErr, tc.Name, headerTestName)
	}
}

func TestResetPassword(t *testing.T) {
	service := NewUserService()
	assert.NotNil(t, service, consts.ShouldNotNil, headerTestName)
	repository := userRepository
	assert.NotNil(t, repository, consts.ShouldNotNil, headerTestName)
	ctx := helper.NewFiberCtx().Context()
	assert.NotNil(t, ctx, consts.ShouldNotNil, headerTestName)

	validUser := createUser()
	defer repository.Delete(ctx, validUser.ID)

	err := service.ForgetPassword(ctx, model.UserForgetPassword{Email: validUser.Email})
	assert.Nil(t, err, consts.ShouldNil, headerTestName)

	key := validUser.Email + KEY_FORGET_PASSWORD
	validCode := redisConTest.Get(key).Val()
	assert.True(t, len(validCode) > 0, consts.ShouldNotNil, headerTestName)

	type testCase struct {
		Name    string
		Payload model.UserResetPassword
		WantErr bool
	}

	testCases := []testCase{
		{
			Name: "Success Reset Password -1",
			Payload: model.UserResetPassword{
				Email:       validUser.Email,
				Code:        validCode,
				NewPassword: "new-password",
			},
			WantErr: false,
		},
		{
			Name: "Failed Reset Password -1: code not found",
			Payload: model.UserResetPassword{
				Email:       validUser.Email,
				Code:        validCode,
				NewPassword: "new-password",
			},
			WantErr: true,
		},
		{
			Name: "Failed Reset Password -2: user not found",
			Payload: model.UserResetPassword{
				Email:       helper.RandomEmail(),
				Code:        validCode,
				NewPassword: "new-password",
			},
			WantErr: true,
		},
	}
	for _, tc := range testCases {
		log.Println(tc.Name, headerTestName)

		err := service.ResetPassword(ctx, tc.Payload)
		if tc.WantErr {
			assert.Error(t, err, consts.ShouldErr, tc.Name, headerTestName)
			continue
		}
		assert.NoError(t, err, consts.ShouldNotErr, tc.Name, headerTestName)
	}
}

func TestLogout(t *testing.T) {
	service := NewUserService()
	assert.NotNil(t, service, consts.ShouldNotNil, headerTestName)
	c := helper.NewFiberCtx()
	assert.NotNil(t, c, consts.ShouldNotNil, headerTestName)

	logoutErr := service.Logout(c)
	assert.NoError(t, logoutErr, consts.ShouldErr, headerTestName)
}

func TestGetAll(t *testing.T) {
	service := NewUserService()
	assert.NotNil(t, service, consts.ShouldNotNil, headerTestName)
	repository := userRepository
	assert.NotNil(t, repository, consts.ShouldNotNil, headerTestName)
	ctx := helper.NewFiberCtx().Context()
	assert.NotNil(t, ctx, consts.ShouldNotNil, headerTestName)

	for i := 0; i < 3; i++ {
		validUser := createUser()
		defer repository.Delete(ctx, validUser.ID)
	}

	type testCase struct {
		Name    string
		Payload model.RequestGetAll
		WantErr bool
	}

	testCases := []testCase{
		{
			Name: "Success Get All -1",
			Payload: model.RequestGetAll{
				Limit: 100,
				Page:  1,
			},
			WantErr: false,
		},
		{
			Name: "Success Get All -2",
			Payload: model.RequestGetAll{
				Limit: 12,
				Page:  2,
				Sort:  "name",
			},
			WantErr: false,
		},
		{
			Name: "Failed Get All -1: invalid sort",
			Payload: model.RequestGetAll{
				Limit: 12,
				Page:  2,
				Sort:  "invalid",
			},
			WantErr: true,
		},
	}
	for _, tc := range testCases {
		log.Println(tc.Name, headerTestName)

		_, _, getErr := service.GetAll(ctx, tc.Payload)
		if tc.WantErr {
			assert.Error(t, getErr, consts.ShouldErr, tc.Name, headerTestName)
			continue
		}
		assert.NoError(t, getErr, consts.ShouldNotErr, tc.Name, headerTestName)
	}
}

func TestSoftDelete(t *testing.T) {
	service := NewUserService()
	assert.NotNil(t, service, consts.ShouldNotNil, headerTestName)
	repository := userRepository
	assert.NotNil(t, repository, consts.ShouldNotNil, headerTestName)
	ctx := helper.NewFiberCtx().Context()
	assert.NotNil(t, ctx, consts.ShouldNotNil, headerTestName)

	validUser := createUser()
	defer repository.Delete(ctx, validUser.ID)

	type testCase struct {
		Name    string
		WantErr bool
		ID      int
	}

	testCases := []testCase{
		{
			Name:    "Success Soft Delete -1",
			WantErr: false,
			ID:      validUser.ID,
		},
		{
			Name:    "Failed Soft Delete -1: user not found",
			WantErr: true,
			ID:      validUser.ID + 99,
		},
		{
			Name:    "Failed Soft Delete -1: invalid ID / user not found",
			WantErr: true,
			ID:      -10,
		},
	}

	for _, tc := range testCases {
		log.Println(tc.Name, headerTestName)

		err := service.SoftDelete(ctx, tc.ID)
		if tc.WantErr {
			assert.Error(t, err, consts.ShouldErr, tc.Name, headerTestName)
			continue
		}
		assert.NoError(t, err, consts.ShouldNotErr, tc.Name, headerTestName)

		user, getErr := repository.GetByID(ctx, tc.ID)
		assert.NoError(t, getErr, consts.ShouldNotErr, tc.Name, headerTestName)
		assert.NotNil(t, user, consts.ShouldNotNil, tc.Name, headerTestName)
		assert.NotNil(t, user.DeletedAt, consts.ShouldNotNil, tc.Name, headerTestName)
	}
}

func TestMyProfile(t *testing.T) {
	service := NewUserService()
	assert.NotNil(t, service, consts.ShouldNotNil, headerTestName)
	repository := userRepository
	assert.NotNil(t, repository, consts.ShouldNotNil, headerTestName)
	ctx := helper.NewFiberCtx().Context()
	assert.NotNil(t, ctx, consts.ShouldNotNil, headerTestName)

	validUser := createUser()
	defer repository.Delete(ctx, validUser.ID)
	validUser2 := createUser()
	defer repository.Delete(ctx, validUser2.ID)

	type testCase struct {
		Name    string
		ID      int
		WantErr bool
	}

	testCases := []testCase{
		{
			Name:    "Success Get My Profile -1",
			ID:      validUser.ID,
			WantErr: false,
		},
		{
			Name:    "Success Get My Profile -2",
			ID:      validUser2.ID,
			WantErr: false,
		},
		{
			Name:    "Failed Get My Profile -1: data not found",
			ID:      validUser2.ID * 99,
			WantErr: true,
		},
		{
			Name:    "Failed Get My Profile -2: invalid ID",
			ID:      -1,
			WantErr: true,
		},
		{
			Name:    "Failed Get My Profile -3: invalid ID",
			ID:      0,
			WantErr: true,
		},
	}

	for _, tc := range testCases {
		log.Println(tc.Name, headerTestName)

		user, getErr := service.MyProfile(ctx, tc.ID)
		if tc.WantErr {
			assert.Error(t, getErr, consts.ShouldErr, tc.Name, headerTestName)
			continue
		}
		assert.NoError(t, getErr, consts.ShouldNotErr, tc.Name, headerTestName)
		assert.NotNil(t, user, consts.ShouldNotErr, tc.Name, headerTestName)
	}
}

func TestUpdateProfile(t *testing.T) {
	service := NewUserService()
	assert.NotNil(t, service, consts.ShouldNotNil, headerTestName)
	repository := userRepository
	assert.NotNil(t, repository, consts.ShouldNotNil, headerTestName)
	ctx := helper.NewFiberCtx().Context()
	assert.NotNil(t, ctx, consts.ShouldNotNil, headerTestName)

	validUser := createUser()
	defer repository.Delete(ctx, validUser.ID)

	type testCase struct {
		Name    string
		Payload model.UserUpdate
		WantErr bool
	}

	testCases := []testCase{
		{
			Name: "Success Update -1",
			Payload: model.UserUpdate{
				ID:   validUser.ID,
				Name: helper.RandomString(12) + "xxxx",
			},
			WantErr: false,
		},
		{
			Name: "Success Update -2",
			Payload: model.UserUpdate{
				ID:   validUser.ID,
				Name: helper.RandomString(6),
			},
			WantErr: false,
		},
		{
			Name: "Failed Update -1: invalid ID",
			Payload: model.UserUpdate{
				ID:   -1,
				Name: helper.RandomString(12),
			},
			WantErr: true,
		},
		{
			Name: "Failed Update -2: invalid ID",
			Payload: model.UserUpdate{
				ID:   0,
				Name: helper.RandomString(12),
			},
			WantErr: true,
		},
	}

	for _, tc := range testCases {
		log.Println(tc.Name, headerTestName)

		updateErr := service.UpdateProfile(ctx, tc.Payload)
		if tc.WantErr {
			assert.Error(t, updateErr, consts.ShouldErr, tc.Name, headerTestName)
			continue
		}
		assert.NoError(t, updateErr, consts.ShouldNotErr, tc.Name, headerTestName)

		user, getErr := repository.GetByID(ctx, tc.Payload.ID)
		assert.NoError(t, getErr, consts.ShouldNotErr, tc.Name, headerTestName)
		assert.NotNil(t, user, consts.ShouldNotNil, tc.Name, headerTestName)
		assert.Equal(t, user.Name, tc.Payload.Name, consts.ShouldNotNil, tc.Name, headerTestName)
	}
}

func TestUpdatePassword(t *testing.T) {
	service := NewUserService()
	assert.NotNil(t, service, consts.ShouldNotNil, headerTestName)
	repository := userRepository
	assert.NotNil(t, repository, consts.ShouldNotNil, headerTestName)
	ctx := helper.NewFiberCtx().Context()
	assert.NotNil(t, ctx, consts.ShouldNotNil, headerTestName)

	validUser := createUser()
	defer repository.Delete(ctx, validUser.ID)

	type testCase struct {
		Name    string
		Payload model.UserPasswordUpdate
		WantErr bool
	}

	testCases := []testCase{
		{
			Name: "Success Update Password -1",
			Payload: model.UserPasswordUpdate{
				ID:          validUser.ID,
				OldPassword: validUser.Password,
				NewPassword: helper.RandomString(16),
			},
			WantErr: false,
		},
		{
			Name: "Failed Update Password -1: wrong password / password is already changed",
			Payload: model.UserPasswordUpdate{
				ID:          validUser.ID,
				OldPassword: validUser.Password,
				NewPassword: helper.RandomString(16),
			},
			WantErr: true,
		},
		{
			Name: "Failed Update Password -2: invalid ID",
			Payload: model.UserPasswordUpdate{
				ID:          -1,
				OldPassword: validUser.Password,
			},
			WantErr: true,
		},
		{
			Name: "Failed Update Password -3: invalid ID",
			Payload: model.UserPasswordUpdate{
				ID:          0,
				OldPassword: validUser.Password,
			},
			WantErr: true,
		},
	}

	for _, tc := range testCases {
		log.Println(tc.Name, headerTestName)

		updateErr := service.UpdatePassword(ctx, tc.Payload)
		if tc.WantErr {
			assert.Error(t, updateErr, consts.ShouldErr, tc.Name, headerTestName)
			continue
		}
		assert.NoError(t, updateErr, consts.ShouldNotErr, tc.Name, headerTestName)

		if tc.Payload.ID == validUser.ID {
			token, loginErr := service.Login(ctx, model.UserLogin{
				Email:    validUser.Email,
				Password: tc.Payload.NewPassword,
			})
			assert.NoError(t, loginErr, consts.ShouldNotErr, tc.Name, headerTestName)
			assert.True(t, token != "", consts.ShouldNotNil, tc.Name, headerTestName)
		}
	}
}

func TestDelete(t *testing.T) {
	service := NewUserService()
	assert.NotNil(t, service, consts.ShouldNotNil, headerTestName)
	repository := userRepository
	assert.NotNil(t, repository, consts.ShouldNotNil, headerTestName)
	ctx := helper.NewFiberCtx().Context()
	assert.NotNil(t, ctx, consts.ShouldNotNil, headerTestName)

	users := make([]model.UserDeleteAccount, 2)
	for i := range users {
		user := createUser()
		users[i] = model.UserDeleteAccount{
			ID:       user.ID,
			Password: user.Password,
		}
		defer repository.Delete(ctx, user.ID)
	}

	type testCase struct {
		Name    string
		Payload model.UserDeleteAccount
		WantErr bool
	}

	testCases := []testCase{
		{
			Name: "Failed Delete User -1: invalid ID",
			Payload: model.UserDeleteAccount{
				ID: -1,
			},
			WantErr: true,
		},
		{
			Name: "Failed Delete User -2: data not found",
			Payload: model.UserDeleteAccount{
				ID: users[0].ID * 99,
			},
			WantErr: true,
		},
		{
			Name: "Failed Delete User -3: wrong password",
			Payload: model.UserDeleteAccount{
				ID:       users[0].ID,
				Password: "wrong-password",
			},
			WantErr: true,
		},
	}
	for i, user := range users {
		testCases = append(testCases, testCase{
			Name: "Success Delete User -" + strconv.Itoa(i+1),
			Payload: model.UserDeleteAccount{
				ID:       user.ID,
				Password: user.Password,
			},
			WantErr: false,
		})
		testCases = append(testCases, testCase{
			Name: "Failed Delete User -" + strconv.Itoa(i+3) + ": already deleted",
			Payload: model.UserDeleteAccount{
				ID:       user.ID,
				Password: user.Password,
			},
			WantErr: true,
		})
	}

	for _, tc := range testCases {
		log.Println(tc.Name, headerTestName)

		deleteErr := service.DeleteAccount(ctx, model.UserDeleteAccount{
			ID:       tc.Payload.ID,
			Password: tc.Payload.Password,
		})
		if tc.WantErr {
			assert.Error(t, deleteErr, consts.ShouldErr, tc.Name, headerTestName)
			continue
		}
		assert.NoError(t, deleteErr, consts.ShouldNotErr, tc.Name, headerTestName)

		_, getErr := service.MyProfile(ctx, tc.Payload.ID)
		assert.Error(t, getErr, consts.ShouldErr, tc.Name, headerTestName)
	}
}

func createUser() entity.User {
	pw := helper.RandomString(15)
	pwHashed, _ := hash.Generate(pw)
	repository := userRepository
	ctx := helper.NewFiberCtx().Context()
	data := entity.User{
		Name:        helper.RandomString(15),
		Email:       helper.RandomEmail(),
		Password:    pwHashed,
		ActivatedAt: &timeNow,
	}
	data.SetCreateTime()
	id, err := repository.Create(ctx, data, []int{1})
	if err != nil {
		log.Fatal("failed create a new user", headerTestName)
	}
	data.Password = pw
	data.ID = id
	return data
}
