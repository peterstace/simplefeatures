package geom_test

import (
	"strings"
	"testing"

	. "github.com/peterstace/simplefeatures/geom"
)

func GeomFromWKT(t *testing.T, wkt string) Geometry {
	t.Helper()
	geom, err := UnmarshalWKT(strings.NewReader(wkt))
	if err != nil {
		t.Fatalf("could not unmarshal WKT:\n  wkt: %s\n  err: %v", wkt, err)
	}
	return geom
}

func ExpectPanics(t *testing.T, fn func()) {
	t.Helper()
	defer func() {
		if r := recover(); r != nil {
			return
		}
		t.Errorf("didn't panic")
	}()
	fn()
}

func ExpectNoErr(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func ExpectGeomEq(t *testing.T, got, want Geometry, opts ...EqualsExactOption) {
	t.Helper()
	if !got.EqualsExact(want, opts...) {
		t.Errorf("\ngot:  %v\nwant: %v\n", got.AsText(), want.AsText())
	}
}

func ExpectCoordsEq(t *testing.T, got, want Coordinates) {
	t.Helper()
	if got != want {
		t.Errorf("\ngot:  %v\nwant: %v\n", got, want)
	}
}

func ExpectStringEq(t *testing.T, got, want string) {
	t.Helper()
	if got != want {
		t.Errorf("\ngot:  %q\nwant: %q\n", got, want)
	}
}

func ExpectIntEq(t *testing.T, got, want int) {
	t.Helper()
	if got != want {
		t.Errorf("\ngot:  %d\nwant: %d\n", got, want)
	}
}

func ExpectBoolEq(t *testing.T, got, want bool) {
	t.Helper()
	if got != want {
		t.Errorf("\ngot:  %t\nwant: %t\n", got, want)
	}
}

func ExpectXYEq(t *testing.T, got, want XY) {
	t.Helper()
	if got != want {
		t.Errorf("\ngot:  %v\nwant: %v\n", got, want)
	}
}
