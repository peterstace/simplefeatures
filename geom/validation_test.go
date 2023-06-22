package geom_test

import (
	"fmt"
	"math"
	"strconv"
	"testing"

	. "github.com/peterstace/simplefeatures/geom"
)

func xy(x, y float64) Coordinates {
	return Coordinates{Type: DimXY, XY: XY{x, y}}
}

func TestPointValidation(t *testing.T) {
	nan := math.NaN()
	inf := math.Inf(+1)
	for i, tc := range []struct {
		wantValid bool
		input     Coordinates
	}{
		{true, xy(0, 0)},
		{false, xy(nan, 0)},
		{false, xy(0, nan)},
		{false, xy(nan, nan)},
		{false, xy(inf, 0)},
		{false, xy(0, inf)},
		{false, xy(inf, inf)},
		{false, xy(-inf, 0)},
		{false, xy(0, -inf)},
		{false, xy(-inf, -inf)},
	} {
		t.Run(fmt.Sprintf("point_%d", i), func(t *testing.T) {
			_, err := NewPoint(tc.input)
			if tc.wantValid {
				expectNoErr(t, err)
			} else {
				expectErr(t, err)
			}
		})
	}
}

func TestDisableAllPointValidations(t *testing.T) {
	c := xy(2, math.NaN())

	_, err := NewPoint(c)
	expectErr(t, err)

	_, err = NewPoint(c, DisableAllValidations)
	expectNoErr(t, err)
}

func TestLineStringValidationInvalidFromRawCoords(t *testing.T) {
	nan := math.NaN()
	inf := math.Inf(+1)
	for i, pts := range [][]float64{
		{0, 0},
		{1, 1},
		{0, 0, 0, 0},
		{1, 1, 1, 1},
		{0, 0, 1, 1, 2, nan},
		{0, 0, 1, 1, nan, 2},
		{0, 0, 1, 1, 2, inf},
		{0, 0, 1, 1, inf, 2},
		{0, 0, 1, 1, 2, -inf},
		{0, 0, 1, 1, -inf, 2},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			seq := NewSequence(pts, DimXY)
			_, err := NewLineString(seq)
			expectErr(t, err)
			_, err = NewLineString(seq, DisableAllValidations)
			expectNoErr(t, err)
		})
	}
}

func TestPolygonValidation(t *testing.T) {
	for i, wkt := range []string{
		"POLYGON EMPTY",
		"POLYGON((0 0,1 0,1 1,0 1,0 0))",
		"POLYGON((0 0,3 0,3 3,0 3,0 0),(1 1,2 1,2 2,1 2,1 1))",
		`POLYGON(
			(0 0,5 0,5 5,0 5,0 0),
			(1 1,3 1,3 2,1 1),
			(1 1,4 3,3 4,1 1),
			(1 1,2 3,1 3,1 1)
		)`,
		`POLYGON(
			(0 0,5 0,5 5,0 5,0 0),
			(1 1,2 1,2 2,1 2,1 1),
			(2 1,3 1,3 2,2 1),
			(1 2,2 3,1 3,1 2),
			(2 2,4 3,3 4,2 2)
		)`,
	} {
		t.Run("valid_"+strconv.Itoa(i), func(t *testing.T) {
			_, err := UnmarshalWKT(wkt)
			if err != nil {
				t.Error(err)
			}
		})
	}
	for i, wkt := range []string{
		// not closed
		"POLYGON((0 0,1 1,0 1))",

		// not simple
		"POLYGON((0 0,1 1,0 1,1 0,0 0))",

		// intersect at a line
		"POLYGON((0 0,3 0,3 3,0 3,0 0),(0 1,1 1,1 2,0 2,0 1))",

		// intersect at two points
		"POLYGON((0 0,3 0,3 3,0 3,0 0),(1 0,3 1,2 2,1 0))",

		// inner ring is outside of the outer ring
		"POLYGON((0 0,3 0,3 3,0 3,0 0),(4 0,7 0,7 3,4 3,4 0))",

		// polygons aren't connected
		`POLYGON(
			(0 0, 4 0, 4 4, 0 4, 0 0),
			(2 0, 3 1, 2 2, 1 1, 2 0),
			(2 2, 3 3, 2 4, 1 3, 2 2)
		)`,
		`POLYGON(
			(0 0, 6 0, 6 5, 0 5, 0 0),
			(2 1, 4 1, 4 2, 2 2, 2 1),
			(2 2, 3 3, 2 4, 1 3, 2 2),
			(4 2, 5 3, 4 4, 3 3, 4 2)
		)`,

		// Nested rings
		`POLYGON(
			(0 0,5 0,5 5,0 5,0 0),
			(1 1,4 1,4 4,1 4,1 1),
			(2 2,3 2,3 3,2 3,2 2)
		)`,
		`POLYGON(
			(0 0,5 0,5 5,0 5,0 0),
			(2 2,3 2,3 3,2 3,2 2),
			(1 1,4 1,4 4,1 4,1 1)
		)`,

		// Contains empty rings.
		`POLYGON(EMPTY)`,
		`POLYGON(EMPTY,(0 0,0 1,1 0,0 0))`,
		`POLYGON((0 0,0 1,1 0,0 0),EMPTY)`,
	} {
		t.Run("invalid_"+strconv.Itoa(i), func(t *testing.T) {
			_, err := UnmarshalWKT(wkt)
			if err == nil {
				t.Log("WKT", wkt)
				t.Error("expected error")
			}
		})
	}
}

