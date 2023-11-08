package main

import (
	"testing"
	"time"
)

func TestMain(t *testing.T) {
	defer func() {
		r := recover()
		if r != nil {
			t.Error("should not panic", r)
		}
	}()

	go main()
	time.Sleep(3 * time.Second)
}
