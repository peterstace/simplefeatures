package simplefeatures_test

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"testing"

	. "github.com/peterstace/simplefeatures"
)

func TestPointValidation(t *testing.T) {
	for _, tt := range []struct {
		x, y float64
	}{
		{0, math.Inf(-1)},
		{0, math.Inf(+1)},
		{0, math.NaN()},
		{math.Inf(-1), 0},
		{math.Inf(+1), 0},
		{math.NaN(), 0},
		{math.Inf(-1), math.Inf(-1)},
		{math.Inf(+1), math.Inf(+1)},
		{math.NaN(), math.NaN()},
	} {
		t.Run(fmt.Sprintf("%f_%f", tt.x, tt.y), func(t *testing.T) {
			_, err := NewPoint(tt.x, tt.y)
			if err == nil {
				t.Error("expected error")
			}
		})
	}
}

func TestLineValidation(t *testing.T) {
	xy := func(x, y float64) Coordinates {
		return Coordinates{XY: XY{x, y}}
	}
	for i, pts := range [][2]Coordinates{
		{xy(0, 0), xy(0, 0)},
		{xy(-1, -1), xy(-1, -1)},
		{xy(0, 0), xy(1, math.NaN())},
		{xy(0, 0), xy(math.NaN(), 1)},
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
	xy := func(x, y float64) Coordinates {
		return Coordinates{XY: XY{x, y}}
	}
	for i, pts := range [][]Coordinates{
		{xy(0, 0)},
		{xy(1, 1)},
		{xy(0, 0), xy(0, 0)},
		{xy(1, 1), xy(1, 1)},
		{xy(0, 0), xy(1, math.NaN())},
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
	xy := func(x, y float64) Coordinates {
		return Coordinates{XY: XY{x, y}}
	}
	for i, pts := range [][]Coordinates{
		{xy(0, 0), xy(1, 0), xy(math.NaN(), 1), xy(0, 0)},  // has NaN
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

		// TODO: tests for connectedness
	} {
		t.Run("invalid_"+strconv.Itoa(i), func(t *testing.T) {
			_, err := UnmarshalWKT(strings.NewReader(wkt))
			if err == nil {
				t.Error("expected error")
			}
		})
	}
}
