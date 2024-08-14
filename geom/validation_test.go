package geom_test

import (
	"fmt"
	"math"
	"strconv"
	"testing"

	"github.com/peterstace/simplefeatures/geom"
)

func xy(x, y float64) geom.Coordinates {
	return geom.Coordinates{Type: geom.DimXY, XY: geom.XY{x, y}}
}

func TestPointValidation(t *testing.T) {
	nan := math.NaN()
	inf := math.Inf(+1)
	for i, tc := range []struct {
		reason geom.RuleViolation
		input  geom.Coordinates
	}{
		{"", xy(0, 0)},
		{geom.ViolateNaN, xy(nan, 0)},
		{geom.ViolateNaN, xy(0, nan)},
		{geom.ViolateNaN, xy(nan, nan)},
		{geom.ViolateInf, xy(inf, 0)},
		{geom.ViolateInf, xy(0, inf)},
		{geom.ViolateInf, xy(inf, inf)},
		{geom.ViolateInf, xy(-inf, 0)},
		{geom.ViolateInf, xy(0, -inf)},
		{geom.ViolateInf, xy(-inf, -inf)},
	} {
		t.Run(fmt.Sprintf("point_%d", i), func(t *testing.T) {
			pt := geom.NewPoint(tc.input)
			expectValidity(t, pt, tc.reason)
		})
	}
}

func TestLineStringValidation(t *testing.T) {
	nan := math.NaN()
	inf := math.Inf(+1)
	for i, tc := range []struct {
		reason geom.RuleViolation
		inputs []float64
	}{
		{"", []float64{0, 0, 1, 1}},
		{geom.ViolateTwoPoints, []float64{0, 0}},
		{geom.ViolateTwoPoints, []float64{1, 1}},
		{geom.ViolateTwoPoints, []float64{0, 0, 0, 0}},
		{geom.ViolateTwoPoints, []float64{1, 1, 1, 1}},
		{geom.ViolateNaN, []float64{0, 0, 1, 1, 2, nan}},
		{geom.ViolateNaN, []float64{0, 0, 1, 1, nan, 2}},
		{geom.ViolateInf, []float64{0, 0, 1, 1, 2, inf}},
		{geom.ViolateInf, []float64{0, 0, 1, 1, inf, 2}},
		{geom.ViolateInf, []float64{0, 0, 1, 1, 2, -inf}},
		{geom.ViolateInf, []float64{0, 0, 1, 1, -inf, 2}},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			seq := geom.NewSequence(tc.inputs, geom.DimXY)
			ls := geom.NewLineString(seq)
			expectValidity(t, ls, tc.reason)
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
			poly, err := geom.UnmarshalWKT(wkt)
			if err != nil {
				t.Error(err)
			}
			expectNoErr(t, poly.Validate())
		})
	}

	for i, tc := range []struct {
		reason geom.RuleViolation
		wkt    string
	}{
		{
			geom.ViolateRingClosed,
			"POLYGON((0 0,1 1,0 1))",
		},
		{
			geom.ViolateRingSimple,
			"POLYGON((0 0,1 1,0 1,1 0,0 0))",
		},
		{
			geom.ViolateRingsMultiTouch,
			"POLYGON((0 0,3 0,3 3,0 3,0 0),(0 1,1 1,1 2,0 2,0 1))",
		},
		{
			geom.ViolateRingsMultiTouch,
			"POLYGON((0 0,3 0,3 3,0 3,0 0),(1 0,3 1,2 2,1 0))",
		},
		{
			geom.ViolateInteriorInExterior,
			"POLYGON((0 0,3 0,3 3,0 3,0 0),(4 0,7 0,7 3,4 3,4 0))",
		},
		{
			geom.ViolateInteriorConnected,
			`POLYGON(
				(0 0, 4 0, 4 4, 0 4, 0 0),
				(2 0, 3 1, 2 2, 1 1, 2 0),
				(2 2, 3 3, 2 4, 1 3, 2 2)
			)`,
		},
		{
			geom.ViolateInteriorConnected,
			`POLYGON(
				(0 0, 6 0, 6 5, 0 5, 0 0),
				(2 1, 4 1, 4 2, 2 2, 2 1),
				(2 2, 3 3, 2 4, 1 3, 2 2),
				(4 2, 5 3, 4 4, 3 3, 4 2)
			)`,
		},
		{
			geom.ViolateRingNested,
			`POLYGON(
				(0 0,5 0,5 5,0 5,0 0),
				(1 1,4 1,4 4,1 4,1 1),
				(2 2,3 2,3 3,2 3,2 2)
			)`,
		},
		{
			geom.ViolateRingNested,
			`POLYGON(
				(0 0,5 0,5 5,0 5,0 0),
				(2 2,3 2,3 3,2 3,2 2),
				(1 1,4 1,4 4,1 4,1 1)
			)`,
		},
		{
			geom.ViolateRingEmpty,
			`POLYGON(EMPTY)`,
		},
		{
			geom.ViolateRingEmpty,
			`POLYGON(EMPTY,(0 0,0 1,1 0,0 0))`,
		},
		{
			geom.ViolateRingEmpty,
			`POLYGON((0 0,0 1,1 0,0 0),EMPTY)`,
		},
		{
			// https://github.com/peterstace/simplefeatures/issues/631
			geom.ViolateInteriorInExterior,
			`POLYGON(
				(1 1,5 1,5 5,1 5,1 1),
				(3 1,6 0,6 6,0 6,0 0,3 1)
			)`,
		},
	} {
		t.Run("invalid_"+strconv.Itoa(i), func(t *testing.T) {
			t.Run("Constructor", func(t *testing.T) {
				_, err := geom.UnmarshalWKT(tc.wkt)
				expectValidationErrWithReason(t, err, tc.reason)
			})
			t.Run("Validate", func(t *testing.T) {
				poly, err := geom.UnmarshalWKT(tc.wkt, geom.NoValidate{})
				expectNoErr(t, err)
				expectValidity(t, poly, tc.reason)
			})
		})
	}
}

