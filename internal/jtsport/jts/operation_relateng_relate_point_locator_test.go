package jts_test

import (
	"testing"

	"github.com/peterstace/simplefeatures/internal/jtsport/jts"
)

var gcPLA = "GEOMETRYCOLLECTION (POINT (1 1), POINT (2 1), LINESTRING (3 1, 3 9), LINESTRING (4 1, 5 4, 7 1, 4 1), LINESTRING (12 12, 14 14), POLYGON ((6 5, 6 9, 9 9, 9 5, 6 5)), POLYGON ((10 10, 10 16, 16 16, 16 10, 10 10)), POLYGON ((11 11, 11 17, 17 17, 17 11, 11 11)), POLYGON ((12 12, 12 16, 16 16, 16 12, 12 12)))"

func TestRelatePointLocatorPoint(t *testing.T) {
	checkDimLocation(t, gcPLA, 1, 1, jts.OperationRelateng_DimensionLocation_POINT_INTERIOR)
	checkDimLocation(t, gcPLA, 0, 1, jts.OperationRelateng_DimensionLocation_EXTERIOR)
}

func TestRelatePointLocatorPointInLine(t *testing.T) {
	checkDimLocation(t, gcPLA, 3, 8, jts.OperationRelateng_DimensionLocation_LINE_INTERIOR)
}

func TestRelatePointLocatorPointInArea(t *testing.T) {
	checkDimLocation(t, gcPLA, 8, 8, jts.OperationRelateng_DimensionLocation_AREA_INTERIOR)
}

func TestRelatePointLocatorLine(t *testing.T) {
	checkDimLocation(t, gcPLA, 3, 3, jts.OperationRelateng_DimensionLocation_LINE_INTERIOR)
	checkDimLocation(t, gcPLA, 3, 1, jts.OperationRelateng_DimensionLocation_LINE_BOUNDARY)
}

func TestRelatePointLocatorLineInArea(t *testing.T) {
	checkDimLocation(t, gcPLA, 11, 11, jts.OperationRelateng_DimensionLocation_AREA_INTERIOR)
	checkDimLocation(t, gcPLA, 14, 14, jts.OperationRelateng_DimensionLocation_AREA_INTERIOR)
}

func TestRelatePointLocatorArea(t *testing.T) {
	checkDimLocation(t, gcPLA, 8, 8, jts.OperationRelateng_DimensionLocation_AREA_INTERIOR)
	checkDimLocation(t, gcPLA, 9, 9, jts.OperationRelateng_DimensionLocation_AREA_BOUNDARY)
}

func TestRelatePointLocatorAreaInArea(t *testing.T) {
	checkDimLocation(t, gcPLA, 11, 11, jts.OperationRelateng_DimensionLocation_AREA_INTERIOR)
	checkDimLocation(t, gcPLA, 12, 12, jts.OperationRelateng_DimensionLocation_AREA_INTERIOR)
	checkDimLocation(t, gcPLA, 10, 10, jts.OperationRelateng_DimensionLocation_AREA_BOUNDARY)
	checkDimLocation(t, gcPLA, 16, 16, jts.OperationRelateng_DimensionLocation_AREA_INTERIOR)
}

func TestRelatePointLocatorLineNode(t *testing.T) {
	checkNodeLocation(t, gcPLA, 3, 1, jts.Geom_Location_Boundary)
}

func TestRelatePointLocatorLineEndInGCLA(t *testing.T) {
	wkt := "GEOMETRYCOLLECTION (POLYGON ((0 0, 10 0, 10 10, 0 10, 0 0)), LINESTRING (12 2, 0 2, 0 5, 5 5), LINESTRING (12 10, 12 2))"
	checkLineEndDimLocation(t, wkt, 5, 5, jts.OperationRelateng_DimensionLocation_AREA_INTERIOR)
	checkLineEndDimLocation(t, wkt, 12, 2, jts.OperationRelateng_DimensionLocation_LINE_INTERIOR)
	checkLineEndDimLocation(t, wkt, 12, 10, jts.OperationRelateng_DimensionLocation_LINE_BOUNDARY)
}

func checkDimLocation(t *testing.T, wkt string, x, y float64, expectedDimLoc int) {
	t.Helper()
	reader := jts.Io_NewWKTReader()
	geom, err := reader.Read(wkt)
	if err != nil {
		t.Fatalf("Failed to read geometry: %v", err)
	}
	locator := jts.OperationRelateng_NewRelatePointLocator(geom)
	coord := &jts.Geom_Coordinate{X: x, Y: y}
	actual := locator.LocateWithDim(coord)
	if actual != expectedDimLoc {
		t.Errorf("LocateWithDim at (%v, %v): got %v, expected %v", x, y, actual, expectedDimLoc)
	}
}

func checkLineEndDimLocation(t *testing.T, wkt string, x, y float64, expectedDimLoc int) {
	t.Helper()
	reader := jts.Io_NewWKTReader()
	geom, err := reader.Read(wkt)
	if err != nil {
		t.Fatalf("Failed to read geometry: %v", err)
	}
	locator := jts.OperationRelateng_NewRelatePointLocator(geom)
	coord := &jts.Geom_Coordinate{X: x, Y: y}
	actual := locator.LocateLineEndWithDim(coord)
	if actual != expectedDimLoc {
		t.Errorf("LocateLineEndWithDim at (%v, %v): got %v, expected %v", x, y, actual, expectedDimLoc)
	}
}

func checkNodeLocation(t *testing.T, wkt string, x, y float64, expectedLoc int) {
	t.Helper()
	reader := jts.Io_NewWKTReader()
	geom, err := reader.Read(wkt)
	if err != nil {
		t.Fatalf("Failed to read geometry: %v", err)
	}
	locator := jts.OperationRelateng_NewRelatePointLocator(geom)
	coord := &jts.Geom_Coordinate{X: x, Y: y}
	actual := locator.LocateNode(coord, nil)
	if actual != expectedLoc {
		t.Errorf("LocateNode at (%v, %v): got %v, expected %v", x, y, actual, expectedLoc)
	}
}
