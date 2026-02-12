// Package test provides test helpers.
package test

import (
	"errors"
	"math"
	"reflect"
	"testing"

	"github.com/peterstace/simplefeatures/geom"
)

func FromWKT(tb testing.TB, wkt string) geom.Geometry {
	tb.Helper()
	g, err := geom.UnmarshalWKT(wkt)
	NoErr(tb, err)
	return g
}

func Eq[T comparable](tb testing.TB, got, want T) {
	tb.Helper()
	if got != want {
		tb.Fatalf("got:  %v\nwant: %v", got, want)
	}
}

func True(tb testing.TB, cond bool) {
	tb.Helper()
	if !cond {
		tb.Fatal("condition is false, want true")
	}
}

func False(tb testing.TB, cond bool) {
	tb.Helper()
	if cond {
		tb.Fatal("condition is true, want false")
	}
}

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

func ErrAs(tb testing.TB, err error, target any) {
	tb.Helper()
	if !errors.As(err, target) {
		tb.Fatalf("expected error '%v' to be 'As' of type %T", err, target)
	}
}

func ExactEquals(tb testing.TB, got, want geom.Geometry, opts ...geom.ExactEqualsOption) {
	tb.Helper()
	if !geom.ExactEquals(got, want, opts...) {
		tb.Fatalf("geometries should be exactly equal:\n got: %v\nwant: %v", got.AsText(), want.AsText())
	}
}

func NotExactEquals(tb testing.TB, got, doNotWant geom.Geometry, opts ...geom.ExactEqualsOption) {
	tb.Helper()
	if geom.ExactEquals(got, doNotWant, opts...) {
		tb.Fatalf("geometries should not be exactly equal:\n      got: %v\ndoNotWant: %v", got.AsText(), doNotWant.AsText())
	}
}

func ExactEqualsWKT(tb testing.TB, got geom.Geometry, wantWKT string, opts ...geom.ExactEqualsOption) {
	tb.Helper()
	want := FromWKT(tb, wantWKT)
	ExactEquals(tb, got, want, opts...)
}

func DeepEqual(tb testing.TB, a, b any) {
	tb.Helper()
	if !reflect.DeepEqual(a, b) {
		tb.Fatalf("values should be deeply equal:\n  a: %#v\n  b: %#v", a, b)
	}
}

func NotDeepEqual(tb testing.TB, a, b any) {
	tb.Helper()
	if reflect.DeepEqual(a, b) {
		tb.Fatalf("values should not be deeply equal:\n  a: %#v\n  b: %#v", a, b)
	}
}

// Tolerance specifies tolerances for approximate float comparison.
type Tolerance struct {
	Rel float64 // Relative tolerance: diff must be <= Rel * max(|got|, |want|).
	Abs float64 // Absolute tolerance: diff must be <= Abs.
}

// ApproxEqual asserts that two float64 values are approximately equal. The
// comparison passes if the difference is within either the relative tolerance
// or the absolute tolerance.
func ApproxEqual(tb testing.TB, got, want float64, tol Tolerance) {
	tb.Helper()
	diff := math.Abs(got - want)
	maxVal := math.Max(math.Abs(got), math.Abs(want))
	if diff > tol.Rel*maxVal && diff > tol.Abs {
		tb.Fatalf("values not approximately equal (tol=%+v):\n  got:  %v\n  want: %v\n  diff: %v", tol, got, want, diff)
	}
}
