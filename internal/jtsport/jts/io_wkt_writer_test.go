package jts_test

import (
	"testing"

	"github.com/peterstace/simplefeatures/internal/jtsport/java"
	"github.com/peterstace/simplefeatures/internal/jtsport/jts"
)

func TestWKTWriterProperties(t *testing.T) {
	// Tests ported from WKTWriterTest.java testProperties.
	writer := jts.Io_NewWKTWriter()
	writer3D := jts.Io_NewWKTWriterWithDimension(3)
	writer2DM := jts.Io_NewWKTWriterWithDimension(3)
	writer2DM.SetOutputOrdinates(jts.Io_Ordinate_CreateXYM())

	// Check default output ordinates.
	if !writer.GetOutputOrdinates().Equals(jts.Io_Ordinate_CreateXY()) {
		t.Errorf("default writer: expected XY ordinates")
	}
	if !writer3D.GetOutputOrdinates().Equals(jts.Io_Ordinate_CreateXYZ()) {
		t.Errorf("3D writer: expected XYZ ordinates")
	}
	if !writer2DM.GetOutputOrdinates().Equals(jts.Io_Ordinate_CreateXYM()) {
		t.Errorf("2DM writer: expected XYM ordinates")
	}

	// Test 4D writer.
	writer4D := jts.Io_NewWKTWriterWithDimension(4)
	if !writer4D.GetOutputOrdinates().Equals(jts.Io_Ordinate_CreateXYZM()) {
		t.Errorf("4D writer: expected XYZM ordinates")
	}

	// Test SetOutputOrdinates.
	writer4D.SetOutputOrdinates(jts.Io_Ordinate_CreateXY())
	if !writer4D.GetOutputOrdinates().Equals(jts.Io_Ordinate_CreateXY()) {
		t.Errorf("after set XY: expected XY ordinates")
	}

	writer4D.SetOutputOrdinates(jts.Io_Ordinate_CreateXYZ())
	if !writer4D.GetOutputOrdinates().Equals(jts.Io_Ordinate_CreateXYZ()) {
		t.Errorf("after set XYZ: expected XYZ ordinates")
	}

	writer4D.SetOutputOrdinates(jts.Io_Ordinate_CreateXYM())
	if !writer4D.GetOutputOrdinates().Equals(jts.Io_Ordinate_CreateXYM()) {
		t.Errorf("after set XYM: expected XYM ordinates")
	}

	writer4D.SetOutputOrdinates(jts.Io_Ordinate_CreateXYZM())
	if !writer4D.GetOutputOrdinates().Equals(jts.Io_Ordinate_CreateXYZM()) {
		t.Errorf("after set XYZM: expected XYZM ordinates")
	}
}

