package geos_test

import (
	"reflect"
	"testing"
)

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

func expectDeepEq(t testing.TB, want, got interface{}) {
	t.Helper()
	if !reflect.DeepEqual(want, got) {
		t.Logf("want: %v", want)
		t.Logf("got:  %v", got)
		t.Fatal("expected equal")
	}
}
