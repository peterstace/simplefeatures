package geom_test

import (
	"testing"

	"github.com/peterstace/simplefeatures/geom"
)

func TestEqualsExact(t *testing.T) {
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

		"ls_empty": "LINESTRING EMPTY",

		"ls_a": "LINESTRING(1 2,3 4,5 6)",
		"ls_b": "LINESTRING(1 2,3 4,5 6.1)",
		"ls_c": "LINESTRING(5 6,3 4,1 2)",

		// ccw rings
		"ls_m": "LINESTRING(0 0,1 0,0 1,0 0)",
		"ls_n": "LINESTRING(1 0,0 1,0 0,1 0)",
		"ls_o": "LINESTRING(0 1,0 0,1 0,0 1)",
		// cw rings
		"ls_p": "LINESTRING(0 0,0 1,1 0,0 0)",
		"ls_q": "LINESTRING(1 0,0 0,0 1,1 0)",
		"ls_r": "LINESTRING(0 1,1 0,0 0,0 1)",

		"lr_a": "LINEARRING(0 0,0 1,1 1,1 0,0 0)",
		"lr_b": "LINEARRING(0 0,1 0,1 1,0 1,0 0)",
		"lr_c": "LINESTRING(0 0,1 0,1 1,0 1,0 0)",

		"lr_d": "LINEARRING(1 1,0 1,1 0,1 1)",
		"lr_e": "LINEARRING(1 1,0 1,1 0.1,1 1)",

		"p_empty": "POLYGON EMPTY",
		"p_a":     "POLYGON((0 0,0 1,1 0,0 0))",
		"p_b":     "POLYGON((0 0,1 0,0 1,0 0))",
		"p_c":     "POLYGON((0 0,0 1,1 1,1 0,0 0))",
		"p_d":     "POLYGON((0 0,0 1,1 1,1 0.1,0 0))",

		"p_e": `POLYGON(
			(0 0,5 0,5 3,0 3,0 0),
			(1 1,2 1,2 2,1 2,1 1),
			(3 1,4 1,4 2,3 2,3 1)
		)`,
		"p_f": `POLYGON(
			(0 0,5 0,5 3,0 3,0 0),
			(3 1,4 1,4 2,3 2,3 1),
			(1 1,2 1,2 2,1 2,1 1)
		)`,

		"mp_empty": "MULTIPOINT EMPTY",

		"mp_1_a": "MULTIPOINT(4 8)",
		"mp_1_b": "MULTIPOINT(4 8.1)",
		"mp_1_c": "MULTIPOINT(2 5)",

		"mp_2_a": "MULTIPOINT(4 2,7 5)",
		"mp_2_b": "MULTIPOINT(4 1.9,7.1 5)",
		"mp_2_c": "MULTIPOINT(3 8,2 5)",
		"mp_2_d": "MULTIPOINT(2 5,3 8)",

		"mp_2_e": "MULTIPOINT(2 5,2 5)",

		"mp_3_a": "MULTIPOINT(1 1,1 2,2 1)",
		"mp_3_b": "MULTIPOINT(1 1,2 1,1 2)",
		"mp_3_c": "MULTIPOINT(1 2,1 1,2 1)",
		"mp_3_d": "MULTIPOINT(1 2,2 1,1 1)",
		"mp_3_e": "MULTIPOINT(2 1,1 1,1 2)",
		"mp_3_f": "MULTIPOINT(2 1,1 2,1 1)",

		"mp_3_g": "MULTIPOINT(3 3,3 3,7 6)",
		"mp_3_h": "MULTIPOINT(7 6,3 3,3 3)",
		"mp_3_i": "MULTIPOINT(3 3,7 6,3 3)",

		"mls_empty": "MULTILINESTRING EMPTY",

		"mls_a": "MULTILINESTRING((0 1,2 3,4 5))",
		"mls_b": "MULTILINESTRING((4 5,2 3,0 1))",

		"mls_c": `MULTILINESTRING(
			(5 3,4 8,1 2,9 8),
			(8 4,6 1,3 9,0 2)
		)`,
		"mls_d": `MULTILINESTRING(
			(8 4,6 1,3 9,0 2),
			(5 3,4 8,1 2,9 8)
		)`,

		"mpo_empty": "MULTIPOLYGON EMPTY",
		"mpo_1_a":   "MULTIPOLYGON(((0 0,0 1,1 0,0 0)))",
		"mpo_1_b":   "MULTIPOLYGON(((0 0,1 0,0 1,0 0)))",
		"mpo_1_c":   "MULTIPOLYGON(((0 0,0 1,1 1,1 0,0 0)))",

		"g_empty": "GEOMETRYCOLLECTION EMPTY",
		"g_1_a":   "GEOMETRYCOLLECTION(POINT(1 2))",
		"g_1_b":   "GEOMETRYCOLLECTION(POINT(1 3))",
		"g_1_c":   "GEOMETRYCOLLECTION(POINT(1.1 9))",
		"g_1_d":   "GEOMETRYCOLLECTION(POINT(1.0 9))",
		"g_2_a":   "GEOMETRYCOLLECTION(POINT(1 3),LINESTRING(1 2,3 4))",
		"g_2_b":   "GEOMETRYCOLLECTION(LINESTRING(1 2,3 4),POINT(1 3))",
		"g_2_c":   "GEOMETRYCOLLECTION(GEOMETRYCOLLECTION(POINT(1 5),LINESTRING(1 2,3 4)))",
		"g_2_d":   "GEOMETRYCOLLECTION(GEOMETRYCOLLECTION(LINESTRING(1 2,3 4),POINT(1 5)))",
	}

	type pair struct{ keyA, keyB string }
	eqWithTolerance := []pair{
		{"pt_a", "pt_d"},
		{"pt_c", "pt_d"},
		{"pt_f", "pt_b"},

		{"ln_a", "ln_b"},
		{"ln_b", "ln_c"},
		{"ln_a", "ln_c"},

		{"ls_a", "ls_b"},

		{"lr_b", "lr_c"},
		{"lr_d", "lr_e"},

		{"mp_1_a", "mp_1_b"},
		{"mp_2_a", "mp_2_b"},

		{"p_c", "p_d"},

		{"g_1_c", "g_1_d"},
	}

	eqWithoutOrder := []pair{
		{"ln_a", "ln_d"},
		{"ls_a", "ls_c"},

		{"ls_m", "ls_n"},
		{"ls_m", "ls_o"},
		{"ls_m", "ls_p"},
		{"ls_m", "ls_q"},
		{"ls_m", "ls_r"},
		{"ls_n", "ls_o"},
		{"ls_n", "ls_p"},
		{"ls_n", "ls_q"},
		{"ls_n", "ls_r"},
		{"ls_o", "ls_p"},
		{"ls_o", "ls_q"},
		{"ls_o", "ls_r"},
		{"ls_p", "ls_q"},
		{"ls_p", "ls_r"},
		{"ls_q", "ls_r"},

		{"lr_a", "lr_b"},
		{"lr_b", "lr_c"},
		{"lr_c", "lr_a"},

		{"mp_2_c", "mp_2_d"},

		{"mp_3_a", "mp_3_b"},
		{"mp_3_a", "mp_3_c"},
		{"mp_3_a", "mp_3_d"},
		{"mp_3_a", "mp_3_e"},
		{"mp_3_a", "mp_3_f"},
		{"mp_3_b", "mp_3_c"},
		{"mp_3_b", "mp_3_d"},
		{"mp_3_b", "mp_3_e"},
		{"mp_3_b", "mp_3_f"},
		{"mp_3_c", "mp_3_d"},
		{"mp_3_c", "mp_3_e"},
		{"mp_3_c", "mp_3_f"},
		{"mp_3_d", "mp_3_e"},
		{"mp_3_d", "mp_3_f"},
		{"mp_3_e", "mp_3_f"},

		{"mp_3_g", "mp_3_h"},
		{"mp_3_h", "mp_3_i"},
		{"mp_3_i", "mp_3_g"},

		{"p_a", "p_b"},
		{"p_e", "p_f"},

		{"mls_a", "mls_b"},
		{"mls_c", "mls_d"},

		{"mpo_1_a", "mpo_1_b"},

		{"g_2_a", "g_2_b"},
		{"g_2_c", "g_2_d"},
	}

	for _, p := range append(eqWithTolerance, eqWithoutOrder...) {
		if _, ok := wkts[p.keyA]; !ok {
			t.Fatalf("bad test setup: %v doesn't exist", p.keyA)
		}
		if _, ok := wkts[p.keyB]; !ok {
			t.Fatalf("bad test setup: %v doesn't exist", p.keyB)
		}
	}

	t.Run("reflexive", func(t *testing.T) {
		for key, wkt := range wkts {
			t.Run(key, func(t *testing.T) {
				g := geomFromWKT(t, wkt)
				t.Run("no options", func(t *testing.T) {
					if !g.EqualsExact(g) {
						t.Logf("WKT: %v", wkt)
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
						t.Logf("WKT A: %v", wkts[keyA])
						t.Logf("WKT B: %v", wkts[keyB])
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
						t.Logf("WKT A: %v", wkts[keyA])
						t.Logf("WKT B: %v", wkts[keyB])
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
