package service

import (
	"log"
	"strconv"
	"testing"
	"time"

	"github.com/Lukmanern/gost/domain/entity"
	"github.com/Lukmanern/gost/domain/model"
	"github.com/Lukmanern/gost/internal/consts"
	"github.com/Lukmanern/gost/internal/env"
	"github.com/Lukmanern/gost/internal/hash"
	"github.com/Lukmanern/gost/internal/helper"
	repository "github.com/Lukmanern/gost/repository/user"
	"github.com/stretchr/testify/assert"
)

const (
	headerTestName string = "at UserServiceTest"
)

var (
	timeNow        time.Time
	userRepository repository.UserRepository
)

func init() {
	envFilePath := "./../../.env"
	env.ReadConfig(envFilePath)
	timeNow = time.Now()
	userRepository = repository.NewUserRepository()
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

		deleteErr := service.DeleteAccount(ctx, id)
		assert.NoError(t, deleteErr, consts.ShouldNotErr, tc.Name, headerTestName)

		// value reset
		user = nil
		getErr = nil
		user, getErr = repository.GetByID(ctx, id)
		assert.Error(t, getErr, consts.ShouldErr, tc.Name, headerTestName)
		assert.Nil(t, user, consts.ShouldNil, tc.Name, headerTestName)
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

// func TestLogout(t *testing.T) {
// 	service := NewUserService()
// 	assert.NotNil(t, service, consts.ShouldNotNil, headerTestName)
// 	repository := userRepository
// 	assert.NotNil(t, repository, consts.ShouldNotNil, headerTestName)
// 	c := helper.NewFiberCtx()
// 	assert.NotNil(t, c, consts.ShouldNotNil, headerTestName)

// 	logoutErr := service.Logout(c)
// 	assert.Error(t, logoutErr, consts.ShouldErr, headerTestName)
// }

func TestGetAll(t *testing.T) {
	service := NewUserService()
	assert.NotNil(t, service, consts.ShouldNotNil, headerTestName)
	repository := userRepository
	assert.NotNil(t, repository, consts.ShouldNotNil, headerTestName)
	ctx := helper.NewFiberCtx().Context()
	assert.NotNil(t, ctx, consts.ShouldNotNil, headerTestName)

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

// func TestUpdatePassword(t *testing.T) {
// 	service := NewUserService()
// 	assert.NotNil(t, service, consts.ShouldNotNil, headerTestName)
// 	repository := userRepository
// 	assert.NotNil(t, repository, consts.ShouldNotNil, headerTestName)
// 	ctx := helper.NewFiberCtx().Context()
// 	assert.NotNil(t, ctx, consts.ShouldNotNil, headerTestName)

// 	validUser := createUser()
// 	defer repository.Delete(ctx, validUser.ID)

// 	type testCase struct {
// 		Name    string
// 		Payload model.UserPasswordUpdate
// 		WantErr bool
// 	}

// 	testCases := []testCase{
// 		{
// 			Name: "Success Update -1",
// 			Payload: model.UserPasswordUpdate{
// 				ID:          validUser.ID,
// 				OldPassword: validUser.Password,
// 				NewPassword: helper.RandomString(16),
// 			},
// 			WantErr: false,
// 		},
// 		{
// 			Name: "Failed Update -1: wrong password / password is already changed",
// 			Payload: model.UserPasswordUpdate{
// 				ID:          validUser.ID,
// 				OldPassword: validUser.Password,
// 				NewPassword: helper.RandomString(16),
// 			},
// 			WantErr: true,
// 		},
// 		{
// 			Name: "Failed Update -2: invalid ID",
// 			Payload: model.UserPasswordUpdate{
// 				ID:          -1,
// 				OldPassword: validUser.Password,
// 			},
// 			WantErr: true,
// 		},
// 		{
// 			Name: "Failed Update -3: invalid ID",
// 			Payload: model.UserPasswordUpdate{
// 				ID:          0,
// 				OldPassword: validUser.Password,
// 			},
// 			WantErr: true,
// 		},
// 	}

// 	for _, tc := range testCases {
// 		log.Println(tc.Name, headerTestName)

// 		updateErr := service.UpdatePassword(ctx, tc.Payload)
// 		if tc.WantErr {
// 			assert.Error(t, updateErr, consts.ShouldErr, tc.Name, headerTestName)
// 			continue
// 		}
// 		assert.NoError(t, updateErr, consts.ShouldNotErr, tc.Name, headerTestName)

// 		if tc.Payload.ID == validUser.ID {
// 			token, loginErr := service.Login(ctx, model.UserLogin{
// 				Email:    validUser.Email,
// 				Password: tc.Payload.NewPassword,
// 			})
// 			assert.NoError(t, loginErr, consts.ShouldNotErr, tc.Name, headerTestName)
// 			assert.True(t, token != "", consts.ShouldNotNil, tc.Name, headerTestName)
// 		}
// 	}
// }

func TestDelete(t *testing.T) {
	service := NewUserService()
	assert.NotNil(t, service, consts.ShouldNotNil, headerTestName)
	repository := userRepository
	assert.NotNil(t, repository, consts.ShouldNotNil, headerTestName)
	ctx := helper.NewFiberCtx().Context()
	assert.NotNil(t, ctx, consts.ShouldNotNil, headerTestName)

	userIDs := make([]int, 2)
	for i := range userIDs {
		user := createUser()
		userIDs[i] = user.ID
		defer repository.Delete(ctx, user.ID)
	}

	type testCase struct {
		Name    string
		ID      int
		WantErr bool
	}

	testCases := []testCase{
		{
			Name:    "Failed Delete User -1: invalid ID",
			ID:      -1,
			WantErr: true,
		},
		{
			Name:    "Failed Delete User -2: data not found",
			ID:      userIDs[0] * 99,
			WantErr: true,
		},
	}
	for i, id := range userIDs {
		testCases = append(testCases, testCase{
			Name:    "Success Delete User -" + strconv.Itoa(i+1),
			ID:      id,
			WantErr: false,
		})
		testCases = append(testCases, testCase{
			Name:    "Failed Delete User -" + strconv.Itoa(i+3) + ": already deleted",
			ID:      id,
			WantErr: true,
		})
	}

	for _, tc := range testCases {
		log.Println(tc.Name, headerTestName)

		deleteErr := service.DeleteAccount(ctx, tc.ID)
		if tc.WantErr {
			assert.Error(t, deleteErr, consts.ShouldErr, tc.Name, headerTestName)
			continue
		}
		assert.NoError(t, deleteErr, consts.ShouldNotErr, tc.Name, headerTestName)

		_, getErr := service.MyProfile(ctx, tc.ID)
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
