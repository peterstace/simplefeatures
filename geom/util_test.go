package geom_test

import (
	"bytes"
	"math"
	"testing"

	. "github.com/peterstace/simplefeatures/geom"
)

func geomFromWKT(t testing.TB, wkt string) Geometry {
	t.Helper()
	geom, err := UnmarshalWKT(wkt)
	if err != nil {
		t.Fatalf("could not unmarshal WKT:\n  wkt: %s\n  err: %v", wkt, err)
	}
	return geom
}

func geomsFromWKTs(t testing.TB, wkts []string) []Geometry {
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

func expectPanics(t testing.TB, fn func()) {
	t.Helper()
	defer func() {
		if r := recover(); r != nil {
			return
		}
		t.Errorf("didn't panic")
	}()
	fn()
}

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

func expectGeomEq(t testing.TB, got, want Geometry, opts ...ExactEqualsOption) {
	t.Helper()
	if !ExactEquals(got, want, opts...) {
		t.Errorf("\ngot:  %v\nwant: %v\n", got.AsText(), want.AsText())
	}
}

func expectGeomEqWKT(t testing.TB, got Geometry, wantWKT string, opts ...ExactEqualsOption) {
	t.Helper()
	want := geomFromWKT(t, wantWKT)
	expectGeomEq(t, got, want, opts...)
}

func expectGeomsEq(t testing.TB, got, want []Geometry, opts ...ExactEqualsOption) {
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

func expectCoordsEq(t testing.TB, got, want Coordinates) {
	t.Helper()
	if got != want {
		t.Errorf("\ngot:  %v\nwant: %v\n", got, want)
	}
}

func expectStringEq(t testing.TB, got, want string) {
	t.Helper()
	if got != want {
		t.Errorf("\ngot:  %q\nwant: %q\n", got, want)
	}
}

func expectIntEq(t testing.TB, got, want int) {
	t.Helper()
	if got != want {
		t.Errorf("\ngot:  %d\nwant: %d\n", got, want)
	}
}

func expectBoolEq(t testing.TB, got, want bool) {
	t.Helper()
	if got != want {
		t.Errorf("\ngot:  %t\nwant: %t\n", got, want)
	}
}

func expectTrue(t testing.TB, got bool) {
	t.Helper()
	expectBoolEq(t, got, true)
}

//nolint:deadcode,unused
func expectFalse(t testing.TB, got bool) {
	t.Helper()
	expectBoolEq(t, got, false)
}

func expectXYEq(t testing.TB, got, want XY) {
	t.Helper()
	if got != want {
		t.Errorf("\ngot:  %v\nwant: %v\n", got, want)
	}
}
func expectXYWithinTolerance(t testing.TB, got, want XY, tolerance float64) {
	t.Helper()
	if delta := math.Abs(got.Sub(want).Length()); delta > tolerance {
		t.Errorf("\ngot:  %v\nwant: %v\n", got, want)
	}
}

func expectCoordinatesTypeEq(t testing.TB, got, want CoordinatesType) {
	t.Helper()
	if got != want {
		t.Errorf("\ngot:  %v\nwant: %v\n", got, want)
	}
}

func expectBytesEq(t testing.TB, got, want []byte) {
	t.Helper()
	if !bytes.Equal(got, want) {
		t.Errorf("\ngot:  %v\nwant: %v\n", got, want)
	}
}

func expectFloat64Eq(t testing.TB, got, want float64) {
	t.Helper()
	if got != want {
		t.Errorf("\ngot:  %v\nwant: %v\n", got, want)
	}
}

func expectEnvEq(t testing.TB, got, want Envelope) {
	t.Helper()
	if ExactEquals(got.Min().AsGeometry(), want.Min().AsGeometry()) &&
		ExactEquals(got.Max().AsGeometry(), want.Max().AsGeometry()) {
		return
	}
	t.Errorf(
		"\ngot:  %v\nwant: %v\n",
		got.AsGeometry().AsText(),
		want.AsGeometry().AsText(),
	)
}

func expectSequenceEq(t testing.TB, got, want Sequence) {
	t.Helper()
	show := func() {
		t.Logf("len(got): %d, ct(got): %s", got.Length(), got.CoordinatesType())
		for i := 0; i < got.Length(); i++ {
			t.Logf("got[%d]: %v", i, got.Get(i))
		}
		t.Logf("len(want): %d, ct(want): %s", want.Length(), want.CoordinatesType())
		for i := 0; i < want.Length(); i++ {
			t.Logf("want[%d]: %v", i, want.Get(i))
		}
	}
	if got.CoordinatesType() != want.CoordinatesType() {
		t.Errorf("mismatched coordinate type")
		show()
		return
	}
	if got.Length() != want.Length() {
		t.Errorf("length mismatch")
		show()
		return
	}
	for i := 0; i < got.Length(); i++ {
		w := want.Get(i)
		g := got.Get(i)
		if g != w {
			t.Errorf("mismatch at %d: got:%v want:%v", i, g, w)
		}
	}
}
