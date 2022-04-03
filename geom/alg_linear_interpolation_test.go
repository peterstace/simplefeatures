package geom_test

import (
	"fmt"
	"strconv"
	"testing"
)

func TestInterpolatePointEmpty(t *testing.T) {
	for _, variant := range []string{"", "Z", "M", "ZM"} {
		for _, ratio := range []float64{-0.5, 0, 0.5, 1, 1.5} {
			t.Run(fmt.Sprintf("%v_%v", variant, ratio), func(t *testing.T) {
				inputWKT := "LINESTRING " + variant + " EMPTY"
				input := geomFromWKT(t, inputWKT).MustAsLineString()
				wantWKT := "POINT " + variant + " EMPTY"
				got := input.InterpolatePoint(ratio).AsGeometry()
				expectGeomEqWKT(t, got, wantWKT)
			})
		}
	}
}

func TestInterpolatePoint(t *testing.T) {
	for i, tc := range []struct {
		lsWKT   string
		frac    float64
		wantWKT string
	}{
		// Single segment XY
		{"LINESTRING(2 1,1 3)", -0.5, "POINT(2.00 1.0)"},
		{"LINESTRING(2 1,1 3)", 0.00, "POINT(2.00 1.0)"},
		{"LINESTRING(2 1,1 3)", 0.25, "POINT(1.75 1.5)"},
		{"LINESTRING(2 1,1 3)", 0.50, "POINT(1.50 2.0)"},
		{"LINESTRING(2 1,1 3)", 0.75, "POINT(1.25 2.5)"},
		{"LINESTRING(2 1,1 3)", 1.00, "POINT(1.00 3.0)"},
		{"LINESTRING(2 1,1 3)", 1.50, "POINT(1.00 3.0)"},

		// Single segment Z
		{"LINESTRING Z (2 1 5,1 3 7)", -0.5, "POINT Z (2.00 1.0 5.0)"},
		{"LINESTRING Z (2 1 5,1 3 7)", 0.00, "POINT Z (2.00 1.0 5.0)"},
		{"LINESTRING Z (2 1 5,1 3 7)", 0.25, "POINT Z (1.75 1.5 5.5)"},
		{"LINESTRING Z (2 1 5,1 3 7)", 0.50, "POINT Z (1.50 2.0 6.0)"},
		{"LINESTRING Z (2 1 5,1 3 7)", 0.75, "POINT Z (1.25 2.5 6.5)"},
		{"LINESTRING Z (2 1 5,1 3 7)", 1.00, "POINT Z (1.00 3.0 7.0)"},
		{"LINESTRING Z (2 1 5,1 3 7)", 1.50, "POINT Z (1.00 3.0 7.0)"},

		// Single segment M
		{"LINESTRING M (2 1 5,1 3 7)", -0.5, "POINT M (2.00 1.0 5.0)"},
		{"LINESTRING M (2 1 5,1 3 7)", 0.00, "POINT M (2.00 1.0 5.0)"},
		{"LINESTRING M (2 1 5,1 3 7)", 0.25, "POINT M (1.75 1.5 5.5)"},
		{"LINESTRING M (2 1 5,1 3 7)", 0.50, "POINT M (1.50 2.0 6.0)"},
		{"LINESTRING M (2 1 5,1 3 7)", 0.75, "POINT M (1.25 2.5 6.5)"},
		{"LINESTRING M (2 1 5,1 3 7)", 1.00, "POINT M (1.00 3.0 7.0)"},
		{"LINESTRING M (2 1 5,1 3 7)", 1.50, "POINT M (1.00 3.0 7.0)"},

		// Single segment ZM
		{"LINESTRING ZM (2 1 5 7,1 3 7 5)", -0.5, "POINT ZM (2.00 1.0 5.0 7.0)"},
		{"LINESTRING ZM (2 1 5 7,1 3 7 5)", 0.00, "POINT ZM (2.00 1.0 5.0 7.0)"},
		{"LINESTRING ZM (2 1 5 7,1 3 7 5)", 0.25, "POINT ZM (1.75 1.5 5.5 6.5)"},
		{"LINESTRING ZM (2 1 5 7,1 3 7 5)", 0.50, "POINT ZM (1.50 2.0 6.0 6.0)"},
		{"LINESTRING ZM (2 1 5 7,1 3 7 5)", 0.75, "POINT ZM (1.25 2.5 6.5 5.5)"},
		{"LINESTRING ZM (2 1 5 7,1 3 7 5)", 1.00, "POINT ZM (1.00 3.0 7.0 5.0)"},
		{"LINESTRING ZM (2 1 5 7,1 3 7 5)", 1.50, "POINT ZM (1.00 3.0 7.0 5.0)"},

		// Multiple Segments (all equal length)
		{"LINESTRING(0 0,1 0,2 0,3 0,3 1)", 0.000, "POINT(0.0 0.0)"},
		{"LINESTRING(0 0,1 0,2 0,3 0,3 1)", 0.125, "POINT(0.5 0.0)"},
		{"LINESTRING(0 0,1 0,2 0,3 0,3 1)", 0.250, "POINT(1.0 0.0)"},
		{"LINESTRING(0 0,1 0,2 0,3 0,3 1)", 0.375, "POINT(1.5 0.0)"},
		{"LINESTRING(0 0,1 0,2 0,3 0,3 1)", 0.875, "POINT(3.0 0.5)"},

		// Multiple Segments (different lengths)
		{"LINESTRING(0 0,3 0,3 1)", 0.000, "POINT(0.0 0.0)"},
		{"LINESTRING(0 0,3 0,3 1)", 0.125, "POINT(0.5 0.0)"},
		{"LINESTRING(0 0,3 0,3 1)", 0.250, "POINT(1.0 0.0)"},
		{"LINESTRING(0 0,3 0,3 1)", 0.375, "POINT(1.5 0.0)"},
		{"LINESTRING(0 0,3 0,3 1)", 0.875, "POINT(3.0 0.5)"},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			ls := geomFromWKT(t, tc.lsWKT).MustAsLineString()
			t.Logf("ls:   %v", ls.AsText())
			t.Logf("frac: %v", tc.frac)
			got := ls.InterpolatePoint(tc.frac).AsGeometry()
			expectGeomEqWKT(t, got, tc.wantWKT)
		})
	}
}

