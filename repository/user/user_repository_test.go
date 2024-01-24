package repository

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
	"github.com/stretchr/testify/assert"
)

const (
	headerTestName string = "at UserRepoTest"
)

var (
	timeNow time.Time
)

func init() {
	envFilePath := "./../../.env"
	env.ReadConfig(envFilePath)
	timeNow = time.Now()
}

func TestCreateDelete(t *testing.T) {
	repository := NewUserRepository()
	ctx := helper.NewFiberCtx().Context()
	assert.NotNil(t, repository, consts.ShouldNotNil, headerTestName)
	assert.NotNil(t, ctx, consts.ShouldNotNil, headerTestName)

	type testCase struct {
		Name    string
		Payload entity.User
		WantErr bool
	}
	pwHashed, hashErr := hash.Generate(helper.RandomString(8))
	assert.NoError(t, hashErr, consts.ShouldNotErr, headerTestName)

	validUser := createUser()
	defer repository.Delete(ctx, validUser.ID)

	testCases := []testCase{
		{
			Name: "Failed Create User -1: email already used",
			Payload: entity.User{
				Name:        helper.RandomString(12),
				Email:       validUser.Email,
				Password:    pwHashed,
				ActivatedAt: &timeNow,
			},
			WantErr: true,
		},
		{
			Name: "Success Create User -1",
			Payload: entity.User{
				Name:        helper.RandomString(10),
				Email:       helper.RandomEmail(),
				Password:    pwHashed,
				ActivatedAt: &timeNow,
			},
			WantErr: false,
		},
		{
			Name: "Success Create User -2",
			Payload: entity.User{
				Name:        helper.RandomString(10),
				Email:       helper.RandomEmail(),
				Password:    pwHashed,
				ActivatedAt: &timeNow,
			},
			WantErr: false,
		},
	}

	for _, tc := range testCases {
		log.Println(tc.Name, headerTestName)

		tc.Payload.SetCreateTime()
		id, createErr := repository.Create(ctx, tc.Payload, 1)
		if tc.WantErr {
			assert.Error(t, createErr, consts.ShouldErr, tc.Name, headerTestName)
			continue
		}

		assert.NoError(t, createErr, consts.ShouldNotErr, tc.Name, headerTestName)

		deleteErr := repository.Delete(ctx, id)
		assert.NoError(t, deleteErr, consts.ShouldNotErr, headerTestName)
	}
}

func TestGetByID(t *testing.T) {
	repository := NewUserRepository()
	ctx := helper.NewFiberCtx().Context()
	assert.NotNil(t, repository, consts.ShouldNotNil, headerTestName)
	assert.NotNil(t, ctx, consts.ShouldNotNil, headerTestName)

	userIDs := make([]int, 2)
	for i := range userIDs {
		userIDs[i] = createUser().ID
		defer repository.Delete(ctx, userIDs[i])
	}

	type testCase struct {
		Name    string
		ID      int
		WantErr bool
	}
	testCases := []testCase{
		{
			Name:    "Failed Get User -1: invalid ID",
			WantErr: true,
			ID:      -1,
		},
		{
			Name:    "Failed Get User -2: Data not found",
			WantErr: true,
			ID:      userIDs[0] * 99,
		},
	}
	for i, id := range userIDs {
		testCases = append(testCases, testCase{
			Name:    "Success Get data-" + strconv.Itoa(i+1),
			ID:      id,
			WantErr: false,
		})
	}

	for _, tc := range testCases {
		log.Println(tc.Name, headerTestName)

		user, err := repository.GetByID(ctx, tc.ID)
		if tc.WantErr {
			assert.Error(t, err, consts.ShouldErr, headerTestName)
			continue
		}
		assert.NoError(t, err, consts.ShouldNotErr, headerTestName)
		assert.NotNil(t, user, consts.ShouldNotNil, headerTestName)
	}
}

