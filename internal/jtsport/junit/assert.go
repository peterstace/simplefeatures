// Package junit provides JUnit-style assertion helpers for ported tests.
// These helpers enable 1-1 line mapping between Java JUnit tests and Go tests.
package junit

import (
	"reflect"
	"testing"
)

// AssertEquals checks that expected equals actual.
func AssertEquals[T comparable](t *testing.T, expected, actual T) {
	t.Helper()
	if expected != actual {
		t.Errorf("expected %v but was %v", expected, actual)
	}
}

// AssertEqualsDeep checks that expected equals actual using reflect.DeepEqual.
// Use this for comparing struct values through pointers.
func AssertEqualsDeep(t *testing.T, expected, actual any) {
	t.Helper()
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("expected %v but was %v", expected, actual)
	}
}

// AssertEqualsNaN checks that actual is NaN when expected is NaN.
// For JUnit assertEquals compatibility with NaN values.
func AssertEqualsNaN(t *testing.T, expected, actual float64) {
	t.Helper()
	expectedIsNaN := expected != expected // NaN != NaN
	actualIsNaN := actual != actual
	if expectedIsNaN && actualIsNaN {
		return // Both NaN, equals
	}
	if expectedIsNaN != actualIsNaN || expected != actual {
		t.Errorf("expected %v but was %v", expected, actual)
	}
}

// AssertEqualsFloat64 checks that expected equals actual within the given tolerance.
func AssertEqualsFloat64(t *testing.T, expected, actual, tolerance float64) {
	t.Helper()
	diff := expected - actual
	if diff < 0 {
		diff = -diff
	}
	if diff > tolerance {
		t.Errorf("expected %v but was %v (tolerance %v)", expected, actual, tolerance)
	}
}

// AssertTrue checks that the condition is true.
func AssertTrue(t *testing.T, condition bool) {
	t.Helper()
	if !condition {
		t.Error("expected true but was false")
	}
}

// AssertFalse checks that the condition is false.
func AssertFalse(t *testing.T, condition bool) {
	t.Helper()
	if condition {
		t.Error("expected false but was true")
	}
}

// AssertNull checks that the value is nil.
// Uses reflection to properly handle typed nil pointers.
func AssertNull(t *testing.T, value any) {
	t.Helper()
	if value == nil {
		return
	}
	rv := reflect.ValueOf(value)
	if rv.Kind() == reflect.Ptr && rv.IsNil() {
		return
	}
	t.Errorf("expected nil but was %v", value)
}

// AssertNotNull checks that the value is not nil.
// Uses reflection to properly handle typed nil pointers.
func AssertNotNull(t *testing.T, value any) {
	t.Helper()
	if value == nil {
		t.Error("expected non-nil but was nil")
		return
	}
	rv := reflect.ValueOf(value)
	if rv.Kind() == reflect.Ptr && rv.IsNil() {
		t.Error("expected non-nil but was nil")
	}
}

// Fail fails the test with the given message.
func Fail(t *testing.T, message string) {
	t.Helper()
	t.Error(message)
}
