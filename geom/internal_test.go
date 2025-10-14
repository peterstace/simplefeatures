package geom

import (
	"testing"
)

// These test helpers should only be used by internal tests in the geom
// package. They exist in order to avoid import cycles (otherwise, the
// internal/test package would be used instead).

func wktToGeom(tb testing.TB, wkt string) Geometry {
	tb.Helper()
	g, err := UnmarshalWKT(wkt)
	testNoErr(tb, err)
	return g
}

func testExactEqualsWKT(tb testing.TB, g Geometry, wkt string, opts ...ExactEqualsOption) {
	tb.Helper()
	want := wktToGeom(tb, wkt)
	testExactEquals(tb, g, want, opts...)
}

func testExactEquals(tb testing.TB, g1, g2 Geometry, opts ...ExactEqualsOption) {
	tb.Helper()
	if !ExactEquals(g1, g2, opts...) {
		tb.Fatalf("geometries should be exactly equal:\n  g1: %v\n  g2: %v", g1.AsText(), g2.AsText())
	}
}

func testNoErr(tb testing.TB, err error) {
	tb.Helper()
	if err != nil {
		tb.Fatalf("unexpected error: %v", err)
	}
}
