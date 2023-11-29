package rawgeos

import (
	"errors"
	"testing"
)

func TestWrapNil(t *testing.T) {
	if wrap(nil, "format") != nil {
		t.Fatalf("Expected nil but got error")
	}
}

func TestWrapNonNilNoArgs(t *testing.T) {
	original := errors.New("original error")
	got := wrap(original, "context").Error()
	const want = "context: original error"
	if got != want {
		t.Fatalf("got: %v want: %v", got, want)
	}
}

func TestWrapNonNilWithArgs(t *testing.T) {
	original := errors.New("original error")
	got := wrap(original, "context foo=%v bar=%v", "baz", 42).Error()
	const want = "context foo=baz bar=42: original error"
	if got != want {
		t.Fatalf("got: %v want: %v", got, want)
	}
}
