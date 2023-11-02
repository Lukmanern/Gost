# DevNote

### Todo / OnDev

1. Add file management (upload and download)
2. Make code more clean
   controller dev Done

```

// uint8 is the lowest memory cost in Golang
// maximum value length is 255
type (
	PermissionMap     = map[uint8]uint8
	PermissionNameMap = map[string]uint8
)

var (
	PermissionHashMap     PermissionMap
	PermissionNameHashMap PermissionNameMap
)

// Run once at app.go setupfunc
func PermissionsHashMap() PermissionMap {
	PermissionHashMap := make(PermissionMap, 0)
	permissions := AllPermissions()
	for i := range permissions {
		PermissionHashMap[uint8(i+1)] = 0b_0001
	}

	return PermissionHashMap
}

```

### Done / OnTest

1. Reset Password
2. Email Confirmation For Forget Password
3. Email Confirmation For Registration
4. User request to delete the account
5. Migrate from MySQL to Supabase PostgreSQL
6. Add more test for controller and service
7. add Max Retry/Jail in Login
8. Modify Middleware Test
9. Add Test file
10. Fixing RBAC

### Read List

1. Database connection conf : https://www.alexedwards.net/blog/configuring-sqldb
2. GoFiber Test https://dev.to/koddr/go-fiber-by-examples-testing-the-application-1ldf
3. https://aleksei-kornev.medium.com/production-readiness-checklist-for-backend-applications-8d2b0c57ccec/
4. https://last9.io/blog/deployment-readiness-checklists/
5. https://github.com/gorrion-io/production-readiness-checklist/
6. https://www.cockroachlabs.com/docs/cockroachcloud/production-checklist/
7. https://www.alexedwards.net/blog/ci-with-go-and-github-actions
8. https://roadmap.sh/best-practices/api-security/
