package geom_test

import (
	"testing"

	"github.com/peterstace/simplefeatures/geom"
)

func TestClipByRect(t *testing.T) {
	type d4Transform struct {
		name string
		fn   func(geom.XY) geom.XY
	}
	d4Transforms := []d4Transform{
		{"identity", func(xy geom.XY) geom.XY { return xy }},
		{"rot90", func(xy geom.XY) geom.XY { return geom.XY{X: -xy.Y, Y: xy.X} }},
		{"rot180", func(xy geom.XY) geom.XY { return geom.XY{X: -xy.X, Y: -xy.Y} }},
		{"rot270", func(xy geom.XY) geom.XY { return geom.XY{X: xy.Y, Y: -xy.X} }},
		{"reflect_x", func(xy geom.XY) geom.XY { return geom.XY{X: -xy.X, Y: xy.Y} }},
		{"reflect_y", func(xy geom.XY) geom.XY { return geom.XY{X: xy.X, Y: -xy.Y} }},
		{"reflect_diag", func(xy geom.XY) geom.XY { return geom.XY{X: xy.Y, Y: xy.X} }},
		{"reflect_anti", func(xy geom.XY) geom.XY { return geom.XY{X: -xy.Y, Y: -xy.X} }},
	}
	// R is a non-square rectangle so that the D4 transforms produce distinct
	// configurations.
	rect := geom.NewEnvelope(geom.XY{X: 1, Y: 2}, geom.XY{X: 5, Y: 4})

	for _, tt := range []struct {
		name  string
		input string
		want  string
		opts  []geom.ExactEqualsOption
	}{
		// PT1: Empty Point
		{"PT1", "POINT EMPTY", "POINT EMPTY", nil},
		// PT2: Point strictly inside R
		{"PT2", "POINT(3 3)", "POINT(3 3)", nil},
		// PT3: Point strictly outside R
		{"PT3", "POINT(0 0)", "POINT EMPTY", nil},
		// PT4: Point on edge of R (on left edge, x=1)
		{"PT4", "POINT(1 3)", "POINT(1 3)", nil},
		// PT5: Point on corner of R (bottom-left corner)
		{"PT5", "POINT(1 2)", "POINT(1 2)", nil},

		// MP1: Empty MultiPoint
		{"MP1", "MULTIPOINT EMPTY", "MULTIPOINT EMPTY", nil},
		// MP2: All points inside R
		{"MP2", "MULTIPOINT(2 3,3 3)", "MULTIPOINT(2 3,3 3)", nil},
		// MP3: All points outside R
		{"MP3", "MULTIPOINT(0 0,6 6)", "MULTIPOINT EMPTY", nil},
		// MP4: Some points inside, some outside
		{"MP4", "MULTIPOINT(3 3,0 0)", "MULTIPOINT(3 3)", nil},
		// MP5: Single point inside R
		{"MP5", "MULTIPOINT(3 3)", "MULTIPOINT(3 3)", nil},
		// MP6: Points on edges and corners of R
		{"MP6", "MULTIPOINT(1 3,1 2)", "MULTIPOINT(1 3,1 2)", nil},
		// MP7: Mix of inside, on-boundary, and outside
		{"MP7", "MULTIPOINT(3 3,1 3,0 0)", "MULTIPOINT(3 3,1 3)", nil},
		// MP8: MultiPoint containing empty points
		{"MP8", "MULTIPOINT(3 3,EMPTY)", "MULTIPOINT(3 3)", nil},
		// MP9: XYZ MultiPoint, all points outside R
		{"MP9", "MULTIPOINT Z(0 0 7,6 6 8)", "MULTIPOINT Z EMPTY", nil},

		// LS1: Empty LineString
		{"LS1", "LINESTRING EMPTY", "LINESTRING EMPTY", nil},
		// LS2: Entirely inside R
		{"LS2", "LINESTRING(2 3,4 3)", "LINESTRING(2 3,4 3)", nil},
		// LS3: Entirely outside R (no overlap)
		{"LS3", "LINESTRING(6 5,7 6)", "LINESTRING EMPTY", nil},
		// LS4: Entirely outside R on one side (all left of R)
		{"LS4", "LINESTRING(0 3,0 3.5)", "LINESTRING EMPTY", nil},
		// LS5: Crosses R, entering and exiting once
		{"LS5", "LINESTRING(0 3,2 3,6 5)", "LINESTRING(1 3,2 3,4 4)", nil},
		// LS6: Crosses R, entering and exiting multiple times
		{"LS6", "LINESTRING(0 3,3 3,3 5,4 5,4 3,6 3)", "MULTILINESTRING((1 3,3 3,3 4),(4 4,4 3,5 3))", nil},
		// LS7: One endpoint inside, one outside
		{"LS7", "LINESTRING(3 3,6 3)", "LINESTRING(3 3,5 3)", nil},
		// LS8: One endpoint outside, one inside
		{"LS8", "LINESTRING(0 3,3 3)", "LINESTRING(1 3,3 3)", nil},
		// LS9: Both endpoints outside, segment passes through R
		{"LS9", "LINESTRING(0 3,6 3)", "LINESTRING(1 3,5 3)", nil},
		// LS10: Endpoint exactly on edge of R, other inside
		{"LS10", "LINESTRING(1 3,3 3)", "LINESTRING(1 3,3 3)", nil},
		// LS11: Endpoint exactly on corner of R, other inside
		{"LS11", "LINESTRING(1 2,3 3)", "LINESTRING(1 2,3 3)", nil},
		// LS12: Both endpoints on boundary of R
		{"LS12", "LINESTRING(1 3,5 3)", "LINESTRING(1 3,5 3)", nil},
		// LS13: LineString lies entirely along one edge of R
		{"LS13", "LINESTRING(2 2,4 2)", "LINESTRING(2 2,4 2)", nil},
		// LS14: LineString lies entirely along two adjacent edges (L-shaped)
		{"LS14", "LINESTRING(1 3,1 2,3 2)", "LINESTRING(1 3,1 2,3 2)", nil},
		// LS15: Segment touches corner of R but does not enter (V-shape)
		{"LS15", "LINESTRING(0 1,1 2,0 3)", "LINESTRING EMPTY", nil},
		// LS16: Segment touches edge of R tangentially
		{"LS16", "LINESTRING(2 5,3 4,4 5)", "LINESTRING EMPTY", nil},
		// LS17: Diagonal line crossing two edges of R
		{"LS17", "LINESTRING(0 0,6 6)", "LINESTRING(2 2,4 4)", nil},
		// LS18: Axis-aligned line crossing two opposite edges of R
		{"LS18", "LINESTRING(3 0,3 6)", "LINESTRING(3 2,3 4)", nil},
		// LS19: Closed LineString (ring) inside R
		{"LS19", "LINESTRING(2 2.5,4 2.5,3 3.5,2 2.5)", "LINESTRING(2 2.5,4 2.5,3 3.5,2 2.5)", nil},
		// LS20: Closed LineString (ring) partially overlapping R
		{"LS20", "LINESTRING(2 1,4 1,4 5,2 5,2 1)", "MULTILINESTRING((4 2,4 4),(2 4,2 2))", nil},
		// LS21: Multi-segment LineString with some segments inside, some outside
		{"LS21", "LINESTRING(3 3,4 3,6 3)", "LINESTRING(3 3,4 3,5 3)", nil},
		// LS22: Zigzag LineString entering and exiting R many times
		{"LS22", "LINESTRING(0 3,2 3,2 5,3 5,3 3,4 3,4 5,5 5,5 3,7 3)", "MULTILINESTRING((1 3,2 3,2 4),(3 4,3 3,4 3,4 4),(5 4,5 3))", nil},
		// LS23: Vertex exactly on R boundary, adjacent edges inside
		{"LS23", "LINESTRING(2 3,3 2,4 3)", "LINESTRING(2 3,3 2,4 3)", nil},
		// LS24: Vertex exactly on R boundary, adjacent edges outside
		{"LS24", "LINESTRING(0 1,1 3,0 5)", "LINESTRING EMPTY", nil},
		// LS25: LineString passes through two opposite corners of R
		{"LS25", "LINESTRING(0 1.5,6 4.5)", "LINESTRING(1 2,5 4)", nil},

		// MLS1: Empty MultiLineString
		{"MLS1", "MULTILINESTRING EMPTY", "MULTILINESTRING EMPTY", nil},
		// MLS2: All component LineStrings inside R
		{"MLS2", "MULTILINESTRING((2 3,4 3),(3 2.5,3 3.5))", "MULTILINESTRING((2 3,4 3),(3 2.5,3 3.5))", nil},
		// MLS3: All component LineStrings outside R
		{"MLS3", "MULTILINESTRING((6 5,7 6),(0 0,0 1))", "MULTILINESTRING EMPTY", nil},
		// MLS4: Some components inside, some outside
		{"MLS4", "MULTILINESTRING((2 3,4 3),(6 5,7 6))", "MULTILINESTRING((2 3,4 3))", nil},
		// MLS5: Single component, clipped to a single segment
		{"MLS5", "MULTILINESTRING((0 3,6 3))", "MULTILINESTRING((1 3,5 3))", nil},
		// MLS6: One component crosses R, another is inside R
		{"MLS6", "MULTILINESTRING((0 3,6 3),(3 2.5,3 3.5))", "MULTILINESTRING((1 3,5 3),(3 2.5,3 3.5))", nil},
		// MLS7: Component that produces multiple fragments when clipped
		{"MLS7", "MULTILINESTRING((0 3,3 3,3 5,4 5,4 3,6 3))", "MULTILINESTRING((1 3,3 3,3 4),(4 4,4 3,5 3))", nil},
		// MLS8: MultiLineString containing empty LineStrings
		{"MLS8", "MULTILINESTRING((2 3,4 3),EMPTY)", "MULTILINESTRING((2 3,4 3))", nil},
		// MLS9: XYZ MultiLineString, all components outside R
		{"MLS9", "MULTILINESTRING Z((6 5 1,7 6 2),(0 0 3,0 1 4))", "MULTILINESTRING Z EMPTY", nil},

		// PG1: Empty Polygon
		{"PG1", "POLYGON EMPTY", "POLYGON EMPTY", nil},
		// PG2: Entirely inside R
		{"PG2", "POLYGON((2 2.5,4 2.5,4 3.5,2 3.5,2 2.5))", "POLYGON((2 2.5,4 2.5,4 3.5,2 3.5,2 2.5))", []geom.ExactEqualsOption{geom.IgnoreOrder}},
		// PG3: Entirely outside R (no overlap)
		{"PG3", "POLYGON((6 5,8 5,8 7,6 7,6 5))", "POLYGON EMPTY", nil},
		// PG4: R entirely inside Polygon
		{"PG4", "POLYGON((0 0,6 0,6 6,0 6,0 0))", "POLYGON((1 2,5 2,5 4,1 4,1 2))", []geom.ExactEqualsOption{geom.IgnoreOrder}},
		// PG5: Partially overlapping one edge (left) of R
		{"PG5", "POLYGON((0 2.5,3 2.5,3 3.5,0 3.5,0 2.5))", "POLYGON((1 2.5,3 2.5,3 3.5,1 3.5,1 2.5))", []geom.ExactEqualsOption{geom.IgnoreOrder}},
		// PG6: Partially overlapping two adjacent edges (bottom-left corner clip)
		{"PG6", "POLYGON((0 1,3 1,3 3,0 3,0 1))", "POLYGON((1 2,3 2,3 3,1 3,1 2))", []geom.ExactEqualsOption{geom.IgnoreOrder}},
		// PG7: Partially overlapping two opposite edges (strip left-right)
		{"PG7", "POLYGON((0 2.5,6 2.5,6 3.5,0 3.5,0 2.5))", "POLYGON((1 2.5,5 2.5,5 3.5,1 3.5,1 2.5))", []geom.ExactEqualsOption{geom.IgnoreOrder}},
		// PG8: Partially overlapping three edges (left, bottom, right)
		{"PG8", "POLYGON((0 1,6 1,6 3,0 3,0 1))", "POLYGON((1 2,5 2,5 3,1 3,1 2))", []geom.ExactEqualsOption{geom.IgnoreOrder}},
		// PG9: Polygon shares entire bottom edge with R
		{"PG9", "POLYGON((1 2,5 2,5 3,1 3,1 2))", "POLYGON((1 2,5 2,5 3,1 3,1 2))", []geom.ExactEqualsOption{geom.IgnoreOrder}},
		// PG10: Polygon vertex exactly on R left edge (x=1)
		{"PG10", "POLYGON((1 3,2 2.5,2 3.5,1 3))", "POLYGON((1 3,2 2.5,2 3.5,1 3))", []geom.ExactEqualsOption{geom.IgnoreOrder}},
		// PG11: Polygon vertex exactly on R corner (1,2)
		{"PG11", "POLYGON((1 2,3 2.5,3 3.5,1 2))", "POLYGON((1 2,3 2.5,3 3.5,1 2))", []geom.ExactEqualsOption{geom.IgnoreOrder}},
		// PG12: Polygon edge collinear with R bottom edge, polygon inside R
		{"PG12", "POLYGON((2 2,4 2,3 3,2 2))", "POLYGON((2 2,4 2,3 3,2 2))", []geom.ExactEqualsOption{geom.IgnoreOrder}},
		// PG13: Polygon edge collinear with R bottom edge, polygon outside R
		{"PG13", "POLYGON((2 2,4 2,3 1,2 2))", "POLYGON EMPTY", nil},
		// PG14: Polygon touches R at single point on left edge (vertex-to-edge)
		{"PG14", "POLYGON((0 2.5,1 3,0 3.5,0 2.5))", "POLYGON EMPTY", nil},
		// PG15: Polygon touches R at single corner point (1,2)
		{"PG15", "POLYGON((0 1,1 2,0 3,0 1))", "POLYGON EMPTY", nil},
		// PG16: Convex polygon (large square) clipped at all 4 edges
		{"PG16", "POLYGON((-1 0,7 0,7 6,-1 6,-1 0))", "POLYGON((1 2,5 2,5 4,1 4,1 2))", []geom.ExactEqualsOption{geom.IgnoreOrder}},
		// PG17: Concave polygon, concavity inside R
		{"PG17", "POLYGON((2 2.5,4 2.5,4 3.5,3 3,2 3.5,2 2.5))", "POLYGON((2 2.5,4 2.5,4 3.5,3 3,2 3.5,2 2.5))", []geom.ExactEqualsOption{geom.IgnoreOrder}},
		// PG18: Concave polygon, concavity facing left R boundary
		{"PG18", "POLYGON((0 2.5,3 2.5,3 3,2 3,2 3.5,0 3.5,0 2.5))", "POLYGON((1 2.5,3 2.5,3 3,2 3,2 3.5,1 3.5,1 2.5))", []geom.ExactEqualsOption{geom.IgnoreOrder}},
		// PG19: C-shaped polygon, concavity crosses left edge → MultiPolygon
		{"PG19",
			"POLYGON((0 2.5,3 2.5,3 2.8,0.5 2.8,0.5 3.2,3 3.2,3 3.5,0 3.5,0 2.5))",
			"MULTIPOLYGON(((1 2.5,3 2.5,3 2.8,1 2.8,1 2.5)),((1 3.2,3 3.2,3 3.5,1 3.5,1 3.2)))",
			[]geom.ExactEqualsOption{geom.IgnoreOrder}},
		// PG20: Very thin sliver polygon partially inside R
		{"PG20", "POLYGON((0 2.999,6 2.999,6 3.001,0 3.001,0 2.999))", "POLYGON((1 2.999,5 2.999,5 3.001,1 3.001,1 2.999))", []geom.ExactEqualsOption{geom.IgnoreOrder}},
		// PG21: Triangle clipped at all 4 rect edges producing a polygon
		{"PG21", "POLYGON((3 0,7 3,3 6,3 0))", "POLYGON((3 4,3 2,5 2,5 4,3 4))", []geom.ExactEqualsOption{geom.IgnoreOrder}},
		// PG22: Counter-clockwise exterior ring (explicitly CCW, same as PG2 but winding verified by D4 reflections)
		{"PG22", "POLYGON((2 2.5,2 3.5,4 3.5,4 2.5,2 2.5))", "POLYGON((2 2.5,2 3.5,4 3.5,4 2.5,2 2.5))", []geom.ExactEqualsOption{geom.IgnoreOrder}},

		// PH1: Exterior and hole both inside R
		{"PH1",
			"POLYGON((2 2.5,4 2.5,4 3.5,2 3.5,2 2.5),(2.5 2.8,2.5 3.2,3.5 3.2,3.5 2.8,2.5 2.8))",
			"POLYGON((2 2.5,4 2.5,4 3.5,2 3.5,2 2.5),(2.5 2.8,2.5 3.2,3.5 3.2,3.5 2.8,2.5 2.8))",
			[]geom.ExactEqualsOption{geom.IgnoreOrder}},
		// PH2: Exterior clipped, hole entirely inside clipped region
		{"PH2",
			"POLYGON((0 2.5,4 2.5,4 3.5,0 3.5,0 2.5),(2 2.8,2 3.2,3 3.2,3 2.8,2 2.8))",
			"POLYGON((1 2.5,4 2.5,4 3.5,1 3.5,1 2.5),(2 2.8,2 3.2,3 3.2,3 2.8,2 2.8))",
			[]geom.ExactEqualsOption{geom.IgnoreOrder}},
		// PH3: Multiple holes, all inside R
		{"PH3",
			"POLYGON((0 0,6 0,6 6,0 6,0 0),(2 2.5,2 3,3 3,3 2.5,2 2.5),(3.5 2.5,3.5 3,4.5 3,4.5 2.5,3.5 2.5))",
			"POLYGON((1 2,5 2,5 4,1 4,1 2),(2 2.5,2 3,3 3,3 2.5,2 2.5),(3.5 2.5,3.5 3,4.5 3,4.5 2.5,3.5 2.5))",
			[]geom.ExactEqualsOption{geom.IgnoreOrder}},
		// PH4: Hole entirely outside R (inside exterior outside R)
		{"PH4",
			"POLYGON((0 0,6 0,6 6,0 6,0 0),(0.2 0.2,0.2 0.8,0.8 0.8,0.8 0.2,0.2 0.2))",
			"POLYGON((1 2,5 2,5 4,1 4,1 2))",
			[]geom.ExactEqualsOption{geom.IgnoreOrder}},
		// PH5: Hole in part of exterior that is clipped away
		{"PH5",
			"POLYGON((2 0,4 0,4 3,2 3,2 0),(2.5 0.5,2.5 1.5,3.5 1.5,3.5 0.5,2.5 0.5))",
			"POLYGON((2 2,4 2,4 3,2 3,2 2))",
			[]geom.ExactEqualsOption{geom.IgnoreOrder}},
		// PH6: Hole crosses one edge of R (left edge)
		{"PH6",
			"POLYGON((-2 -2,8 -2,8 8,-2 8,-2 -2),(0 2.5,0 3.5,2 3.5,2 2.5,0 2.5))",
			"POLYGON((1 4,1 3.5,2 3.5,2 2.5,1 2.5,1 2,5 2,5 4,1 4))",
			[]geom.ExactEqualsOption{geom.IgnoreOrder}},
		// PH7: Hole crosses two adjacent edges of R (bottom-left corner)
		{"PH7",
			"POLYGON((-2 -2,8 -2,8 8,-2 8,-2 -2),(0 1,0 3,2 3,2 1,0 1))",
			"POLYGON((1 3,2 3,2 2,5 2,5 4,1 4,1 3))",
			[]geom.ExactEqualsOption{geom.IgnoreOrder}},
		// PH8: Hole crosses two opposite edges of R (left and right) — splits polygon
		{"PH8",
			"POLYGON((-2 -2,8 -2,8 8,-2 8,-2 -2),(0 2.8,0 3.2,6 3.2,6 2.8,0 2.8))",
			"MULTIPOLYGON(((1 2,5 2,5 2.8,1 2.8,1 2)),((1 3.2,5 3.2,5 4,1 4,1 3.2)))",
			[]geom.ExactEqualsOption{geom.IgnoreOrder}},
		// PH9: Hole crosses all four edges of R, leaving top and bottom strips
		{"PH9",
			"POLYGON((-2 -2,8 -2,8 8,-2 8,-2 -2),(0.5 2.5,0.5 3.5,5.5 3.5,5.5 2.5,0.5 2.5))",
			"MULTIPOLYGON(((1 2,5 2,5 2.5,1 2.5,1 2)),((1 3.5,5 3.5,5 4,1 4,1 3.5)))",
			[]geom.ExactEqualsOption{geom.IgnoreOrder}},
		// PH10: Hole on R boundary, exterior extends beyond R — becomes concavity
		{"PH10",
			"POLYGON((-2 -2,8 -2,8 8,-2 8,-2 -2),(1 2.5,1 3.5,2 3.5,2 2.5,1 2.5))",
			"POLYGON((1 4,1 3.5,2 3.5,2 2.5,1 2.5,1 2,5 2,5 4,1 4))",
			[]geom.ExactEqualsOption{geom.IgnoreOrder}},
		// PH11: One hole inside R, another outside R
		{"PH11",
			"POLYGON((-2 -2,8 -2,8 8,-2 8,-2 -2),(2.5 2.8,2.5 3.2,3.5 3.2,3.5 2.8,2.5 2.8),(-1.5 -1.5,-1.5 -0.5,-0.5 -0.5,-0.5 -1.5,-1.5 -1.5))",
			"POLYGON((1 2,5 2,5 4,1 4,1 2),(2.5 2.8,2.5 3.2,3.5 3.2,3.5 2.8,2.5 2.8))",
			[]geom.ExactEqualsOption{geom.IgnoreOrder}},
		// PH12: One hole inside R, another splits polygon
		{"PH12",
			"POLYGON((-2 -2,8 -2,8 8,-2 8,-2 -2),(2.5 2.2,2.5 2.4,3.5 2.4,3.5 2.2,2.5 2.2),(0 2.5,0 3.5,6 3.5,6 2.5,0 2.5))",
			"MULTIPOLYGON(((1 2,5 2,5 2.5,1 2.5,1 2),(2.5 2.2,2.5 2.4,3.5 2.4,3.5 2.2,2.5 2.2)),((1 3.5,5 3.5,5 4,1 4,1 3.5)))",
			[]geom.ExactEqualsOption{geom.IgnoreOrder}},
		// PH13: Multiple holes each crossing one edge of R (left edge)
		{"PH13",
			"POLYGON((-2 -2,8 -2,8 8,-2 8,-2 -2),(0 2.2,0 2.6,2 2.6,2 2.2,0 2.2),(0 3.4,0 3.8,2 3.8,2 3.4,0 3.4))",
			"POLYGON((1 4,1 3.8,2 3.8,2 3.4,1 3.4,1 2.6,2 2.6,2 2.2,1 2.2,1 2,5 2,5 4,1 4))",
			[]geom.ExactEqualsOption{geom.IgnoreOrder}},
		// PH14: Two holes each crossing two opposite edges, splitting into three pieces
		{"PH14",
			"POLYGON((-2 -2,8 -2,8 8,-2 8,-2 -2),(0 2.5,0 2.8,6 2.8,6 2.5,0 2.5),(0 3.2,0 3.5,6 3.5,6 3.2,0 3.2))",
			"MULTIPOLYGON(((1 2,5 2,5 2.5,1 2.5,1 2)),((1 2.8,5 2.8,5 3.2,1 3.2,1 2.8)),((1 3.5,5 3.5,5 4,1 4,1 3.5)))",
			[]geom.ExactEqualsOption{geom.IgnoreOrder}},
		// PH15: Hole covers entire clipped area
		{"PH15",
			"POLYGON((-2 -2,8 -2,8 8,-2 8,-2 -2),(0.5 1.5,0.5 4.5,5.5 4.5,5.5 1.5,0.5 1.5))",
			"POLYGON EMPTY",
			nil},
	} {
		t.Run(tt.name, func(t *testing.T) {
			for _, tr := range d4Transforms {
				t.Run(tr.name, func(t *testing.T) {
					input := geomFromWKT(t, tt.input).TransformXY(tr.fn)
					want := geomFromWKT(t, tt.want).TransformXY(tr.fn)
					r := rect.TransformXY(tr.fn)
					got := geom.ClipByRect(input, r)
					expectGeomEq(t, got, want, tt.opts...)
				})
			}
		})
	}
}
