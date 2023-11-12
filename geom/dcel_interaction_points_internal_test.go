package geom

import (
	"strconv"
	"testing"
)

func TestFindInteractionPoints(t *testing.T) {
	for i, tt := range []struct {
		inputWKTs           []string
		multiPointOutputWKT string
	}{
		// Single LineString cases (all interactions are self-interactions).
		{
			// LineString start and endpoints are included.
			[]string{"LINESTRING(0 0,1 1)"},
			"MULTIPOINT(0 0,1 1)",
		},
		{
			// LineString intermediate points are NOT included.
			[]string{"LINESTRING(0 0,1 1,2 2)"},
			"MULTIPOINT(0 0,2 2)",
		},
		{
			// For closed LineStrings, the start/end point is the same so we
			// just get a single interaction point.
			[]string{"LINESTRING(0 0,0 1,1 0,0 0)"},
			"MULTIPOINT(0 0)",
		},
		{
			// LineStrings self-intersections are interaction points.
			[]string{"LINESTRING(0 0,1 1,2 2,2 0,1 1,0 2)"},
			"MULTIPOINT(0 0,1 1,0 2)",
		},
		{
			// Combination of closed and self-intersecting LineString.
			[]string{"LINESTRING(0 0,1 1,2 2,2 0,1 1,0 2,0 0)"},
			"MULTIPOINT(0 0,1 1)",
		},
		{
			// Self intersections at endpoints are interaction points.
			[]string{"LINESTRING(0 0,1 1,2 2,2 0,1 1)"},
			"MULTIPOINT(0 0,1 1)",
		},
		{
			// LineStrings that reverse back on themselves have an interaction
			// point at the at the reversal point.
			[]string{"LINESTRING(0 0,1 1,0 0)"},
			"MULTIPOINT(0 0,1 1)",
		},
		{
			// Even when a point appears multiple times, it's NOT an action
			// point if the prev/next points are the same.
			[]string{"LINESTRING(0 0,1 1,2 2,1 1,0 0)"},
			"MULTIPOINT(0 0,2 2)",
		},
		{
			// Bowtie shape (brings a few cases together).
			[]string{"LINESTRING(0 0,1 0,2 0,3 1,3 -1,2 0,1 0,0 0,-1 -1,-1 1,0 0)"},
			"MULTIPOINT(0 0,2 0)",
		},

		// Interaction between multiple LineStrings.
		{
			[]string{"MULTILINESTRING((0 0,1 1,2 2),(0 2,1 1,2 0))"},
			"MULTIPOINT(0 0,1 1,2 2,0 2,2 0)",
		},
		{
			[]string{"LINESTRING(0 0,1 1,2 2)", "LINESTRING(0 2,1 1,2 0)"},
			"MULTIPOINT(0 0,1 1,2 2,0 2,2 0)",
		},
		{
			[]string{"LINESTRING(0 0,0 1,1 1,2 1,2 0)", "LINESTRING(0 2,0 1,1 1,2 1,2 2)"},
			"MULTIPOINT(0 0,0 1,0 2,2 0,2 1,2 2)",
		},

		// Point/MultiPoint cases.
		{
			[]string{"POINT(1 2)"},
			"MULTIPOINT(1 2)",
		},
		{
			[]string{"POINT(1 2)", "POINT(1 2)"},
			"MULTIPOINT(1 2)",
		},
		{
			[]string{"POINT(1 2)", "POINT(2 1)"},
			"MULTIPOINT(1 2,2 1)",
		},
		{
			[]string{"MULTIPOINT(1 2,1 2)"},
			"MULTIPOINT(1 2)",
		},
		{
			[]string{"MULTIPOINT(1 2,2 1)"},
			"MULTIPOINT(1 2,2 1)",
		},
		{
			[]string{"MULTIPOINT(1 2,EMPTY)"},
			"MULTIPOINT(1 2)",
		},
		{
			[]string{"MULTIPOINT(EMPTY,1 2)"},
			"MULTIPOINT(1 2)",
		},

		// Points introduce interaction points where there wouldn't normally be one.
		{
			[]string{"LINESTRING(0 0,1 1,2 2)", "POINT(1 1)"},
			"MULTIPOINT(0 0,1 1,2 2)",
		},
		{
			[]string{"POINT(1 1)", "LINESTRING(0 0,1 1,2 2)"},
			"MULTIPOINT(0 0,1 1,2 2)",
		},

		// Polygons and MultiPolygons work the same way as LineStrings, but
		// using the boundaries.
		{
			[]string{"POLYGON((0 0,3 0,3 3,0 3,0 0),(2 1,3 3,1 2,2 1))"},
			"MULTIPOINT(0 0,2 1,3 3)",
		},
		{
			[]string{"MULTIPOLYGON(((0 0,3 0,3 3,0 3,0 0)),((4 1,5 2,3 3,4 1)))"},
			"MULTIPOINT(0 0,4 1,3 3)",
		},

		// Multiple Polygons
		{
			[]string{"POLYGON((0 0,0 2,1 2,2 2,2 1,2 0,0 0))", "POLYGON((1 1,2 1,3 1,3 3,1 3,1 2,1 1))"},
			"MULTIPOINT(0 0,1 1,1 2,2 1)",
		},
		{
			[]string{"POLYGON((0 0,0 1,1 1,2 1,2 0,1 0,0 0))", "POLYGON((0 1,0 2,1 2,2 2,2 1,1 1,0 1))"},
			"MULTIPOINT(0 0,0 1,2 1)",
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			inputs := make([]Geometry, len(tt.inputWKTs))
			for i, wkt := range tt.inputWKTs {
				var err error
				inputs[i], err = UnmarshalWKT(wkt)
				if err != nil {
					t.Fatal(err)
				}
			}

			want, err := UnmarshalWKT(tt.multiPointOutputWKT)
			if err != nil {
				t.Fatal(err)
			}

			gotXYs := findInteractionPoints(inputs)
			var gotPoints []Point
			for xy := range gotXYs {
				gotPoints = append(gotPoints, xy.AsPoint())
			}
			got := NewMultiPoint(gotPoints).AsGeometry()

			if !ExactEquals(want, got, IgnoreOrder) {
				for _, input := range tt.inputWKTs {
					t.Logf("input: %v", input)
				}
				t.Logf("want:  %v", tt.multiPointOutputWKT)
				t.Logf("got:   %v", got.AsText())
				t.Error("mismatch")
			}
		})
	}
}
