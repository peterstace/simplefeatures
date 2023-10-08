package check

import (
	"cmp"
	"testing"

	"github.com/peterstace/simplefeatures/geom"
)

func GeomFromWKT(t testing.TB, wkt string, nv ...geom.NoValidate) geom.Geometry {
	t.Helper()
	geom, err := geom.UnmarshalWKT(wkt, nv...)
	if err != nil {
		t.Fatalf("could not unmarshal WKT:\n  wkt: %s\n  err: %v", wkt, err)
	}
	return geom
}

func NoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func True(t *testing.T, b bool) {
	t.Helper()
	if !b {
		t.Errorf("\ngot:  false\nwant: true\n")
	}
}

func Eq[V comparable](t *testing.T, got, want V) {
	t.Helper()
	if got != want {
		t.Errorf("\ngot:  %v\nwant: %v\n", got, want)
	}
}

func GE[V cmp.Ordered](t *testing.T, lhs, rhs V) {
	t.Helper()
	if !(lhs >= rhs) {
		t.Errorf("\ngot:  %v < %v\nwant: %v >= %v\n", lhs, rhs)
	}
}

func GT[V cmp.Ordered](t *testing.T, lhs, rhs V) {
	t.Helper()
	if !(lhs > rhs) {
		t.Errorf("\ngot:  %v <= %v\nwant: %v > %v\n", lhs, rhs)
	}
}

func LE[V cmp.Ordered](t *testing.T, lhs, rhs V) {
	t.Helper()
	if !(lhs <= rhs) {
		t.Errorf("\ngot:  %v > %v\nwant: %v <= %v\n", lhs, rhs)
	}
}

func LT[V cmp.Ordered](t *testing.T, lhs, rhs V) {
	t.Helper()
	if !(lhs < rhs) {
		t.Errorf("\ngot:  %v >= %v\nwant: %v < %v\n", lhs, rhs)
	}
}

func GeomEq(t testing.TB, got, want geom.Geometry, opts ...geom.ExactEqualsOption) {
	t.Helper()
	if !geom.ExactEquals(got, want, opts...) {
		t.Errorf("\ngot:  %v\nwant: %v\n", got.AsText(), want.AsText())
	}
}
