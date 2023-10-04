package service

import "context"

type UserAuthService interface {
	Login(ctx context.Context)
	Logout(ctx context.Context)
	ForgotPassword(ctx context.Context)
	UpdatePassword(ctx context.Context)
	UpdateProfile(ctx context.Context)
}
