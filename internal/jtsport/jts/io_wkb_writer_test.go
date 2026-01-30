package jts_test

import (
	"testing"

	"github.com/peterstace/simplefeatures/internal/jtsport/java"
	"github.com/peterstace/simplefeatures/internal/jtsport/jts"
)

func TestWKBWriterSRID(t *testing.T) {
	gf := jts.Geom_NewGeometryFactoryDefault()
	p1 := gf.CreatePointFromCoordinate(jts.Geom_NewCoordinateWithXY(1, 2))
	p1.SetSRID(1234)

	// First write out without srid set.
	w := jts.Io_NewWKBWriter()
	wkb := w.Write(p1.Geom_Geometry)

	// Check the 3rd bit of the second byte, should be unset.
	b := wkb[1] & 0x20
	if b != 0 {
		t.Errorf("expected SRID bit unset, got %02X", b)
	}

	// Read geometry back in.
	r := jts.Io_NewWKBReaderWithFactory(gf)
	p2, err := r.ReadBytes(wkb)
	if err != nil {
		t.Fatalf("reading WKB: %v", err)
	}

	if !p1.EqualsExact(p2) {
		t.Errorf("geometries not equal")
	}
	if p2.GetSRID() != 0 {
		t.Errorf("expected SRID 0, got %d", p2.GetSRID())
	}

	// Now write out with srid set.
	w = jts.Io_NewWKBWriterWithDimensionAndSRID(2, true)
	wkb = w.Write(p1.Geom_Geometry)

	// Check the 3rd bit of the second byte, should be set.
	b = wkb[1] & 0x20
	if b != 0x20 {
		t.Errorf("expected SRID bit set, got %02X", b)
	}

	srid := (int(wkb[5]&0xff) << 24) | (int(wkb[6]&0xff) << 16) |
		(int(wkb[7]&0xff) << 8) | int(wkb[8]&0xff)
	if srid != 1234 {
		t.Errorf("expected SRID 1234, got %d", srid)
	}

	r = jts.Io_NewWKBReaderWithFactory(gf)
	p2, err = r.ReadBytes(wkb)
	if err != nil {
		t.Fatalf("reading WKB: %v", err)
	}

	// Read the geometry back in.
	if !p1.EqualsExact(p2) {
		t.Errorf("geometries not equal")
	}
	if p2.GetSRID() != 1234 {
		t.Errorf("expected SRID 1234, got %d", p2.GetSRID())
	}
}

func TestWKBWriterPointEmpty2D(t *testing.T) {
	checkWKBWriterOutput(t, "POINT EMPTY", 2, jts.Io_ByteOrderValues_LITTLE_ENDIAN, -1, "0101000000000000000000F87F000000000000F87F")
}

func TestWKBWriterPointEmpty3D(t *testing.T) {
	checkWKBWriterOutput(t, "POINT EMPTY", 3, jts.Io_ByteOrderValues_LITTLE_ENDIAN, -1, "0101000080000000000000F87F000000000000F87F000000000000F87F")
}

func TestWKBWriterPolygonEmpty2DSRID(t *testing.T) {
	checkWKBWriterOutput(t, "POLYGON EMPTY", 2, jts.Io_ByteOrderValues_LITTLE_ENDIAN, 4326, "0103000020E610000000000000")
}

func TestWKBWriterPolygonEmpty2D(t *testing.T) {
	checkWKBWriterOutput(t, "POLYGON EMPTY", 2, jts.Io_ByteOrderValues_LITTLE_ENDIAN, -1, "010300000000000000")
}

func TestWKBWriterPolygonEmpty3D(t *testing.T) {
	checkWKBWriterOutput(t, "POLYGON EMPTY", 3, jts.Io_ByteOrderValues_LITTLE_ENDIAN, -1, "010300008000000000")
}

func TestWKBWriterMultiPolygonEmpty2D(t *testing.T) {
	checkWKBWriterOutput(t, "MULTIPOLYGON EMPTY", 2, jts.Io_ByteOrderValues_LITTLE_ENDIAN, -1, "010600000000000000")
}

func TestWKBWriterMultiPolygonEmpty3D(t *testing.T) {
	checkWKBWriterOutput(t, "MULTIPOLYGON EMPTY", 3, jts.Io_ByteOrderValues_LITTLE_ENDIAN, -1, "010600008000000000")
}

func TestWKBWriterMultiPolygonEmpty2DSRID(t *testing.T) {
	checkWKBWriterOutput(t, "MULTIPOLYGON EMPTY", 2, jts.Io_ByteOrderValues_LITTLE_ENDIAN, 4326, "0106000020E610000000000000")
}

func TestWKBWriterMultiPolygon(t *testing.T) {
	checkWKBWriterOutput(t,
		"MULTIPOLYGON(((0 0,0 10,10 10,10 0,0 0),(1 1,1 9,9 9,9 1,1 1)),((-9 0,-9 10,-1 10,-1 0,-9 0)))",
		2,
		jts.Io_ByteOrderValues_LITTLE_ENDIAN,
		4326,
		"0106000020E61000000200000001030000000200000005000000000000000000000000000000000000000000000000000000000000000000244000000000000024400000000000002440000000000000244000000000000000000000000000000000000000000000000005000000000000000000F03F000000000000F03F000000000000F03F0000000000002240000000000000224000000000000022400000000000002240000000000000F03F000000000000F03F000000000000F03F0103000000010000000500000000000000000022C0000000000000000000000000000022C00000000000002440000000000000F0BF0000000000002440000000000000F0BF000000000000000000000000000022C00000000000000000")
}

