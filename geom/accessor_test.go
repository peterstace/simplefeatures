package geom_test

import (
	"testing"

	. "github.com/peterstace/simplefeatures/geom"
	. "github.com/peterstace/simplefeatures/internal/geomtest"
)

func TestPointAccessorNonEmpty(t *testing.T) {
	xy, ok := GeomFromWKT(t, "POINT(1 2)").AsPoint().XY()
	ExpectBoolEq(t, ok, true)
	ExpectXYEq(t, xy, XY{1, 2})
}

func TestPointAccessorEmpty(t *testing.T) {
	_, ok := GeomFromWKT(t, "POINT EMPTY").AsPoint().XY()
	ExpectBoolEq(t, ok, false)
}

func TestLineAccessor(t *testing.T) {
	line := GeomFromWKT(t, "LINESTRING(1 2,3 4)").AsLine()
	t.Run("start", func(t *testing.T) {
		got := line.StartPoint()
		want := Coordinates{XY{1, 2}}
		ExpectCoordsEq(t, got, want)
	})
	t.Run("end", func(t *testing.T) {
		got := line.EndPoint()
		want := Coordinates{XY{3, 4}}
		ExpectCoordsEq(t, got, want)
	})
	t.Run("num points", func(t *testing.T) {
		if line.NumPoints() != 2 {
			t.Errorf("wanted 2")
		}
	})
	t.Run("point 0", func(t *testing.T) {
		got := line.PointN(0)
		want := Coordinates{XY{1, 2}}
		ExpectCoordsEq(t, got, want)
	})
	t.Run("point 1", func(t *testing.T) {
		got := line.PointN(1)
		want := Coordinates{XY{3, 4}}
		ExpectCoordsEq(t, got, want)
	})
	t.Run("point 2", func(t *testing.T) {
		ExpectPanics(t, func() {
			line.PointN(2)
		})
	})
	t.Run("point -1", func(t *testing.T) {
		ExpectPanics(t, func() {
			line.PointN(-1)
		})
	})
}

func TestLineStringAccessor(t *testing.T) {
	ls := GeomFromWKT(t, "LINESTRING(1 2,3 4,5 6)").AsLineString()
	pt12 := Coordinates{XY{1, 2}}
	pt34 := Coordinates{XY{3, 4}}
	pt56 := Coordinates{XY{5, 6}}

	t.Run("start", func(t *testing.T) {
		ExpectGeomEq(t, ls.StartPoint().AsGeometry(), NewPointC(pt12).AsGeometry())
	})
	t.Run("end", func(t *testing.T) {
		ExpectGeomEq(t, ls.EndPoint().AsGeometry(), NewPointC(pt56).AsGeometry())
	})
	t.Run("num points", func(t *testing.T) {
		ExpectIntEq(t, ls.NumPoints(), 3)
	})
	t.Run("point n", func(t *testing.T) {
		ExpectPanics(t, func() { ls.PointN(-1) })
		ExpectCoordsEq(t, ls.PointN(0), pt12)
		ExpectCoordsEq(t, ls.PointN(1), pt34)
		ExpectCoordsEq(t, ls.PointN(2), pt56)
		ExpectPanics(t, func() { ls.PointN(3) })
	})
	t.Run("num lines", func(t *testing.T) {
		ExpectIntEq(t, ls.NumLines(), 2)
	})
	t.Run("line n", func(t *testing.T) {
		ExpectPanics(t, func() { ls.LineN(-1) })
		ExpectGeomEq(t,
			ls.LineN(0).AsGeometry(),
			GeomFromWKT(t, "LINESTRING(1 2,3 4)"),
		)
		ExpectGeomEq(t,
			ls.LineN(1).AsGeometry(),
			GeomFromWKT(t, "LINESTRING(3 4,5 6)"),
		)
		ExpectPanics(t, func() { ls.LineN(2) })
	})
}

