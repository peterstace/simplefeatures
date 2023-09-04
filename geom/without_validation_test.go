package geom_test

import (
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
			_, err = geom.UnmarshalWKTWithoutValidation(wkt)
			if err != nil {
				t.Logf("wkt: %v", wkt)
				t.Errorf("disabling validations still gave an error: %v", err)
			}
		})
	}
}