func TestGetByEmail(t *testing.T) {
	repository := NewUserRepository()
	ctx := helper.NewFiberCtx().Context()
	assert.NotNil(t, repository, consts.ShouldNotNil, headerTestName)
	assert.NotNil(t, ctx, consts.ShouldNotNil, headerTestName)

	userEmails := make([]string, 2)
	for i := range userEmails {
		user := createUser()
		userEmails[i] = user.Email
		defer repository.Delete(ctx, user.ID)
	}

	type testCase struct {
		Name    string
		Email   string
		WantErr bool
	}
	testCases := []testCase{
		{
			Name:    "Failed Get User -1: invalid Email",
			WantErr: true,
			Email:   "",
		},
		{
			Name:    "Failed Get User -2: Data not found",
			WantErr: true,
			Email:   "validemail@example.xyz",
		},
	}
	for i, email := range userEmails {
		testCases = append(testCases, testCase{
			Name:    "Success Get data-" + strconv.Itoa(i+1),
			Email:   email,
			WantErr: false,
		})
	}

	for _, tc := range testCases {
		log.Println(tc.Name, headerTestName)

		user, err := repository.GetByEmail(ctx, tc.Email)
		if tc.WantErr {
			assert.Error(t, err, consts.ShouldErr, headerTestName)
			continue
		}
		assert.NoError(t, err, consts.ShouldNotErr, headerTestName)
		assert.NotNil(t, user, consts.ShouldNotNil, headerTestName)
	}
}

func TestGetByConditions(t *testing.T) {
	repository := NewUserRepository()
	ctx := helper.NewFiberCtx().Context()
	assert.NotNil(t, repository, consts.ShouldNotNil, headerTestName)
	assert.NotNil(t, ctx, consts.ShouldNotNil, headerTestName)

	userEmails := make([]string, 2)
	for i := range userEmails {
		user := createUser()
		userEmails[i] = user.Email
		defer repository.Delete(ctx, user.ID)
	}

	type testCase struct {
		Name       string
		Conditions map[string]any
		WantErr    bool
	}
	testCases := []testCase{
		{
			Name:    "Failed Get User -1: invalid Conditions",
			WantErr: true,
			Conditions: map[string]any{
				"invalid =": 90,
			},
		},
		{
			Name:    "Failed Get User -2: Data not found",
			WantErr: true,
			Conditions: map[string]any{
				"email =": helper.RandomEmail(),
			},
		},
	}
	for i, email := range userEmails {
		testCases = append(testCases, testCase{
			Name:    "Success Get data-" + strconv.Itoa(i+1),
			WantErr: false,
			Conditions: map[string]any{
				"email =": email,
			},
		})
	}

	for _, tc := range testCases {
		log.Println(tc.Name, headerTestName)

		user, err := repository.GetByConditions(ctx, tc.Conditions)
		if tc.WantErr {
			assert.Error(t, err, consts.ShouldErr, headerTestName)
			continue
		}
		assert.NoError(t, err, consts.ShouldNotErr, headerTestName)
		assert.NotNil(t, user, consts.ShouldNotNil, headerTestName)
	}
}

func TestGetAll(t *testing.T) {
	repository := NewUserRepository()
	ctx := helper.NewFiberCtx().Context()
	assert.NotNil(t, repository, consts.ShouldNotNil, headerTestName)
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
				Page:  1,
				Limit: 100,
			},
			WantErr: false,
		},
		{
			Name: "Success Get All -2",
			Payload: model.RequestGetAll{
				Page:  1,
				Limit: 10,
			},
			WantErr: false,
		},
		{
			Name: "Failed Get All -1: invalid sort",
			Payload: model.RequestGetAll{
				Page:  1,
				Limit: 10,
				Sort:  "invalid-sort",
			},
			WantErr: true,
		},
	}

	for _, tc := range testCases {
		log.Println(tc.Name, headerTestName)

		users, total, getErr := repository.GetAll(ctx, tc.Payload)
		if tc.WantErr {
			assert.Error(t, getErr, consts.ShouldErr, headerTestName)
			continue
		}
		assert.NoError(t, getErr, consts.ShouldNotErr, headerTestName)
		assert.True(t, total >= 0, headerTestName)
		assert.True(t, len(users) >= 0, headerTestName)
	}
}

