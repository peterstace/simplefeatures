package geom_test

import (
	"reflect"
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
		got := line.StartPoint()
		want := geomFromWKT(t, "POINT(1 2)")
		if !reflect.DeepEqual(got, want) {
			t.Errorf("want=%v got=%v", want, got)
		}
	})
	t.Run("end", func(t *testing.T) {
		got := line.EndPoint()
		want := geomFromWKT(t, "POINT(3 4)")
		if !reflect.DeepEqual(got, want) {
			t.Errorf("want=%v got=%v", want, got)
		}
	})
	t.Run("num points", func(t *testing.T) {
		if line.NumPoints() != 2 {
			t.Errorf("wanted 2")
		}
	})
	t.Run("point 0", func(t *testing.T) {
		got := line.PointN(0)
		want := geomFromWKT(t, "POINT(1 2)")
		expectDeepEqual(t, got, want)
	})
	t.Run("point 1", func(t *testing.T) {
		got := line.PointN(1)
		want := geomFromWKT(t, "POINT(3 4)")
		expectDeepEqual(t, got, want)
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
	pt12 := geomFromWKT(t, "POINT(1 2)")
	pt34 := geomFromWKT(t, "POINT(3 4)")
	pt56 := geomFromWKT(t, "POINT(5 6)")

	t.Run("start", func(t *testing.T) {
		expectDeepEqual(t, ls.StartPoint(), pt12)
	})
	t.Run("end", func(t *testing.T) {
		expectDeepEqual(t, ls.EndPoint(), pt56)
	})
	t.Run("num points", func(t *testing.T) {
		expectDeepEqual(t, ls.NumPoints(), 3)
	})
	t.Run("point n", func(t *testing.T) {
		expectPanics(t, func() { ls.PointN(-1) })
		expectDeepEqual(t, ls.PointN(0), pt12)
		expectDeepEqual(t, ls.PointN(1), pt34)
		expectDeepEqual(t, ls.PointN(2), pt56)
		expectPanics(t, func() { ls.PointN(3) })
	})
}

func TestLinearRingAccessor(t *testing.T) {
	ls := geomFromWKT(t, "LINEARRING(0 0,1 0,0 1,0 0)").(LinearRing)
	pt0 := geomFromWKT(t, "POINT(0 0)")
	pt1 := geomFromWKT(t, "POINT(1 0)")
	pt2 := geomFromWKT(t, "POINT(0 1)")
	pt3 := geomFromWKT(t, "POINT(0 0)")

	t.Run("start", func(t *testing.T) {
		expectDeepEqual(t, ls.StartPoint(), pt0)
	})
	t.Run("end", func(t *testing.T) {
		expectDeepEqual(t, ls.EndPoint(), pt3)
	})
	t.Run("num points", func(t *testing.T) {
		expectDeepEqual(t, ls.NumPoints(), 4)
	})
	t.Run("point n", func(t *testing.T) {
		expectPanics(t, func() { ls.PointN(-1) })
		expectDeepEqual(t, ls.PointN(0), pt0)
		expectDeepEqual(t, ls.PointN(1), pt1)
		expectDeepEqual(t, ls.PointN(2), pt2)
		expectDeepEqual(t, ls.PointN(3), pt3)
		expectPanics(t, func() { ls.PointN(4) })
	})
}

func TestPolygonAccessor(t *testing.T) {
	poly := geomFromWKT(t, "POLYGON((0 0,5 0,5 3,0 3,0 0),(1 1,2 1,2 2,1 2,1 1),(3 1,4 1,4 2,3 2,3 1))").(Polygon)
	outer := geomFromWKT(t, "LINEARRING(0 0,5 0,5 3,0 3,0 0)")
	inner0 := geomFromWKT(t, "LINEARRING(1 1,2 1,2 2,1 2,1 1)")
	inner1 := geomFromWKT(t, "LINEARRING(3 1,4 1,4 2,3 2,3 1)")

	expectDeepEqual(t, poly.ExteriorRing(), outer)
	expectDeepEqual(t, poly.NumInteriorRings(), 2)
	expectPanics(t, func() { poly.InteriorRingN(-1) })
	expectDeepEqual(t, poly.InteriorRingN(0), inner0)
	expectDeepEqual(t, poly.InteriorRingN(1), inner1)
	expectPanics(t, func() { poly.InteriorRingN(2) })
}

func TestMultiPointAccessor(t *testing.T) {
	mp := geomFromWKT(t, "MULTIPOINT((4 5),(2 3),(8 7))").(MultiPoint)
	pt0 := geomFromWKT(t, "POINT(4 5)")
	pt1 := geomFromWKT(t, "POINT(2 3)")
	pt2 := geomFromWKT(t, "POINT(8 7)")

	expectDeepEqual(t, mp.NumPoints(), 3)
	expectPanics(t, func() { mp.PointN(-1) })
	expectDeepEqual(t, mp.PointN(0), pt0)
	expectDeepEqual(t, mp.PointN(1), pt1)
	expectDeepEqual(t, mp.PointN(2), pt2)
	expectPanics(t, func() { mp.PointN(3) })
}

func TestMultiLineStringAccessors(t *testing.T) {
	mls := geomFromWKT(t, "MULTILINESTRING((1 2,3 4,5 6),(7 8,9 10,11 12))").(MultiLineString)
	ls0 := geomFromWKT(t, "LINESTRING(1 2,3 4,5 6)")
	ls1 := geomFromWKT(t, "LINESTRING(7 8,9 10,11 12)")

	expectDeepEqual(t, mls.NumLineStrings(), 2)
	expectPanics(t, func() { mls.LineStringN(-1) })
	expectDeepEqual(t, mls.LineStringN(0), ls0)
	expectDeepEqual(t, mls.LineStringN(1), ls1)
	expectPanics(t, func() { mls.LineStringN(2) })
}

func TestMultiPolygonAccessors(t *testing.T) {
	polys := geomFromWKT(t, "MULTIPOLYGON(((0 0,0 1,1 0,0 0)),((2 0,2 1,3 0,2 0)))").(MultiPolygon)
	poly0 := geomFromWKT(t, "POLYGON((0 0,0 1,1 0,0 0))")
	poly1 := geomFromWKT(t, "POLYGON((2 0,2 1,3 0,2 0))")

	expectDeepEqual(t, polys.NumPolygons(), 2)
	expectPanics(t, func() { polys.PolygonN(-1) })
	expectDeepEqual(t, polys.PolygonN(0), poly0)
	expectDeepEqual(t, polys.PolygonN(1), poly1)
	expectPanics(t, func() { polys.PolygonN(2) })
}

func TestGeometryCollectionAccessors(t *testing.T) {
	geoms := geomFromWKT(t, "GEOMETRYCOLLECTION(POLYGON((0 0,0 1,1 0,0 0)),POLYGON((2 0,2 1,3 0,2 0)))").(GeometryCollection)
	geom0 := geomFromWKT(t, "POLYGON((0 0,0 1,1 0,0 0))")
	geom1 := geomFromWKT(t, "POLYGON((2 0,2 1,3 0,2 0))")

	expectDeepEqual(t, geoms.NumGeometries(), 2)
	expectPanics(t, func() { geoms.GeometryN(-1) })
	expectDeepEqual(t, geoms.GeometryN(0), geom0)
	expectDeepEqual(t, geoms.GeometryN(1), geom1)
	expectPanics(t, func() { geoms.GeometryN(2) })
}
