package jts_test

import (
	"testing"

	"github.com/peterstace/simplefeatures/internal/jtsport/jts"
	"github.com/peterstace/simplefeatures/internal/jtsport/junit"
)

func TestRelateGeometryUniquePoints(t *testing.T) {
	reader := jts.Io_NewWKTReader()
	geom, err := reader.Read("MULTIPOINT ((0 0), (5 5), (5 0), (0 0))")
	junit.AssertNull(t, err)
	rgeom := jts.OperationRelateng_NewRelateGeometry(geom)
	pts := rgeom.GetUniquePoints()
	junit.AssertEquals(t, 3, len(pts))
}

func TestRelateGeometryBoundary(t *testing.T) {
	reader := jts.Io_NewWKTReader()
	geom, err := reader.Read("MULTILINESTRING ((0 0, 9 9), (9 9, 5 1))")
	junit.AssertNull(t, err)
	rgeom := jts.OperationRelateng_NewRelateGeometry(geom)
	junit.AssertTrue(t, rgeom.HasBoundary())
}

func TestRelateGeometryHasDimension(t *testing.T) {
	reader := jts.Io_NewWKTReader()
	geom, err := reader.Read("GEOMETRYCOLLECTION (POLYGON ((1 9, 5 9, 5 5, 1 5, 1 9)), LINESTRING (1 1, 5 4), POINT (6 5))")
	junit.AssertNull(t, err)
	rgeom := jts.OperationRelateng_NewRelateGeometry(geom)
	junit.AssertTrue(t, rgeom.HasDimension(0))
	junit.AssertTrue(t, rgeom.HasDimension(1))
	junit.AssertTrue(t, rgeom.HasDimension(2))
}

func TestRelateGeometryDimension(t *testing.T) {
	tests := []struct {
		wkt             string
		expectedDim     int
		expectedDimReal int
	}{
		{"POINT (0 0)", 0, 0},
		{"LINESTRING (0 0, 0 0)", 1, 0},
		{"LINESTRING (0 0, 9 9)", 1, 1},
		{"LINESTRING (0 0, 0 0, 9 9)", 1, 1},
		{"POLYGON ((1 9, 5 9, 5 5, 1 5, 1 9))", 2, 2},
		{"GEOMETRYCOLLECTION (POLYGON ((1 9, 5 9, 5 5, 1 5, 1 9)), LINESTRING (1 1, 5 4), POINT (6 5))", 2, 2},
		{"GEOMETRYCOLLECTION (POLYGON EMPTY, LINESTRING (1 1, 5 4), POINT (6 5))", 2, 1},
	}

	reader := jts.Io_NewWKTReader()
	for _, tt := range tests {
		geom, err := reader.Read(tt.wkt)
		junit.AssertNull(t, err)
		rgeom := jts.OperationRelateng_NewRelateGeometry(geom)
		junit.AssertEquals(t, tt.expectedDim, rgeom.GetDimension())
		junit.AssertEquals(t, tt.expectedDimReal, rgeom.GetDimensionReal())
	}
}
