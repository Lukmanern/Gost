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

func TestCountRoles(t *testing.T) {
	hashMapRoles := make(map[string]int, 0)

	for _, role := range AllRoles() {
		hashMapRoles[role.Name] += 1
		if hashMapRoles[role.Name] > 1 {
			t.Error("should 1, not more, non-unique role detected : ", role.Name)
		}
	}

	if len(hashMapRoles) != len(AllRoles()) {
		t.Error("should equal len, non-unique role detected")
	}
}
