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

// This Test is to make sure
// that permission value is unique
func TestCountPermissions(t *testing.T) {
	hashMapPermissions := make(map[string]int, 0)

	for _, permission := range AllPermissions() {
		hashMapPermissions[permission.Name] += 1
		if hashMapPermissions[permission.Name] > 1 {
			t.Error("should 1, not more, non-unique permission detected : ", permission.Name)
		}
	}

	if len(hashMapPermissions) != len(AllPermissions()) {
		t.Error("should equal len, non-unique permission detected")
	}
}
