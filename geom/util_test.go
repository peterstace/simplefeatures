package geom_test

import (
	"bytes"
	"math"
	"testing"

	. "github.com/peterstace/simplefeatures/geom"
)

func geomFromWKT(t *testing.T, wkt string) Geometry {
	t.Helper()
	geom, err := UnmarshalWKT(wkt)
	if err != nil {
		t.Fatalf("could not unmarshal WKT:\n  wkt: %s\n  err: %v", wkt, err)
	}
	return geom
}

func geomsFromWKTs(t *testing.T, wkts []string) []Geometry {
	t.Helper()
	var gs []Geometry
	for _, wkt := range wkts {
		g, err := UnmarshalWKT(wkt)
		if err != nil {
			t.Fatalf("could not unmarshal WKT:\n  wkt: %s\n  err: %v", wkt, err)
		}
		gs = append(gs, g)
	}
	return gs
}

func xyCoords(x, y float64) Coordinates {
	return Coordinates{XY: XY{x, y}, Type: DimXY}
}

func upcastPoints(ps []Point) []Geometry {
	gs := make([]Geometry, len(ps))
	for i, p := range ps {
		gs[i] = p.AsGeometry()
	}
	return gs
}

func upcastLineStrings(lss []LineString) []Geometry {
	gs := make([]Geometry, len(lss))
	for i, ls := range lss {
		gs[i] = ls.AsGeometry()
	}
	return gs
}

func upcastPolygons(ps []Polygon) []Geometry {
	gs := make([]Geometry, len(ps))
	for i, p := range ps {
		gs[i] = p.AsGeometry()
	}
	return gs
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

func expectErr(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("expected error but got nil")
	}
}

func expectGeomEq(t *testing.T, got, want Geometry, opts ...ExactEqualsOption) {
	t.Helper()
	if !ExactEquals(got, want, opts...) {
		t.Errorf("\ngot:  %v\nwant: %v\n", got.AsText(), want.AsText())
	}
}

func expectGeomsEq(t *testing.T, got, want []Geometry, opts ...ExactEqualsOption) {
	t.Helper()
	if len(got) != len(want) {
		t.Errorf("\ngot:  len %d\nwant: len %d\n", len(got), len(want))
	}
	for i := range got {
		if !ExactEquals(got[i], want[i], opts...) {
			t.Errorf("\ngot:  %v\nwant: %v\n", got[i].AsText(), want[i].AsText())
		}
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

func expectTrue(t *testing.T, got bool) {
	t.Helper()
	expectBoolEq(t, got, true)
}

func expectFalse(t *testing.T, got bool) {
	t.Helper()
	expectBoolEq(t, got, false)
}

func expectXYEq(t *testing.T, got, want XY) {
	t.Helper()
	if got != want {
		t.Errorf("\ngot:  %v\nwant: %v\n", got, want)
	}
}
func expectXYWithinTolerance(t *testing.T, got, want XY, tolerance float64) {
	t.Helper()
	if delta := math.Abs(got.Sub(want).Length()); delta > tolerance {
		t.Errorf("\ngot:  %v\nwant: %v\n", got, want)
	}
}

func expectCoordinatesTypeEq(t *testing.T, got, want CoordinatesType) {
	t.Helper()
	if got != want {
		t.Errorf("\ngot:  %v\nwant: %v\n", got, want)
	}
}

func expectBytesEq(t *testing.T, got, want []byte) {
	t.Helper()
	if !bytes.Equal(got, want) {
		t.Errorf("\ngot:  %v\nwant: %v\n", got, want)
	}
}
