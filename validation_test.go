package simplefeatures_test

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	. "github.com/peterstace/simplefeatures"
)

func xy(x, y float64) Coordinates {
	return Coordinates{XY: XY{
		NewScalarFromFloat64(x),
		NewScalarFromFloat64(y),
	}}
}

func TestLineValidation(t *testing.T) {
	for i, pts := range [][2]Coordinates{
		{xy(0, 0), xy(0, 0)},
		{xy(-1, -1), xy(-1, -1)},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			_, err := NewLine(pts[0], pts[1])
			if err == nil {
				t.Error("expected error")
			}
		})
	}
}

func TestLineStringValidation(t *testing.T) {
	for i, pts := range [][]Coordinates{
		{xy(0, 0)},
		{xy(1, 1)},
		{xy(0, 0), xy(0, 0)},
		{xy(1, 1), xy(1, 1)},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			_, err := NewLineString(pts)
			if err == nil {
				t.Error("expected error")
			}
		})
	}
}

func TestLinearRingValidation(t *testing.T) {
	for i, pts := range [][]Coordinates{
		{xy(0, 0), xy(1, 1), xy(0, 1)},                     // not closed
		{xy(0, 0), xy(1, 1), xy(0, 1), xy(1, 0), xy(0, 0)}, // not simple
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			_, err := NewLinearRing(pts)
			if err == nil {
				t.Error("expected error")
			}
		})
	}
}

func TestPolygonValidation(t *testing.T) {
	for i, wkt := range []string{
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
			_, err := UnmarshalWKT(strings.NewReader(wkt))
			if err != nil {
				t.Error(err)
			}
		})
	}
	for i, wkt := range []string{
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
	} {
		t.Run("invalid_"+strconv.Itoa(i), func(t *testing.T) {
			_, err := UnmarshalWKT(strings.NewReader(wkt))
			if err == nil {
				t.Error("expected error")
			}
		})
	}
}

func TestMultiPolygonValidation(t *testing.T) {
	for i, wkt := range []string{
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
	} {
		t.Run(fmt.Sprintf("valid_%d", i), func(t *testing.T) {
			geomFromWKT(t, wkt)
		})
	}
	for i, wkt := range []string{
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
			((0 0,0 1,1 0,0 0)),
			((0 0,0 1,1 0,0 0))
		)`,
	} {
		t.Run(fmt.Sprintf("invalid_%d", i), func(t *testing.T) {
			_, err := UnmarshalWKT(strings.NewReader(wkt))
			if err == nil {
				t.Error("expected error")
			}
		})
	}
}
