package jts

import (
	"encoding/hex"
	"math"
	"testing"
)

const elevationModelTestTolerance = 0.00001

func TestElevationModelBox(t *testing.T) {
	checkElevation(t, "POLYGON Z ((1 6 50, 9 6 60, 9 4 50, 1 4 40, 1 6 50))", "",
		0, 10, 50, 5, 10, 50, 10, 10, 60,
		0, 5, 50, 5, 5, 50, 10, 5, 50,
		0, 4, 40, 5, 4, 50, 10, 4, 50,
		0, 0, 40, 5, 0, 50, 10, 0, 50,
	)
}

func TestElevationModelLine(t *testing.T) {
	checkElevation(t, "LINESTRING Z (0 0 0, 10 10 10)", "",
		-1, 11, 5, 11, 11, 10,
		0, 10, 5, 5, 10, 5, 10, 10, 10,
		0, 5, 5, 5, 5, 5, 10, 5, 5,
		0, 0, 0, 5, 0, 5, 10, 0, 5,
		-1, -1, 0, 5, -1, 5, 11, -1, 5,
	)
}

func TestElevationModelPopulateZLine(t *testing.T) {
	checkElevationPopulateZ(t, "LINESTRING Z (0 0 0, 10 10 10)",
		"LINESTRING (1 1, 9 9)",
		"LINESTRING (1 1 0, 9 9 10)",
	)
}

func TestElevationModelPopulateZBox(t *testing.T) {
	checkElevationPopulateZ(t, "LINESTRING Z (0 0 0, 10 10 10)",
		"POLYGON ((1 9, 9 9, 9 1, 1 1, 1 9))",
		"POLYGON Z ((1 1 0, 1 9 5, 9 9 10, 9 1 5, 1 1 0))",
	)
}

func TestElevationModelMultiLine(t *testing.T) {
	checkElevation(t, "MULTILINESTRING Z ((0 0 0, 10 10 8), (1 2 2, 9 8 6))", "",
		-1, 11, 4, 11, 11, 7,
		0, 10, 4, 5, 10, 4, 10, 10, 7,
		0, 5, 4, 5, 5, 4, 10, 5, 4,
		0, 0, 1, 5, 0, 4, 10, 0, 4,
		-1, -1, 1, 5, -1, 4, 11, -1, 4,
	)
}

func TestElevationModelTwoLines(t *testing.T) {
	checkElevation(t, "LINESTRING Z (0 0 0, 10 10 8)", "LINESTRING Z (1 2 2, 9 8 6)",
		-1, 11, 4, 11, 11, 7,
		0, 10, 4, 5, 10, 4, 10, 10, 7,
		0, 5, 4, 5, 5, 4, 10, 5, 4,
		0, 0, 1, 5, 0, 4, 10, 0, 4,
		-1, -1, 1, 5, -1, 4, 11, -1, 4,
	)
}

func TestElevationModelLineHorizontal(t *testing.T) {
	checkElevation(t, "LINESTRING Z (0 5 0, 10 5 10)", "",
		0, 10, 0, 5, 10, 5, 10, 10, 10,
		0, 5, 0, 5, 5, 5, 10, 5, 10,
		0, 0, 0, 5, 0, 5, 10, 0, 10,
	)
}

func TestElevationModelLineVertical(t *testing.T) {
	checkElevation(t, "LINESTRING Z (5 0 0, 5 10 10)", "",
		0, 10, 10, 5, 10, 10, 10, 10, 10,
		0, 5, 5, 5, 5, 5, 10, 5, 5,
		0, 0, 0, 5, 0, 0, 10, 0, 0,
	)
}

func TestElevationModelPoint(t *testing.T) {
	checkElevation(t, "POINT Z (5 5 5)", "",
		0, 9, 5, 5, 9, 5, 9, 9, 5,
		0, 5, 5, 5, 5, 5, 9, 5, 5,
		0, 0, 5, 5, 0, 5, 9, 0, 5,
	)
}

func TestElevationModelMultiPointSame(t *testing.T) {
	checkElevation(t, "MULTIPOINT Z ((5 5 5), (5 5 9))", "",
		0, 9, 7, 5, 9, 7, 9, 9, 7,
		0, 5, 7, 5, 5, 7, 9, 5, 7,
		0, 0, 7, 5, 0, 7, 9, 0, 7,
	)
}

