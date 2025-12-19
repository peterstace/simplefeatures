package jts_test

import (
	"math"
	"testing"

	"github.com/peterstace/simplefeatures/internal/jtsport/jts"
	"github.com/peterstace/simplefeatures/internal/jtsport/junit"
)

var ioWKBTestGeomFactory = jts.Geom_NewGeometryFactoryDefault()
var ioWKBTestRdr = jts.Io_NewWKTReaderWithFactory(ioWKBTestGeomFactory)

func TestWKBFirst(t *testing.T) {
	runWKBTest(t, "MULTIPOINT ((0 0), (1 4), (100 200))")
}

func TestWKBPointPCS(t *testing.T) {
	runWKBTestPackedCoordinate(t, "POINT (1 2)")
}

func TestWKBPoint(t *testing.T) {
	runWKBTest(t, "POINT (1 2)")
}

func TestWKBPointEmpty(t *testing.T) {
	runWKBTest(t, "POINT EMPTY")
}

func TestWKBLineString(t *testing.T) {
	runWKBTest(t, "LINESTRING (1 2, 10 20, 100 200)")
}

func TestWKBPolygon(t *testing.T) {
	runWKBTest(t, "POLYGON ((0 0, 100 0, 100 100, 0 100, 0 0))")
}

func TestWKBPolygonWithHole(t *testing.T) {
	runWKBTest(t, "POLYGON ((0 0, 100 0, 100 100, 0 100, 0 0), (1 1, 1 10, 10 10, 10 1, 1 1) )")
}

func TestWKBMultiPoint(t *testing.T) {
	runWKBTest(t, "MULTIPOINT ((0 0), (1 4), (100 200))")
}

func TestWKBMultiLineString(t *testing.T) {
	runWKBTest(t, "MULTILINESTRING ((0 0, 1 10), (10 10, 20 30), (123 123, 456 789))")
}

func TestWKBMultiPolygon(t *testing.T) {
	runWKBTest(t, "MULTIPOLYGON ( ((0 0, 100 0, 100 100, 0 100, 0 0), (1 1, 1 10, 10 10, 10 1, 1 1) ), ((200 200, 200 250, 250 250, 250 200, 200 200)) )")
}

func TestWKBGeometryCollection(t *testing.T) {
	runWKBTest(t, "GEOMETRYCOLLECTION ( POINT ( 1 1), LINESTRING (0 0, 10 10), POLYGON ((0 0, 100 0, 100 100, 0 100, 0 0)) )")
}

func TestWKBNestedGeometryCollection(t *testing.T) {
	runWKBTest(t, "GEOMETRYCOLLECTION ( POINT (20 20), GEOMETRYCOLLECTION ( POINT ( 1 1), LINESTRING (0 0, 10 10), POLYGON ((0 0, 100 0, 100 100, 0 100, 0 0)) ) )")
}

func TestWKBLineStringEmpty(t *testing.T) {
	runWKBTest(t, "LINESTRING EMPTY")
}

func TestWKBGeometryCollectionContainingEmptyGeometries(t *testing.T) {
	runWKBTest(t, "GEOMETRYCOLLECTION (LINESTRING EMPTY, MULTIPOINT EMPTY)")
}

func TestWKBBigPolygon(t *testing.T) {
	t.Skip("Skipping: Util_GeometricShapeFactory not yet ported")
	// shapeFactory := jts.Util_NewGeometricShapeFactory(ioWKBTestGeomFactory)
	// shapeFactory.SetBase(jts.Geom_NewCoordinateWithXY(0, 0))
	// shapeFactory.SetSize(1000)
	// shapeFactory.SetNumPoints(1000)
	// geom := shapeFactory.CreateRectangle()
	// ioWKBTestRunWKBTest(t, geom, 2, false)
}

func TestWKBPolygonEmpty(t *testing.T) {
	runWKBTest(t, "POLYGON EMPTY")
}

func TestWKBMultiPointEmpty(t *testing.T) {
	runWKBTest(t, "MULTIPOINT EMPTY")
}

