package geom_test

import (
	"testing"

	. "github.com/peterstace/simplefeatures/geom"
)

func TestPointAccessor(t *testing.T) {
	pt := geomFromWKT(t, "POINT(1 2)").AsPoint()
	want := XY{1, 2}
	got := pt.XY()
	if !want.Equals(got) {
		t.Errorf("got=%v want=%v", got, want)
	}
}

func TestLineAccessor(t *testing.T) {
	line := geomFromWKT(t, "LINESTRING(1 2,3 4)").AsLine()
	t.Run("start", func(t *testing.T) {
		got := line.StartPoint()
		want := geomFromWKT(t, "POINT(1 2)")
		expectGeomEq(t, got, want)
	})
	t.Run("end", func(t *testing.T) {
		got := line.EndPoint()
		want := geomFromWKT(t, "POINT(3 4)")
		expectGeomEq(t, got, want)
	})
	t.Run("num points", func(t *testing.T) {
		if line.NumPoints() != 2 {
			t.Errorf("wanted 2")
		}
	})
	t.Run("point 0", func(t *testing.T) {
		got := line.PointN(0)
		want := geomFromWKT(t, "POINT(1 2)")
		expectGeomEq(t, got, want)
	})
	t.Run("point 1", func(t *testing.T) {
		got := line.PointN(1)
		want := geomFromWKT(t, "POINT(3 4)")
		expectGeomEq(t, got, want)
	})
	t.Run("point 2", func(t *testing.T) {
		expectPanics(t, func() {
			line.PointN(2)
		})
	})
	t.Run("point -1", func(t *testing.T) {
		expectPanics(t, func() {
			line.PointN(-1)
		})
	})
}

func TestLineStringAccessor(t *testing.T) {
	ls := geomFromWKT(t, "LINESTRING(1 2,3 4,5 6)").AsLineString()
	pt12 := geomFromWKT(t, "POINT(1 2)")
	pt34 := geomFromWKT(t, "POINT(3 4)")
	pt56 := geomFromWKT(t, "POINT(5 6)")

	t.Run("start", func(t *testing.T) {
		expectGeomEq(t, ls.StartPoint(), pt12)
	})
	t.Run("end", func(t *testing.T) {
		expectGeomEq(t, ls.EndPoint(), pt56)
	})
	t.Run("num points", func(t *testing.T) {
		expectIntEq(t, ls.NumPoints(), 3)
	})
	t.Run("point n", func(t *testing.T) {
		expectPanics(t, func() { ls.PointN(-1) })
		expectGeomEq(t, ls.PointN(0), pt12)
		expectGeomEq(t, ls.PointN(1), pt34)
		expectGeomEq(t, ls.PointN(2), pt56)
		expectPanics(t, func() { ls.PointN(3) })
	})
}

func TestPolygonAccessor(t *testing.T) {
	poly := geomFromWKT(t, "POLYGON((0 0,5 0,5 3,0 3,0 0),(1 1,2 1,2 2,1 2,1 1),(3 1,4 1,4 2,3 2,3 1))").AsPolygon()
	outer := geomFromWKT(t, "LINESTRING(0 0,5 0,5 3,0 3,0 0)")
	inner0 := geomFromWKT(t, "LINESTRING(1 1,2 1,2 2,1 2,1 1)")
	inner1 := geomFromWKT(t, "LINESTRING(3 1,4 1,4 2,3 2,3 1)")

	expectGeomEq(t, poly.ExteriorRing(), outer)
	expectIntEq(t, poly.NumInteriorRings(), 2)
	expectPanics(t, func() { poly.InteriorRingN(-1) })
	expectGeomEq(t, poly.InteriorRingN(0), inner0)
	expectGeomEq(t, poly.InteriorRingN(1), inner1)
	expectPanics(t, func() { poly.InteriorRingN(2) })
}

func TestMultiPointAccessor(t *testing.T) {
	mp := geomFromWKT(t, "MULTIPOINT((4 5),(2 3),(8 7))").AsMultiPoint()
	pt0 := geomFromWKT(t, "POINT(4 5)")
	pt1 := geomFromWKT(t, "POINT(2 3)")
	pt2 := geomFromWKT(t, "POINT(8 7)")

	expectIntEq(t, mp.NumPoints(), 3)
	expectPanics(t, func() { mp.PointN(-1) })
	expectGeomEq(t, mp.PointN(0), pt0)
	expectGeomEq(t, mp.PointN(1), pt1)
	expectGeomEq(t, mp.PointN(2), pt2)
	expectPanics(t, func() { mp.PointN(3) })
}

func TestMultiLineStringAccessors(t *testing.T) {
	mls := geomFromWKT(t, "MULTILINESTRING((1 2,3 4,5 6),(7 8,9 10,11 12))").AsMultiLineString()
	ls0 := geomFromWKT(t, "LINESTRING(1 2,3 4,5 6)")
	ls1 := geomFromWKT(t, "LINESTRING(7 8,9 10,11 12)")

	expectIntEq(t, mls.NumLineStrings(), 2)
	expectPanics(t, func() { mls.LineStringN(-1) })
	expectGeomEq(t, mls.LineStringN(0), ls0)
	expectGeomEq(t, mls.LineStringN(1), ls1)
	expectPanics(t, func() { mls.LineStringN(2) })
}

func TestMultiPolygonAccessors(t *testing.T) {
	polys := geomFromWKT(t, "MULTIPOLYGON(((0 0,0 1,1 0,0 0)),((2 0,2 1,3 0,2 0)))").AsMultiPolygon()
	poly0 := geomFromWKT(t, "POLYGON((0 0,0 1,1 0,0 0))")
	poly1 := geomFromWKT(t, "POLYGON((2 0,2 1,3 0,2 0))")

	expectIntEq(t, polys.NumPolygons(), 2)
	expectPanics(t, func() { polys.PolygonN(-1) })
	expectGeomEq(t, polys.PolygonN(0), poly0)
	expectGeomEq(t, polys.PolygonN(1), poly1)
	expectPanics(t, func() { polys.PolygonN(2) })
}

func TestGeometryCollectionAccessors(t *testing.T) {
	geoms := geomFromWKT(t, "GEOMETRYCOLLECTION(POLYGON((0 0,0 1,1 0,0 0)),POLYGON((2 0,2 1,3 0,2 0)))").AsGeometryCollection()
	geom0 := geomFromWKT(t, "POLYGON((0 0,0 1,1 0,0 0))")
	geom1 := geomFromWKT(t, "POLYGON((2 0,2 1,3 0,2 0))")

	expectIntEq(t, geoms.NumGeometries(), 2)
	expectPanics(t, func() { geoms.GeometryN(-1) })
	expectGeomEq(t, geoms.GeometryN(0), geom0)
	expectGeomEq(t, geoms.GeometryN(1), geom1)
	expectPanics(t, func() { geoms.GeometryN(2) })
}
