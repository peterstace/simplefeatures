package jts_test

import (
	"testing"

	"github.com/peterstace/simplefeatures/internal/jtsport/java"
	"github.com/peterstace/simplefeatures/internal/jtsport/jts"
)

func TestWKTReaderPoint(t *testing.T) {
	reader := jts.Io_NewWKTReader()
	geom, err := reader.Read("POINT (10 20)")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	pt := java.Cast[*jts.Geom_Point](geom)
	if pt.GetX() != 10 {
		t.Errorf("expected X=10, got %v", pt.GetX())
	}
	if pt.GetY() != 20 {
		t.Errorf("expected Y=20, got %v", pt.GetY())
	}
}

func TestWKTReaderPointEmpty(t *testing.T) {
	reader := jts.Io_NewWKTReader()
	geom, err := reader.Read("POINT EMPTY")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !geom.IsEmpty() {
		t.Errorf("expected empty geometry")
	}
}

func TestWKTReaderLineString(t *testing.T) {
	reader := jts.Io_NewWKTReader()
	geom, err := reader.Read("LINESTRING (0 0, 10 10, 20 20)")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ls := java.Cast[*jts.Geom_LineString](geom)
	if ls.GetNumPoints() != 3 {
		t.Errorf("expected 3 points, got %d", ls.GetNumPoints())
	}
}

func TestWKTReaderPolygon(t *testing.T) {
	// Tests ported from WKTReaderTest.java testPolygon.
	reader := jts.Io_NewWKTReader()

	// Test basic 2D polygon with 2 holes.
	geom, err := reader.Read("POLYGON ((10 10, 10 20, 20 20, 20 15, 10 10), (11 11, 12 11, 12 12, 12 11, 11 11), (11 19, 11 18, 12 18, 12 19, 11 19))")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	poly := java.Cast[*jts.Geom_Polygon](geom)
	if poly.GetNumInteriorRing() != 2 {
		t.Errorf("expected 2 interior rings, got %d", poly.GetNumInteriorRing())
	}
	shellCS := poly.GetExteriorRing().GetCoordinateSequence()
	checkCoordXY(t, shellCS, 0, 10, 10)
	checkCoordXY(t, shellCS, 1, 10, 20)
	checkCoordXY(t, shellCS, 2, 20, 20)
	ring0CS := poly.GetInteriorRingN(0).GetCoordinateSequence()
	checkCoordXY(t, ring0CS, 0, 11, 11)
	ring1CS := poly.GetInteriorRingN(1).GetCoordinateSequence()
	checkCoordXY(t, ring1CS, 0, 11, 19)

	// Test EMPTY.
	geom, err = reader.Read("POLYGON EMPTY")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	poly = java.Cast[*jts.Geom_Polygon](geom)
	if !poly.IsEmpty() {
		t.Errorf("expected empty polygon")
	}

	// Test XYZ.
	geom, err = reader.Read("POLYGON Z((10 10 10, 10 20 10, 20 20 10, 20 15 10, 10 10 10), (11 11 10, 12 11 10, 12 12 10, 12 11 10, 11 11 10))")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	poly = java.Cast[*jts.Geom_Polygon](geom)
	shellCS = poly.GetExteriorRing().GetCoordinateSequence()
	if !shellCS.HasZ() {
		t.Errorf("expected coordinate sequence to have Z")
	}
	checkCoordXYZ(t, shellCS, 0, 10, 10, 10)

	// Test XYM.
	geom, err = reader.Read("POLYGON M((10 10 11, 10 20 11, 20 20 11, 20 15 11, 10 10 11), (11 11 11, 12 11 11, 12 12 11, 12 11 11, 11 11 11))")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	poly = java.Cast[*jts.Geom_Polygon](geom)
	shellCS = poly.GetExteriorRing().GetCoordinateSequence()
	if !shellCS.HasM() {
		t.Errorf("expected coordinate sequence to have M")
	}

	// Test XYZM.
	geom, err = reader.Read("POLYGON ZM((10 10 10 11, 10 20 10 11, 20 20 10 11, 20 15 10 11, 10 10 10 11), (11 11 10 11, 12 11 10 11, 12 12 10 11, 12 11 10 11, 11 11 10 11))")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	poly = java.Cast[*jts.Geom_Polygon](geom)
	shellCS = poly.GetExteriorRing().GetCoordinateSequence()
	if !shellCS.HasZ() || !shellCS.HasM() {
		t.Errorf("expected coordinate sequence to have Z and M")
	}
}

