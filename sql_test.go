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
	expectDeepEqual(t, any.Geom, geomFromWKT(t, val.(string)))
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
		expectDeepEqual(t, any.Geom, geomFromWKT(t, wkt))
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
