package helper

import (
	"net"
	"strings"
	"testing"
)

func TestRandomString(t *testing.T) {
	for i := 0; i < 25; i++ {
		s := RandomString(uint(i))
		if len(s) != i {
			t.Error("len of string should equal")
		}
	}
}

func TestRandomEmails(t *testing.T) {
	for i := 1; i <= 20; i++ {
		emails := RandomEmails(uint(i))
		if len(emails) != i {
			t.Error("total of emails should equal")
		}
		for _, email := range emails {
			if len(email) < 10 {
				t.Error("length of an email should not less than 10")
			}
			if email != strings.ToLower(email) {
				t.Error("email should lower by results")
			}
		}
	}
}

func TestRandomEmail(t *testing.T) {
	for i := 0; i < 25; i++ {
		email := RandomEmail()
		validateErr := ValidateEmails(email)
		if validateErr != nil {
			t.Error("should not error")
		}
		if len(email) < 25 {
			t.Error("should more than 25")
		}
	}
}

func TestNewFiberCtx(t *testing.T) {
	c := NewFiberCtx()
	if c == nil {
		t.Error("should not nil")
	}
}

func TestRandomIPAddress(t *testing.T) {
	for i := 0; i < 20; i++ {
		ipRand := RandomIPAddress()
		ip := net.ParseIP(ipRand)
		if ip == nil {
			t.Error("should not nil")
		}
	}
}

func TestValidateEmails(t *testing.T) {
	err1 := ValidateEmails("f", "a")
	if err1 == nil {
		t.Error("should err not nil")
	}

	err2 := ValidateEmails("validemail0987@gmail.com")
	if err2 != nil {
		t.Error("should err not nil")
	}

	err3 := ValidateEmails("validemail0987@gmail.com", "invalidemail0987@.gmail.com")
	if err3 == nil {
		t.Error("should err not nil")
	}

	err4 := ValidateEmails("validemail0987@gmail.com", "validemail0987@gmail.com", "invalidemail0987@gmail.com.")
	if err4 == nil {
		t.Error("should err not nil")
	}
}