func TestWKBMultiLineStringEmpty(t *testing.T) {
	runWKBTest(t, "MULTILINESTRING EMPTY")
}

func TestWKBMultiPolygonEmpty(t *testing.T) {
	runWKBTest(t, "MULTIPOLYGON EMPTY")
}

func TestWKBGeometryCollectionEmpty(t *testing.T) {
	runWKBTest(t, "GEOMETRYCOLLECTION EMPTY")
}

func TestWKBWriteAndReadM(t *testing.T) {
	wkt := "MULTILINESTRING M((1 1 1, 2 2 2))"
	wktReader := jts.Io_NewWKTReader()
	geometryBefore, err := wktReader.Read(wkt)
	if err != nil {
		t.Fatalf("parsing WKT: %v", err)
	}

	wkbWriter := jts.Io_NewWKBWriterWithDimension(3)
	outputOrdinates := jts.Io_Ordinate_CreateXY()
	outputOrdinates.Add(jts.Io_Ordinate_M)
	wkbWriter.SetOutputOrdinates(outputOrdinates)
	write := wkbWriter.Write(geometryBefore)

	wkbReader := jts.Io_NewWKBReader()
	geometryAfter, err := wkbReader.ReadBytes(write)
	if err != nil {
		t.Fatalf("reading WKB: %v", err)
	}

	junit.AssertEquals(t, 1.0, geometryAfter.GetCoordinates()[0].GetX())
	junit.AssertEquals(t, 1.0, geometryAfter.GetCoordinates()[0].GetY())
	junit.AssertTrue(t, math.IsNaN(geometryAfter.GetCoordinates()[0].GetZ()))
	junit.AssertEquals(t, 1.0, geometryAfter.GetCoordinates()[0].GetM())
}

func TestWKBWriteAndReadZ(t *testing.T) {
	wkt := "MULTILINESTRING ((1 1 1, 2 2 2))"
	wktReader := jts.Io_NewWKTReader()
	geometryBefore, err := wktReader.Read(wkt)
	if err != nil {
		t.Fatalf("parsing WKT: %v", err)
	}

	wkbWriter := jts.Io_NewWKBWriterWithDimension(3)
	write := wkbWriter.Write(geometryBefore)

	wkbReader := jts.Io_NewWKBReader()
	geometryAfter, err := wkbReader.ReadBytes(write)
	if err != nil {
		t.Fatalf("reading WKB: %v", err)
	}

	junit.AssertEquals(t, 1.0, geometryAfter.GetCoordinates()[0].GetX())
	junit.AssertEquals(t, 1.0, geometryAfter.GetCoordinates()[0].GetY())
	junit.AssertEquals(t, 1.0, geometryAfter.GetCoordinates()[0].GetZ())
	junit.AssertTrue(t, math.IsNaN(geometryAfter.GetCoordinates()[0].GetM()))
}

func runWKBTest(t *testing.T, wkt string) {
	t.Helper()
	runWKBTestCoordinateArray(t, wkt)
}

func runWKBTestPackedCoordinate(t *testing.T, wkt string) {
	t.Helper()
	geomFactory := jts.Geom_NewGeometryFactoryWithCoordinateSequenceFactory(
		jts.GeomImpl_NewPackedCoordinateSequenceFactoryWithType(jts.GeomImpl_PackedCoordinateSequenceFactory_DOUBLE))
	rdr := jts.Io_NewWKTReaderWithFactory(geomFactory)
	g, err := rdr.Read(wkt)
	if err != nil {
		t.Fatalf("parsing WKT: %v", err)
	}
	// Since we are using a PCS of dim=2, only check 2-dimensional storage.
	ioWKBTestRunWKBTest(t, g, 2, true)
	ioWKBTestRunWKBTest(t, g, 2, false)
}

