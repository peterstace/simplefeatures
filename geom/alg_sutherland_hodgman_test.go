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