func TestWKTWriterWritePoint(t *testing.T) {
	precisionModel := jts.Geom_NewPrecisionModelWithScale(1)
	geometryFactory := jts.Geom_NewGeometryFactoryWithPrecisionModelAndSRID(precisionModel, 0)
	writer := jts.Io_NewWKTWriter()

	point := geometryFactory.CreatePointFromCoordinate(jts.Geom_NewCoordinateWithXY(10, 10))
	got := writer.Write(point.Geom_Geometry)
	want := "POINT (10 10)"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestWKTWriterWriteLineString(t *testing.T) {
	precisionModel := jts.Geom_NewPrecisionModelWithScale(1)
	geometryFactory := jts.Geom_NewGeometryFactoryWithPrecisionModelAndSRID(precisionModel, 0)
	writer := jts.Io_NewWKTWriter()

	coordinates := []*jts.Geom_Coordinate{
		jts.Geom_NewCoordinateWithXYZ(10, 10, 0),
		jts.Geom_NewCoordinateWithXYZ(20, 20, 0),
		jts.Geom_NewCoordinateWithXYZ(30, 40, 0),
	}
	lineString := geometryFactory.CreateLineStringFromCoordinates(coordinates)
	got := writer.Write(lineString.Geom_Geometry)
	want := "LINESTRING (10 10, 20 20, 30 40)"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestWKTWriterWritePolygon(t *testing.T) {
	precisionModel := jts.Geom_NewPrecisionModelWithScale(1)
	geometryFactory := jts.Geom_NewGeometryFactoryWithPrecisionModelAndSRID(precisionModel, 0)
	writer := jts.Io_NewWKTWriter()

	coordinates := []*jts.Geom_Coordinate{
		jts.Geom_NewCoordinateWithXYZ(10, 10, 0),
		jts.Geom_NewCoordinateWithXYZ(10, 20, 0),
		jts.Geom_NewCoordinateWithXYZ(20, 20, 0),
		jts.Geom_NewCoordinateWithXYZ(20, 15, 0),
		jts.Geom_NewCoordinateWithXYZ(10, 10, 0),
	}
	linearRing := geometryFactory.CreateLinearRingFromCoordinates(coordinates)
	polygon := geometryFactory.CreatePolygonWithLinearRingAndHoles(linearRing, []*jts.Geom_LinearRing{})
	got := writer.Write(polygon.Geom_Geometry)
	want := "POLYGON ((10 10, 10 20, 20 20, 20 15, 10 10))"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestWKTWriterWriteMultiPoint(t *testing.T) {
	precisionModel := jts.Geom_NewPrecisionModelWithScale(1)
	geometryFactory := jts.Geom_NewGeometryFactoryWithPrecisionModelAndSRID(precisionModel, 0)
	writer := jts.Io_NewWKTWriter()

	points := []*jts.Geom_Point{
		geometryFactory.CreatePointFromCoordinate(jts.Geom_NewCoordinateWithXYZ(10, 10, 0)),
		geometryFactory.CreatePointFromCoordinate(jts.Geom_NewCoordinateWithXYZ(20, 20, 0)),
	}
	multiPoint := geometryFactory.CreateMultiPointFromPoints(points)
	got := writer.Write(multiPoint.Geom_Geometry)
	want := "MULTIPOINT ((10 10), (20 20))"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestWKTWriterWriteMultiLineString(t *testing.T) {
	precisionModel := jts.Geom_NewPrecisionModelWithScale(1)
	geometryFactory := jts.Geom_NewGeometryFactoryWithPrecisionModelAndSRID(precisionModel, 0)
	writer := jts.Io_NewWKTWriter()

	coordinates1 := []*jts.Geom_Coordinate{
		jts.Geom_NewCoordinateWithXYZ(10, 10, 0),
		jts.Geom_NewCoordinateWithXYZ(20, 20, 0),
	}
	lineString1 := geometryFactory.CreateLineStringFromCoordinates(coordinates1)

	coordinates2 := []*jts.Geom_Coordinate{
		jts.Geom_NewCoordinateWithXYZ(15, 15, 0),
		jts.Geom_NewCoordinateWithXYZ(30, 15, 0),
	}
	lineString2 := geometryFactory.CreateLineStringFromCoordinates(coordinates2)

	multiLineString := geometryFactory.CreateMultiLineStringFromLineStrings([]*jts.Geom_LineString{lineString1, lineString2})
	got := writer.Write(multiLineString.Geom_Geometry)
	want := "MULTILINESTRING ((10 10, 20 20), (15 15, 30 15))"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestWKTWriterWriteMultiPolygon(t *testing.T) {
	precisionModel := jts.Geom_NewPrecisionModelWithScale(1)
	geometryFactory := jts.Geom_NewGeometryFactoryWithPrecisionModelAndSRID(precisionModel, 0)
	writer := jts.Io_NewWKTWriter()

	coordinates1 := []*jts.Geom_Coordinate{
		jts.Geom_NewCoordinateWithXYZ(10, 10, 0),
		jts.Geom_NewCoordinateWithXYZ(10, 20, 0),
		jts.Geom_NewCoordinateWithXYZ(20, 20, 0),
		jts.Geom_NewCoordinateWithXYZ(20, 15, 0),
		jts.Geom_NewCoordinateWithXYZ(10, 10, 0),
	}
	linearRing1 := geometryFactory.CreateLinearRingFromCoordinates(coordinates1)
	polygon1 := geometryFactory.CreatePolygonWithLinearRingAndHoles(linearRing1, []*jts.Geom_LinearRing{})

	coordinates2 := []*jts.Geom_Coordinate{
		jts.Geom_NewCoordinateWithXYZ(60, 60, 0),
		jts.Geom_NewCoordinateWithXYZ(70, 70, 0),
		jts.Geom_NewCoordinateWithXYZ(80, 60, 0),
		jts.Geom_NewCoordinateWithXYZ(60, 60, 0),
	}
	linearRing2 := geometryFactory.CreateLinearRingFromCoordinates(coordinates2)
	polygon2 := geometryFactory.CreatePolygonWithLinearRingAndHoles(linearRing2, []*jts.Geom_LinearRing{})

	multiPolygon := geometryFactory.CreateMultiPolygonFromPolygons([]*jts.Geom_Polygon{polygon1, polygon2})
	got := writer.Write(multiPolygon.Geom_Geometry)
	want := "MULTIPOLYGON (((10 10, 10 20, 20 20, 20 15, 10 10)), ((60 60, 70 70, 80 60, 60 60)))"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestWKTWriterWriteGeometryCollection(t *testing.T) {
	precisionModel := jts.Geom_NewPrecisionModelWithScale(1)
	geometryFactory := jts.Geom_NewGeometryFactoryWithPrecisionModelAndSRID(precisionModel, 0)
	writer := jts.Io_NewWKTWriter()

	point1 := geometryFactory.CreatePointFromCoordinate(jts.Geom_NewCoordinateWithXY(10, 10))
	point2 := geometryFactory.CreatePointFromCoordinate(jts.Geom_NewCoordinateWithXY(30, 30))
	coordinates := []*jts.Geom_Coordinate{
		jts.Geom_NewCoordinateWithXYZ(15, 15, 0),
		jts.Geom_NewCoordinateWithXYZ(20, 20, 0),
	}
	lineString1 := geometryFactory.CreateLineStringFromCoordinates(coordinates)

	geometries := []*jts.Geom_Geometry{point1.Geom_Geometry, point2.Geom_Geometry, lineString1.Geom_Geometry}
	geometryCollection := geometryFactory.CreateGeometryCollectionFromGeometries(geometries)
	got := writer.Write(geometryCollection.Geom_Geometry)
	want := "GEOMETRYCOLLECTION (POINT (10 10), POINT (30 30), LINESTRING (15 15, 20 20))"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestWKTWriterWriteLargeNumbers1(t *testing.T) {
	precisionModel := jts.Geom_NewPrecisionModelWithScale(1e9)
	geometryFactory := jts.Geom_NewGeometryFactoryWithPrecisionModelAndSRID(precisionModel, 0)

	point1 := geometryFactory.CreatePointFromCoordinate(jts.Geom_NewCoordinateWithXY(123456789012345678, 10e9))
	got := point1.ToText()
	want := "POINT (123456789012345680 10000000000)"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestWKTWriterWriteLargeNumbers2(t *testing.T) {
	precisionModel := jts.Geom_NewPrecisionModelWithScale(1e9)
	geometryFactory := jts.Geom_NewGeometryFactoryWithPrecisionModelAndSRID(precisionModel, 0)

	point1 := geometryFactory.CreatePointFromCoordinate(jts.Geom_NewCoordinateWithXY(1234, 10e9))
	got := point1.ToText()
	want := "POINT (1234 10000000000)"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestWKTWriterWriteLargeNumbers3(t *testing.T) {
	precisionModel := jts.Geom_NewPrecisionModelWithScale(1e9)
	geometryFactory := jts.Geom_NewGeometryFactoryWithPrecisionModelAndSRID(precisionModel, 0)

	point1 := geometryFactory.CreatePointFromCoordinate(jts.Geom_NewCoordinateWithXY(123456789012345678000000e9, 10e9))
	got := point1.ToText()
	want := "POINT (123456789012345690000000000000000 10000000000)"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestWKTWriterWrite3D(t *testing.T) {
	geometryFactory := jts.Geom_NewGeometryFactoryDefault()
	writer3D := jts.Io_NewWKTWriterWithDimension(3)

	point := geometryFactory.CreatePointFromCoordinate(jts.Geom_NewCoordinateWithXYZ(1, 1, 1))
	got := writer3D.Write(point.Geom_Geometry)
	want := "POINT Z(1 1 1)"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}

	writer2DM := jts.Io_NewWKTWriterWithDimension(3)
	writer2DM.SetOutputOrdinates(jts.Io_Ordinate_CreateXYM())
	got = writer2DM.Write(point.Geom_Geometry)
	want = "POINT (1 1)"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestWKTWriterWrite3DWithNaN(t *testing.T) {
	geometryFactory := jts.Geom_NewGeometryFactoryDefault()
	writer3D := jts.Io_NewWKTWriterWithDimension(3)

	coordinates := []*jts.Geom_Coordinate{
		jts.Geom_NewCoordinateWithXY(1, 1),
		jts.Geom_NewCoordinateWithXYZ(2, 2, 2),
	}
	line := geometryFactory.CreateLineStringFromCoordinates(coordinates)
	got := writer3D.Write(line.Geom_Geometry)
	want := "LINESTRING Z(1 1 NaN, 2 2 2)"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}

	writer2DM := jts.Io_NewWKTWriterWithDimension(3)
	writer2DM.SetOutputOrdinates(jts.Io_Ordinate_CreateXYM())
	got = writer2DM.Write(line.Geom_Geometry)
	want = "LINESTRING (1 1, 2 2)"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestWKTWriterLineStringZM(t *testing.T) {
	geometryFactory := jts.Geom_NewGeometryFactoryDefault()
	writer4D := jts.Io_NewWKTWriterWithDimension(4)
	reader := jts.Io_NewWKTReader()

	coordinates := []*jts.Geom_Coordinate{
		jts.Geom_NewCoordinateXYZM4DWithXYZM(1, 2, 3, 4).Geom_Coordinate,
		jts.Geom_NewCoordinateXYZM4DWithXYZM(5, 6, 7, 8).Geom_Coordinate,
	}
	lineZM := geometryFactory.CreateLineStringFromCoordinates(coordinates)
	wkt := writer4D.Write(lineZM.Geom_Geometry)

	deserialized, err := reader.Read(wkt)
	if err != nil {
		t.Fatalf("failed to read WKT: %v", err)
	}

	deserializedLine := java.Cast[*jts.Geom_LineString](deserialized)
	p0 := deserializedLine.GetPointN(0).GetCoordinate()
	p1 := deserializedLine.GetPointN(1).GetCoordinate()

	if p0.GetX() != 1.0 {
		t.Errorf("p0.X: got %v, want 1.0", p0.GetX())
	}
	if p0.GetY() != 2.0 {
		t.Errorf("p0.Y: got %v, want 2.0", p0.GetY())
	}
	if p0.GetZ() != 3.0 {
		t.Errorf("p0.Z: got %v, want 3.0", p0.GetZ())
	}
	if p0.GetM() != 4.0 {
		t.Errorf("p0.M: got %v, want 4.0", p0.GetM())
	}

	if p1.GetX() != 5.0 {
		t.Errorf("p1.X: got %v, want 5.0", p1.GetX())
	}
	if p1.GetY() != 6.0 {
		t.Errorf("p1.Y: got %v, want 6.0", p1.GetY())
	}
	if p1.GetZ() != 7.0 {
		t.Errorf("p1.Z: got %v, want 7.0", p1.GetZ())
	}
	if p1.GetM() != 8.0 {
		t.Errorf("p1.M: got %v, want 8.0", p1.GetM())
	}
}
