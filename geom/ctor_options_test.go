package geom_test

import (
	"math"
	"strconv"
	"testing"

	"github.com/peterstace/simplefeatures/geom"
)

func TestDisableValidation(t *testing.T) {
	for i, wkt := range []string{
		// Point -- has no geometric validations
		"LINESTRING(1 2,1 2)",                // same point
		"LINESTRING(1 2,1 2,1 2)",            // same point
		"POLYGON((1 2,1 2,1 2))",             // same point
		"POLYGON((0 0,0 1,1 0))",             // not closed
		"POLYGON((0 0,2 0,2 1,1 0,0 1,0 0))", // not simple
		// Exterior ring inside interior ring
		`POLYGON(
			(5 0,0 6,6 6,6 0,0 0),
			(1 1,1 9,9 9,9 1,1 1)
		)`,
		// MultiPoint -- has no validations
		"MULTILINESTRING((1 2,3 4),(1 1,1 1))",
		// Sub-Polygons overlap
		`MULTIPOLYGON(
			((0 0,2 0,2 2,0 2,0 0)),
			((1 1,3 1,3 3,1 3,1 1))
		)`,
		"GEOMETRYCOLLECTION(LINESTRING(0 1,0 1))",
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			_, err := geom.UnmarshalWKT(wkt)
			if err == nil {
				t.Logf("wkt: %v", wkt)
				t.Fatal("expected validation error unmarshalling wkt")
			}
			_, err = geom.UnmarshalWKT(wkt, geom.DisableAllValidations)
			if err != nil {
				t.Logf("wkt: %v", wkt)
				t.Errorf("disabling validations still gave an error: %v", err)
			}
		})
	}
}

func TestOmitInvalid(t *testing.T) {
	for i, tt := range []struct {
		input  string
		output string
	}{
		{
			"LINESTRING(1 1)",
			"LINESTRING EMPTY",
		},
		{"LINESTRING Z (1 1 1)", "LINESTRING Z EMPTY"},
		{"LINESTRING M (1 1 1)", "LINESTRING M EMPTY"},
		{"LINESTRING ZM (1 1 1 1)", "LINESTRING ZM EMPTY"},
		{
			"LINESTRING(2 2,2 2)",
			"LINESTRING EMPTY",
		},
		{
			"MULTILINESTRING((3 3))",
			"MULTILINESTRING(EMPTY)",
		},
		{"MULTILINESTRING Z ((3 3 0))", "MULTILINESTRING Z (EMPTY)"},
		{"MULTILINESTRING M ((3 3 0))", "MULTILINESTRING M (EMPTY)"},
		{"MULTILINESTRING ZM ((3 3 0 0))", "MULTILINESTRING ZM (EMPTY)"},
		{
			"MULTILINESTRING((4 4,5 5),(6 6,6 6))",
			"MULTILINESTRING((4 4,5 5),EMPTY)",
		},
		{
			"MULTILINESTRING((7 7,7 7),(8 8,9 9))",
			"MULTILINESTRING(EMPTY,(8 8,9 9))",
		},
		{
			"POLYGON((0 0,1 1,0 1,1 0,0 0))",
			"POLYGON EMPTY",
		},
		{"POLYGON Z ((0 0 0,1 1 0,0 1 0,1 0 0,0 0 0))", "POLYGON Z EMPTY"},
		{"POLYGON M ((0 0 0,1 1 0,0 1 0,1 0 0,0 0 0))", "POLYGON M EMPTY"},
		{"POLYGON ZM ((0 0 0 0,1 1 0 0,0 1 0 0,1 0 0 0,0 0 0 0))", "POLYGON ZM EMPTY"},
		{
			"MULTIPOLYGON(((0 0,1 1,0 1,1 0,0 0)))",
			"MULTIPOLYGON(EMPTY)",
		},
		{
			"MULTIPOLYGON(((0 0,1 1,0 1,1 0,0 0)),((0 0,0 1,1 0,0 0)))",
			"MULTIPOLYGON(EMPTY,((0 0,0 1,1 0,0 0)))",
		},
		{
			"MULTIPOLYGON(((0 0,0 1,1 0,0 0)),((0 0,1 1,0 1,1 0,0 0)))",
			"MULTIPOLYGON(((0 0,0 1,1 0,0 0)),EMPTY)",
		},
		{
			"MULTIPOLYGON(((0 0,0 2,2 2,2 0,0 0)),((1 1,1 3,3 3,3 1,1 1)))",
			"MULTIPOLYGON EMPTY",
		},
		{
			"GEOMETRYCOLLECTION(LINESTRING(2 2,2 2))",
			"GEOMETRYCOLLECTION(LINESTRING EMPTY)",
		},
		{
			"GEOMETRYCOLLECTION(GEOMETRYCOLLECTION(LINESTRING(2 2,2 2)))",
			"GEOMETRYCOLLECTION(GEOMETRYCOLLECTION(LINESTRING EMPTY))",
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			g := geomFromWKT(t, tt.input, geom.OmitInvalid)
			expectGeomEq(t, g, geomFromWKT(t, tt.output))
		})
	}
}

func TestOmitInvalidDirectPointConstruction(t *testing.T) {
	// This test is separate from the cases specified using WKTs because it's
	// not possible to unmarshal an invalid Point WKT due to it not matching
	// the WKT grammar.
	for _, ct := range []geom.CoordinatesType{geom.DimXYZ, geom.DimXYM, geom.DimXYZM} {
		t.Run(ct.String(), func(t *testing.T) {
			coords := geom.Coordinates{
				XY:   geom.XY{X: math.NaN(), Y: math.NaN()},
				Type: ct,
			}
			pt, err := geom.NewPoint(coords, geom.OmitInvalid)
			expectNoErr(t, err)
			expectTrue(t, pt.IsEmpty())
			expectCoordinatesTypeEq(t, pt.CoordinatesType(), ct)
		})
	}
}