func runWKBTestCoordinateArray(t *testing.T, wkt string) {
	t.Helper()
	geomFactory := jts.Geom_NewGeometryFactoryDefault()
	rdr := jts.Io_NewWKTReaderWithFactory(geomFactory)
	g, err := rdr.Read(wkt)
	if err != nil {
		t.Fatalf("parsing WKT: %v", err)
	}

	// CoordinateArrays support dimension 3, so test both dimensions.
	ioWKBTestRunWKBTest(t, g, 2, true)
	ioWKBTestRunWKBTest(t, g, 2, false)
	ioWKBTestRunWKBTest(t, g, 3, true)
	ioWKBTestRunWKBTest(t, g, 3, false)
}

func ioWKBTestRunWKBTest(t *testing.T, g *jts.Geom_Geometry, dimension int, toHex bool) {
	t.Helper()
	ioWKBTestSetZ(g)
	ioWKBTestRunWKBTestWithByteOrder(t, g, dimension, jts.Io_ByteOrderValues_LITTLE_ENDIAN, toHex)
	ioWKBTestRunWKBTestWithByteOrder(t, g, dimension, jts.Io_ByteOrderValues_BIG_ENDIAN, toHex)
}

func ioWKBTestRunWKBTestWithByteOrder(t *testing.T, g *jts.Geom_Geometry, dimension, byteOrder int, toHex bool) {
	t.Helper()
	ioWKBTestRunGeometry(t, g, dimension, byteOrder, toHex, 100)
	ioWKBTestRunGeometry(t, g, dimension, byteOrder, toHex, 0)
	ioWKBTestRunGeometry(t, g, dimension, byteOrder, toHex, 101010)
	ioWKBTestRunGeometry(t, g, dimension, byteOrder, toHex, -1)
}

func ioWKBTestSetZ(g *jts.Geom_Geometry) {
	g.ApplyCoordinateFilter(ioWKBTestNewAverageZFilter())
}

var ioWKBTestComp2 = jts.Geom_NewCoordinateSequenceComparatorWithDimensionLimit(2)
var ioWKBTestComp3 = jts.Geom_NewCoordinateSequenceComparatorWithDimensionLimit(3)

var ioWKBTestWKBReader = jts.Io_NewWKBReaderWithFactory(ioWKBTestGeomFactory)

func ioWKBTestRunGeometry(t *testing.T, g *jts.Geom_Geometry, dimension, byteOrder int, toHex bool, srid int) {
	t.Helper()

	includeSRID := false
	if srid >= 0 {
		includeSRID = true
		g.SetSRID(srid)
	}

	wkbWriter := jts.Io_NewWKBWriterWithDimensionOrderAndSRID(dimension, byteOrder, includeSRID)
	wkb := wkbWriter.Write(g)
	var wkbHex string
	if toHex {
		wkbHex = jts.Io_WKBWriter_ToHex(wkb)
	}

	if toHex {
		wkb = jts.Io_WKBReader_HexToBytes(wkbHex)
	}
	g2, err := ioWKBTestWKBReader.ReadBytes(wkb)
	if err != nil {
		t.Fatalf("reading WKB: %v", err)
	}

	var comp *jts.Geom_CoordinateSequenceComparator
	if dimension == 2 {
		comp = ioWKBTestComp2
	} else {
		comp = ioWKBTestComp3
	}
	isEqual := g.CompareToWithComparator(g2, comp) == 0
	junit.AssertTrue(t, isEqual)

	if includeSRID {
		isSRIDEqual := g.GetSRID() == g2.GetSRID()
		junit.AssertTrue(t, isSRIDEqual)
	}
}

type ioWKBTestAverageZFilter struct{}

var _ jts.Geom_CoordinateFilter = (*ioWKBTestAverageZFilter)(nil)

func (f *ioWKBTestAverageZFilter) IsGeom_CoordinateFilter() {}

func ioWKBTestNewAverageZFilter() *ioWKBTestAverageZFilter {
	return &ioWKBTestAverageZFilter{}
}

func (f *ioWKBTestAverageZFilter) Filter(coord *jts.Geom_Coordinate) {
	coord.SetZ((coord.GetX() + coord.GetY()) / 2)
}
