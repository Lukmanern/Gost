package errors

import "errors"

const (
	RedisNil    = "redis nil value"
	NotFound    = "data not found"
	ServerErr   = "internal server error: "
	InvalidID   = "invalid id"
	InvalidBody = "invalid json body: "

	ShouldUnauthorized = "should unauthorized"
	ShouldErr          = "should error"
	ShouldNotErr       = "should not error"
	ShouldNil          = "should nil"
	ShouldNotNil       = "should not nil"
	ShouldEqual        = "should equal"
	ShouldNotEqual     = "should not equal"

	LoginShouldSuccess = "login should success"
)

func New(message string) error {
	return errors.New(message)
}
