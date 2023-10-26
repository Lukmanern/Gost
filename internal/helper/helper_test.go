package helper

import (
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
			t.Error("len of emails should equal")
		}
		for _, email := range emails {
			if len(email) < 33 {
				t.Error("len of an email should not less than 33")
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
