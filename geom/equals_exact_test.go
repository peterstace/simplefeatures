package geom_test

import (
	"testing"

	"github.com/peterstace/simplefeatures/geom"
)

func TestEqualsExact(t *testing.T) {
	// Always equal to itself.
	// Always not-equal to another type.
	// Sometimes equal to others.

	// List a bunch of different geometries. Indicate which ones should be equal to which other ones?

	wkts := map[string]string{
		"pt_a": "POINT(2 3)",
		"pt_b": "POINT(3 -1)",
		"pt_c": "POINT(2.09 2.91)",
		"pt_d": "POINT(2.08 2.92)",
		"pt_e": "POINT EMPTY",
		"pt_f": "POINT(3.125 -1)",

		"ln_a": "LINESTRING(1 2,3 4)",
		"ln_b": "LINESTRING(1 2,3 3.9)",
		"ln_c": "LINESTRING(1.1 2,3 4)",
		"ln_d": "LINESTRING(3 4,1 2)",

		// TODO: LineString
		// TODO: LinearRing
		// TODO: EmptySet (LineString)

		// TODO: Polygon
		// TODO: EmptySet (Polygon)

		// TODO: MultiPoint
		// TODO: MultiLineString
		// TODO: MultiPolygon
	}

	type pair struct{ keyA, keyB string }
	eqWithTolerance := []pair{
		{"pt_a", "pt_d"},
		{"pt_c", "pt_d"},
		{"pt_f", "pt_b"},

		{"ln_a", "ln_b"},
		{"ln_b", "ln_c"},
		{"ln_a", "ln_c"},
	}

	eqWithoutOrder := []pair{
		{"ln_a", "ln_d"},
	}

	t.Run("reflexive", func(t *testing.T) {
		for key, wkt := range wkts {
			t.Run(key, func(t *testing.T) {
				t.Logf("WKT: %v", wkt)
				g := geomFromWKT(t, wkt)
				t.Run("no options", func(t *testing.T) {
					if !g.EqualsExact(g) {
						t.Errorf("should be equal to itself")
					}
				})
			})
		}
	})
	t.Run("equal with tolerance", func(t *testing.T) {
		for keyA := range wkts {
			for keyB := range wkts {
				t.Run(keyA+" and "+keyB, func(t *testing.T) {
					var want bool
					if keyA == keyB {
						want = true
					}
					for _, p := range eqWithTolerance {
						if (keyA == p.keyA && keyB == p.keyB) || (keyA == p.keyB && keyB == p.keyA) {
							want = true
							break
						}
					}
					gA := geomFromWKT(t, wkts[keyA])
					gB := geomFromWKT(t, wkts[keyB])
					got := gA.EqualsExact(gB, geom.Tolerance(0.125))
					if got != want {
						t.Errorf("got=%v want=%v", got, want)
					}
				})
			}
		}
	})
	t.Run("equal ignoring order", func(t *testing.T) {
		for keyA := range wkts {
			for keyB := range wkts {
				t.Run(keyA+" and "+keyB, func(t *testing.T) {
					var want bool
					if keyA == keyB {
						want = true
					}
					for _, p := range eqWithoutOrder {
						if (keyA == p.keyA && keyB == p.keyB) || (keyA == p.keyB && keyB == p.keyA) {
							want = true
							break
						}
					}
					gA := geomFromWKT(t, wkts[keyA])
					gB := geomFromWKT(t, wkts[keyB])
					got := gA.EqualsExact(gB, geom.IgnoreOrder)
					if got != want {
						t.Errorf("got=%v want=%v", got, want)
					}
				})
			}
		}
	})
}

func TestEqualsExactOrthogonal(t *testing.T) {
	// TODO: check that the two options don't interact with each other badly.
}