func TestLineStringEmptyAccessor(t *testing.T) {
	ls := GeomFromWKT(t, "LINESTRING EMPTY").AsLineString()
	emptyPoint := GeomFromWKT(t, "POINT EMPTY")

	t.Run("start", func(t *testing.T) {
		ExpectGeomEq(t, ls.StartPoint().AsGeometry(), emptyPoint)
	})
	t.Run("end", func(t *testing.T) {
		ExpectGeomEq(t, ls.EndPoint().AsGeometry(), emptyPoint)
	})
	t.Run("num points", func(t *testing.T) {
		ExpectIntEq(t, ls.NumPoints(), 0)
	})
	t.Run("point n", func(t *testing.T) {
		ExpectPanics(t, func() { ls.PointN(-1) })
		ExpectPanics(t, func() { ls.PointN(0) })
		ExpectPanics(t, func() { ls.PointN(1) })
	})
	t.Run("num lines", func(t *testing.T) {
		ExpectIntEq(t, ls.NumLines(), 0)
	})
	t.Run("line n", func(t *testing.T) {
		ExpectPanics(t, func() { ls.LineN(-1) })
		ExpectPanics(t, func() { ls.LineN(0) })
		ExpectPanics(t, func() { ls.LineN(1) })
	})
}

func TestLineStringAccessorWithDuplicates(t *testing.T) {
	ls := GeomFromWKT(t, "LINESTRING(1 2,3 4,3 4,5 6)").AsLineString()
	pt12 := Coordinates{XY{1, 2}}
	pt34 := Coordinates{XY{3, 4}}
	pt56 := Coordinates{XY{5, 6}}

	t.Run("num points", func(t *testing.T) {
		ExpectIntEq(t, ls.NumPoints(), 4)
	})
	t.Run("point n", func(t *testing.T) {
		ExpectPanics(t, func() { ls.PointN(-1) })
		ExpectCoordsEq(t, ls.PointN(0), pt12)
		ExpectCoordsEq(t, ls.PointN(1), pt34)
		ExpectCoordsEq(t, ls.PointN(2), pt34)
		ExpectCoordsEq(t, ls.PointN(3), pt56)
		ExpectPanics(t, func() { ls.PointN(4) })
	})
}

func TestLineStringAccessorWithMoreDuplicates(t *testing.T) {
	ls := GeomFromWKT(t, "LINESTRING(1 2,1 2,3 4,3 4,3 4,5 6,5 6)").AsLineString()
	pt12 := Coordinates{XY{1, 2}}
	pt34 := Coordinates{XY{3, 4}}
	pt56 := Coordinates{XY{5, 6}}

	t.Run("num points", func(t *testing.T) {
		ExpectIntEq(t, ls.NumPoints(), 7)
	})
	t.Run("point n", func(t *testing.T) {
		ExpectPanics(t, func() { ls.PointN(-1) })
		ExpectCoordsEq(t, ls.PointN(0), pt12)
		ExpectCoordsEq(t, ls.PointN(1), pt12)
		ExpectCoordsEq(t, ls.PointN(2), pt34)
		ExpectCoordsEq(t, ls.PointN(3), pt34)
		ExpectCoordsEq(t, ls.PointN(4), pt34)
		ExpectCoordsEq(t, ls.PointN(5), pt56)
		ExpectCoordsEq(t, ls.PointN(6), pt56)
		ExpectPanics(t, func() { ls.PointN(7) })
	})
}

func TestPolygonAccessor(t *testing.T) {
	poly := GeomFromWKT(t, "POLYGON((0 0,5 0,5 3,0 3,0 0),(1 1,2 1,2 2,1 2,1 1),(3 1,4 1,4 2,3 2,3 1))").AsPolygon()
	outer := GeomFromWKT(t, "LINESTRING(0 0,5 0,5 3,0 3,0 0)")
	inner0 := GeomFromWKT(t, "LINESTRING(1 1,2 1,2 2,1 2,1 1)")
	inner1 := GeomFromWKT(t, "LINESTRING(3 1,4 1,4 2,3 2,3 1)")

	ExpectGeomEq(t, poly.ExteriorRing().AsGeometry(), outer)
	ExpectIntEq(t, poly.NumInteriorRings(), 2)
	ExpectPanics(t, func() { poly.InteriorRingN(-1) })
	ExpectGeomEq(t, poly.InteriorRingN(0).AsGeometry(), inner0)
	ExpectGeomEq(t, poly.InteriorRingN(1).AsGeometry(), inner1)
	ExpectPanics(t, func() { poly.InteriorRingN(2) })
}

