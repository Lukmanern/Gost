package rbac

import (
	"testing"
)

func TestAllPermissions(t *testing.T) {
	defer func() {
		r := recover()
		if r != nil {
			t.Error("should not panic, but got:", r)
		}
	}()

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

func TestAllRoles(t *testing.T) {
	for _, role := range AllRoles() {
		if role.Name == "" {
			t.Error("name should not string-nil")
		}
	}
}

func TestCountRoles(t *testing.T) {
	hashMapRoles := make(map[string]int, 0)

	for _, role := range AllRoles() {
		hashMapRoles[role.Name] += 1
		if hashMapRoles[role.Name] > 1 {
			t.Error("should 1, not more, non-unique role (role:name) detected : ", role.Name)
		}
	}

	if len(hashMapRoles) != len(AllRoles()) {
		t.Error("should equal len, non-unique role detected")
	}
}
