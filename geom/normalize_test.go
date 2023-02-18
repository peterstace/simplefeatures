package geom_test

import (
	"math/rand"
	"strconv"
	"testing"

	"github.com/peterstace/simplefeatures/geom"
)

func TestNormalize(t *testing.T) {
	for i, tc := range []struct {
		inputWKT string
		wantWKT  string
	}{
		{"POINT EMPTY", "POINT EMPTY"},
		{"POINT(1 2)", "POINT(1 2)"},

		// MultiPoint with only XY.
		{"MULTIPOINT EMPTY", "MULTIPOINT EMPTY"},
		{"MULTIPOINT(EMPTY)", "MULTIPOINT(EMPTY)"},
		{"MULTIPOINT(1 2)", "MULTIPOINT(1 2)"},
		{"MULTIPOINT(1 2,3 4)", "MULTIPOINT(1 2,3 4)"},
		{"MULTIPOINT(3 4,1 2)", "MULTIPOINT(1 2,3 4)"},
		{"MULTIPOINT(3 4,EMPTY)", "MULTIPOINT(EMPTY,3 4)"},
		{"MULTIPOINT(1 1,1 -1)", "MULTIPOINT(1 -1,1 1)"},
		{"MULTIPOINT(2 1,3 1,1 1)", "MULTIPOINT(1 1,2 1,3 1)"},

		// MultiPoint with 3D and Measure.
		{"MULTIPOINT Z EMPTY", "MULTIPOINT Z EMPTY"},
		{"MULTIPOINT M EMPTY", "MULTIPOINT M EMPTY"},
		{"MULTIPOINT ZM EMPTY", "MULTIPOINT ZM EMPTY"},
		{"MULTIPOINT Z (0 0 1,0 0 0)", "MULTIPOINT Z (0 0 0,0 0 1)"},
		{"MULTIPOINT M (0 0 1,0 0 0)", "MULTIPOINT M (0 0 0,0 0 1)"},
		{"MULTIPOINT ZM (0 0 1 2,0 0 1 1)", "MULTIPOINT ZM (0 0 1 1,0 0 1 2)"},

		// LineString with XY only.
		{"LINESTRING EMPTY", "LINESTRING EMPTY"},
		{"LINESTRING(1 2,3 4)", "LINESTRING(1 2,3 4)"},
		{"LINESTRING(3 4,1 2)", "LINESTRING(1 2,3 4)"},
		{"LINESTRING(1 2,5 6,0 5,3 4)", "LINESTRING(1 2,5 6,0 5,3 4)"},
		{"LINESTRING(3 4,5 6,0 5,1 2)", "LINESTRING(1 2,0 5,5 6,3 4)"},
		{"LINESTRING(0 0,0 1,1 0,0 0)", "LINESTRING(0 0,0 1,1 0,0 0)"},
		{"LINESTRING(0 0,1 0,0 1,0 0)", "LINESTRING(0 0,0 1,1 0,0 0)"},

		// LineString with 3D and Measure.
		{"LINESTRING Z (0 0 1, 1 1 0, 0 0 0)", "LINESTRING Z (0 0 0, 1 1 0, 0 0 1)"},
		{"LINESTRING M (0 0 1, 1 1 0, 0 0 0)", "LINESTRING M (0 0 0, 1 1 0, 0 0 1)"},
		{"LINESTRING ZM (0 0 0 1, 1 1 0 0, 0 0 0 0)", "LINESTRING ZM (0 0 0 0, 1 1 0 0, 0 0 0 1)"},

		{"MULTILINESTRING EMPTY", "MULTILINESTRING EMPTY"},
		{"MULTILINESTRING((3 4,1 2))", "MULTILINESTRING((1 2,3 4))"},
		{"MULTILINESTRING((5 6,7 8),(1 2,3 4))", "MULTILINESTRING((1 2,3 4),(5 6,7 8))"},
		{"MULTILINESTRING((5 6,7 8),EMPTY,(1 2,3 4))", "MULTILINESTRING(EMPTY,(1 2,3 4),(5 6,7 8))"},
		{"MULTILINESTRING((5 6,7 8),(1 2,5 6),(1 2,3 4))", "MULTILINESTRING((1 2,3 4),(1 2,5 6),(5 6,7 8))"},
		{"MULTILINESTRING((5 6,7 8),(1 2,3 4),(1 2,5 6))", "MULTILINESTRING((1 2,3 4),(1 2,5 6),(5 6,7 8))"},

		{"POLYGON EMPTY", "POLYGON EMPTY"},

		// Normalises outer ring orientation:
		{"POLYGON((0 0,0 1,1 0,0 0))", "POLYGON((0 0,1 0,0 1,0 0))"},
		{"POLYGON((0 0,1 0,0 1,0 0))", "POLYGON((0 0,1 0,0 1,0 0))"},

		// Normalises inner ring orientations:
		{
			"POLYGON((0 0,3 0,3 3,0 3,0 0),(1 1,1 2,2 2,2 1,1 1))",
			"POLYGON((0 0,3 0,3 3,0 3,0 0),(1 1,1 2,2 2,2 1,1 1))",
		},
		{
			"POLYGON((0 0,3 0,3 3,0 3,0 0),(1 1,2 1,2 2,1 2,1 1))",
			"POLYGON((0 0,3 0,3 3,0 3,0 0),(1 1,1 2,2 2,2 1,1 1))",
		},

		// Normalises ring starting points:
		{"POLYGON((0 0,1 0,0 1,0 0))", "POLYGON((0 0,1 0,0 1,0 0))"},
		{"POLYGON((1 0,0 1,0 0,1 0))", "POLYGON((0 0,1 0,0 1,0 0))"},
		{"POLYGON((0 1,0 0,1 0,0 1))", "POLYGON((0 0,1 0,0 1,0 0))"},

		// Ring ordering is by lexicographical order rather than order of just
		// the first point:
		{
			`POLYGON(
				(0 0,4 0,4 4,0 4,0 0),
				(1 1,3 1,3 2,1 1),
				(1 1,1 3,2 3,1 1))`,
			`POLYGON(
				(0 0,4 0,4 4,0 4,0 0),
				(1 1,1 3,2 3,1 1),
				(1 1,3 2,3 1,1 1))`,
		},

		// MultiPolygons have their child Polygons ordered by outer ring.
		{
			`MULTIPOLYGON(((1 1,2 1,1 2,1 1)),((0 0,1 0,0 1,0 0)))`,
			`MULTIPOLYGON(((0 0,1 0,0 1,0 0)),((1 1,2 1,1 2,1 1)))`,
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			got := geomFromWKT(t, tc.inputWKT).Normalize()
			expectGeomEqWKT(t, got, tc.wantWKT)
		})
	}
}

