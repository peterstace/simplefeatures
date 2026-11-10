package geom_test

import (
	"testing"

	"github.com/peterstace/simplefeatures/geom"
	"github.com/peterstace/simplefeatures/internal/test"
)

func nestInCollections(g geom.Geometry, depth int) geom.Geometry {
	for i := 0; i < depth; i++ {
		g = geom.NewGeometryCollection([]geom.Geometry{g}).AsGeometry()
	}
	return g
}

var collectionSink geom.GeometryCollection

func TestNewGeometryCollectionDoesNotCopyNestedMembers(t *testing.T) {
	leaf := test.FromWKT(t, "MULTIPOINT(1 2,3 4,5 6)")
	wrap := func(g geom.Geometry) func() {
		return func() { collectionSink = geom.NewGeometryCollection([]geom.Geometry{g}) }
	}
	shallow := testing.AllocsPerRun(10, wrap(nestInCollections(leaf, 1)))
	deep := testing.AllocsPerRun(10, wrap(nestInCollections(leaf, 100)))
	test.Eq(t, deep, shallow)
}

func TestGeometryCollectionForceCoordinatesTypeUnchanged(t *testing.T) {
	gc := nestInCollections(test.FromWKT(t, "MULTIPOINT(1 2,3 4,5 6)"), 100).MustAsGeometryCollection()
	allocs := testing.AllocsPerRun(10, func() { collectionSink = gc.ForceCoordinatesType(geom.DimXY) })
	test.Eq(t, allocs, 0.0)
	test.DeepEqual(t, gc.ForceCoordinatesType(geom.DimXY), gc)
}

func TestNewGeometryCollectionForcesNestedMembers(t *testing.T) {
	nestedZ := nestInCollections(test.FromWKT(t, "POINT Z(1 2 3)"), 3)
	gc := geom.NewGeometryCollection([]geom.Geometry{nestedZ, test.FromWKT(t, "POINT(4 5)")})
	test.ExactEqualsWKT(t, gc.AsGeometry(),
		"GEOMETRYCOLLECTION(GEOMETRYCOLLECTION(GEOMETRYCOLLECTION(GEOMETRYCOLLECTION(POINT(1 2)))),POINT(4 5))")
	test.Eq(t, gc.GeometryN(0).CoordinatesType(), geom.DimXY)
}