func TestWKTReaderPolygonWithHole(t *testing.T) {
	reader := jts.Io_NewWKTReader()
	geom, err := reader.Read("POLYGON ((0 0, 100 0, 100 100, 0 100, 0 0), (10 10, 20 10, 20 20, 10 20, 10 10))")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	poly := java.Cast[*jts.Geom_Polygon](geom)
	if poly.GetNumInteriorRing() != 1 {
		t.Errorf("expected 1 interior ring, got %d", poly.GetNumInteriorRing())
	}
}

func TestWKTReaderMultiPoint(t *testing.T) {
	reader := jts.Io_NewWKTReader()
	geom, err := reader.Read("MULTIPOINT ((10 10), (20 20))")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	mp := java.Cast[*jts.Geom_MultiPoint](geom)
	if mp.GetNumGeometries() != 2 {
		t.Errorf("expected 2 points, got %d", mp.GetNumGeometries())
	}
}

func TestWKTReaderMultiPointOldSyntax(t *testing.T) {
	reader := jts.Io_NewWKTReader()
	geom, err := reader.Read("MULTIPOINT (10 10, 20 20)")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	mp := java.Cast[*jts.Geom_MultiPoint](geom)
	if mp.GetNumGeometries() != 2 {
		t.Errorf("expected 2 points, got %d", mp.GetNumGeometries())
	}
}

func TestWKTReaderMultiLineString(t *testing.T) {
	reader := jts.Io_NewWKTReader()
	geom, err := reader.Read("MULTILINESTRING ((0 0, 10 10), (20 20, 30 30))")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	mls := java.Cast[*jts.Geom_MultiLineString](geom)
	if mls.GetNumGeometries() != 2 {
		t.Errorf("expected 2 linestrings, got %d", mls.GetNumGeometries())
	}
}

func TestWKTReaderMultiPolygon(t *testing.T) {
	// Tests ported from WKTReaderTest.java testMultiPolygonXY.
	reader := jts.Io_NewWKTReader()

	// Test MultiPolygon with first polygon having a hole.
	geom, err := reader.Read("MULTIPOLYGON (((10 10, 10 20, 20 20, 20 15, 10 10), (11 11, 12 11, 12 12, 12 11, 11 11)), ((60 60, 70 70, 80 60, 60 60)))")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	mpoly := java.Cast[*jts.Geom_MultiPolygon](geom)
	if mpoly.GetNumGeometries() != 2 {
		t.Errorf("expected 2 polygons, got %d", mpoly.GetNumGeometries())
	}

	// Verify first polygon exterior ring coordinates.
	poly0 := java.Cast[*jts.Geom_Polygon](mpoly.GetGeometryN(0))
	shell0CS := poly0.GetExteriorRing().GetCoordinateSequence()
	checkCoordXY(t, shell0CS, 0, 10, 10)
	checkCoordXY(t, shell0CS, 1, 10, 20)
	checkCoordXY(t, shell0CS, 2, 20, 20)
	checkCoordXY(t, shell0CS, 3, 20, 15)
	checkCoordXY(t, shell0CS, 4, 10, 10)

	// Verify first polygon interior ring coordinates.
	hole0CS := poly0.GetInteriorRingN(0).GetCoordinateSequence()
	checkCoordXY(t, hole0CS, 0, 11, 11)
	checkCoordXY(t, hole0CS, 1, 12, 11)

	// Verify second polygon exterior ring coordinates.
	poly1 := java.Cast[*jts.Geom_Polygon](mpoly.GetGeometryN(1))
	shell1CS := poly1.GetExteriorRing().GetCoordinateSequence()
	checkCoordXY(t, shell1CS, 0, 60, 60)
	checkCoordXY(t, shell1CS, 1, 70, 70)
	checkCoordXY(t, shell1CS, 2, 80, 60)
}

func TestWKTReaderGeometryCollection(t *testing.T) {
	reader := jts.Io_NewWKTReader()
	geom, err := reader.Read("GEOMETRYCOLLECTION (POINT (10 10), LINESTRING (0 0, 10 10))")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	gc := java.Cast[*jts.Geom_GeometryCollection](geom)
	if gc.GetNumGeometries() != 2 {
		t.Errorf("expected 2 geometries, got %d", gc.GetNumGeometries())
	}
}