func TestNormalizeGeometryCollection(t *testing.T) {
	var geoms []geom.Geometry
	for _, wkt := range []string{
		"POINT EMPTY",
		"POINT(0 0)",
		"POINT(1 1)",
		"POINT(1 2)",
		"MULTIPOINT EMPTY",
		"MULTIPOINT((0 0))",
		"MULTIPOINT(3 4,1 2)",
		"MULTIPOINT(1 2,3 5)",
		"LINESTRING EMPTY",
		"LINESTRING(4 3,1 2)",
		"LINESTRING(1 2,4 5)",
		"MULTILINESTRING EMPTY",
		"MULTILINESTRING((3 4,1 2))",
		"MULTILINESTRING((5 6,7 8),(1 2,5 6),(1 2,3 4))",
		"POLYGON EMPTY",
		"POLYGON((0 0,0 5,5 0,0 0))",
		"POLYGON((0 0,0 5,5 0,0 0),(1 1,1 2,2 1,1 1))",
		"MULTIPOLYGON EMPTY",
		"MULTIPOLYGON(((1 1,2 1,1 2,1 1)))",
		"MULTIPOLYGON(((1 1,2 1,1 2,1 1)),((0 0,1 0,0 1,0 0)))",
		"GEOMETRYCOLLECTION EMPTY",
		"GEOMETRYCOLLECTION(POINT(1 2))",
		"GEOMETRYCOLLECTION(LINESTRING(0 0,1 1),POINT(1 2))",
	} {
		geoms = append(geoms, geomFromWKT(t, wkt))
	}
	rand.Shuffle(len(geoms), func(i, j int) {
		geoms[i], geoms[j] = geoms[j], geoms[i]
	})
	got := geom.NewGeometryCollection(geoms).Normalize().AsGeometry()

	expectGeomEqWKT(t, got, `
		GEOMETRYCOLLECTION(
			POINT EMPTY,
			POINT(0 0),
			POINT(1 1),
			POINT(1 2),
			LINESTRING EMPTY,
			LINESTRING(1 2,4 3),
			LINESTRING(1 2,4 5),
			POLYGON EMPTY,
			POLYGON((0 0,5 0,0 5,0 0)),
			POLYGON((0 0,5 0,0 5,0 0),(1 1,1 2,2 1,1 1)),
			MULTIPOINT EMPTY,
			MULTIPOINT((0 0)),
			MULTIPOINT(1 2,3 4),
			MULTIPOINT(1 2,3 5),
			MULTILINESTRING EMPTY,
			MULTILINESTRING((1 2,3 4)),
			MULTILINESTRING((1 2,3 4),(1 2,5 6),(5 6,7 8)),
			MULTIPOLYGON EMPTY,
			MULTIPOLYGON(((0 0,1 0,0 1,0 0)),((1 1,2 1,1 2,1 1))),
			MULTIPOLYGON(((1 1,2 1,1 2,1 1))),
			GEOMETRYCOLLECTION EMPTY,
			GEOMETRYCOLLECTION(POINT(1 2)),
			GEOMETRYCOLLECTION(POINT(1 2),LINESTRING(0 0,1 1))
		)`,
	)
}
