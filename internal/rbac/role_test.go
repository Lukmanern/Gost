package rbac

import (
	"testing"
)

func TestAllRoles(t *testing.T) {
	for _, role := range AllRoles() {
		if role.Name == "" {
			t.Error("name should not string-nil")
		}
	}
}
