package simplefeatures_test

import (
	"testing"

	. "github.com/peterstace/simplefeatures"
)

func TestValuer(t *testing.T) {
	any := AnyGeometry{geomFromWKT(t, "POINT(1 2)")}
	val, err := any.Value()
	if err != nil {
		t.Fatal(err)
	}
	valGeom := geomFromWKT(t, val.(string)).(Point)
	expectDeepEqual(t, valGeom.XY().X.AsFloat(), 1.0)
	expectDeepEqual(t, valGeom.XY().Y.AsFloat(), 2.0)
}

func TestValuerZero(t *testing.T) {
	var any AnyGeometry
	if _, err := any.Value(); err == nil {
		t.Fatal("expected an error")
	}
}

func TestScanner(t *testing.T) {
	const wkt = "POINT(2 3)"
	var any AnyGeometry
	check := func(t *testing.T, err error) {
		if err != nil {
			t.Fatal(err)
		}
		got := any.Geom.(Point).XY()
		expectDeepEqual(t, got.X.AsFloat(), 2.0)
		expectDeepEqual(t, got.Y.AsFloat(), 3.0)
	}

	t.Run("string", func(t *testing.T) {
		any = AnyGeometry{}
		check(t, any.Scan(string(wkt)))
	})
	t.Run("byte", func(t *testing.T) {
		any = AnyGeometry{}
		check(t, any.Scan([]byte(wkt)))
	})
}
