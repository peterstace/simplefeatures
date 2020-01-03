package geom_test

import (
	"testing"

	. "github.com/peterstace/simplefeatures/geom"
)

func TestPointAccessor(t *testing.T) {
	pt := geomFromWKT(t, "POINT(1 2)").(Point)
	want := XY{1, 2}
	got := pt.XY()
	if !want.Equals(got) {
		t.Errorf("got=%v want=%v", got, want)
	}
}

func TestLineAccessor(t *testing.T) {
	line := geomFromWKT(t, "LINESTRING(1 2,3 4)").(Line)
	t.Run("start", func(t *testing.T) {
		got := line.StartPoint().AsGeometry()
		want := gFromWKT(t, "POINT(1 2)")
		expectGeomEq(t, got, want)
	})
	t.Run("end", func(t *testing.T) {
		got := line.EndPoint().AsGeometry()
		want := gFromWKT(t, "POINT(3 4)")
		expectGeomEq(t, got, want)
	})
	t.Run("num points", func(t *testing.T) {
		if line.NumPoints() != 2 {
			t.Errorf("wanted 2")
		}
	})
	t.Run("point 0", func(t *testing.T) {
		got := line.PointN(0).AsGeometry()
		want := gFromWKT(t, "POINT(1 2)")
		expectGeomEq(t, got, want)
	})
	t.Run("point 1", func(t *testing.T) {
		got := line.PointN(1).AsGeometry()
		want := gFromWKT(t, "POINT(3 4)")
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
	ls := geomFromWKT(t, "LINESTRING(1 2,3 4,5 6)").(LineString)
	pt12 := gFromWKT(t, "POINT(1 2)")
	pt34 := gFromWKT(t, "POINT(3 4)")
	pt56 := gFromWKT(t, "POINT(5 6)")

	t.Run("start", func(t *testing.T) {
		expectGeomEq(t, ls.StartPoint().AsGeometry(), pt12)
	})
	t.Run("end", func(t *testing.T) {
		expectGeomEq(t, ls.EndPoint().AsGeometry(), pt56)
	})
	t.Run("num points", func(t *testing.T) {
		expectIntEq(t, ls.NumPoints(), 3)
	})
	t.Run("point n", func(t *testing.T) {
		expectPanics(t, func() { ls.PointN(-1) })
		expectGeomEq(t, ls.PointN(0).AsGeometry(), pt12)
		expectGeomEq(t, ls.PointN(1).AsGeometry(), pt34)
		expectGeomEq(t, ls.PointN(2).AsGeometry(), pt56)
		expectPanics(t, func() { ls.PointN(3) })
	})
}

func TestPolygonAccessor(t *testing.T) {
	poly := gFromWKT(t, "POLYGON((0 0,5 0,5 3,0 3,0 0),(1 1,2 1,2 2,1 2,1 1),(3 1,4 1,4 2,3 2,3 1))").AsPolygon()
	outer := gFromWKT(t, "LINESTRING(0 0,5 0,5 3,0 3,0 0)")
	inner0 := gFromWKT(t, "LINESTRING(1 1,2 1,2 2,1 2,1 1)")
	inner1 := gFromWKT(t, "LINESTRING(3 1,4 1,4 2,3 2,3 1)")

	expectGeomEq(t, poly.ExteriorRing().AsGeometry(), outer)
	expectIntEq(t, poly.NumInteriorRings(), 2)
	expectPanics(t, func() { poly.InteriorRingN(-1) })
	expectGeomEq(t, poly.InteriorRingN(0).AsGeometry(), inner0)
	expectGeomEq(t, poly.InteriorRingN(1).AsGeometry(), inner1)
	expectPanics(t, func() { poly.InteriorRingN(2) })
}

func TestMultiPointAccessor(t *testing.T) {
	mp := gFromWKT(t, "MULTIPOINT((4 5),(2 3),(8 7))").AsMultiPoint()
	pt0 := gFromWKT(t, "POINT(4 5)")
	pt1 := gFromWKT(t, "POINT(2 3)")
	pt2 := gFromWKT(t, "POINT(8 7)")

	expectIntEq(t, mp.NumPoints(), 3)
	expectPanics(t, func() { mp.PointN(-1) })
	expectGeomEq(t, mp.PointN(0).AsGeometry(), pt0)
	expectGeomEq(t, mp.PointN(1).AsGeometry(), pt1)
	expectGeomEq(t, mp.PointN(2).AsGeometry(), pt2)
	expectPanics(t, func() { mp.PointN(3) })
}

func TestMultiLineStringAccessors(t *testing.T) {
	mls := gFromWKT(t, "MULTILINESTRING((1 2,3 4,5 6),(7 8,9 10,11 12))").AsMultiLineString()
	ls0 := gFromWKT(t, "LINESTRING(1 2,3 4,5 6)")
	ls1 := gFromWKT(t, "LINESTRING(7 8,9 10,11 12)")

	expectIntEq(t, mls.NumLineStrings(), 2)
	expectPanics(t, func() { mls.LineStringN(-1) })
	expectGeomEq(t, mls.LineStringN(0).AsGeometry(), ls0)
	expectGeomEq(t, mls.LineStringN(1).AsGeometry(), ls1)
	expectPanics(t, func() { mls.LineStringN(2) })
}

func TestMultiPolygonAccessors(t *testing.T) {
	polys := gFromWKT(t, "MULTIPOLYGON(((0 0,0 1,1 0,0 0)),((2 0,2 1,3 0,2 0)))").AsMultiPolygon()
	poly0 := gFromWKT(t, "POLYGON((0 0,0 1,1 0,0 0))")
	poly1 := gFromWKT(t, "POLYGON((2 0,2 1,3 0,2 0))")

	expectIntEq(t, polys.NumPolygons(), 2)
	expectPanics(t, func() { polys.PolygonN(-1) })
	expectGeomEq(t, polys.PolygonN(0).AsGeometry(), poly0)
	expectGeomEq(t, polys.PolygonN(1).AsGeometry(), poly1)
	expectPanics(t, func() { polys.PolygonN(2) })
}

func TestGeometryCollectionAccessors(t *testing.T) {
	geoms := gFromWKT(t, "GEOMETRYCOLLECTION(POLYGON((0 0,0 1,1 0,0 0)),POLYGON((2 0,2 1,3 0,2 0)))").AsGeometryCollection()
	geom0 := gFromWKT(t, "POLYGON((0 0,0 1,1 0,0 0))")
	geom1 := gFromWKT(t, "POLYGON((2 0,2 1,3 0,2 0))")

	expectIntEq(t, geoms.NumGeometries(), 2)
	expectPanics(t, func() { geoms.GeometryN(-1) })
	expectGeomEq(t, geoms.GeometryN(0), geom0)
	expectGeomEq(t, geoms.GeometryN(1), geom1)
	expectPanics(t, func() { geoms.GeometryN(2) })
}
