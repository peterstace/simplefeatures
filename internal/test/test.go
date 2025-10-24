// Package test provides test helpers.
package test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/peterstace/simplefeatures/geom"
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

func ErrAs(tb testing.TB, err error, target any) {
	tb.Helper()
	if !errors.As(err, target) {
		tb.Fatalf("expected error '%v' to be 'As' of type %T", err, target)
	}
}

func ExactEquals(tb testing.TB, g1, g2 geom.Geometry, opts ...geom.ExactEqualsOption) {
	tb.Helper()
	if !geom.ExactEquals(g1, g2, opts...) {
		tb.Fatalf("geometries should be exactly equal:\n  g1: %v\n  g2: %v", g1.AsText(), g2.AsText())
	}
}

func NotExactEquals(tb testing.TB, g1, g2 geom.Geometry, opts ...geom.ExactEqualsOption) {
	tb.Helper()
	if geom.ExactEquals(g1, g2, opts...) {
		tb.Fatalf("geometries should not be exactly equal:\n  g1: %v\n  g2: %v", g1.AsText(), g2.AsText())
	}
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
