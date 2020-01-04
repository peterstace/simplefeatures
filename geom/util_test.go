package geom_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/peterstace/simplefeatures/geom"
	. "github.com/peterstace/simplefeatures/geom"
)

func must(t *testing.T) func(geom.Geometry, error) geom.Geometry {
	return func(g geom.Geometry, err error) geom.Geometry {
		if err != nil {
			t.Fatalf("must have no error but got: %v", err)
		}
		return g
	}
}

func gFromWKT(t *testing.T, wkt string) Geometry {
	t.Helper()
	geom, err := UnmarshalWKT(strings.NewReader(wkt))
	if err != nil {
		t.Fatalf("could not unmarshal WKT:\n  wkt: %s\n  err: %v", wkt, err)
	}
	return geom
}

func expectDeepEqual(t *testing.T, got, want interface{}) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		args := []interface{}{got, want}
		format := `
expected to be equal, but aren't:
	got:  %+v
	want: %+v
`
		// Special cases for geometries:
		gotGeom, okGot := got.(Geometry)
		if okGot {
			format += "    got  (WKT): %s\n"
			args = append(args, gotGeom.AsText())
		}
		wantGeom, okWant := want.(Geometry)
		if okWant {
			format += "    want (WKT): %s\n"
			args = append(args, wantGeom.AsText())
		}
		t.Errorf(format, args...)
	}
}

func expectPanics(t *testing.T, fn func()) {
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
