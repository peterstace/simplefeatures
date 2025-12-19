package jts_test

import (
	"testing"

	"github.com/peterstace/simplefeatures/internal/jtsport/jts"
	"github.com/peterstace/simplefeatures/internal/jtsport/junit"
)

func TestGeometryExtracterExtract(t *testing.T) {
	reader := jts.Io_NewWKTReader()
	gc, err := reader.Read("GEOMETRYCOLLECTION ( POINT (1 1), LINESTRING (0 0, 10 10), LINESTRING (10 10, 20 20), LINEARRING (10 10, 20 20, 15 15, 10 10), POLYGON ((0 0, 100 0, 100 100, 0 100, 0 0)), GEOMETRYCOLLECTION ( POINT (1 1) ) )")
	if err != nil {
		t.Fatalf("failed to parse WKT: %v", err)
	}

	// Verify that LinearRings are included when extracting LineStrings.
	lineStringsAndLinearRings := jts.GeomUtil_GeometryExtracter_Extract(gc, jts.Geom_Geometry_TypeNameLineString)
	junit.AssertEquals(t, 3, len(lineStringsAndLinearRings))

	// Verify that only LinearRings are extracted.
	linearRings := jts.GeomUtil_GeometryExtracter_Extract(gc, jts.Geom_Geometry_TypeNameLinearRing)
	junit.AssertEquals(t, 1, len(linearRings))

	// Verify that nested geometries are extracted.
	points := jts.GeomUtil_GeometryExtracter_Extract(gc, jts.Geom_Geometry_TypeNamePoint)
	junit.AssertEquals(t, 2, len(points))
}
