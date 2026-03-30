package geom_test

import (
	"testing"

	"github.com/peterstace/simplefeatures/geom"
	"github.com/peterstace/simplefeatures/internal/test"
)

func TestPointAccessorNonEmpty(t *testing.T) {
	xy, ok := test.FromWKT(t, "POINT(1 2)").MustAsPoint().XY()
	test.Eq(t, ok, true)
	test.Eq(t, xy, geom.XY{1, 2})
}

func TestPointAccessorEmpty(t *testing.T) {
	_, ok := test.FromWKT(t, "POINT EMPTY").MustAsPoint().XY()
	test.Eq(t, ok, false)
}

func xyCoords(x, y float64) geom.Coordinates {
	return geom.Coordinates{XY: geom.XY{x, y}, Type: geom.DimXY}
}

func TestLineStringAccessor(t *testing.T) {
	ls := test.FromWKT(t, "LINESTRING(1 2,3 4,5 6)").MustAsLineString()
	seq := ls.Coordinates()
	pt12 := xyCoords(1, 2)
	pt34 := xyCoords(3, 4)
	pt56 := xyCoords(5, 6)

	t.Run("start", func(t *testing.T) {
		want := geom.NewPoint(pt12)
		test.ExactEquals(t, ls.StartPoint().AsGeometry(), want.AsGeometry())
	})
	t.Run("end", func(t *testing.T) {
		want := geom.NewPoint(pt56)
		test.ExactEquals(t, ls.EndPoint().AsGeometry(), want.AsGeometry())
	})
	t.Run("num points", func(t *testing.T) {
		test.Eq(t, seq.Length(), 3)
	})
	t.Run("point n", func(t *testing.T) {
		test.Panics(t, func() { seq.Get(-1) })
		test.Eq(t, seq.Get(0), pt12)
		test.Eq(t, seq.Get(1), pt34)
		test.Eq(t, seq.Get(2), pt56)
		test.Panics(t, func() { seq.Get(3) })
	})
}

func TestLineStringEmptyAccessor(t *testing.T) {
	ls := test.FromWKT(t, "LINESTRING EMPTY").MustAsLineString()
	seq := ls.Coordinates()
	emptyPoint := test.FromWKT(t, "POINT EMPTY")

	t.Run("start", func(t *testing.T) {
		test.ExactEquals(t, ls.StartPoint().AsGeometry(), emptyPoint)
	})
	t.Run("end", func(t *testing.T) {
		test.ExactEquals(t, ls.EndPoint().AsGeometry(), emptyPoint)
	})
	t.Run("num points", func(t *testing.T) {
		test.Eq(t, seq.Length(), 0)
	})
	t.Run("point n", func(t *testing.T) {
		test.Panics(t, func() { seq.Get(-1) })
		test.Panics(t, func() { seq.Get(0) })
		test.Panics(t, func() { seq.Get(1) })
	})
}

func TestLineStringAccessorWithDuplicates(t *testing.T) {
	ls := test.FromWKT(t, "LINESTRING(1 2,3 4,3 4,5 6)").MustAsLineString()
	seq := ls.Coordinates()
	pt12 := xyCoords(1, 2)
	pt34 := xyCoords(3, 4)
	pt56 := xyCoords(5, 6)

	t.Run("num points", func(t *testing.T) {
		test.Eq(t, seq.Length(), 4)
	})
	t.Run("point n", func(t *testing.T) {
		test.Panics(t, func() { seq.Get(-1) })
		test.Eq(t, seq.Get(0), pt12)
		test.Eq(t, seq.Get(1), pt34)
		test.Eq(t, seq.Get(2), pt34)
		test.Eq(t, seq.Get(3), pt56)
		test.Panics(t, func() { seq.Get(4) })
	})
}

func TestLineStringAccessorWithMoreDuplicates(t *testing.T) {
	ls := test.FromWKT(t, "LINESTRING(1 2,1 2,3 4,3 4,3 4,5 6,5 6)").MustAsLineString()
	seq := ls.Coordinates()
	pt12 := xyCoords(1, 2)
	pt34 := xyCoords(3, 4)
	pt56 := xyCoords(5, 6)

	t.Run("num points", func(t *testing.T) {
		test.Eq(t, seq.Length(), 7)
	})
	t.Run("point n", func(t *testing.T) {
		test.Panics(t, func() { seq.Get(-1) })
		test.Eq(t, seq.Get(0), pt12)
		test.Eq(t, seq.Get(1), pt12)
		test.Eq(t, seq.Get(2), pt34)
		test.Eq(t, seq.Get(3), pt34)
		test.Eq(t, seq.Get(4), pt34)
		test.Eq(t, seq.Get(5), pt56)
		test.Eq(t, seq.Get(6), pt56)
		test.Panics(t, func() { seq.Get(7) })
	})
}

