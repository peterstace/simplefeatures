package geom_test

import (
	"fmt"
	"math"
	"strconv"
	"testing"

	"github.com/peterstace/simplefeatures/geom"
	"github.com/peterstace/simplefeatures/internal/test"
)

func xy(x, y float64) geom.Coordinates {
	return geom.Coordinates{Type: geom.DimXY, XY: geom.XY{x, y}}
}

func TestPointValidation(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		pt := geom.NewPoint(xy(0, 0))
		test.NoErr(t, pt.Validate())
	})

	nan := math.NaN()
	inf := math.Inf(+1)
	for i, tc := range []struct {
		reason geom.RuleViolation
		input  geom.Coordinates
	}{
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
		t.Run(fmt.Sprintf("invalid_%d", i), func(t *testing.T) {
			pt := geom.NewPoint(tc.input)
			expectRuleViolation(t, pt, tc.reason)
		})
	}
}

func TestLineStringValidation(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		seq := geom.NewSequence([]float64{0, 0, 1, 1}, geom.DimXY)
		ls := geom.NewLineString(seq)
		test.NoErr(t, ls.Validate())
	})

	nan := math.NaN()
	inf := math.Inf(+1)
	for i, tc := range []struct {
		reason geom.RuleViolation
		inputs []float64
	}{
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
		t.Run(fmt.Sprintf("invalid_%d", i), func(t *testing.T) {
			seq := geom.NewSequence(tc.inputs, geom.DimXY)
			ls := geom.NewLineString(seq)
			expectRuleViolation(t, ls, tc.reason)
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
			test.NoErr(t, poly.Validate())
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
				var ve geom.ValidationError
				test.ErrAs(t, err, &ve)
				test.Eq(t, string(ve.RuleViolation), string(tc.reason))
			})
			t.Run("Validate", func(t *testing.T) {
				poly, err := geom.UnmarshalWKT(tc.wkt, geom.NoValidate{})
				test.NoErr(t, err)
				expectRuleViolation(t, poly, tc.reason)
			})
		})
	}
}

func TestMultiPointValidation(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		mp := geom.NewMultiPoint([]geom.Point{
			geom.NewPoint(xy(0, 1)),
			geom.NewPoint(xy(2, 3)),
		})
		test.NoErr(t, mp.Validate())
	})

	nan := math.NaN()
	for i, tc := range []struct {
		reason geom.RuleViolation
		coords []geom.Coordinates
	}{
		{geom.ViolateNaN, []geom.Coordinates{xy(0, 1), xy(2, nan)}},
		{geom.ViolateNaN, []geom.Coordinates{xy(nan, 1), xy(2, 3)}},
	} {
		t.Run(fmt.Sprintf("invalid_%d", i), func(t *testing.T) {
			var pts []geom.Point
			for _, c := range tc.coords {
				pt := geom.NewPoint(c)
				pts = append(pts, pt)
			}
			mp := geom.NewMultiPoint(pts)
			expectRuleViolation(t, mp, tc.reason)
		})
	}
}

func TestMultiLineStringValidation(t *testing.T) {
	newMLS := func(coords [][]float64) geom.MultiLineString {
		var lss []geom.LineString
		for _, c := range coords {
			seq := geom.NewSequence(c, geom.DimXY)
			ls := geom.NewLineString(seq)
			lss = append(lss, ls)
		}
		return geom.NewMultiLineString(lss)
	}
	t.Run("valid_empty", func(t *testing.T) {
		test.NoErr(t, newMLS([][]float64{}).Validate())
	})
	t.Run("valid_single", func(t *testing.T) {
		test.NoErr(t, newMLS([][]float64{{0, 1, 2, 3}}).Validate())
	})

	nan := math.NaN()
	for i, tc := range []struct {
		reason geom.RuleViolation
		coords [][]float64
	}{
		{geom.ViolateTwoPoints, [][]float64{{0, 1}}},
		{geom.ViolateNaN, [][]float64{{0, 1, 2, nan}}},
	} {
		t.Run(fmt.Sprintf("invalid_%d", i), func(t *testing.T) {
			expectRuleViolation(t, newMLS(tc.coords), tc.reason)
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
			test.FromWKT(t, wkt)
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
			g := test.FromWKT(t, wkt, geom.NoValidate{})
			expectRuleViolation(t, g, geom.ViolatePolysMultiTouch)
		})
	}
}

func TestMultiPolygonConstraintValidation(t *testing.T) {
	poly, err := geom.UnmarshalWKT("POLYGON((0 0,1 1,0 1,1 0,0 0))", geom.NoValidate{})
	test.NoErr(t, err)
	expectRuleViolation(t, poly, geom.ViolateRingSimple)

	mp := geom.NewMultiPolygon([]geom.Polygon{poly.MustAsPolygon()})
	expectRuleViolation(t, mp, geom.ViolateRingSimple)
}

func TestGeometryCollectionValidation(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		gc := test.FromWKT(t, "GEOMETRYCOLLECTION(LINESTRING(0 1,2 3))", geom.NoValidate{})
		test.NoErr(t, gc.Validate())
	})
	t.Run("invalid", func(t *testing.T) {
		gc := test.FromWKT(t, "GEOMETRYCOLLECTION(LINESTRING(0 1))", geom.NoValidate{})
		expectRuleViolation(t, gc, geom.ViolateTwoPoints)
	})
}
