package geom_test

import (
	"testing"

	"github.com/peterstace/simplefeatures/geom"
	"github.com/peterstace/simplefeatures/internal/test"
)

func TestZeroValueGeometries(t *testing.T) {
	t.Run("Point", func(t *testing.T) {
		var pt geom.Point
		test.Eq(t, pt.IsEmpty(), true)
		test.Eq(t, pt.CoordinatesType(), geom.DimXY)
	})
	t.Run("LineString", func(t *testing.T) {
		var ls geom.LineString
		test.Eq(t, ls.Coordinates().Length(), 0)
		test.Eq(t, ls.CoordinatesType(), geom.DimXY)
	})
	t.Run("Polygon", func(t *testing.T) {
		var p geom.Polygon
		test.Eq(t, p.IsEmpty(), true)
		test.Eq(t, p.CoordinatesType(), geom.DimXY)
	})
	t.Run("MultiPoint", func(t *testing.T) {
		var mp geom.MultiPoint
		test.Eq(t, mp.NumPoints(), 0)
		test.Eq(t, mp.CoordinatesType(), geom.DimXY)
	})
	t.Run("MultiLineString", func(t *testing.T) {
		var mls geom.MultiLineString
		test.Eq(t, mls.NumLineStrings(), 0)
		test.Eq(t, mls.CoordinatesType(), geom.DimXY)
	})
	t.Run("MultiPolygon", func(t *testing.T) {
		var mp geom.MultiPolygon
		test.Eq(t, mp.NumPolygons(), 0)
		test.Eq(t, mp.CoordinatesType(), geom.DimXY)
	})
	t.Run("GeometryCollection", func(t *testing.T) {
		var gc geom.GeometryCollection
		test.Eq(t, gc.NumGeometries(), 0)
		test.Eq(t, gc.CoordinatesType(), geom.DimXY)
	})
}

func TestEmptySliceConstructors(t *testing.T) {
	t.Run("Polygon", func(t *testing.T) {
		p := geom.NewPolygon(nil)
		test.NoErr(t, p.Validate())
		test.Eq(t, p.IsEmpty(), true)
		test.Eq(t, p.CoordinatesType(), geom.DimXY)
	})
	t.Run("MultiPoint", func(t *testing.T) {
		mp := geom.NewMultiPoint(nil)
		test.NoErr(t, mp.Validate())
		test.Eq(t, mp.NumPoints(), 0)
		test.Eq(t, mp.CoordinatesType(), geom.DimXY)
	})
	t.Run("MultiLineString", func(t *testing.T) {
		mls := geom.NewMultiLineString(nil)
		test.NoErr(t, mls.Validate())
		test.Eq(t, mls.NumLineStrings(), 0)
		test.Eq(t, mls.CoordinatesType(), geom.DimXY)
	})
	t.Run("MultiPolygon", func(t *testing.T) {
		mp := geom.NewMultiPolygon(nil)
		test.NoErr(t, mp.Validate())
		test.Eq(t, mp.NumPolygons(), 0)
		test.Eq(t, mp.CoordinatesType(), geom.DimXY)
	})
	t.Run("GeometryCollection", func(t *testing.T) {
		gc := geom.NewGeometryCollection(nil)
		test.NoErr(t, gc.Validate())
		test.Eq(t, gc.NumGeometries(), 0)
		test.Eq(t, gc.CoordinatesType(), geom.DimXY)
	})
}
