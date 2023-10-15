package response

import (
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestSuccessNoContent(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{"TestSuccessNoContentValid", false},
		// Add more test cases as needed
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &fiber.Ctx{}
			if err := SuccessNoContent(c); (err != nil) != tt.wantErr {
				t.Errorf("SuccessNoContent() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCreateResponse(t *testing.T) {
	type args struct {
		c          *fiber.Ctx
		statusCode int
		success    bool
		message    string
		data       interface{}
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := CreateResponse(tt.args.c, tt.args.statusCode, tt.args.success, tt.args.message, tt.args.data); (err != nil) != tt.wantErr {
				t.Errorf("CreateResponse() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSuccessLoaded(t *testing.T) {
	type args struct {
		c    *fiber.Ctx
		data interface{}
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := SuccessLoaded(tt.args.c, tt.args.data); (err != nil) != tt.wantErr {
				t.Errorf("SuccessLoaded() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSuccessCreated(t *testing.T) {
	type args struct {
		c    *fiber.Ctx
		data interface{}
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := SuccessCreated(tt.args.c, tt.args.data); (err != nil) != tt.wantErr {
				t.Errorf("SuccessCreated() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBadRequest(t *testing.T) {
	type args struct {
		c       *fiber.Ctx
		message string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := BadRequest(tt.args.c, tt.args.message); (err != nil) != tt.wantErr {
				t.Errorf("BadRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUnauthorized(t *testing.T) {
	type args struct {
		c *fiber.Ctx
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Unauthorized(tt.args.c); (err != nil) != tt.wantErr {
				t.Errorf("Unauthorized() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDataNotFound(t *testing.T) {
	type args struct {
		c       *fiber.Ctx
		message string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := DataNotFound(tt.args.c, tt.args.message); (err != nil) != tt.wantErr {
				t.Errorf("DataNotFound() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestError(t *testing.T) {
	type args struct {
		c       *fiber.Ctx
		message string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Error(tt.args.c, tt.args.message); (err != nil) != tt.wantErr {
				t.Errorf("Error() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestErrorWithData(t *testing.T) {
	type args struct {
		c       *fiber.Ctx
		message string
		data    interface{}
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ErrorWithData(tt.args.c, tt.args.message, tt.args.data); (err != nil) != tt.wantErr {
				t.Errorf("ErrorWithData() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