func TestPolygonAccessor(t *testing.T) {
	poly := test.FromWKT(t, "POLYGON((0 0,5 0,5 3,0 3,0 0),(1 1,2 1,2 2,1 2,1 1),(3 1,4 1,4 2,3 2,3 1))").MustAsPolygon()
	outer := test.FromWKT(t, "LINESTRING(0 0,5 0,5 3,0 3,0 0)")
	inner0 := test.FromWKT(t, "LINESTRING(1 1,2 1,2 2,1 2,1 1)")
	inner1 := test.FromWKT(t, "LINESTRING(3 1,4 1,4 2,3 2,3 1)")

	test.ExactEquals(t, poly.ExteriorRing().AsGeometry(), outer)
	test.Eq(t, poly.NumInteriorRings(), 2)
	test.Panics(t, func() { poly.InteriorRingN(-1) })
	test.ExactEquals(t, poly.InteriorRingN(0).AsGeometry(), inner0)
	test.ExactEquals(t, poly.InteriorRingN(1).AsGeometry(), inner1)
	test.Panics(t, func() { poly.InteriorRingN(2) })
}

func TestMultiPointAccessor(t *testing.T) {
	mp := test.FromWKT(t, "MULTIPOINT((4 5),(2 3),(8 7))").MustAsMultiPoint()
	pt0 := test.FromWKT(t, "POINT(4 5)")
	pt1 := test.FromWKT(t, "POINT(2 3)")
	pt2 := test.FromWKT(t, "POINT(8 7)")

	test.Eq(t, mp.NumPoints(), 3)
	test.Panics(t, func() { mp.PointN(-1) })
	test.ExactEquals(t, mp.PointN(0).AsGeometry(), pt0)
	test.ExactEquals(t, mp.PointN(1).AsGeometry(), pt1)
	test.ExactEquals(t, mp.PointN(2).AsGeometry(), pt2)
	test.Panics(t, func() { mp.PointN(3) })
}

func TestMultiLineStringAccessors(t *testing.T) {
	mls := test.FromWKT(t, "MULTILINESTRING((1 2,3 4,5 6),(7 8,9 10,11 12))").MustAsMultiLineString()
	ls0 := test.FromWKT(t, "LINESTRING(1 2,3 4,5 6)")
	ls1 := test.FromWKT(t, "LINESTRING(7 8,9 10,11 12)")

	test.Eq(t, mls.NumLineStrings(), 2)
	test.Panics(t, func() { mls.LineStringN(-1) })
	test.ExactEquals(t, mls.LineStringN(0).AsGeometry(), ls0)
	test.ExactEquals(t, mls.LineStringN(1).AsGeometry(), ls1)
	test.Panics(t, func() { mls.LineStringN(2) })
}

func TestMultiPolygonAccessors(t *testing.T) {
	polys := test.FromWKT(t, "MULTIPOLYGON(((0 0,0 1,1 0,0 0)),((2 0,2 1,3 0,2 0)))").MustAsMultiPolygon()
	poly0 := test.FromWKT(t, "POLYGON((0 0,0 1,1 0,0 0))")
	poly1 := test.FromWKT(t, "POLYGON((2 0,2 1,3 0,2 0))")

	test.Eq(t, polys.NumPolygons(), 2)
	test.Panics(t, func() { polys.PolygonN(-1) })
	test.ExactEquals(t, polys.PolygonN(0).AsGeometry(), poly0)
	test.ExactEquals(t, polys.PolygonN(1).AsGeometry(), poly1)
	test.Panics(t, func() { polys.PolygonN(2) })
}

func TestGeometryCollectionAccessors(t *testing.T) {
	geoms := test.FromWKT(t, "GEOMETRYCOLLECTION(POLYGON((0 0,0 1,1 0,0 0)),POLYGON((2 0,2 1,3 0,2 0)))").MustAsGeometryCollection()
	geom0 := test.FromWKT(t, "POLYGON((0 0,0 1,1 0,0 0))")
	geom1 := test.FromWKT(t, "POLYGON((2 0,2 1,3 0,2 0))")

	test.Eq(t, geoms.NumGeometries(), 2)
	test.Panics(t, func() { geoms.GeometryN(-1) })
	test.ExactEquals(t, geoms.GeometryN(0), geom0)
	test.ExactEquals(t, geoms.GeometryN(1), geom1)
	test.Panics(t, func() { geoms.GeometryN(2) })
}
