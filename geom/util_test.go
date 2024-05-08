package geom_test

import (
	"bytes"
	"math"
	"reflect"
	"strconv"
	"testing"

	"github.com/peterstace/simplefeatures/geom"
)

//nolint:unparam
func geomFromWKT(tb testing.TB, wkt string, nv ...geom.NoValidate) geom.Geometry {
	tb.Helper()
	geom, err := geom.UnmarshalWKT(wkt, nv...)
	if err != nil {
		tb.Fatalf("could not unmarshal WKT:\n  wkt: %s\n  err: %v", wkt, err)
	}
	return geom
}

//nolint:unparam
func geomsFromWKTs(tb testing.TB, wkts []string, nv ...geom.NoValidate) []geom.Geometry {
	tb.Helper()
	var gs []geom.Geometry
	for _, wkt := range wkts {
		g, err := geom.UnmarshalWKT(wkt, nv...)
		if err != nil {
			tb.Fatalf("could not unmarshal WKT:\n  wkt: %s\n  err: %v", wkt, err)
		}
		gs = append(gs, g)
	}
	return gs
}

func geomFromGeoJSON(tb testing.TB, geojson string, nv ...geom.NoValidate) geom.Geometry {
	tb.Helper()
	g, err := geom.UnmarshalGeoJSON([]byte(geojson), nv...)
	if err != nil {
		tb.Fatalf("could not unmarshal GeoJSON:\n geojson: %s\n     err: %v", geojson, err)
	}
	return g
}

func xyCoords(x, y float64) geom.Coordinates {
	return geom.Coordinates{XY: geom.XY{x, y}, Type: geom.DimXY}
}

func upcastPoints(ps []geom.Point) []geom.Geometry {
	gs := make([]geom.Geometry, len(ps))
	for i, p := range ps {
		gs[i] = p.AsGeometry()
	}
	return gs
}

func upcastLineStrings(lss []geom.LineString) []geom.Geometry {
	gs := make([]geom.Geometry, len(lss))
	for i, ls := range lss {
		gs[i] = ls.AsGeometry()
	}
	return gs
}

func upcastPolygons(ps []geom.Polygon) []geom.Geometry {
	gs := make([]geom.Geometry, len(ps))
	for i, p := range ps {
		gs[i] = p.AsGeometry()
	}
	return gs
}

func expectPanics(tb testing.TB, fn func()) {
	tb.Helper()
	defer func() {
		if r := recover(); r != nil {
			return
		}
		tb.Errorf("didn't panic")
	}()
	fn()
}

func expectNoErr(tb testing.TB, err error) {
	tb.Helper()
	if err != nil {
		tb.Fatalf("unexpected error: %v", err)
	}
}

func expectErr(tb testing.TB, err error) {
	tb.Helper()
	if err == nil {
		tb.Fatal("expected error but got nil")
	}
}

func expectDeepEq(tb testing.TB, got, want interface{}) {
	tb.Helper()
	if !reflect.DeepEqual(got, want) {
		tb.Errorf("\ngot:  %v\nwant: %v\n", got, want)
	}
}

func expectGeomEq(tb testing.TB, got, want geom.Geometry, opts ...geom.ExactEqualsOption) {
	tb.Helper()
	if !geom.ExactEquals(got, want, opts...) {
		tb.Errorf("\ngot:  %v\nwant: %v\n", got.AsText(), want.AsText())
	}
}

//nolint:unused
func expectGeomApproxEq(tb testing.TB, got, want geom.Geometry) {
	tb.Helper()
	eq, err := geom.Equals(got, want)
	if err != nil {
		tb.Errorf("\ngot:  %v\nwant: %v\nerr: %v\n", got.AsText(), want.AsText(), err)
	}
	if !eq {
		tb.Errorf("\ngot:  %v\nwant: %v\n", got.AsText(), want.AsText())
	}
}

func expectGeomEqWKT(tb testing.TB, got geom.Geometry, wantWKT string, opts ...geom.ExactEqualsOption) {
	tb.Helper()
	want := geomFromWKT(tb, wantWKT)
	expectGeomEq(tb, got, want, opts...)
}

