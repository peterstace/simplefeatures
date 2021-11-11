package geom_test

import (
	"testing"

	. "github.com/peterstace/simplefeatures/geom"
)

func TestPointAccessorNonEmpty(t *testing.T) {
	xy, ok := geomFromWKT(t, "POINT(1 2)").MustAsPoint().XY()
	expectBoolEq(t, ok, true)
	expectXYEq(t, xy, XY{1, 2})
}

func TestPointAccessorEmpty(t *testing.T) {
	_, ok := geomFromWKT(t, "POINT EMPTY").MustAsPoint().XY()
	expectBoolEq(t, ok, false)
}

func TestLineStringAccessor(t *testing.T) {
	ls := geomFromWKT(t, "LINESTRING(1 2,3 4,5 6)").MustAsLineString()
	seq := ls.Coordinates()
	pt12 := xyCoords(1, 2)
	pt34 := xyCoords(3, 4)
	pt56 := xyCoords(5, 6)

	t.Run("start", func(t *testing.T) {
		want, err := NewPoint(pt12)
		expectNoErr(t, err)
		expectGeomEq(t, ls.StartPoint().AsGeometry(), want.AsGeometry())
	})
	t.Run("end", func(t *testing.T) {
		want, err := NewPoint(pt56)
		expectNoErr(t, err)
		expectGeomEq(t, ls.EndPoint().AsGeometry(), want.AsGeometry())
	})
	t.Run("num points", func(t *testing.T) {
		expectIntEq(t, seq.Length(), 3)
	})
	t.Run("point n", func(t *testing.T) {
		expectPanics(t, func() { seq.Get(-1) })
		expectCoordsEq(t, seq.Get(0), pt12)
		expectCoordsEq(t, seq.Get(1), pt34)
		expectCoordsEq(t, seq.Get(2), pt56)
		expectPanics(t, func() { seq.Get(3) })
	})
}

func TestLineStringEmptyAccessor(t *testing.T) {
	ls := geomFromWKT(t, "LINESTRING EMPTY").MustAsLineString()
	seq := ls.Coordinates()
	emptyPoint := geomFromWKT(t, "POINT EMPTY")

	t.Run("start", func(t *testing.T) {
		expectGeomEq(t, ls.StartPoint().AsGeometry(), emptyPoint)
	})
	t.Run("end", func(t *testing.T) {
		expectGeomEq(t, ls.EndPoint().AsGeometry(), emptyPoint)
	})
	t.Run("num points", func(t *testing.T) {
		expectIntEq(t, seq.Length(), 0)
	})
	t.Run("point n", func(t *testing.T) {
		expectPanics(t, func() { seq.Get(-1) })
		expectPanics(t, func() { seq.Get(0) })
		expectPanics(t, func() { seq.Get(1) })
	})
}

func TestLineStringAccessorWithDuplicates(t *testing.T) {
	ls := geomFromWKT(t, "LINESTRING(1 2,3 4,3 4,5 6)").MustAsLineString()
	seq := ls.Coordinates()
	pt12 := xyCoords(1, 2)
	pt34 := xyCoords(3, 4)
	pt56 := xyCoords(5, 6)

	t.Run("num points", func(t *testing.T) {
		expectIntEq(t, seq.Length(), 4)
	})
	t.Run("point n", func(t *testing.T) {
		expectPanics(t, func() { seq.Get(-1) })
		expectCoordsEq(t, seq.Get(0), pt12)
		expectCoordsEq(t, seq.Get(1), pt34)
		expectCoordsEq(t, seq.Get(2), pt34)
		expectCoordsEq(t, seq.Get(3), pt56)
		expectPanics(t, func() { seq.Get(4) })
	})
}

func TestLineStringAccessorWithMoreDuplicates(t *testing.T) {
	ls := geomFromWKT(t, "LINESTRING(1 2,1 2,3 4,3 4,3 4,5 6,5 6)").MustAsLineString()
	seq := ls.Coordinates()
	pt12 := xyCoords(1, 2)
	pt34 := xyCoords(3, 4)
	pt56 := xyCoords(5, 6)

	t.Run("num points", func(t *testing.T) {
		expectIntEq(t, seq.Length(), 7)
	})
	t.Run("point n", func(t *testing.T) {
		expectPanics(t, func() { seq.Get(-1) })
		expectCoordsEq(t, seq.Get(0), pt12)
		expectCoordsEq(t, seq.Get(1), pt12)
		expectCoordsEq(t, seq.Get(2), pt34)
		expectCoordsEq(t, seq.Get(3), pt34)
		expectCoordsEq(t, seq.Get(4), pt34)
		expectCoordsEq(t, seq.Get(5), pt56)
		expectCoordsEq(t, seq.Get(6), pt56)
		expectPanics(t, func() { seq.Get(7) })
	})
}

