// Package test provides test helpers.
package test

import (
	"errors"
	"testing"
)

func NoErr(tb testing.TB, err error) {
	tb.Helper()
	if err != nil {
		tb.Fatalf("unexpected error: %v", err)
	}
}

func Err(tb testing.TB, err error) {
	tb.Helper()
	if err == nil {
		tb.Fatal("expected error but got nil")
	}
}

func ErrAs(tb testing.TB, err error, target interface{}) {
	tb.Helper()
	if !errors.As(err, target) {
		tb.Fatalf("expected error '%v' to be 'As' of type %T", err, target)
	}
}