func TestWKTReaderLinearRing(t *testing.T) {
	// Tests ported from WKTReaderTest.java testLinearRing.
	reader := jts.Io_NewWKTReader()

	// Test basic 2D LinearRing.
	geom, err := reader.Read("LINEARRING (10 10, 20 20, 30 40, 10 10)")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	lr := java.Cast[*jts.Geom_LinearRing](geom)
	if lr.GetNumPoints() != 4 {
		t.Errorf("expected 4 points, got %d", lr.GetNumPoints())
	}
	cs := lr.GetCoordinateSequence()
	checkCoordXY(t, cs, 0, 10, 10)
	checkCoordXY(t, cs, 1, 20, 20)
	checkCoordXY(t, cs, 2, 30, 40)
	checkCoordXY(t, cs, 3, 10, 10)

	// Test EMPTY.
	geom, err = reader.Read("LINEARRING EMPTY")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	lr = java.Cast[*jts.Geom_LinearRing](geom)
	if !lr.IsEmpty() {
		t.Errorf("expected empty linearring")
	}

	// Test XYZ.
	geom, err = reader.Read("LINEARRING Z(10 10 10, 20 20 10, 30 40 10, 10 10 10)")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	lr = java.Cast[*jts.Geom_LinearRing](geom)
	cs = lr.GetCoordinateSequence()
	if !cs.HasZ() {
		t.Errorf("expected coordinate sequence to have Z")
	}
	checkCoordXYZ(t, cs, 0, 10, 10, 10)

	// Test XYM.
	geom, err = reader.Read("LINEARRING M(10 10 11, 20 20 11, 30 40 11, 10 10 11)")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	lr = java.Cast[*jts.Geom_LinearRing](geom)
	cs = lr.GetCoordinateSequence()
	if !cs.HasM() {
		t.Errorf("expected coordinate sequence to have M")
	}

	// Test XYZM.
	geom, err = reader.Read("LINEARRING ZM(10 10 10 11, 20 20 10 11, 30 40 10 11, 10 10 10 11)")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	lr = java.Cast[*jts.Geom_LinearRing](geom)
	cs = lr.GetCoordinateSequence()
	if !cs.HasZ() || !cs.HasM() {
		t.Errorf("expected coordinate sequence to have Z and M")
	}
}

func checkCoordXY(t *testing.T, cs jts.Geom_CoordinateSequence, idx int, x, y float64) {
	t.Helper()
	if cs.GetX(idx) != x {
		t.Errorf("coord %d: expected X=%v, got %v", idx, x, cs.GetX(idx))
	}
	if cs.GetY(idx) != y {
		t.Errorf("coord %d: expected Y=%v, got %v", idx, y, cs.GetY(idx))
	}
}

func checkCoordXYZ(t *testing.T, cs jts.Geom_CoordinateSequence, idx int, x, y, z float64) {
	t.Helper()
	checkCoordXY(t, cs, idx, x, y)
	if cs.GetZ(idx) != z {
		t.Errorf("coord %d: expected Z=%v, got %v", idx, z, cs.GetZ(idx))
	}
}

func TestWKTReaderCaseInsensitive(t *testing.T) {
	reader := jts.Io_NewWKTReader()
	geom, err := reader.Read("point (10 20)")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	pt := java.Cast[*jts.Geom_Point](geom)
	if pt.GetX() != 10 || pt.GetY() != 20 {
		t.Errorf("unexpected coordinates")
	}
}

func TestWKTReaderPointZ(t *testing.T) {
	reader := jts.Io_NewWKTReader()
	geom, err := reader.Read("POINT Z(10 20 30)")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	pt := java.Cast[*jts.Geom_Point](geom)
	coord := pt.GetCoordinate()
	if coord.GetX() != 10 || coord.GetY() != 20 || coord.GetZ() != 30 {
		t.Errorf("unexpected coordinates: X=%v, Y=%v, Z=%v", coord.GetX(), coord.GetY(), coord.GetZ())
	}
}

func TestWKTReaderLinearRingNotClosed(t *testing.T) {
	reader := jts.Io_NewWKTReader()
	// In Go, the linear ring construction panics when the ring is not closed.
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic for unclosed ring")
		}
	}()
	_, _ = reader.Read("LINEARRING (10 10, 20 20, 30 40, 10 99)")
}

