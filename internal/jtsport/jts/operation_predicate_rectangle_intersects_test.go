package jts

import (
	"testing"

	"github.com/peterstace/simplefeatures/internal/jtsport/java"
	"github.com/peterstace/simplefeatures/internal/jtsport/junit"
)

func TestRectangleIntersects_XYZM(t *testing.T) {
	geomFact := Geom_NewGeometryFactoryWithCoordinateSequenceFactory(GeomImpl_PackedCoordinateSequenceFactory_DOUBLE_FACTORY)
	rdr := Io_NewWKTReaderWithFactory(geomFact)
	rectGeom, err := rdr.Read("POLYGON ZM ((1 9 2 3, 9 9 2 3, 9 1 2 3, 1 1 2 3, 1 9 2 3))")
	if err != nil {
		t.Fatalf("failed to read rect: %v", err)
	}
	rect := java.Cast[*Geom_Polygon](rectGeom)
	line, err := rdr.Read("LINESTRING ZM (5 15 5 5, 15 5 5 5)")
	if err != nil {
		t.Fatalf("failed to read line: %v", err)
	}
	rectIntersects := OperationPredicate_RectangleIntersects_Intersects(rect, line)
	junit.AssertEquals(t, false, rectIntersects)
}
