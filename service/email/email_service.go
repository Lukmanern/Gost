package service

import (
	"errors"
	"fmt"
	"net/smtp"
	"sync"

	"github.com/Lukmanern/gost/internal/env"
	"github.com/Lukmanern/gost/internal/helper"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type EmailService interface {
	Send(emails []string, subject string, message string) (res map[string]bool, err error)
	getAuth() smtp.Auth
	getSMTPAddr() string
	getMime() string
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

func (svc EmailServiceImpl) Send(emails []string, subject string, message string) (map[string]bool, error) {
	if validateErr := helper.ValidateEmails(emails...); validateErr != nil {
		return nil, validateErr
	}

	subject = cases.Title(language.Und).String(subject)
	lenEmails := len(emails)
	errorSends := make([]error, lenEmails)
	var wg sync.WaitGroup

	addr := svc.getSMTPAddr()
	auth := svc.getAuth()
	mime := svc.getMime()
	for i, email := range emails {
		body := "From: " + svc.Email + "\n" +
			"To: " + email + "\n" +
			"Subject: " + subject + "\n" + mime +
			message
		wg.Add(1)
		go func(i int, email string) {
			defer wg.Done()
			errSend := smtp.SendMail(addr, auth, svc.Email, []string{email}, []byte(body))
			if errSend != nil {
				errorSends[i] = errSend
			}
		}(i, email)
	}
	wg.Wait()

	var hasError error = nil
	res := make(map[string]bool, lenEmails)
	for i, email := range emails {
		if errorSends[i] != nil {
			res[email] = false
			errMsg := "emails may have failed, check $res for detail, in $res true for success"
			hasError = errors.New(errMsg)
			continue
		}
		res[email] = true
	}
	if hasError != nil {
		return res, hasError
	}
	return res, nil
}

func (svc EmailServiceImpl) getAuth() smtp.Auth {
	return smtp.PlainAuth("", svc.Email, svc.Password, svc.Server)
}

func (svc EmailServiceImpl) getSMTPAddr() string {
	return fmt.Sprintf("%s:%d", svc.Server, svc.Port)
}

func (svc EmailServiceImpl) getMime() string {
	return "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\r\n"
}