func TestWKTReaderMultiPointEmpty(t *testing.T) {
	reader := jts.Io_NewWKTReader()
	geom, err := reader.Read("MULTIPOINT EMPTY")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	mp := java.Cast[*jts.Geom_MultiPoint](geom)
	if !mp.IsEmpty() {
		t.Errorf("expected empty multipoint")
	}
	if mp.GetNumGeometries() != 0 {
		t.Errorf("expected 0 geometries, got %d", mp.GetNumGeometries())
	}
}

func TestWKTReaderMultiPointWithEmpty(t *testing.T) {
	reader := jts.Io_NewWKTReader()
	geom, err := reader.Read("MULTIPOINT ((10 10), EMPTY, (20 20))")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	mp := java.Cast[*jts.Geom_MultiPoint](geom)
	if mp.GetNumGeometries() != 3 {
		t.Errorf("expected 3 geometries, got %d", mp.GetNumGeometries())
	}
	pt0 := java.Cast[*jts.Geom_Point](mp.GetGeometryN(0))
	if pt0.GetX() != 10 || pt0.GetY() != 10 {
		t.Errorf("unexpected coordinates for point 0")
	}
	pt1 := java.Cast[*jts.Geom_Point](mp.GetGeometryN(1))
	if !pt1.IsEmpty() {
		t.Errorf("expected point 1 to be empty")
	}
	pt2 := java.Cast[*jts.Geom_Point](mp.GetGeometryN(2))
	if pt2.GetX() != 20 || pt2.GetY() != 20 {
		t.Errorf("unexpected coordinates for point 2")
	}
}

func TestWKTReaderMultiPointXYZ(t *testing.T) {
	reader := jts.Io_NewWKTReader()
	geom, err := reader.Read("MULTIPOINT Z((10 10 10), (20 20 10))")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	mp := java.Cast[*jts.Geom_MultiPoint](geom)
	if mp.GetNumGeometries() != 2 {
		t.Errorf("expected 2 points, got %d", mp.GetNumGeometries())
	}
	pt0 := java.Cast[*jts.Geom_Point](mp.GetGeometryN(0))
	cs0 := pt0.GetCoordinateSequence()
	if cs0.GetZ(0) != 10 {
		t.Errorf("expected Z=10, got %v", cs0.GetZ(0))
	}
}

func TestWKTReaderMultiPointXYM(t *testing.T) {
	reader := jts.Io_NewWKTReader()
	geom, err := reader.Read("MULTIPOINT M((10 10 11), (20 20 11))")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	mp := java.Cast[*jts.Geom_MultiPoint](geom)
	if mp.GetNumGeometries() != 2 {
		t.Errorf("expected 2 points, got %d", mp.GetNumGeometries())
	}
	pt0 := java.Cast[*jts.Geom_Point](mp.GetGeometryN(0))
	cs0 := pt0.GetCoordinateSequence()
	if !cs0.HasM() {
		t.Errorf("expected coordinate sequence to have M")
	}
}

func TestWKTReaderMultiPointXYZM(t *testing.T) {
	reader := jts.Io_NewWKTReader()
	geom, err := reader.Read("MULTIPOINT ZM((10 10 10 11), (20 20 10 11))")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	mp := java.Cast[*jts.Geom_MultiPoint](geom)
	if mp.GetNumGeometries() != 2 {
		t.Errorf("expected 2 points, got %d", mp.GetNumGeometries())
	}
	pt0 := java.Cast[*jts.Geom_Point](mp.GetGeometryN(0))
	cs0 := pt0.GetCoordinateSequence()
	if !cs0.HasZ() || !cs0.HasM() {
		t.Errorf("expected coordinate sequence to have Z and M")
	}
}

func TestWKTReaderMultiLineStringEmpty(t *testing.T) {
	reader := jts.Io_NewWKTReader()
	geom, err := reader.Read("MULTILINESTRING EMPTY")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	mls := java.Cast[*jts.Geom_MultiLineString](geom)
	if !mls.IsEmpty() {
		t.Errorf("expected empty multilinestring")
	}
	if mls.GetNumGeometries() != 0 {
		t.Errorf("expected 0 geometries, got %d", mls.GetNumGeometries())
	}
}

