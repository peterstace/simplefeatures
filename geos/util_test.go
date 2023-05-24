package geos_test

import "testing"

func expectNoErr(t testing.TB, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func expectErr(t testing.TB, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("expected error but got nil")
	}
}