func TestElevationModelLine2D(t *testing.T) {
	// Tests that XY geometries are scanned correctly (avoiding reading Z)
	// and that they produce a model Z value of NaN.
	// LINESTRING (0 0, 10 10)
	wkbHex := "0102000000020000000000000000000000000000000000000000000000000024400000000000002440"
	wkbBytes, err := hex.DecodeString(wkbHex)
	if err != nil {
		t.Fatalf("failed to decode hex: %v", err)
	}
	wkbReader := Io_NewWKBReader()
	geom, err := wkbReader.ReadBytes(wkbBytes)
	if err != nil {
		t.Fatalf("failed to read WKB: %v", err)
	}
	model := OperationOverlayng_ElevationModel_Create(geom, nil)
	z := model.GetZ(5, 5)
	if !math.IsNaN(z) {
		t.Errorf("expected NaN for 2D geometry, got %v", z)
	}
}

func checkElevation(t *testing.T, wkt1, wkt2 string, ords ...float64) {
	t.Helper()
	reader := Io_NewWKTReader()
	geom1, err := reader.Read(wkt1)
	if err != nil {
		t.Fatalf("failed to read wkt1: %v", err)
	}
	var geom2 *Geom_Geometry
	if wkt2 != "" {
		geom2, err = reader.Read(wkt2)
		if err != nil {
			t.Fatalf("failed to read wkt2: %v", err)
		}
	}

	model := OperationOverlayng_ElevationModel_Create(geom1, geom2)
	numPts := len(ords) / 3
	if 3*numPts != len(ords) {
		t.Fatalf("Incorrect number of ordinates")
	}
	for i := 0; i < numPts; i++ {
		x := ords[3*i]
		y := ords[3*i+1]
		expectedZ := ords[3*i+2]
		actualZ := model.GetZ(x, y)
		if math.IsNaN(expectedZ) && math.IsNaN(actualZ) {
			continue
		}
		if math.Abs(actualZ-expectedZ) > elevationModelTestTolerance {
			t.Errorf("Point (%v, %v): expected Z=%v, got Z=%v", x, y, expectedZ, actualZ)
		}
	}
}

func checkElevationPopulateZ(t *testing.T, wkt, wktNoZ, wktZExpected string) {
	t.Helper()
	reader := Io_NewWKTReader()
	geom, err := reader.Read(wkt)
	if err != nil {
		t.Fatalf("failed to read wkt: %v", err)
	}
	model := OperationOverlayng_ElevationModel_Create(geom, nil)

	geomNoZ, err := reader.Read(wktNoZ)
	if err != nil {
		t.Fatalf("failed to read wktNoZ: %v", err)
	}
	model.PopulateZ(geomNoZ)

	geomZExpected, err := reader.Read(wktZExpected)
	if err != nil {
		t.Fatalf("failed to read wktZExpected: %v", err)
	}
	checkEqualXYZ(t, geomZExpected, geomNoZ)
}

func checkEqualXYZ(t *testing.T, expected, actual *Geom_Geometry) {
	t.Helper()
	expectedNorm := expected.Norm()
	actualNorm := actual.Norm()
	expectedCoords := expectedNorm.GetCoordinates()
	actualCoords := actualNorm.GetCoordinates()
	if len(expectedCoords) != len(actualCoords) {
		t.Errorf("coordinate count mismatch: expected %d, got %d", len(expectedCoords), len(actualCoords))
		return
	}
	for i := range expectedCoords {
		ex, ey, ez := expectedCoords[i].GetX(), expectedCoords[i].GetY(), expectedCoords[i].GetZ()
		ax, ay, az := actualCoords[i].GetX(), actualCoords[i].GetY(), actualCoords[i].GetZ()
		if math.Abs(ex-ax) > elevationModelTestTolerance ||
			math.Abs(ey-ay) > elevationModelTestTolerance ||
			math.Abs(ez-az) > elevationModelTestTolerance {
			t.Errorf("coordinate %d mismatch: expected (%v, %v, %v), got (%v, %v, %v)",
				i, ex, ey, ez, ax, ay, az)
		}
	}
}
