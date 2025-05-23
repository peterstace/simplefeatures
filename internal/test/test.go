// Package test provides test helpers.
package test

import "testing"

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
