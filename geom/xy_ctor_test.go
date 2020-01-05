package geom_test

import (
	"strconv"
	"testing"

	. "github.com/peterstace/simplefeatures/geom"
)

func TestXYConstructors(t *testing.T) {
	type geometer interface {
		AsGeometry() Geometry
	}
	must := func(t *testing.T) func(
		ctor geometer,
		err error,
	) geometer {
		return func(ctor geometer, err error) geometer {
			if err != nil {
				t.Fatal(err)
			}
			return ctor
		}
	}
	for i, tt := range []struct {
		Geom geometer
		WKT  string
	}{
		{
			NewPointXY(XY{X: 1, Y: 2}),
			"POINT(1 2)",
		},
		{
			must(t)(NewLineXY(XY{1, 2}, XY{3, 4})),
			"LINESTRING(1 2,3 4)",
		},
		{
			must(t)(NewLineStringXY([]XY{{1, 2}, {3, 4}, {5, 6}})),
			"LINESTRING(1 2,3 4,5 6)",
		},
		{
			must(t)(NewPolygonXY([][]XY{
				{{0, 0}, {3, 0}, {3, 3}, {0, 3}, {0, 0}},
				{{1, 1}, {2, 1}, {2, 2}, {1, 2}, {1, 1}},
			})),
			"POLYGON((0 0,3 0,3 3,0 3,0 0),(1 1,2 1,2 2,1 2,1 1))",
		},
		{
			NewMultiPointXY([]XY{{1, 2}, {3, 4}, {5, 6}}),
			"MULTIPOINT(1 2,3 4,5 6)",
		},
		{
			must(t)(NewMultiLineStringXY([][]XY{
				{{1, 2}, {3, 4}, {5, 6}},
				{{7, 8}, {9, 0}},
			})),
			"MULTILINESTRING((1 2,3 4,5 6),(7 8,9 0))",
		},
		{
			must(t)(NewMultiPolygonXY([][][]XY{
				{
					{{0, 0}, {3, 0}, {3, 3}, {0, 3}, {0, 0}},
					{{1, 1}, {2, 1}, {2, 2}, {1, 2}, {1, 1}},
				},
				{
					{{4, 0}, {7, 0}, {7, 3}, {4, 3}, {4, 0}},
					{{5, 1}, {6, 1}, {6, 2}, {5, 2}, {5, 1}},
				},
			})),
			`MULTIPOLYGON(
				((0 0,3 0,3 3,0 3,0 0),(1 1,2 1,2 2,1 2,1 1)),
				((4 0,7 0,7 3,4 3,4 0),(5 1,6 1,6 2,5 2,5 1))
			)`,
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			want := geomFromWKT(t, tt.WKT)
			expectGeomEq(t, tt.Geom.AsGeometry(), want)
		})
	}
}