func TestPolygonAccessor(t *testing.T) {
	poly := geomFromWKT(t, "POLYGON((0 0,5 0,5 3,0 3,0 0),(1 1,2 1,2 2,1 2,1 1),(3 1,4 1,4 2,3 2,3 1))").MustAsPolygon()
	outer := geomFromWKT(t, "LINESTRING(0 0,5 0,5 3,0 3,0 0)")
	inner0 := geomFromWKT(t, "LINESTRING(1 1,2 1,2 2,1 2,1 1)")
	inner1 := geomFromWKT(t, "LINESTRING(3 1,4 1,4 2,3 2,3 1)")

	expectGeomEq(t, poly.ExteriorRing().AsGeometry(), outer)
	expectIntEq(t, poly.NumInteriorRings(), 2)
	expectPanics(t, func() { poly.InteriorRingN(-1) })
	expectGeomEq(t, poly.InteriorRingN(0).AsGeometry(), inner0)
	expectGeomEq(t, poly.InteriorRingN(1).AsGeometry(), inner1)
	expectPanics(t, func() { poly.InteriorRingN(2) })
}

func TestMultiPointAccessor(t *testing.T) {
	mp := geomFromWKT(t, "MULTIPOINT((4 5),(2 3),(8 7))").MustAsMultiPoint()
	pt0 := geomFromWKT(t, "POINT(4 5)")
	pt1 := geomFromWKT(t, "POINT(2 3)")
	pt2 := geomFromWKT(t, "POINT(8 7)")

	expectIntEq(t, mp.NumPoints(), 3)
	expectPanics(t, func() { mp.PointN(-1) })
	expectGeomEq(t, mp.PointN(0).AsGeometry(), pt0)
	expectGeomEq(t, mp.PointN(1).AsGeometry(), pt1)
	expectGeomEq(t, mp.PointN(2).AsGeometry(), pt2)
	expectPanics(t, func() { mp.PointN(3) })
}

func TestMultiLineStringAccessors(t *testing.T) {
	mls := geomFromWKT(t, "MULTILINESTRING((1 2,3 4,5 6),(7 8,9 10,11 12))").MustAsMultiLineString()
	ls0 := geomFromWKT(t, "LINESTRING(1 2,3 4,5 6)")
	ls1 := geomFromWKT(t, "LINESTRING(7 8,9 10,11 12)")

	expectIntEq(t, mls.NumLineStrings(), 2)
	expectPanics(t, func() { mls.LineStringN(-1) })
	expectGeomEq(t, mls.LineStringN(0).AsGeometry(), ls0)
	expectGeomEq(t, mls.LineStringN(1).AsGeometry(), ls1)
	expectPanics(t, func() { mls.LineStringN(2) })
}

func TestMultiPolygonAccessors(t *testing.T) {
	polys := geomFromWKT(t, "MULTIPOLYGON(((0 0,0 1,1 0,0 0)),((2 0,2 1,3 0,2 0)))").MustAsMultiPolygon()
	poly0 := geomFromWKT(t, "POLYGON((0 0,0 1,1 0,0 0))")
	poly1 := geomFromWKT(t, "POLYGON((2 0,2 1,3 0,2 0))")

	expectIntEq(t, polys.NumPolygons(), 2)
	expectPanics(t, func() { polys.PolygonN(-1) })
	expectGeomEq(t, polys.PolygonN(0).AsGeometry(), poly0)
	expectGeomEq(t, polys.PolygonN(1).AsGeometry(), poly1)
	expectPanics(t, func() { polys.PolygonN(2) })
}

func TestGeometryCollectionAccessors(t *testing.T) {
	geoms := geomFromWKT(t, "GEOMETRYCOLLECTION(POLYGON((0 0,0 1,1 0,0 0)),POLYGON((2 0,2 1,3 0,2 0)))").MustAsGeometryCollection()
	geom0 := geomFromWKT(t, "POLYGON((0 0,0 1,1 0,0 0))")
	geom1 := geomFromWKT(t, "POLYGON((2 0,2 1,3 0,2 0))")

	expectIntEq(t, geoms.NumGeometries(), 2)
	expectPanics(t, func() { geoms.GeometryN(-1) })
	expectGeomEq(t, geoms.GeometryN(0), geom0)
	expectGeomEq(t, geoms.GeometryN(1), geom1)
	expectPanics(t, func() { geoms.GeometryN(2) })
}
