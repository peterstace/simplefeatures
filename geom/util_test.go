package geom_test

import (
	"strings"
	"testing"

	. "github.com/peterstace/simplefeatures/geom"
)

func geomFromWKT(t *testing.T, wkt string) Geometry {
	t.Helper()
	geom, err := UnmarshalWKT(strings.NewReader(wkt))
	if err != nil {
		t.Fatalf("could not unmarshal WKT:\n  wkt: %s\n  err: %v", wkt, err)
	}
	return geom
}

func xyCoords(x, y float64) Coordinates {
	return Coordinates{XY: XY{x, y}, Type: DimXY}
}

func expectPanics(t *testing.T, fn func()) {
	t.Helper()
	defer func() {
		if r := recover(); r != nil {
			return
		}
		t.Errorf("didn't panic")
	}()
	fn()
}

func expectNoErr(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func expectGeomEq(t *testing.T, got, want Geometry, opts ...EqualsExactOption) {
	t.Helper()
	if !got.EqualsExact(want, opts...) {
		t.Errorf("\ngot:  %v\nwant: %v\n", got.AsText(), want.AsText())
	}
}

func expectCoordsEq(t *testing.T, got, want Coordinates) {
	t.Helper()
	if got != want {
		t.Errorf("\ngot:  %v\nwant: %v\n", got, want)
	}
}

func expectStringEq(t *testing.T, got, want string) {
	t.Helper()
	if got != want {
		t.Errorf("\ngot:  %q\nwant: %q\n", got, want)
	}
}

func expectIntEq(t *testing.T, got, want int) {
	t.Helper()
	if got != want {
		t.Errorf("\ngot:  %d\nwant: %d\n", got, want)
	}
}

func expectBoolEq(t *testing.T, got, want bool) {
	t.Helper()
	if got != want {
		t.Errorf("\ngot:  %t\nwant: %t\n", got, want)
	}
}

func expectXYEq(t *testing.T, got, want XY) {
	t.Helper()
	if got != want {
		t.Errorf("\ngot:  %v\nwant: %v\n", got, want)
	}
}

func expectCoordinatesTypeEq(t *testing.T, got, want CoordinatesType) {
	t.Helper()
	if got != want {
		t.Errorf("\ngot:  %v\nwant: %v\n", got, want)
	}
}
