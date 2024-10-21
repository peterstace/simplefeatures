package carto_test

import (
	"testing"
)

func expectNoErr(tb testing.TB, err error) {
	tb.Helper()
	if err != nil {
		tb.Fatalf("unexpected error: %v", err)
	}
}

func expectTrue(tb testing.TB, b bool) {
	tb.Helper()
	if !b {
		tb.Fatalf("expected true, got false")
	}
}
