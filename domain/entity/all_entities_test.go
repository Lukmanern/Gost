package entity

import (
	"testing"
	"time"

	"github.com/Lukmanern/gost/internal/errors"
)

func TestAllTablesName(t *testing.T) {
	type tableNamer interface {
		TableName() string
	}
	allTables := AllTables()

	for _, table := range allTables {
		strct, ok := table.(tableNamer)
		if !ok {
			t.Error("error while getting tableNamer")
		}
		name := strct.TableName()
		if name == "" {
			t.Errorf("TableName for %T should not be empty: " + name)
		}
	}
}

func TestUserSetActivateAccount(t *testing.T) {
	timeNow := time.Now()
	code := "example-code"
	user := User{
		VerificationCode: &code,
		ActivatedAt:      nil,
		TimeFields: TimeFields{
			CreatedAt: &timeNow,
			UpdatedAt: &timeNow,
		},
	}

	user.SetActivateAccount()
	if user.VerificationCode != nil {
		t.Error(errors.ShouldNil)
	}
	if user.ActivatedAt == nil {
		t.Error(errors.ShouldNil)
	}
}