//nolint:unparam
func expectGeomsEq(tb testing.TB, got, want []geom.Geometry, opts ...geom.ExactEqualsOption) {
	tb.Helper()
	if len(got) != len(want) {
		tb.Errorf("\ngot:  len %d\nwant: len %d\n", len(got), len(want))
	}
	for i := range got {
		if !geom.ExactEquals(got[i], want[i], opts...) {
			tb.Errorf("\ngot:  %v\nwant: %v\n", got[i].AsText(), want[i].AsText())
		}
	}
}

func expectCoordsEq(tb testing.TB, got, want geom.Coordinates) {
	tb.Helper()
	if got != want {
		tb.Errorf("\ngot:  %v\nwant: %v\n", got, want)
	}
}

func expectStringEq(tb testing.TB, got, want string) {
	tb.Helper()
	if got != want {
		tb.Errorf("\ngot:  %s\nwant: %s\n", quotedString(got), quotedString(want))
	}
}

func quotedString(s string) string {
	if strconv.CanBackquote(s) {
		return "`" + s + "`"
	}
	return strconv.Quote(s)
}

func expectIntEq(tb testing.TB, got, want int) {
	tb.Helper()
	if got != want {
		tb.Errorf("\ngot:  %d\nwant: %d\n", got, want)
	}
}

func expectBoolEq(tb testing.TB, got, want bool) {
	tb.Helper()
	if got != want {
		tb.Errorf("\ngot:  %t\nwant: %t\n", got, want)
	}
}

func expectTrue(tb testing.TB, got bool) {
	tb.Helper()
	expectBoolEq(tb, got, true)
}

func expectFalse(tb testing.TB, got bool) {
	tb.Helper()
	expectBoolEq(tb, got, false)
}

func expectXYEq(tb testing.TB, got, want geom.XY) {
	tb.Helper()
	if got != want {
		tb.Errorf("\ngot:  %v\nwant: %v\n", got, want)
	}
}

func expectXYWithinTolerance(tb testing.TB, got, want geom.XY, tolerance float64) {
	tb.Helper()
	if delta := math.Abs(got.Sub(want).Length()); delta > tolerance {
		tb.Errorf("\ngot:  %v\nwant: %v\n", got, want)
	}
}

func expectCoordinatesTypeEq(tb testing.TB, got, want geom.CoordinatesType) {
	tb.Helper()
	if got != want {
		tb.Errorf("\ngot:  %v\nwant: %v\n", got, want)
	}
}

func expectBytesEq(tb testing.TB, got, want []byte) {
	tb.Helper()
	if !bytes.Equal(got, want) {
		tb.Errorf("\ngot:  %v\nwant: %v\n", got, want)
	}
}

func expectFloat64Eq(tb testing.TB, got, want float64) {
	tb.Helper()
	if got != want {
		tb.Errorf("\ngot:  %v\nwant: %v\n", got, want)
	}
}

func expectEnvEq(tb testing.TB, got, want geom.Envelope) {
	tb.Helper()
	if geom.ExactEquals(got.Min().AsGeometry(), want.Min().AsGeometry()) &&
		geom.ExactEquals(got.Max().AsGeometry(), want.Max().AsGeometry()) {
		return
	}
	tb.Errorf("\ngot:  %v\nwant: %v", got, want)
}

func expectSequenceEq(tb testing.TB, got, want geom.Sequence) {
	tb.Helper()
	show := func() {
		tb.Logf("len(got): %d, ct(got): %s", got.Length(), got.CoordinatesType())
		for i := 0; i < got.Length(); i++ {
			tb.Logf("got[%d]: %v", i, got.Get(i))
		}
		tb.Logf("len(want): %d, ct(want): %s", want.Length(), want.CoordinatesType())
		for i := 0; i < want.Length(); i++ {
			tb.Logf("want[%d]: %v", i, want.Get(i))
		}
	}
	if got.CoordinatesType() != want.CoordinatesType() {
		tb.Errorf("mismatched coordinate type")
		show()
		return
	}
	if got.Length() != want.Length() {
		tb.Errorf("length mismatch")
		show()
		return
	}
	for i := 0; i < got.Length(); i++ {
		w := want.Get(i)
		g := got.Get(i)
		if g != w {
			tb.Errorf("mismatch at %d: got:%v want:%v", i, g, w)
		}
	}
}
