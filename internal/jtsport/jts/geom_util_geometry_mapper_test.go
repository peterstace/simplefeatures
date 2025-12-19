package jts_test

import (
	"testing"

	"github.com/peterstace/simplefeatures/internal/jtsport/java"
	"github.com/peterstace/simplefeatures/internal/jtsport/jts"
)

// keepLineOp is a MapOp that:
// - LineString -> LineString
// - Point -> empty LineString
// - Polygon -> nil
type keepLineOp struct{}

func (keepLineOp) Map(geom *jts.Geom_Geometry) *jts.Geom_Geometry {
	if java.InstanceOf[*jts.Geom_Point](geom) {
		return geom.GetFactory().CreateEmpty(1)
	}
	if java.InstanceOf[*jts.Geom_LineString](geom) {
		return geom
	}
	return nil
}

func TestGeometryMapperFlatMapInputEmpty(t *testing.T) {
	checkFlatMap(t,
		"GEOMETRYCOLLECTION( POINT EMPTY, LINESTRING EMPTY)",
		1, keepLineOp{},
		"LINESTRING EMPTY",
	)
}

func TestGeometryMapperFlatMapInputMulti(t *testing.T) {
	checkFlatMap(t,
		"GEOMETRYCOLLECTION( MULTILINESTRING((0 0, 1 1), (1 1, 2 2)), LINESTRING(2 2, 3 3))",
		1, keepLineOp{},
		"MULTILINESTRING ((0 0, 1 1), (1 1, 2 2), (2 2, 3 3))",
	)
}

func TestGeometryMapperFlatMapResultEmpty(t *testing.T) {
	checkFlatMap(t,
		"GEOMETRYCOLLECTION( LINESTRING(0 0, 1 1), LINESTRING(1 1, 2 2))",
		1, keepLineOp{},
		"MULTILINESTRING((0 0, 1 1), (1 1, 2 2))",
	)

	checkFlatMap(t,
		"GEOMETRYCOLLECTION( POINT(0 0), POINT(0 0), LINESTRING(0 0, 1 1))",
		1, keepLineOp{},
		"LINESTRING(0 0, 1 1)",
	)

	checkFlatMap(t,
		"MULTIPOINT((0 0), (1 1))",
		1, keepLineOp{},
		"LINESTRING EMPTY",
	)
}

func TestGeometryMapperFlatMapResultNull(t *testing.T) {
	checkFlatMap(t,
		"GEOMETRYCOLLECTION( POINT(0 0), LINESTRING(0 0, 1 1), POLYGON ((1 1, 1 2, 2 1, 1 1)))",
		1, keepLineOp{},
		"LINESTRING(0 0, 1 1)",
	)
}

// boundaryOp is a MapOp that returns the boundary of a geometry.
type boundaryOp struct{}

func (boundaryOp) Map(geom *jts.Geom_Geometry) *jts.Geom_Geometry {
	return geom.GetBoundary()
}

func TestGeometryMapperFlatMapBoundary(t *testing.T) {
	checkFlatMap(t,
		"GEOMETRYCOLLECTION( POINT(0 0), LINESTRING(0 0, 1 1), POLYGON ((1 1, 1 2, 2 1, 1 1)))",
		0, boundaryOp{},
		"GEOMETRYCOLLECTION (POINT (0 0), POINT (1 1), LINEARRING (1 1, 1 2, 2 1, 1 1))",
	)

	checkFlatMap(t,
		"LINESTRING EMPTY",
		0, boundaryOp{},
		"POINT EMPTY",
	)
}

func checkFlatMap(t *testing.T, wkt string, dim int, op jts.GeomUtil_GeometryMapper_MapOp, wktExpected string) {
	t.Helper()
	reader := jts.Io_NewWKTReader()
	geom, err := reader.Read(wkt)
	if err != nil {
		t.Fatalf("failed to parse input WKT: %v", err)
	}
	actual := jts.GeomUtil_GeometryMapper_FlatMap(geom, dim, op)
	expected, err := reader.Read(wktExpected)
	if err != nil {
		t.Fatalf("failed to parse expected WKT: %v", err)
	}
	if !expected.EqualsExact(actual) {
		t.Errorf("expected %v, got %v", expected, actual)
	}
}
