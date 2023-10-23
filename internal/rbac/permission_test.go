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

func Test_AllPermissionsIDHashMap(t *testing.T) {
	hashMap := AllPermissionsIDHashMap()
	if len(hashMap) < 1 {
		t.Error("len of $hashMap should more than one")
	}
	func() {
		v, ok := hashMap[0]
		if ok {
			t.Error("should not ok")
		}
		if v == 1 {
			t.Error("should not equal to one")
		}
	}()
	func() {
		v, ok := hashMap[1]
		if !ok {
			t.Error("should ok")
		}
		if v != 1 {
			t.Error("should equal to one")
		}
	}()
}
