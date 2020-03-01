package geom_test

import (
	"strconv"
	"testing"

	. "github.com/peterstace/simplefeatures/internal/geomtest"
)

func TestConvexHull(t *testing.T) {
	for i, tt := range []struct {
		input  string
		output string
	}{
		// The following tests exercise the 'plumbing' between geometry
		// types and the convex hull algorithm.

		{
			input:  "POINT EMPTY",
			output: "POINT EMPTY",
		},
		{
			input:  "POINT(1 2)",
			output: "POINT(1 2)",
		},
		{
			input:  "LINESTRING EMPTY",
			output: "LINESTRING EMPTY",
		},
		{
			input:  "LINESTRING(1 2,3 4)",
			output: "LINESTRING(1 2,3 4)",
		},
		{
			input:  "LINESTRING(0 0,1 1,1 0,0 1)",
			output: "POLYGON((0 0,1 0,1 1,0 1,0 0))",
		},
		{
			input:  "POLYGON((0 0,0 1,1 0,0 0))",
			output: "POLYGON((0 0,1 0,0 1,0 0))",
		},
		{
			input:  "POLYGON EMPTY",
			output: "POLYGON EMPTY",
		},
		{
			input:  "MULTIPOINT(0 0,0 1,1 0)",
			output: "POLYGON((0 0,1 0,0 1,0 0))",
		},
		{
			input:  "MULTIPOINT EMPTY",
			output: "MULTIPOINT EMPTY",
		},
		{
			input:  "MULTIPOINT(1 2,3 4,EMPTY)",
			output: "LINESTRING(1 2,3 4)",
		},
		{
			input:  "MULTILINESTRING EMPTY",
			output: "MULTILINESTRING EMPTY",
		},
		{
			input:  "MULTILINESTRING((0 1,2 3))",
			output: "LINESTRING(0 1,2 3)",
		},
		{
			input:  "MULTIPOLYGON EMPTY",
			output: "MULTIPOLYGON EMPTY",
		},
		{
			input:  "MULTIPOLYGON(((0 0,1 0,0 1,0 0)))",
			output: "POLYGON((0 0,1 0,0 1,0 0))",
		},
		{
			input:  "GEOMETRYCOLLECTION EMPTY",
			output: "GEOMETRYCOLLECTION EMPTY",
		},
		{
			input:  "GEOMETRYCOLLECTION(POINT(1 2))",
			output: "POINT(1 2)",
		},
		{
			input:  "GEOMETRYCOLLECTION(POINT EMPTY,POINT(1 2))",
			output: "POINT(1 2)",
		},
		{
			input:  "GEOMETRYCOLLECTION(GEOMETRYCOLLECTION(POINT(1 2)),POINT(2 1))",
			output: "LINESTRING(2 1,1 2)",
		},
		{
			input:  "GEOMETRYCOLLECTION(LINESTRING(1 2,3 4))",
			output: "LINESTRING(1 2,3 4)",
		},

		// The following tests exercise various cases in the covex hull
		// algorithm itself.

		{
			// 2 points - duplicated.
			// (2 points - distinct case is already covered by plumbing tests).
			input:  "MULTIPOINT(1 2,1 2)",
			output: "POINT(1 2)",
		},
		{
			// 3 points - colinear and distinct
			input:  "MULTIPOINT(2 1,4 2,0 0)",
			output: "LINESTRING(0 0,4 2)",
		},
		{
			// 3 points - non-colinear and distinct - counterclockwise case
			input:  "MULTIPOINT(1 2,2 2,3 4)",
			output: "POLYGON((1 2,2 2,3 4,1 2))",
		},
		{
			// 3 points - non-colinear and distinct - clockwise case
			input:  "MULTIPOINT(2 1,2 2,4 3)",
			output: "POLYGON((2 1,4 3,2 2,2 1))",
		},
		{
			// 3 points - two are coincident
			input:  "MULTIPOINT(2 1,3 6,2 1)",
			output: "LINESTRING(2 1,3 6)",
		},
		{
			// 3 points - all are coincident
			input:  "MULTIPOINT(3 8,3 8,3 8)",
			output: "POINT(3 8)",
		},
		{
			// 4 points - aligned square
			input:  "MULTIPOINT(0 0,1 0,0 1,1 1)",
			output: "POLYGON((0 0,1 0,1 1,0 1,0 0))",
		},
		{
			// 4 points - rotated square
			input:  "MULTIPOINT(4 2,2 1,1 3,3 4)",
			output: "POLYGON((2 1,4 2,3 4,1 3,2 1))",
		},
		{
			// 4 points - triangle (two points coincident)
			input:  "MULTIPOINT(1 1,3 1,1 1,2 5)",
			output: "POLYGON((1 1,3 1,2 5,1 1))",
		},
		{
			// 4 points - line (2 pairs of points coincident)
			input:  "MULTIPOINT(2 3,6 7,2 3,6 7)",
			output: "LINESTRING(2 3,6 7)",
		},
		{
			// 4 points - all coincident
			input:  "MULTIPOINT(2 3,2 3,2 3,2 3)",
			output: "POINT(2 3)",
		},
		{
			// 5 points - convex pentagon
			input:  "MULTIPOINT(2 0,0 0,1 3,0 2,2 2)",
			output: "POLYGON((0 0,2 0,2 2,1 3,0 2,0 0))",
		},
		{
			// 5 points - concave pentagon
			input:  "MULTIPOINT(2 0,0 0,1 1,0 2,2 2)",
			output: "POLYGON((0 0,2 0,2 2,0 2,0 0))",
		},
		{
			// no collinear points
			input:  "MULTIPOINT(7 7,8 5,7 2,6 5,5 5,4 6,4 2,3 7,2 1)",
			output: "POLYGON((2 1,7 2,8 5,7 7,3 7,2 1))",
		},
		{
			// there are collinear points",
			input:  "MULTIPOINT(7 7,8 5,7 2,6 5,5 5,4 6,4 2,3 7,2 1,14 9)",
			output: "POLYGON((2 1,7 2,14 9,3 7,2 1))",
		},
		{
			// reproduced a bug
			input:  "MULTIPOINT(0 0,1 1,2 2,1 3,1 0,0 1,2 1,0 2,2 0)",
			output: "POLYGON((0 0,2 0,2 2,1 3,0 2,0 0))",
		},
		{
			// grid of 3x3 points
			input:  "MULTIPOINT(0 0,1 1,2 2,1 2,1 0,0 1,2 1,0 2,2 0)",
			output: "POLYGON((0 0,2 0,2 2,0 2,0 0))",
		},
		{
			// grid of 4x4 points
			input:  "MULTIPOINT(0 0,1 0,2 0,3 0,0 1,1 1,2 1,3 1,0 2,1 2,2 2,3 2,0 3,1 3,2 3,3 3)",
			output: "POLYGON((0 0,3 0,3 3,0 3,0 0))",
		},
		{
			input:  `MULTIPOINT((0.532 0.548),(0.385 0.378),(0.428 0.463),(0.506 0.443),(0.372 0.613),(0.648 0.636),(0.417 0.447))`,
			output: `POLYGON((0.385 0.378,0.506 0.443,0.648 0.636,0.372 0.613,0.385 0.378))`,
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Logf("input: %s", tt.input)
			got := GeomFromWKT(t, tt.input).ConvexHull()
			ExpectGeomEq(t, got, GeomFromWKT(t, tt.output))
		})
	}
}
