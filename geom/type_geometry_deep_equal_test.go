package geom_test

import (
	"testing"

	"github.com/peterstace/simplefeatures/geom"
	"github.com/peterstace/simplefeatures/internal/test"
)

// TestReflectDeepEqualWorks proves that reflect.DeepEqual works for comparing
// geom.Geometry values.
func TestReflectDeepEqualWorks(t *testing.T) {
	tests := []struct {
		name string
		wkt  string
	}{
		{"Point", "POINT(1 2)"},
		{"LineString", "LINESTRING(1 2,3 4)"},
		{"Polygon", "POLYGON((0 0,0 1,1 0,0 0))"},
		{"MultiPoint", "MULTIPOINT((1 2),(3 4))"},
		{"MultiLineString", "MULTILINESTRING((1 2,3 4),(5 6,7 8))"},
		{"MultiPolygon", "MULTIPOLYGON(((0 0,0 1,1 0,0 0)))"},
		{"GeometryCollection", "GEOMETRYCOLLECTION(POINT(1 2))"},
		{"EmptyPoint", "POINT EMPTY"},
		{"EmptyLineString", "LINESTRING EMPTY"},
		{"EmptyPolygon", "POLYGON EMPTY"},
		{"EmptyMultiPoint", "MULTIPOINT EMPTY"},
		{"EmptyMultiLineString", "MULTILINESTRING EMPTY"},
		{"EmptyMultiPolygon", "MULTIPOLYGON EMPTY"},
		{"EmptyGeometryCollection", "GEOMETRYCOLLECTION EMPTY"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g1 := geomFromWKT(t, tt.wkt)
			g2 := geomFromWKT(t, tt.wkt)
			test.ExactEquals(t, g1, g2) // The geometries are semantically equal.
			test.DeepEqual(t, g1, g2)   // And reflect.DeepEqual works as expected.
		})
	}
}

// TestReflectDeepEqualWorksForZeroValue proves that reflect.DeepEqual handles
// the special zero value semantics of geom.Geometry. The zero value of
// Geometry is an empty GeometryCollection with a nil impl, and this is
// semantically equal to manually constructed empty GeometryCollections.
func TestReflectDeepEqualWorksForZeroValue(t *testing.T) {
	// Zero value has impl == nil.
	var g1 geom.Geometry

	// Constructed from WKT.
	g2 := geomFromWKT(t, "GEOMETRYCOLLECTION EMPTY")

	// Manually constructed with nil slice.
	g3 := geom.NewGeometryCollection(nil).AsGeometry()

	// Manually constructed with empty slice.
	g4 := geom.NewGeometryCollection([]geom.Geometry{}).AsGeometry()

	// All four geometries are semantically equal.
	test.ExactEquals(t, g1, g2)
	test.ExactEquals(t, g1, g3)
	test.ExactEquals(t, g1, g4)
	test.ExactEquals(t, g2, g3)
	test.ExactEquals(t, g2, g4)
	test.ExactEquals(t, g3, g4)

	// And reflect.DeepEqual works for all pairs.
	test.DeepEqual(t, g1, g2)
	test.DeepEqual(t, g1, g3)
	test.DeepEqual(t, g1, g4)
	test.DeepEqual(t, g2, g3)
	test.DeepEqual(t, g2, g4)
	test.DeepEqual(t, g3, g4)
}
