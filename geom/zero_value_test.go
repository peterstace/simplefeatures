package geom_test

import (
	"testing"

	"github.com/peterstace/simplefeatures/geom"
)

func TestZeroValueGeometries(t *testing.T) {
	t.Run("Point", func(t *testing.T) {
		var pt geom.Point
		expectBoolEq(t, pt.IsEmpty(), true)
		expectCoordinatesTypeEq(t, pt.CoordinatesType(), geom.DimXY)
	})
	t.Run("LineString", func(t *testing.T) {
		var ls geom.LineString
		expectIntEq(t, ls.Coordinates().Length(), 0)
		expectCoordinatesTypeEq(t, ls.CoordinatesType(), geom.DimXY)
	})
	t.Run("Polygon", func(t *testing.T) {
		var p geom.Polygon
		expectBoolEq(t, p.IsEmpty(), true)
		expectCoordinatesTypeEq(t, p.CoordinatesType(), geom.DimXY)
	})
	t.Run("MultiPoint", func(t *testing.T) {
		var mp geom.MultiPoint
		expectIntEq(t, mp.NumPoints(), 0)
		expectCoordinatesTypeEq(t, mp.CoordinatesType(), geom.DimXY)
	})
	t.Run("MultiLineString", func(t *testing.T) {
		var mls geom.MultiLineString
		expectIntEq(t, mls.NumLineStrings(), 0)
		expectCoordinatesTypeEq(t, mls.CoordinatesType(), geom.DimXY)
	})
	t.Run("MultiPolygon", func(t *testing.T) {
		var mp geom.MultiPolygon
		expectIntEq(t, mp.NumPolygons(), 0)
		expectCoordinatesTypeEq(t, mp.CoordinatesType(), geom.DimXY)
	})
	t.Run("GeometryCollection", func(t *testing.T) {
		var gc geom.GeometryCollection
		expectIntEq(t, gc.NumGeometries(), 0)
		expectCoordinatesTypeEq(t, gc.CoordinatesType(), geom.DimXY)
	})
}

func TestEmptySliceConstructors(t *testing.T) {
	t.Run("Polygon", func(t *testing.T) {
		p := geom.NewPolygon(nil)
		expectNoErr(t, p.Validate())
		expectBoolEq(t, p.IsEmpty(), true)
		expectCoordinatesTypeEq(t, p.CoordinatesType(), geom.DimXY)
	})
	t.Run("MultiPoint", func(t *testing.T) {
		mp := geom.NewMultiPoint(nil)
		expectNoErr(t, mp.Validate())
		expectIntEq(t, mp.NumPoints(), 0)
		expectCoordinatesTypeEq(t, mp.CoordinatesType(), geom.DimXY)
	})
	t.Run("MultiLineString", func(t *testing.T) {
		mls := geom.NewMultiLineString(nil)
		expectNoErr(t, mls.Validate())
		expectIntEq(t, mls.NumLineStrings(), 0)
		expectCoordinatesTypeEq(t, mls.CoordinatesType(), geom.DimXY)
	})
	t.Run("MultiPolygon", func(t *testing.T) {
		mp := geom.NewMultiPolygon(nil)
		expectNoErr(t, mp.Validate())
		expectIntEq(t, mp.NumPolygons(), 0)
		expectCoordinatesTypeEq(t, mp.CoordinatesType(), geom.DimXY)
	})
	t.Run("GeometryCollection", func(t *testing.T) {
		gc := geom.NewGeometryCollection(nil)
		expectNoErr(t, gc.Validate())
		expectIntEq(t, gc.NumGeometries(), 0)
		expectCoordinatesTypeEq(t, gc.CoordinatesType(), geom.DimXY)
	})
}