func TestWKBWriterGeometryCollection(t *testing.T) {
	checkWKBWriterOutput(t,
		"GEOMETRYCOLLECTION(POINT(0 1),POINT(0 1),POINT(2 3),LINESTRING(2 3,4 5),LINESTRING(0 1,2 3),LINESTRING(4 5,6 7),POLYGON((0 0,0 10,10 10,10 0,0 0),(1 1,1 9,9 9,9 1,1 1)),POLYGON((0 0,0 10,10 10,10 0,0 0),(1 1,1 9,9 9,9 1,1 1)),POLYGON((-9 0,-9 10,-1 10,-1 0,-9 0)))",
		2,
		jts.Io_ByteOrderValues_LITTLE_ENDIAN,
		4326,
		"0107000020E61000000900000001010000000000000000000000000000000000F03F01010000000000000000000000000000000000F03F01010000000000000000000040000000000000084001020000000200000000000000000000400000000000000840000000000000104000000000000014400102000000020000000000000000000000000000000000F03F000000000000004000000000000008400102000000020000000000000000001040000000000000144000000000000018400000000000001C4001030000000200000005000000000000000000000000000000000000000000000000000000000000000000244000000000000024400000000000002440000000000000244000000000000000000000000000000000000000000000000005000000000000000000F03F000000000000F03F000000000000F03F0000000000002240000000000000224000000000000022400000000000002240000000000000F03F000000000000F03F000000000000F03F01030000000200000005000000000000000000000000000000000000000000000000000000000000000000244000000000000024400000000000002440000000000000244000000000000000000000000000000000000000000000000005000000000000000000F03F000000000000F03F000000000000F03F0000000000002240000000000000224000000000000022400000000000002240000000000000F03F000000000000F03F000000000000F03F0103000000010000000500000000000000000022C0000000000000000000000000000022C00000000000002440000000000000F0BF0000000000002440000000000000F0BF000000000000000000000000000022C00000000000000000")
}

func TestWKBWriterLineStringZM(t *testing.T) {
	gf := jts.Geom_NewGeometryFactoryDefault()
	coords := []*jts.Geom_Coordinate{
		jts.Geom_NewCoordinateXYZM4DWithXYZM(1, 2, 3, 4).Geom_Coordinate,
		jts.Geom_NewCoordinateXYZM4DWithXYZM(5, 6, 7, 8).Geom_Coordinate,
	}
	lineZM := gf.CreateLineStringFromCoordinates(coords)

	wkbWriter := jts.Io_NewWKBWriterWithDimension(4)
	write := wkbWriter.Write(lineZM.Geom_Geometry)

	wkbReader := jts.Io_NewWKBReader()
	deserialized, err := wkbReader.ReadBytes(write)
	if err != nil {
		t.Fatalf("reading WKB: %v", err)
	}

	deserializedLS := java.Cast[*jts.Geom_LineString](deserialized)

	if !lineZM.EqualsGeometry(deserialized) {
		t.Errorf("geometries not equal")
	}

	coord0 := deserializedLS.GetPointN(0).GetCoordinate()
	if coord0.GetX() != 1.0 {
		t.Errorf("expected X=1.0, got %v", coord0.GetX())
	}
	if coord0.GetY() != 2.0 {
		t.Errorf("expected Y=2.0, got %v", coord0.GetY())
	}
	if coord0.GetZ() != 3.0 {
		t.Errorf("expected Z=3.0, got %v", coord0.GetZ())
	}
	if coord0.GetM() != 4.0 {
		t.Errorf("expected M=4.0, got %v", coord0.GetM())
	}

	coord1 := deserializedLS.GetPointN(1).GetCoordinate()
	if coord1.GetX() != 5.0 {
		t.Errorf("expected X=5.0, got %v", coord1.GetX())
	}
	if coord1.GetY() != 6.0 {
		t.Errorf("expected Y=6.0, got %v", coord1.GetY())
	}
	if coord1.GetZ() != 7.0 {
		t.Errorf("expected Z=7.0, got %v", coord1.GetZ())
	}
	if coord1.GetM() != 8.0 {
		t.Errorf("expected M=8.0, got %v", coord1.GetM())
	}
}

func checkWKBWriterOutput(t *testing.T, wkt string, dimension, byteOrder, srid int, expectedWKBHex string) {
	t.Helper()
	rdr := jts.Io_NewWKTReader()
	geom, err := rdr.Read(wkt)
	if err != nil {
		t.Fatalf("parsing WKT: %v", err)
	}

	// Set SRID if not -1.
	includeSRID := false
	if srid >= 0 {
		includeSRID = true
		geom.SetSRID(srid)
	}

	wkbWriter := jts.Io_NewWKBWriterWithDimensionOrderAndSRID(dimension, byteOrder, includeSRID)
	wkb := wkbWriter.Write(geom)
	wkbHex := jts.Io_WKBWriter_ToHex(wkb)

	if wkbHex != expectedWKBHex {
		t.Errorf("WKB hex mismatch\nexpected: %s\ngot:      %s", expectedWKBHex, wkbHex)
	}
}
