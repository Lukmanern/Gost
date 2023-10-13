package service

import (
	"fmt"
	"net/smtp"
	"sync"

	"github.com/Lukmanern/gost/internal/env"
)

type EmailService interface {
	Send(to string, subject string, message string) (err error)
	SendBulk(to []string, subject string, message string) (err error)
	SendResetPassword(to string, subject string, message string) (err error)
}

type EmailServiceImpl struct {
	Server   string
	Port     int
	Email    string
	Password string
}

var (
	emailService     *EmailServiceImpl
	emailServiceOnce sync.Once
)

func NewEmailService() EmailService {
	emailServiceOnce.Do(func() {
		config := env.Configuration()
		emailService = &EmailServiceImpl{
			Server:   config.SMTPServer,
			Port:     config.SMTPPort,
			Email:    config.SMTPEmail,
			Password: config.SMTPPassword,
		}
	})

	return emailService
}

func auth(email, password, server string) (auth smtp.Auth) {
	return smtp.PlainAuth("", email, password, server)
}

func (svc EmailServiceImpl) Send(to string, subject string, message string) (err error) {
	smtpAddr := fmt.Sprintf("%s:%d", svc.Server, svc.Port)
	auth := auth(svc.Email, svc.Password, svc.Server)
	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\r\n"
	body := "From: " + svc.Email + "\n" +
		"To: " + to + "\n" +
		"Subject: " + subject + "\n" + mime +
		message
	return smtp.SendMail(smtpAddr, auth, svc.Email, []string{to}, []byte(body))
}

func (svc EmailServiceImpl) SendBulk(to []string, subject string, message string) (err error) {
	smtpAddr := fmt.Sprintf("%s:%d", svc.Server, svc.Port)
	auth := auth(svc.Email, svc.Password, svc.Server)
	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\r\n"
	body := "From: " + svc.Email + "\n" +
		"Subject: " + subject + "\n" + mime +
		message
	return smtp.SendMail(smtpAddr, auth, svc.Email, to, []byte(body))
}

func (svc EmailServiceImpl) SendResetPassword(to string, subject string, message string) (err error) {
	smtpAddr := fmt.Sprintf("%s:%d", svc.Server, svc.Port)
	auth := auth(svc.Email, svc.Password, svc.Server)
	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\r\n"
	body := "From: " + svc.Email + "\n" +
		"To: " + to + "\n" +
		"Subject: " + subject + "\n" + mime +
		message
	return smtp.SendMail(smtpAddr, auth, svc.Email, []string{to}, []byte(body))
}
