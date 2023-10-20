package rbac

import (
	"testing"
)

func TestAllPermissions(t *testing.T) {
	for _, permission := range AllPermissions() {
		if permission.Name == "" {
			t.Error("permission name should not string-nil")
		}
	}
}