func TestInterpolateEvenlySpacedPointsEmpty(t *testing.T) {
	for _, variant := range []string{"", "Z", "M", "ZM"} {
		for n, want := range map[int]string{
			-1: "MULTIPOINT " + variant + " EMPTY",
			0:  "MULTIPOINT " + variant + " EMPTY",
			1:  "MULTIPOINT " + variant + " (EMPTY)",
			2:  "MULTIPOINT " + variant + " (EMPTY,EMPTY)",
		} {
			t.Run(fmt.Sprintf("%v_%v", variant, n), func(t *testing.T) {
				inputWKT := "LINESTRING " + variant + " EMPTY"
				input := geomFromWKT(t, inputWKT).MustAsLineString()
				got := input.InterpolateEvenlySpacedPoints(n).AsGeometry()
				expectGeomEqWKT(t, got, want)
			})
		}
	}
}

func TestInterpolateEvenlySpacedPoints(t *testing.T) {
	for i, tc := range []struct {
		lsWKT   string
		n       int
		wantWKT string
	}{
		{"LINESTRING(1 1,2 3)", 0, "MULTIPOINT EMPTY"},
		{"LINESTRING(1 1,2 3)", 1, "MULTIPOINT((1.5 2))"},
		{"LINESTRING(1 1,2 3)", 2, "MULTIPOINT((1 1),(2 3))"},
		{"LINESTRING(1 1,2 3)", 3, "MULTIPOINT((1 1),(1.5 2),(2 3))"},
		{"LINESTRING(1 1,2 3)", 5, "MULTIPOINT((1 1),(1.25 1.5),(1.5 2),(1.75 2.5),(2 3))"},

		{"LINESTRING(0 0,1 0,2 0,3 0,3 1)", 5, "MULTIPOINT(0 0,1 0,2 0,3 0,3 1)"},
		{"LINESTRING(0 0,        3 0,3 1)", 5, "MULTIPOINT(0 0,1 0,2 0,3 0,3 1)"},
		{"LINESTRING(0 0,        3 0,3 1)", 1, "MULTIPOINT(2 0)"},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			ls := geomFromWKT(t, tc.lsWKT).MustAsLineString()
			got := ls.InterpolateEvenlySpacedPoints(tc.n).AsGeometry()
			expectGeomEqWKT(t, got, tc.wantWKT)
		})
	}
}
