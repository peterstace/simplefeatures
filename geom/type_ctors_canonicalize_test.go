package geom_test

import (
	"testing"

	"github.com/peterstace/simplefeatures/geom"
	"github.com/peterstace/simplefeatures/internal/test"
)

// TestConstructorsCanonicalize tests that geometry constructors canonicalize
// nil vs empty slice inputs so that semantically equivalent geometries are
// reflect.DeepEqual. The canonical representation is nil, in order to match
// the zero value.
func TestConstructorsCanonicalize(t *testing.T) {
	t.Run("MultiPoint", func(t *testing.T) {
		var zero geom.MultiPoint
		fromNil := geom.NewMultiPoint(nil)
		fromEmpty := geom.NewMultiPoint([]geom.Point{})
		fromXYNil := geom.NewMultiPointXY()
		fromXYEmpty := geom.NewMultiPointXY([]float64{}...)

		test.DeepEqual(t, zero, fromNil)
		test.DeepEqual(t, zero, fromEmpty)
		test.DeepEqual(t, zero, fromXYNil)
		test.DeepEqual(t, zero, fromXYEmpty)
	})

	t.Run("Sequence", func(t *testing.T) {
		var zero geom.Sequence
		fromNil := geom.NewSequence(nil, geom.DimXY)
		fromEmpty := geom.NewSequence([]float64{}, geom.DimXY)

		test.DeepEqual(t, zero, fromNil)
		test.DeepEqual(t, zero, fromEmpty)
		test.DeepEqual(t, fromNil, fromEmpty)
	})

	t.Run("LineString", func(t *testing.T) {
		var zero geom.LineString
		fromNil := geom.NewLineString(geom.NewSequence(nil, geom.DimXY))
		fromEmpty := geom.NewLineString(geom.NewSequence([]float64{}, geom.DimXY))
		fromXYNil := geom.NewLineStringXY()
		fromXYEmpty := geom.NewLineStringXY([]float64{}...)

		test.DeepEqual(t, zero, fromNil)
		test.DeepEqual(t, zero, fromEmpty)
		test.DeepEqual(t, zero, fromXYNil)
		test.DeepEqual(t, zero, fromXYEmpty)
	})

	t.Run("Polygon", func(t *testing.T) {
		var zero geom.Polygon
		fromNil := geom.NewPolygon(nil)
		fromEmpty := geom.NewPolygon([]geom.LineString{})
		fromXYNil := geom.NewPolygonXY()
		fromXYEmpty := geom.NewPolygonXY([][]float64{}...)

		test.DeepEqual(t, zero, fromNil)
		test.DeepEqual(t, zero, fromEmpty)
		test.DeepEqual(t, zero, fromXYNil)
		test.DeepEqual(t, zero, fromXYEmpty)
	})

	t.Run("MultiLineString", func(t *testing.T) {
		var zero geom.MultiLineString
		fromNil := geom.NewMultiLineString(nil)
		fromEmpty := geom.NewMultiLineString([]geom.LineString{})
		fromXYNil := geom.NewMultiLineStringXY()
		fromXYEmpty := geom.NewMultiLineStringXY([][]float64{}...)

		test.DeepEqual(t, zero, fromNil)
		test.DeepEqual(t, zero, fromEmpty)
		test.DeepEqual(t, zero, fromXYNil)
		test.DeepEqual(t, zero, fromXYEmpty)
	})

	t.Run("MultiPolygon", func(t *testing.T) {
		var zero geom.MultiPolygon
		fromNil := geom.NewMultiPolygon(nil)
		fromEmpty := geom.NewMultiPolygon([]geom.Polygon{})
		fromXYNil := geom.NewMultiPolygonXY()
		fromXYEmpty := geom.NewMultiPolygonXY([][][]float64{}...)

		test.DeepEqual(t, zero, fromNil)
		test.DeepEqual(t, zero, fromEmpty)
		test.DeepEqual(t, zero, fromXYNil)
		test.DeepEqual(t, zero, fromXYEmpty)
	})

	t.Run("GeometryCollection", func(t *testing.T) {
		var zero geom.GeometryCollection
		fromNil := geom.NewGeometryCollection(nil)
		fromEmpty := geom.NewGeometryCollection([]geom.Geometry{})

		test.DeepEqual(t, zero, fromNil)
		test.DeepEqual(t, zero, fromEmpty)
		test.DeepEqual(t, fromNil, fromEmpty)
	})
}
