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
	}{
		// PT1: Empty Point
		{"PT1", "POINT EMPTY", "POINT EMPTY"},
		// PT2: Point strictly inside R
		{"PT2", "POINT(3 3)", "POINT(3 3)"},
		// PT3: Point strictly outside R
		{"PT3", "POINT(0 0)", "POINT EMPTY"},
		// PT4: Point on edge of R (on left edge, x=1)
		{"PT4", "POINT(1 3)", "POINT(1 3)"},
		// PT5: Point on corner of R (bottom-left corner)
		{"PT5", "POINT(1 2)", "POINT(1 2)"},

		// MP1: Empty MultiPoint
		{"MP1", "MULTIPOINT EMPTY", "MULTIPOINT EMPTY"},
		// MP2: All points inside R
		{"MP2", "MULTIPOINT(2 3,3 3)", "MULTIPOINT(2 3,3 3)"},
		// MP3: All points outside R
		{"MP3", "MULTIPOINT(0 0,6 6)", "MULTIPOINT EMPTY"},
		// MP4: Some points inside, some outside
		{"MP4", "MULTIPOINT(3 3,0 0)", "MULTIPOINT(3 3)"},
		// MP5: Single point inside R
		{"MP5", "MULTIPOINT(3 3)", "MULTIPOINT(3 3)"},
		// MP6: Points on edges and corners of R
		{"MP6", "MULTIPOINT(1 3,1 2)", "MULTIPOINT(1 3,1 2)"},
		// MP7: Mix of inside, on-boundary, and outside
		{"MP7", "MULTIPOINT(3 3,1 3,0 0)", "MULTIPOINT(3 3,1 3)"},
		// MP8: MultiPoint containing empty points
		{"MP8", "MULTIPOINT(3 3,EMPTY)", "MULTIPOINT(3 3)"},
		// MP9: XYZ MultiPoint, all points outside R
		{"MP9", "MULTIPOINT Z(0 0 7,6 6 8)", "MULTIPOINT Z EMPTY"},

		// LS1: Empty LineString
		{"LS1", "LINESTRING EMPTY", "LINESTRING EMPTY"},
		// LS2: Entirely inside R
		{"LS2", "LINESTRING(2 3,4 3)", "LINESTRING(2 3,4 3)"},
		// LS3: Entirely outside R (no overlap)
		{"LS3", "LINESTRING(6 5,7 6)", "LINESTRING EMPTY"},
		// LS4: Entirely outside R on one side (all left of R)
		{"LS4", "LINESTRING(0 3,0 3.5)", "LINESTRING EMPTY"},
		// LS5: Crosses R, entering and exiting once
		{"LS5", "LINESTRING(0 3,2 3,6 5)", "LINESTRING(1 3,2 3,4 4)"},
		// LS6: Crosses R, entering and exiting multiple times
		{"LS6", "LINESTRING(0 3,3 3,3 5,4 5,4 3,6 3)", "MULTILINESTRING((1 3,3 3,3 4),(4 4,4 3,5 3))"},
		// LS7: One endpoint inside, one outside
		{"LS7", "LINESTRING(3 3,6 3)", "LINESTRING(3 3,5 3)"},
		// LS8: One endpoint outside, one inside
		{"LS8", "LINESTRING(0 3,3 3)", "LINESTRING(1 3,3 3)"},
		// LS9: Both endpoints outside, segment passes through R
		{"LS9", "LINESTRING(0 3,6 3)", "LINESTRING(1 3,5 3)"},
		// LS10: Endpoint exactly on edge of R, other inside
		{"LS10", "LINESTRING(1 3,3 3)", "LINESTRING(1 3,3 3)"},
		// LS11: Endpoint exactly on corner of R, other inside
		{"LS11", "LINESTRING(1 2,3 3)", "LINESTRING(1 2,3 3)"},
		// LS12: Both endpoints on boundary of R
		{"LS12", "LINESTRING(1 3,5 3)", "LINESTRING(1 3,5 3)"},
		// LS13: LineString lies entirely along one edge of R
		{"LS13", "LINESTRING(2 2,4 2)", "LINESTRING(2 2,4 2)"},
		// LS14: LineString lies entirely along two adjacent edges (L-shaped)
		{"LS14", "LINESTRING(1 3,1 2,3 2)", "LINESTRING(1 3,1 2,3 2)"},
		// LS15: Segment touches corner of R but does not enter (V-shape)
		{"LS15", "LINESTRING(0 1,1 2,0 3)", "LINESTRING EMPTY"},
		// LS16: Segment touches edge of R tangentially
		{"LS16", "LINESTRING(2 5,3 4,4 5)", "LINESTRING EMPTY"},
		// LS17: Diagonal line crossing two edges of R
		{"LS17", "LINESTRING(0 0,6 6)", "LINESTRING(2 2,4 4)"},
		// LS18: Axis-aligned line crossing two opposite edges of R
		{"LS18", "LINESTRING(3 0,3 6)", "LINESTRING(3 2,3 4)"},
		// LS19: Closed LineString (ring) inside R
		{"LS19", "LINESTRING(2 2.5,4 2.5,3 3.5,2 2.5)", "LINESTRING(2 2.5,4 2.5,3 3.5,2 2.5)"},
		// LS20: Closed LineString (ring) partially overlapping R
		{"LS20", "LINESTRING(2 1,4 1,4 5,2 5,2 1)", "MULTILINESTRING((4 2,4 4),(2 4,2 2))"},
		// LS21: Multi-segment LineString with some segments inside, some outside
		{"LS21", "LINESTRING(3 3,4 3,6 3)", "LINESTRING(3 3,4 3,5 3)"},
		// LS22: Zigzag LineString entering and exiting R many times
		{"LS22", "LINESTRING(0 3,2 3,2 5,3 5,3 3,4 3,4 5,5 5,5 3,7 3)", "MULTILINESTRING((1 3,2 3,2 4),(3 4,3 3,4 3,4 4),(5 4,5 3))"},
		// LS23: Vertex exactly on R boundary, adjacent edges inside
		{"LS23", "LINESTRING(2 3,3 2,4 3)", "LINESTRING(2 3,3 2,4 3)"},
		// LS24: Vertex exactly on R boundary, adjacent edges outside
		{"LS24", "LINESTRING(0 1,1 3,0 5)", "LINESTRING EMPTY"},
		// LS25: LineString passes through two opposite corners of R
		{"LS25", "LINESTRING(0 1.5,6 4.5)", "LINESTRING(1 2,5 4)"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			for _, tr := range d4Transforms {
				t.Run(tr.name, func(t *testing.T) {
					input := geomFromWKT(t, tt.input).TransformXY(tr.fn)
					want := geomFromWKT(t, tt.want).TransformXY(tr.fn)
					r := rect.TransformXY(tr.fn)
					got := geom.ClipByRect(input, r)
					expectGeomEq(t, got, want)
				})
			}
		})
	}
}
