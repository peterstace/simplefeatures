package geom_test

import (
	"testing"

	. "github.com/peterstace/simplefeatures/geom"
)

func TestZeroValueGeometries(t *testing.T) {
	t.Run("Point", func(t *testing.T) {
		var pt Point
		expectBoolEq(t, pt.IsEmpty(), true)
		expectCoordinatesTypeEq(t, pt.CoordinatesType(), DimXY)
	})
	t.Run("LineString", func(t *testing.T) {
		var ls LineString
		expectIntEq(t, ls.Coordinates().Length(), 0)
		expectCoordinatesTypeEq(t, ls.CoordinatesType(), DimXY)
	})
	t.Run("Polygon", func(t *testing.T) {
		var p Polygon
		expectBoolEq(t, p.IsEmpty(), true)
		expectCoordinatesTypeEq(t, p.CoordinatesType(), DimXY)
	})
	t.Run("MultiPoint", func(t *testing.T) {
		var mp MultiPoint
		expectIntEq(t, mp.NumPoints(), 0)
		expectCoordinatesTypeEq(t, mp.CoordinatesType(), DimXY)
	})
	t.Run("MultiLineString", func(t *testing.T) {
		var mls MultiLineString
		expectIntEq(t, mls.NumLineStrings(), 0)
		expectCoordinatesTypeEq(t, mls.CoordinatesType(), DimXY)
	})
	t.Run("MultiPolygon", func(t *testing.T) {
		var mp MultiPolygon
		expectIntEq(t, mp.NumPolygons(), 0)
		expectCoordinatesTypeEq(t, mp.CoordinatesType(), DimXY)
	})
	t.Run("GeometryCollection", func(t *testing.T) {
		var gc GeometryCollection
		expectIntEq(t, gc.NumGeometries(), 0)
		expectCoordinatesTypeEq(t, gc.CoordinatesType(), DimXY)
	})
}

func TestEmptySliceConstructors(t *testing.T) {
	t.Run("Polygon", func(t *testing.T) {
		p, err := NewPolygonFromRings(nil)
		expectNoErr(t, err)
		expectBoolEq(t, p.IsEmpty(), true)
		expectCoordinatesTypeEq(t, p.CoordinatesType(), DimXY)
	})
	t.Run("MultiPoint", func(t *testing.T) {
		mp, err := NewMultiPointFromPoints(nil)
		expectNoErr(t, err)
		expectIntEq(t, mp.NumPoints(), 0)
		expectCoordinatesTypeEq(t, mp.CoordinatesType(), DimXY)
	})
	t.Run("MultiLineString", func(t *testing.T) {
		mls := NewMultiLineStringFromLineStrings(nil)
		expectIntEq(t, mls.NumLineStrings(), 0)
		expectCoordinatesTypeEq(t, mls.CoordinatesType(), DimXY)
	})
	t.Run("MultiPolygon", func(t *testing.T) {
		mp, err := NewMultiPolygonFromPolygons(nil)
		expectNoErr(t, err)
		expectIntEq(t, mp.NumPolygons(), 0)
		expectCoordinatesTypeEq(t, mp.CoordinatesType(), DimXY)
	})
	t.Run("GeometryCollection", func(t *testing.T) {
		gc := NewGeometryCollection(nil)
		expectIntEq(t, gc.NumGeometries(), 0)
		expectCoordinatesTypeEq(t, gc.CoordinatesType(), DimXY)
	})
}
