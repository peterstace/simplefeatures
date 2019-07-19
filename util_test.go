package simplefeatures_test

import (
	"reflect"
	"strings"
	"testing"

	. "github.com/peterstace/simplefeatures"
)

func geomFromWKT(t *testing.T, wkt string) Geometry {
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
		wantGeom, okWant := want.(Geometry)
		if okGot && okWant {
			format += "    got  (WKT): %s\n"
			format += "    want (WKT): %s\n"
			args = append(args, gotGeom.AsText(), wantGeom.AsText())
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