func TestWKTReaderMultiLineStringWithEmpty(t *testing.T) {
	reader := jts.Io_NewWKTReader()
	geom, err := reader.Read("MULTILINESTRING ((10 10, 20 20), EMPTY, (15 15, 30 15))")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	mls := java.Cast[*jts.Geom_MultiLineString](geom)
	if mls.GetNumGeometries() != 3 {
		t.Errorf("expected 3 geometries, got %d", mls.GetNumGeometries())
	}
	ls0 := java.Cast[*jts.Geom_LineString](mls.GetGeometryN(0))
	if ls0.GetNumPoints() != 2 {
		t.Errorf("expected 2 points in line 0, got %d", ls0.GetNumPoints())
	}
	ls1 := java.Cast[*jts.Geom_LineString](mls.GetGeometryN(1))
	if !ls1.IsEmpty() {
		t.Errorf("expected line 1 to be empty")
	}
	ls2 := java.Cast[*jts.Geom_LineString](mls.GetGeometryN(2))
	if ls2.GetNumPoints() != 2 {
		t.Errorf("expected 2 points in line 2, got %d", ls2.GetNumPoints())
	}
}

func TestWKTReaderMultiLineStringXYZ(t *testing.T) {
	reader := jts.Io_NewWKTReader()
	geom, err := reader.Read("MULTILINESTRING Z((10 10 10, 20 20 10), (15 15 10, 30 15 10))")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	mls := java.Cast[*jts.Geom_MultiLineString](geom)
	if mls.GetNumGeometries() != 2 {
		t.Errorf("expected 2 linestrings, got %d", mls.GetNumGeometries())
	}
	ls0 := java.Cast[*jts.Geom_LineString](mls.GetGeometryN(0))
	cs0 := ls0.GetCoordinateSequence()
	if cs0.GetZ(0) != 10 {
		t.Errorf("expected Z=10, got %v", cs0.GetZ(0))
	}
}

func TestWKTReaderMultiLineStringXYM(t *testing.T) {
	reader := jts.Io_NewWKTReader()
	geom, err := reader.Read("MULTILINESTRING M((10 10 11, 20 20 11), (15 15 11, 30 15 11))")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	mls := java.Cast[*jts.Geom_MultiLineString](geom)
	if mls.GetNumGeometries() != 2 {
		t.Errorf("expected 2 linestrings, got %d", mls.GetNumGeometries())
	}
	ls0 := java.Cast[*jts.Geom_LineString](mls.GetGeometryN(0))
	cs0 := ls0.GetCoordinateSequence()
	if !cs0.HasM() {
		t.Errorf("expected coordinate sequence to have M")
	}
}

func TestWKTReaderMultiLineStringXYZM(t *testing.T) {
	reader := jts.Io_NewWKTReader()
	geom, err := reader.Read("MULTILINESTRING ZM((10 10 10 11, 20 20 10 11), (15 15 10 11, 30 15 10 11))")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	mls := java.Cast[*jts.Geom_MultiLineString](geom)
	if mls.GetNumGeometries() != 2 {
		t.Errorf("expected 2 linestrings, got %d", mls.GetNumGeometries())
	}
	ls0 := java.Cast[*jts.Geom_LineString](mls.GetGeometryN(0))
	cs0 := ls0.GetCoordinateSequence()
	if !cs0.HasZ() || !cs0.HasM() {
		t.Errorf("expected coordinate sequence to have Z and M")
	}
}

func TestWKTReaderMultiPolygonEmpty(t *testing.T) {
	reader := jts.Io_NewWKTReader()
	geom, err := reader.Read("MULTIPOLYGON EMPTY")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	mpoly := java.Cast[*jts.Geom_MultiPolygon](geom)
	if !mpoly.IsEmpty() {
		t.Errorf("expected empty multipolygon")
	}
	if mpoly.GetNumGeometries() != 0 {
		t.Errorf("expected 0 geometries, got %d", mpoly.GetNumGeometries())
	}
}

func TestWKTReaderMultiPolygonWithEmpty(t *testing.T) {
	reader := jts.Io_NewWKTReader()
	geom, err := reader.Read("MULTIPOLYGON (((10 10, 10 20, 20 20, 20 15, 10 10), (11 11, 12 11, 12 12, 12 11, 11 11)), EMPTY, ((60 60, 70 70, 80 60, 60 60)))")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	mpoly := java.Cast[*jts.Geom_MultiPolygon](geom)
	if mpoly.GetNumGeometries() != 3 {
		t.Errorf("expected 3 geometries, got %d", mpoly.GetNumGeometries())
	}
	poly0 := java.Cast[*jts.Geom_Polygon](mpoly.GetGeometryN(0))
	if poly0.GetNumInteriorRing() != 1 {
		t.Errorf("expected 1 interior ring in poly 0, got %d", poly0.GetNumInteriorRing())
	}
	poly1 := java.Cast[*jts.Geom_Polygon](mpoly.GetGeometryN(1))
	if !poly1.IsEmpty() {
		t.Errorf("expected poly 1 to be empty")
	}
	poly2 := java.Cast[*jts.Geom_Polygon](mpoly.GetGeometryN(2))
	if poly2.GetExteriorRing().GetNumPoints() != 4 {
		t.Errorf("expected 4 points in poly 2 exterior ring, got %d", poly2.GetExteriorRing().GetNumPoints())
	}
}