func TestMultiPointValidation(t *testing.T) {
	nan := math.NaN()
	for i, tc := range []struct {
		reason geom.RuleViolation
		coords []geom.Coordinates
	}{
		{"", []geom.Coordinates{xy(0, 1), xy(2, 3)}},
		{geom.ViolateNaN, []geom.Coordinates{xy(0, 1), xy(2, nan)}},
		{geom.ViolateNaN, []geom.Coordinates{xy(nan, 1), xy(2, 3)}},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			var pts []geom.Point
			for _, c := range tc.coords {
				pt := geom.NewPoint(c)
				pts = append(pts, pt)
			}
			mp := geom.NewMultiPoint(pts)
			expectValidity(t, mp, tc.reason)
		})
	}
}

func TestMultiLineStringValidation(t *testing.T) {
	nan := math.NaN()
	for i, tc := range []struct {
		reason geom.RuleViolation
		coords [][]float64
	}{
		{"", [][]float64{}},
		{geom.ViolateTwoPoints, [][]float64{{0, 1}}},
		{"", [][]float64{{0, 1, 2, 3}}},
		{geom.ViolateNaN, [][]float64{{0, 1, 2, nan}}},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			var lss []geom.LineString
			for _, coords := range tc.coords {
				seq := geom.NewSequence(coords, geom.DimXY)
				ls := geom.NewLineString(seq)
				lss = append(lss, ls)
			}
			mls := geom.NewMultiLineString(lss)
			expectValidity(t, mls, tc.reason)
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
			g := geomFromWKT(t, wkt, geom.NoValidate{})
			expectValidity(t, g, geom.ViolatePolysMultiTouch)
		})
	}
}

func TestMultiPolygonConstraintValidation(t *testing.T) {
	poly, err := geom.UnmarshalWKT("POLYGON((0 0,1 1,0 1,1 0,0 0))", geom.NoValidate{})
	expectNoErr(t, err)
	expectValidity(t, poly, geom.ViolateRingSimple)

	mp := geom.NewMultiPolygon([]geom.Polygon{poly.MustAsPolygon()})
	expectValidity(t, mp, geom.ViolateRingSimple)
}

func TestGeometryCollectionValidation(t *testing.T) {
	for i, tc := range []struct {
		reason geom.RuleViolation
		wkt    string
	}{
		{"", "GEOMETRYCOLLECTION(LINESTRING(0 1,2 3))"},
		{geom.ViolateTwoPoints, "GEOMETRYCOLLECTION(LINESTRING(0 1))"},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			gc := geomFromWKT(t, tc.wkt, geom.NoValidate{})
			expectValidity(t, gc, tc.reason)
		})
	}
}
