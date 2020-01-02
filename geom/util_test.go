package geom_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/peterstace/simplefeatures/geom"
	. "github.com/peterstace/simplefeatures/geom"
)

func must(t *testing.T) func(geom.GeometryX, error) geom.GeometryX {
	return func(g geom.GeometryX, err error) geom.GeometryX {
		if err != nil {
			t.Fatalf("must have no error but got: %v", err)
		}
		return g
	}
}

func geomFromWKT(t *testing.T, wkt string) GeometryX {
	t.Helper()
	geom, err := UnmarshalWKT(strings.NewReader(wkt))
	if err != nil {
		t.Fatalf("could not unmarshal WKT:\n  wkt: %s\n  err: %v", wkt, err)
	}
	return geom.AsGeometryX()
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
		gotGeom, okGot := got.(GeometryX)
		if okGot {
			format += "    got  (WKT): %s\n"
			args = append(args, ToGeometry(gotGeom).AsText())
		}
		wantGeom, okWant := want.(GeometryX)
		if okWant {
			format += "    want (WKT): %s\n"
			args = append(args, ToGeometry(wantGeom).AsText())
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
