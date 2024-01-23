package consts

const (
	SuccessCreated = "data successfully created"
	SuccessLoaded  = "data successfully loaded"
	Unauthorized   = "unauthorized"
	BadRequest     = "bad request, please check your request and try again"
	NotFound       = "data not found"
)

const (
	InvalidJSONBody = "invalid JSON body"
	InvalidUserID   = "invalid user ID"
	InvalidID       = "invalid ID"

	RedisNil = "redis nil value"

	ErrGetIDFromJWT = "error while getting user ID from JWT-Claims"
	ErrHashing      = "error while hashing password, please try again"
)

const (
	ShouldErr      = "should error"
	ShouldNotErr   = "should not error"
	ShouldNil      = "should nil"
	ShouldNotNil   = "should not nil"
	ShouldEqual    = "should equal"
	ShouldNotEqual = "should not equal"

	LoginShouldSuccess = "login should success"
)
