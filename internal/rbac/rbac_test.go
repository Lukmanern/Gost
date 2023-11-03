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

// func Test_AllPermissionsIDHashMap(t *testing.T) {
// 	hashMap := PermissionIDsHashMap()
// 	if len(hashMap) < 1 {
// 		t.Error("len of $hashMap should more than one")
// 	}
// 	func() {
// 		v, ok := hashMap[0]
// 		if ok {
// 			t.Error("should not ok")
// 		}
// 		if v == 1 {
// 			t.Error("should not equal to one")
// 		}
// 	}()
// 	func() {
// 		v, ok := hashMap[1]
// 		if !ok {
// 			t.Error("should ok")
// 		}
// 		if v != 1 {
// 			t.Error("should equal to one")
// 		}
// 	}()
// }

// func Test_PermissionNamesHashMap(t *testing.T) {
// 	if len(PermissionNamesHashMap()) != len(AllPermissions()) {
// 		t.Error("len should same")
// 	}
// }

func Test_AllRoles(t *testing.T) {
	for _, role := range AllRoles() {
		if role.Name == "" {
			t.Error("name should not string-nil")
		}
	}
}

func Test_CountRoles(t *testing.T) {
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