func TestMultiPointAccessor(t *testing.T) {
	mp := GeomFromWKT(t, "MULTIPOINT((4 5),(2 3),(8 7))").AsMultiPoint()
	pt0 := GeomFromWKT(t, "POINT(4 5)")
	pt1 := GeomFromWKT(t, "POINT(2 3)")
	pt2 := GeomFromWKT(t, "POINT(8 7)")

	ExpectIntEq(t, mp.NumPoints(), 3)
	ExpectPanics(t, func() { mp.PointN(-1) })
	ExpectGeomEq(t, mp.PointN(0).AsGeometry(), pt0)
	ExpectGeomEq(t, mp.PointN(1).AsGeometry(), pt1)
	ExpectGeomEq(t, mp.PointN(2).AsGeometry(), pt2)
	ExpectPanics(t, func() { mp.PointN(3) })
}

func TestMultiLineStringAccessors(t *testing.T) {
	mls := GeomFromWKT(t, "MULTILINESTRING((1 2,3 4,5 6),(7 8,9 10,11 12))").AsMultiLineString()
	ls0 := GeomFromWKT(t, "LINESTRING(1 2,3 4,5 6)")
	ls1 := GeomFromWKT(t, "LINESTRING(7 8,9 10,11 12)")

	ExpectIntEq(t, mls.NumLineStrings(), 2)
	ExpectPanics(t, func() { mls.LineStringN(-1) })
	ExpectGeomEq(t, mls.LineStringN(0).AsGeometry(), ls0)
	ExpectGeomEq(t, mls.LineStringN(1).AsGeometry(), ls1)
	ExpectPanics(t, func() { mls.LineStringN(2) })
}

func TestMultiPolygonAccessors(t *testing.T) {
	polys := GeomFromWKT(t, "MULTIPOLYGON(((0 0,0 1,1 0,0 0)),((2 0,2 1,3 0,2 0)))").AsMultiPolygon()
	poly0 := GeomFromWKT(t, "POLYGON((0 0,0 1,1 0,0 0))")
	poly1 := GeomFromWKT(t, "POLYGON((2 0,2 1,3 0,2 0))")

	ExpectIntEq(t, polys.NumPolygons(), 2)
	ExpectPanics(t, func() { polys.PolygonN(-1) })
	ExpectGeomEq(t, polys.PolygonN(0).AsGeometry(), poly0)
	ExpectGeomEq(t, polys.PolygonN(1).AsGeometry(), poly1)
	ExpectPanics(t, func() { polys.PolygonN(2) })
}

func TestGeometryCollectionAccessors(t *testing.T) {
	geoms := GeomFromWKT(t, "GEOMETRYCOLLECTION(POLYGON((0 0,0 1,1 0,0 0)),POLYGON((2 0,2 1,3 0,2 0)))").AsGeometryCollection()
	geom0 := GeomFromWKT(t, "POLYGON((0 0,0 1,1 0,0 0))")
	geom1 := GeomFromWKT(t, "POLYGON((2 0,2 1,3 0,2 0))")

	ExpectIntEq(t, geoms.NumGeometries(), 2)
	ExpectPanics(t, func() { geoms.GeometryN(-1) })
	ExpectGeomEq(t, geoms.GeometryN(0), geom0)
	ExpectGeomEq(t, geoms.GeometryN(1), geom1)
	ExpectPanics(t, func() { geoms.GeometryN(2) })
}
