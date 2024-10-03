package geom_test

import (
	"strconv"
	"testing"

	"github.com/peterstace/simplefeatures/geom"
)

// TODO: Test panics.

func TestCoordinateConstructors(t *testing.T) {
	for i, tc := range []struct {
		got     interface{ AsGeometry() geom.Geometry }
		wantWKT string
	}{
		{geom.NewPointXY(1, 2), "POINT(1 2)"},
		{geom.NewPointXYZ(1, 2, 3), "POINT Z(1 2 3)"},
		{geom.NewPointXYM(1, 2, 3), "POINT M(1 2 3)"},
		{geom.NewPointXYZM(1, 2, 3, 4), "POINT ZM(1 2 3 4)"},

		{geom.NewMultiPointXY(), "MULTIPOINT EMPTY"},
		{geom.NewMultiPointXY(1, 2), "MULTIPOINT(1 2)"},
		{geom.NewMultiPointXY(1, 2, 3, 4), "MULTIPOINT(1 2,3 4)"},
		{geom.NewMultiPointXYZ(), "MULTIPOINT Z EMPTY"},
		{geom.NewMultiPointXYZ(1, 2, 3), "MULTIPOINT Z(1 2 3)"},
		{geom.NewMultiPointXYZ(1, 2, 3, 4, 5, 6), "MULTIPOINT Z(1 2 3,4 5 6)"},
		{geom.NewMultiPointXYM(), "MULTIPOINT M EMPTY"},
		{geom.NewMultiPointXYM(1, 2, 3), "MULTIPOINT M(1 2 3)"},
		{geom.NewMultiPointXYM(1, 2, 3, 4, 5, 6), "MULTIPOINT M(1 2 3,4 5 6)"},
		{geom.NewMultiPointXYZM(), "MULTIPOINT ZM EMPTY"},
		{geom.NewMultiPointXYZM(1, 2, 3, 4), "MULTIPOINT ZM(1 2 3 4)"},
		{geom.NewMultiPointXYZM(1, 2, 3, 4, 5, 6, 7, 8), "MULTIPOINT ZM(1 2 3 4,5 6 7 8)"},

		{geom.NewLineStringXY(), "LINESTRING EMPTY"},
		{geom.NewLineStringXY(1, 2, 3, 4), "LINESTRING(1 2,3 4)"},
		{geom.NewLineStringXYZ(), "LINESTRING Z EMPTY"},
		{geom.NewLineStringXYZ(1, 2, 3, 4, 5, 6), "LINESTRING Z(1 2 3,4 5 6)"},
		{geom.NewLineStringXYM(), "LINESTRING M EMPTY"},
		{geom.NewLineStringXYM(1, 2, 3, 4, 5, 6), "LINESTRING M(1 2 3,4 5 6)"},
		{geom.NewLineStringXYZM(), "LINESTRING ZM EMPTY"},
		{geom.NewLineStringXYZM(1, 2, 3, 4, 5, 6, 7, 8), "LINESTRING ZM(1 2 3 4,5 6 7 8)"},

		{geom.NewMultiLineStringXY(), "MULTILINESTRING EMPTY"},
		{geom.NewMultiLineStringXY(nil), "MULTILINESTRING(EMPTY)"},
		{geom.NewMultiLineStringXY([]float64{1, 2, 3, 4}), "MULTILINESTRING((1 2,3 4))"},
		{geom.NewMultiLineStringXY([]float64{1, 2, 3, 4}, nil), "MULTILINESTRING((1 2,3 4),EMPTY)"},
		{geom.NewMultiLineStringXY([]float64{1, 2, 3, 4}, []float64{5, 6, 7, 8}), "MULTILINESTRING((1 2,3 4),(5 6,7 8))"},
		{geom.NewMultiLineStringXYZ(), "MULTILINESTRING Z EMPTY"},
		{geom.NewMultiLineStringXYZ(nil), "MULTILINESTRING Z(EMPTY)"},
		{geom.NewMultiLineStringXYZ([]float64{1, 2, 3, 4, 5, 6}), "MULTILINESTRING Z((1 2 3,4 5 6))"},
		{geom.NewMultiLineStringXYZ([]float64{1, 2, 3, 4, 5, 6}, nil), "MULTILINESTRING Z((1 2 3,4 5 6),EMPTY)"},
		{geom.NewMultiLineStringXYZ([]float64{1, 2, 3, 4, 5, 6}, []float64{7, 8, 9, 10, 11, 12}), "MULTILINESTRING Z((1 2 3,4 5 6),(7 8 9,10 11 12))"},
		{geom.NewMultiLineStringXYM(), "MULTILINESTRING M EMPTY"},
		{geom.NewMultiLineStringXYM(nil), "MULTILINESTRING M(EMPTY)"},
		{geom.NewMultiLineStringXYM([]float64{1, 2, 3, 4, 5, 6}), "MULTILINESTRING M((1 2 3,4 5 6))"},
		{geom.NewMultiLineStringXYM([]float64{1, 2, 3, 4, 5, 6}, nil), "MULTILINESTRING M((1 2 3,4 5 6),EMPTY)"},
		{geom.NewMultiLineStringXYM([]float64{1, 2, 3, 4, 5, 6}, []float64{7, 8, 9, 10, 11, 12}), "MULTILINESTRING M((1 2 3,4 5 6),(7 8 9,10 11 12))"},
		{geom.NewMultiLineStringXYZM(), "MULTILINESTRING ZM EMPTY"},
		{geom.NewMultiLineStringXYZM(nil), "MULTILINESTRING ZM(EMPTY)"},
		{geom.NewMultiLineStringXYZM([]float64{1, 2, 3, 4, 5, 6, 7, 8}), "MULTILINESTRING ZM((1 2 3 4,5 6 7 8))"},
		{geom.NewMultiLineStringXYZM([]float64{1, 2, 3, 4, 5, 6, 7, 8}, nil), "MULTILINESTRING ZM((1 2 3 4,5 6 7 8),EMPTY)"},
		{geom.NewMultiLineStringXYZM([]float64{1, 2, 3, 4, 5, 6, 7, 8}, []float64{9, 10, 11, 12, 13, 14, 15, 16}), "MULTILINESTRING ZM((1 2 3 4,5 6 7 8),(9 10 11 12,13 14 15 16))"},

		{geom.NewPolygonXY(), "POLYGON EMPTY"},
		{geom.NewPolygonXY([]float64{0, 0, 0, 4, 4, 4, 4, 0, 0, 0}), "POLYGON((0 0,0 4,4 4,4 0,0 0))"},
		{geom.NewPolygonXY([]float64{0, 0, 0, 4, 4, 4, 4, 0, 0, 0}, []float64{1, 1, 1, 2, 2, 1, 1, 1}), "POLYGON((0 0,0 4,4 4,4 0,0 0),(1 1,1 2,2 1,1 1))"},
		{geom.NewPolygonXYZ(), "POLYGON Z EMPTY"},
		{geom.NewPolygonXYZ([]float64{0, 0, 10, 0, 4, 11, 4, 4, 12, 4, 0, 13, 0, 0, 14}), "POLYGON Z((0 0 10,0 4 11,4 4 12,4 0 13,0 0 14))"},
		{geom.NewPolygonXYZ([]float64{0, 0, 10, 0, 4, 11, 4, 4, 12, 4, 0, 13, 0, 0, 14}, []float64{1, 1, 20, 1, 2, 21, 2, 2, 22, 2, 1, 23, 1, 1, 24}), "POLYGON Z((0 0 10,0 4 11,4 4 12,4 0 13,0 0 14),(1 1 20,1 2 21,2 2 22,2 1 23,1 1 24))"},
		{geom.NewPolygonXYM(), "POLYGON M EMPTY"},
		{geom.NewPolygonXYM([]float64{0, 0, 10, 0, 4, 11, 4, 4, 12, 4, 0, 13, 0, 0, 14}), "POLYGON M((0 0 10,0 4 11,4 4 12,4 0 13,0 0 14))"},
		{geom.NewPolygonXYM([]float64{0, 0, 10, 0, 4, 11, 4, 4, 12, 4, 0, 13, 0, 0, 14}, []float64{1, 1, 20, 1, 2, 21, 2, 2, 22, 2, 1, 23, 1, 1, 24}), "POLYGON M((0 0 10,0 4 11,4 4 12,4 0 13,0 0 14),(1 1 20,1 2 21,2 2 22,2 1 23,1 1 24))"},
		{geom.NewPolygonXYZM(), "POLYGON ZM EMPTY"},
		{geom.NewPolygonXYZM([]float64{0, 0, 10, 100, 0, 4, 11, 101, 4, 4, 12, 102, 4, 0, 13, 103, 0, 0, 14, 104}), "POLYGON ZM((0 0 10 100,0 4 11 101,4 4 12 102,4 0 13 103,0 0 14 104))"},
		{geom.NewPolygonXYZM([]float64{0, 0, 10, 100, 0, 4, 11, 101, 4, 4, 12, 102, 4, 0, 13, 103, 0, 0, 14, 104}, []float64{1, 1, 20, 200, 1, 2, 21, 201, 2, 2, 22, 202, 2, 1, 23, 203, 1, 1, 24, 204}), "POLYGON ZM((0 0 10 100,0 4 11 101,4 4 12 102,4 0 13 103,0 0 14 104),(1 1 20 200,1 2 21 201,2 2 22 202,2 1 23 203,1 1 24 204))"},

		{geom.NewSingleRingPolygonXY(), "POLYGON EMPTY"},
		{geom.NewSingleRingPolygonXY(0, 0, 0, 4, 4, 4, 4, 0, 0, 0), "POLYGON((0 0,0 4,4 4,4 0,0 0))"},
		{geom.NewSingleRingPolygonXYZ(), "POLYGON Z EMPTY"},
		{geom.NewSingleRingPolygonXYZ(0, 0, 10, 0, 4, 11, 4, 4, 12, 4, 0, 13, 0, 0, 14), "POLYGON Z((0 0 10,0 4 11,4 4 12,4 0 13,0 0 14))"},
		{geom.NewSingleRingPolygonXYM(), "POLYGON M EMPTY"},
		{geom.NewSingleRingPolygonXYM(0, 0, 10, 0, 4, 11, 4, 4, 12, 4, 0, 13, 0, 0, 14), "POLYGON M((0 0 10,0 4 11,4 4 12,4 0 13,0 0 14))"},
		{geom.NewSingleRingPolygonXYZM(), "POLYGON ZM EMPTY"},
		{geom.NewSingleRingPolygonXYZM(0, 0, 10, 100, 0, 4, 11, 101, 4, 4, 12, 102, 4, 0, 13, 103, 0, 0, 14, 104), "POLYGON ZM((0 0 10 100,0 4 11 101,4 4 12 102,4 0 13 103,0 0 14 104))"},

		{geom.NewMultiPolygonXY(), "MULTIPOLYGON EMPTY"},
		{geom.NewMultiPolygonXY(nil), "MULTIPOLYGON(EMPTY)"},
		{geom.NewMultiPolygonXY([][]float64{{0, 0, 0, 1, 1, 1, 1, 0, 0, 0}}, [][]float64{{2, 2, 2, 1, 1, 2, 2, 2}}), "MULTIPOLYGON(((0 0,0 1,1 1,1 0,0 0)),((2 2,2 1,1 2,2 2)))"},
		{geom.NewMultiPolygonXYZ(), "MULTIPOLYGON Z EMPTY"},
		{geom.NewMultiPolygonXYZ(nil), "MULTIPOLYGON Z(EMPTY)"},
		{geom.NewMultiPolygonXYZ([][]float64{{0, 0, 10, 0, 1, 11, 1, 1, 12, 0, 0, 13}}, [][]float64{{2, 2, 20, 2, 1, 21, 1, 2, 22, 2, 2, 23}}), "MULTIPOLYGON Z(((0 0 10,0 1 11,1 1 12,0 0 13)),((2 2 20,2 1 21,1 2 22,2 2 23)))"},
		{geom.NewMultiPolygonXYM(), "MULTIPOLYGON M EMPTY"},
		{geom.NewMultiPolygonXYM(nil), "MULTIPOLYGON M(EMPTY)"},
		{geom.NewMultiPolygonXYM([][]float64{{0, 0, 10, 0, 1, 11, 1, 1, 12, 0, 0, 13}}, [][]float64{{2, 2, 20, 2, 1, 21, 1, 2, 22, 2, 2, 23}}), "MULTIPOLYGON M(((0 0 10,0 1 11,1 1 12,0 0 13)),((2 2 20,2 1 21,1 2 22,2 2 23)))"},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			expectGeomEqWKT(t, tc.got.AsGeometry(), tc.wantWKT)
		})
	}
}