func TestMultiPolygonValidation(t *testing.T) {
	for i, wkt := range []string{
		`MULTIPOLYGON EMPTY`,
		`MULTIPOLYGON(((0 0,0 1,1 1,1 0,0 0)))`,
		`MULTIPOLYGON(
			((0 0,1 0,1 1,0 1,0 0)),
			((2 0,3 0,3 1,2 1,2 0))
		)`,
		`MULTIPOLYGON(
			((0 0,1 0,0 1,0 0)),
			((1 0,2 0,1 1,1 0))
		)`,
		`MULTIPOLYGON(
			((0 0,2 0,2 3,1 1,0 3,0 0)),
			((1 2,2 3,0 3,1 2))
		)`,
		`MULTIPOLYGON(
			((0 0,5 0,5 5,0 5,0 0),(1 1,4 1,4 4,1 4,1 1)),
			((2 2,3 2,3 3,2 3,2 2))
		)`,

		// Child polygons can be empty.
		`MULTIPOLYGON(EMPTY)`,
		`MULTIPOLYGON(((0 0,0 1,1 0,0 0)),EMPTY)`,
		`MULTIPOLYGON(EMPTY,((0 0,0 1,1 0,0 0)))`,

		// Replicates a bug.
		`MULTIPOLYGON(((0 0,0 1,1 1,1 0,0 0)),((2 -1,3 -1,3 0,2 0,2 -1)),((1 1,3 1,3 3,1 3,1 1)))`,
	} {
		t.Run(fmt.Sprintf("valid_%d", i), func(t *testing.T) {
			geomFromWKT(t, wkt)
		})
	}
	for i, wkt := range []string{
		`MULTIPOLYGON(
			((-6 -3,8 4,7 6,-7 -1,-6 -3)),
			((3 -6,5 -5,-2 9,-4 8,3 -6))
		)`,
		`MULTIPOLYGON(
			((0 0,0 1,1 1,1 0,0 0)),
			((1 0,1 1,2 1,2 0,1 0))
		)`,
		`MULTIPOLYGON(
			((0 0,2 0,2 2,0 2,0 0)),
			((1 0,3 0,3 2,1 2,1 0))
		)`,
		`MULTIPOLYGON(
			((1 0,2 0,1 3,1 0)),
			((0 1,3 1,3 2,0 1))
		)`,
		`MULTIPOLYGON(
			((0 0,3 0,3 3,0 3,0 0)),
			((2 1,3 3,1 2,2 1))
		)`,
		`MULTIPOLYGON(
			((2 1,3 3,1 2,2 1)),
			((0 0,3 0,3 3,0 3,0 0))
		)`,
		`MULTIPOLYGON(
			((0 0,0 1,1 0,0 0)),
			((0 0,0 1,1 0,0 0))
		)`,
		`MULTIPOLYGON(
			((0 0,3 0,3 3,0 3,0 0)),
			((1 1,2 1,2 2,1 2,1 1))
		)`,
		`MULTIPOLYGON(
			((1 1,2 1,2 2,1 2,1 1)),
			((0 0,3 0,3 3,0 3,0 0))
		)`,
		`MULTIPOLYGON(
			((0 0,2 0,2 1,0 1,0 0)),
			((0.5 -0.5,1 2,1.5 -0.5,2 2,2 3,0 3,0 2,0.5 -0.5))
		)`,
		`MULTIPOLYGON(
			((0 0,2 0,2 1,0 1,0 0)),
			((0.5 1,1 2,1.5 -0.5,2 2,2 3,0 3,0 2,0.5 1))
		)`,
	} {
		t.Run(fmt.Sprintf("invalid_%d", i), func(t *testing.T) {
			_, err := UnmarshalWKT(wkt)
			if err == nil {
				t.Log(wkt)
				t.Error("expected error")
			}
		})
	}
}
