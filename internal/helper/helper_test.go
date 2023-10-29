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