func TestWKTReaderMultiPolygonXYZ(t *testing.T) {
	reader := jts.Io_NewWKTReader()
	geom, err := reader.Read("MULTIPOLYGON Z(((10 10 10, 10 20 10, 20 20 10, 20 15 10, 10 10 10), (11 11 10, 12 11 10, 12 12 10, 12 11 10, 11 11 10)), ((60 60 10, 70 70 10, 80 60 10, 60 60 10)))")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	mpoly := java.Cast[*jts.Geom_MultiPolygon](geom)
	if mpoly.GetNumGeometries() != 2 {
		t.Errorf("expected 2 polygons, got %d", mpoly.GetNumGeometries())
	}
	poly0 := java.Cast[*jts.Geom_Polygon](mpoly.GetGeometryN(0))
	cs0 := poly0.GetExteriorRing().GetCoordinateSequence()
	if cs0.GetZ(0) != 10 {
		t.Errorf("expected Z=10, got %v", cs0.GetZ(0))
	}
}

func TestWKTReaderMultiPolygonXYM(t *testing.T) {
	reader := jts.Io_NewWKTReader()
	geom, err := reader.Read("MULTIPOLYGON M(((10 10 11, 10 20 11, 20 20 11, 20 15 11, 10 10 11), (11 11 11, 12 11 11, 12 12 11, 12 11 11, 11 11 11)), ((60 60 11, 70 70 11, 80 60 11, 60 60 11)))")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	mpoly := java.Cast[*jts.Geom_MultiPolygon](geom)
	if mpoly.GetNumGeometries() != 2 {
		t.Errorf("expected 2 polygons, got %d", mpoly.GetNumGeometries())
	}
	poly0 := java.Cast[*jts.Geom_Polygon](mpoly.GetGeometryN(0))
	cs0 := poly0.GetExteriorRing().GetCoordinateSequence()
	if !cs0.HasM() {
		t.Errorf("expected coordinate sequence to have M")
	}
}

func TestWKTReaderMultiPolygonXYZM(t *testing.T) {
	reader := jts.Io_NewWKTReader()
	geom, err := reader.Read("MULTIPOLYGON ZM(((10 10 10 11, 10 20 10 11, 20 20 10 11, 20 15 10 11, 10 10 10 11), (11 11 10 11, 12 11 10 11, 12 12 10 11, 12 11 10 11, 11 11 10 11)), ((60 60 10 11, 70 70 10 11, 80 60 10 11, 60 60 10 11)))")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	mpoly := java.Cast[*jts.Geom_MultiPolygon](geom)
	if mpoly.GetNumGeometries() != 2 {
		t.Errorf("expected 2 polygons, got %d", mpoly.GetNumGeometries())
	}
	poly0 := java.Cast[*jts.Geom_Polygon](mpoly.GetGeometryN(0))
	cs0 := poly0.GetExteriorRing().GetCoordinateSequence()
	if !cs0.HasZ() || !cs0.HasM() {
		t.Errorf("expected coordinate sequence to have Z and M")
	}
}

func TestWKTReaderEmptyLineDimOldSyntax(t *testing.T) {
	reader := jts.Io_NewWKTReader()
	geom, err := reader.Read("LINESTRING EMPTY")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ls := java.Cast[*jts.Geom_LineString](geom)
	cs := ls.GetCoordinateSequence()
	// With old JTS syntax allowed (default), dimension should be 3.
	if cs.GetDimension() != 3 {
		t.Errorf("expected dimension 3 with old syntax, got %d", cs.GetDimension())
	}
}

