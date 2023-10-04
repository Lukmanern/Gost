package service

import "context"

type EmailService interface {
	UserResetPassword(ctx context.Context, email string) (err error)
	sendingEmail(ctx context.Context, email, message string)
}