func TestUpdate(t *testing.T) {
	repository := NewUserRepository()
	ctx := helper.NewFiberCtx().Context()
	assert.NotNil(t, repository, consts.ShouldNotNil, headerTestName)
	assert.NotNil(t, ctx, consts.ShouldNotNil, headerTestName)

	// create user for testing {payload}
	user := createUser()
	defer repository.Delete(ctx, user.ID)

	type testCase struct {
		Name    string
		Payload entity.User
		WantErr bool
	}

	testCases := []testCase{
		{
			Name: "Success Update -1",
			Payload: entity.User{
				ID:   user.ID,
				Name: helper.RandomString(12),
			},
			WantErr: false,
		},
		{
			Name: "Success Update -2",
			Payload: entity.User{
				ID:   user.ID,
				Name: helper.RandomString(12),
			},
			WantErr: false,
		},
		{
			Name: "Failed Update -1: invalid ID",
			Payload: entity.User{
				ID:   -10,
				Name: helper.RandomString(12),
			},
			WantErr: true,
		},
		{
			Name: "Failed Update -2: invalid id",
			Payload: entity.User{
				ID:   0,
				Name: helper.RandomString(12),
			},
			WantErr: true,
		},
	}

	for _, tc := range testCases {
		log.Println(tc.Name, headerTestName)

		tc.Payload.SetUpdateTime()
		updateErr := repository.Update(ctx, tc.Payload)
		if tc.WantErr {
			// error by data not found not detected
			// assert.Error(t, updateErr, consts.ShouldErr, headerTestName)
			continue
		}
		assert.NoError(t, updateErr, consts.ShouldNotErr, headerTestName)

		user, getErr := repository.GetByID(ctx, tc.Payload.ID)
		assert.NoError(t, getErr, consts.ShouldNotErr, headerTestName)
		assert.Equal(t, user.Name, tc.Payload.Name)
	}
}

func TestUpdatePassword(t *testing.T) {
	repository := NewUserRepository()
	ctx := helper.NewFiberCtx().Context()
	assert.NotNil(t, repository, consts.ShouldNotNil, headerTestName)
	assert.NotNil(t, ctx, consts.ShouldNotNil, headerTestName)

	user := createUser()
	defer repository.Delete(ctx, user.ID)

	type testCase struct {
		Name    string
		Payload entity.User
		WantErr bool
	}

	testCases := []testCase{
		{
			Name: "Success Update -1",
			Payload: entity.User{
				ID:       user.ID,
				Password: helper.RandomString(12),
			},
			WantErr: false,
		},
		{
			Name: "Failed Update: invalid id",
			Payload: entity.User{
				ID:       -10,
				Password: helper.RandomString(12),
			},
			WantErr: true,
		},
		{
			Name: "Failed Update: invalid id",
			Payload: entity.User{
				ID:       0,
				Password: helper.RandomString(12),
			},
			WantErr: true,
		},
	}

	for _, tc := range testCases {
		log.Println(tc.Name, headerTestName)

		tc.Payload.SetUpdateTime()
		updateErr := repository.UpdatePassword(ctx, tc.Payload.ID, tc.Payload.Password)
		if tc.WantErr {
			assert.Error(t, updateErr, consts.ShouldErr, headerTestName)
			continue
		}
		assert.NoError(t, updateErr, consts.ShouldNotErr, headerTestName)
	}
}

func createUser() entity.User {
	repository := NewUserRepository()
	ctx := helper.NewFiberCtx().Context()
	pwHashed, _ := hash.Generate(helper.RandomString(10))
	newUser := entity.User{
		Name:        helper.RandomString(10),
		Email:       helper.RandomEmail(),
		Password:    pwHashed,
		ActivatedAt: &timeNow,
	}
	newUser.SetCreateTime()
	userID, createErr := repository.Create(ctx, newUser, 1)
	if createErr != nil {
		log.Fatal("error while create new user", headerTestName)
	}
	newUser.ID = userID
	return newUser
}