func TestWKTReaderEmptyLineDim(t *testing.T) {
	reader := jts.Io_NewWKTReader()
	reader.SetIsOldJtsCoordinateSyntaxAllowed(false)
	geom, err := reader.Read("LINESTRING EMPTY")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ls := java.Cast[*jts.Geom_LineString](geom)
	cs := ls.GetCoordinateSequence()
	// With old JTS syntax disallowed, dimension should be 2.
	if cs.GetDimension() != 2 {
		t.Errorf("expected dimension 2, got %d", cs.GetDimension())
	}
}

func TestWKTReaderEmptyPolygonDim(t *testing.T) {
	reader := jts.Io_NewWKTReader()
	reader.SetIsOldJtsCoordinateSyntaxAllowed(false)
	geom, err := reader.Read("POLYGON EMPTY")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	poly := java.Cast[*jts.Geom_Polygon](geom)
	cs := poly.GetExteriorRing().GetCoordinateSequence()
	// With old JTS syntax disallowed, dimension should be 2.
	if cs.GetDimension() != 2 {
		t.Errorf("expected dimension 2, got %d", cs.GetDimension())
	}
}

func TestWKTReaderNaN(t *testing.T) {
	reader := jts.Io_NewWKTReader()
	// Test NaN in uppercase.
	geom, err := reader.Read("POINT (10 10 NaN)")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	pt := java.Cast[*jts.Geom_Point](geom)
	cs := pt.GetCoordinateSequence()
	if !isNaN(cs.GetZ(0)) {
		t.Errorf("expected Z=NaN, got %v", cs.GetZ(0))
	}

	// Test NaN in lowercase.
	geom, err = reader.Read("POINT (10 10 nan)")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	pt = java.Cast[*jts.Geom_Point](geom)
	cs = pt.GetCoordinateSequence()
	if !isNaN(cs.GetZ(0)) {
		t.Errorf("expected Z=NaN, got %v", cs.GetZ(0))
	}

	// Test NaN in mixed case.
	geom, err = reader.Read("POINT (10 10 NAN)")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	pt = java.Cast[*jts.Geom_Point](geom)
	cs = pt.GetCoordinateSequence()
	if !isNaN(cs.GetZ(0)) {
		t.Errorf("expected Z=NaN, got %v", cs.GetZ(0))
	}
}

func TestWKTReaderLargeNumbers(t *testing.T) {
	precisionModel := jts.Geom_NewPrecisionModelWithScale(1e9)
	factory := jts.Geom_NewGeometryFactoryWithPrecisionModel(precisionModel)
	reader := jts.Io_NewWKTReaderWithFactory(factory)
	geom, err := reader.Read("POINT (123456789.01234567890 10)")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	pt := java.Cast[*jts.Geom_Point](geom)
	cs := pt.GetCoordinateSequence()

	// Create expected point with same factory.
	coord := jts.Geom_NewCoordinateWithXY(123456789.01234567890, 10)
	expectedPt := factory.CreatePointFromCoordinate(coord)
	expectedCS := expectedPt.GetCoordinateSequence()

	// Compare with tolerance.
	tolerance := 1e-7
	xDiff := cs.GetX(0) - expectedCS.GetX(0)
	yDiff := cs.GetY(0) - expectedCS.GetY(0)
	if xDiff < -tolerance || xDiff > tolerance {
		t.Errorf("X coordinate mismatch: got %v, expected %v", cs.GetX(0), expectedCS.GetX(0))
	}
	if yDiff < -tolerance || yDiff > tolerance {
		t.Errorf("Y coordinate mismatch: got %v, expected %v", cs.GetY(0), expectedCS.GetY(0))
	}
}

func TestWKTReaderTurkishLocale(t *testing.T) {
	// Go's strconv.ParseFloat is locale-independent, so this test
	// verifies that WKT parsing works correctly regardless of locale.
	// The Java test sets locale to Turkish to verify "i" handling.
	// In Go, we just verify lowercase keywords work.
	reader := jts.Io_NewWKTReader()
	geom, err := reader.Read("point (10 20)")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	pt := java.Cast[*jts.Geom_Point](geom)
	tolerance := 1e-7
	if pt.GetX()-10.0 < -tolerance || pt.GetX()-10.0 > tolerance {
		t.Errorf("expected X=10, got %v", pt.GetX())
	}
	if pt.GetY()-20.0 < -tolerance || pt.GetY()-20.0 > tolerance {
		t.Errorf("expected Y=20, got %v", pt.GetY())
	}
}

func isNaN(f float64) bool {
	return f != f
}
