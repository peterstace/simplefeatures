package jts

import "testing"

func TestSnapRoundingNoderSimple(t *testing.T) {
	wkt := "MULTILINESTRING ((1 1, 9 2), (3 3, 3 0))"
	expected := "MULTILINESTRING ((1 1, 3 1), (3 1, 9 2), (3 3, 3 1), (3 1, 3 0))"
	checkSnapRounding(t, wkt, 1, expected)
}

func TestSnapRoundingNoderSnappedDiagonalLine(t *testing.T) {
	// A diagonal line is snapped to a vertex half a grid cell away.
	wkt := "LINESTRING (2 3, 3 3, 3 2, 2 3)"
	expected := "MULTILINESTRING ((2 3, 3 3), (2 3, 3 3), (3 2, 3 3), (3 2, 3 3))"
	checkSnapRounding(t, wkt, 1.0, expected)
}

func TestSnapRoundingNoderRingsWithParallelNarrowSpikes(t *testing.T) {
	// Rings with parallel narrow spikes are snapped to a simple ring and lines.
	wkt := "MULTILINESTRING ((1 3.3, 1.3 1.4, 3.1 1.4, 3.1 0.9, 1.3 0.9, 1 -0.2, 0.8 1.3, 1 3.3), (1 2.9, 2.9 2.9, 2.9 1.3, 1.7 1, 1.3 0.9, 1 0.4, 1 2.9))"
	expected := "MULTILINESTRING ((1 3, 1 1), (1 1, 2 1), (2 1, 3 1), (3 1, 2 1), (2 1, 1 1), (1 1, 1 0), (1 0, 1 1), (1 1, 1 3), (1 3, 3 3, 3 1), (3 1, 2 1), (2 1, 1 1), (1 1, 1 0), (1 0, 1 1), (1 1, 1 3))"
	checkSnapRounding(t, wkt, 1.0, expected)
}

func TestSnapRoundingNoderHorizontalLinesWithMiddleNode(t *testing.T) {
	// This test checks the HotPixel test for overlapping horizontal line.
	wkt := "MULTILINESTRING ((2.5117493 49.0278625,                      2.5144958 49.0278625), (2.511749 49.027863, 2.513123 49.027863, 2.514496 49.027863))"
	expected := "MULTILINESTRING ((2.511749 49.027863, 2.513123 49.027863), (2.511749 49.027863, 2.513123 49.027863), (2.513123 49.027863, 2.514496 49.027863), (2.513123 49.027863, 2.514496 49.027863))"
	checkSnapRounding(t, wkt, 1_000_000.0, expected)
}

func TestSnapRoundingNoderSlantAndHorizontalLineWithMiddleNode(t *testing.T) {
	wkt := "MULTILINESTRING ((0.1565552 49.5277405, 0.1579285 49.5277405, 0.1593018 49.5277405), (0.1568985 49.5280838, 0.1589584 49.5273972))"
	expected := "MULTILINESTRING ((0.156555 49.527741, 0.157928 49.527741), (0.156899 49.528084, 0.157928 49.527741), (0.157928 49.527741, 0.157929 49.527741, 0.159302 49.527741), (0.157928 49.527741, 0.158958 49.527397))"
	checkSnapRounding(t, wkt, 1_000_000.0, expected)
}

func TestSnapRoundingNoderNearbyCorner(t *testing.T) {
	wkt := "MULTILINESTRING ((0.2 1.1, 1.6 1.4, 1.9 2.9), (0.9 0.9, 2.3 1.7))"
	expected := "MULTILINESTRING ((0 1, 1 1), (1 1, 2 1), (1 1, 2 1), (2 1, 2 2), (2 1, 2 2), (2 2, 2 3))"
	checkSnapRounding(t, wkt, 1.0, expected)
}

func TestSnapRoundingNoderNearbyShape(t *testing.T) {
	wkt := "MULTILINESTRING ((1.3 0.1, 2.4 3.9), (0 1, 1.53 1.48, 0 4))"
	expected := "MULTILINESTRING ((1 0, 2 1), (2 1, 2 4), (0 1, 2 1), (2 1, 0 4))"
	checkSnapRounding(t, wkt, 1.0, expected)
}

func TestSnapRoundingNoderIntOnGridCorner(t *testing.T) {
	// Fixed by ensuring intersections are forced into segments.
	wkt := "MULTILINESTRING ((4.30166242 45.53438188, 4.30166243 45.53438187), (4.3011475 45.5328371, 4.3018341 45.5348969))"
	checkSnapRounding(t, wkt, 100000000, "")
}

func TestSnapRoundingNoderVertexCrossesLine(t *testing.T) {
	wkt := "MULTILINESTRING ((2.2164917 48.8864136, 2.2175217 48.8867569), (2.2175217 48.8867569, 2.2182083 48.8874435), (2.2182083 48.8874435, 2.2161484 48.8853836))"
	checkSnapRounding(t, wkt, 1000000, "")
}

func TestSnapRoundingNoderVertexCrossesLine2(t *testing.T) {
	// Fixed by NOT rounding lines extracted by Overlay.
	wkt := "MULTILINESTRING ((2.276916574988164 49.06082147500638, 2.2769165 49.0608215), (2.2769165 49.0608215, 2.2755432 49.0608215), (2.2762299 49.0615082, 2.276916574988164 49.06082147500638))"
	checkSnapRounding(t, wkt, 1000000, "")
}

func TestSnapRoundingNoderShortLineNodeNotAdded(t *testing.T) {
	// Looks like a very short line is stretched between two grid points.
	wkt := "LINESTRING (2.1279144 48.8445282, 2.126884443750796 48.84555818124935, 2.1268845 48.8455582, 2.1268845 48.8462448)"
	expected := "MULTILINESTRING ((2.127914 48.844528, 2.126885 48.845558), (2.126885 48.845558, 2.126884 48.845558), (2.126884 48.845558, 2.126885 48.845558), (2.126885 48.845558, 2.126885 48.846245))"
	checkSnapRounding(t, wkt, 1000000, expected)
}

func TestSnapRoundingNoderDiagonalNotNodedRightUp(t *testing.T) {
	// This test will fail if the diagonals of hot pixels are not checked.
	wkt := "MULTILINESTRING ((0 0, 10 10), ( 0 2, 4.55 5.4, 9 10 ))"
	checkSnapRounding(t, wkt, 1, "")
}

func TestSnapRoundingNoderDiagonalNotNodedLeftUp(t *testing.T) {
	// Same diagonal test but flipped to test other diagonal.
	wkt := "MULTILINESTRING ((10 0, 0 10), ( 10 2, 5.45 5.45, 1 10 ))"
	checkSnapRounding(t, wkt, 1, "")
}

func TestSnapRoundingNoderDiagonalNotNodedOriginal(t *testing.T) {
	// Original full-precision diagonal line case.
	wkt := "MULTILINESTRING (( 2.45167 48.96709, 2.45768 48.9731 ), (2.4526978 48.968811, 2.4537277 48.9691544, 2.4578476 48.9732742))"
	checkSnapRounding(t, wkt, 100000, "")
}

func TestSnapRoundingNoderLoopBackCreatesNode(t *testing.T) {
	wkt := "LINESTRING (2 2, 5 2, 8 4, 5 6, 4.8 2.3, 2 5)"
	expected := "MULTILINESTRING ((2 2, 5 2), (5 2, 8 4, 5 6, 5 2), (5 2, 2 5))"
	checkSnapRounding(t, wkt, 1, expected)
}

func TestSnapRoundingNoderNearVertexNotNoded(t *testing.T) {
	// An A vertex lies very close to a B segment.
	// Fixed by adding intersection detection for near vertices to segments.
	wkt := "MULTILINESTRING ((2.4829102 48.8726807, 2.4830818249999997 48.873195575, 2.4839401 48.8723373), ( 2.4829102 48.8726807, 2.4832535 48.8737106 ))"
	checkSnapRounding(t, wkt, 100000000, "")
}

func TestSnapRoundingNoderVertexNearHorizSegNotNoded(t *testing.T) {
	// A vertex lies near interior of horizontal segment.
	wkt := "MULTILINESTRING (( 2.5096893 48.9530182, 2.50762932500455 48.95233152500091, 2.5055695 48.9530182 ), ( 2.5090027 48.9523315, 2.5035095 48.9523315 ))"
	checkSnapRounding(t, wkt, 1000000, "")
}

func TestSnapRoundingNoderMCIndexNoderTolerance(t *testing.T) {
	// Tests that MCIndexNoder tolerance is set correctly.
	wkt := "LINESTRING (3670939.6336634574 3396937.3777869204, 3670995.4715200397 3396926.0316904164, 3671077.280213823 3396905.4302639295, 3671203.8838707027 3396908.120176068, 3671334.962571111 3396904.8310892633, 3670037.299066126 3396904.8310892633, 3670037.299066126 3398075.9808747065, 3670939.6336634574 3396937.3777869204)"
	expected := "MULTILINESTRING ((3670776.0631373483 3397212.0584320477, 3670776.0631373483 3396600.058421521), (3670776.0631373483 3396600.058421521, 3671388.063147875 3396600.058421521), (3671388.063147875 3396600.058421521, 3671388.063147875 3397212.0584320477), (3671388.063147875 3397212.0584320477, 3671388.063147875 3396600.058421521), (3671388.063147875 3396600.058421521, 3671388.063147875 3397212.0584320477), (3671388.063147875 3397212.0584320477, 3671388.063147875 3396600.058421521), (3671388.063147875 3396600.058421521, 3670776.0631373483 3396600.058421521), (3670776.0631373483 3396600.058421521, 3670164.063126822 3396600.058421521, 3670164.063126822 3397824.058442574, 3670776.0631373483 3397212.0584320477))"
	checkSnapRounding(t, wkt, 0.0016339869, expected)
}

func checkSnapRounding(t *testing.T, wkt string, scale float64, expectedWKT string) {
	t.Helper()
	geom := readWKT(t, wkt)
	pm := Geom_NewPrecisionModelWithScale(scale)
	noder := NodingSnapround_NewSnapRoundingNoder(pm)
	result := Noding_TestUtil_NodeValidated(geom, nil, noder)

	// Only check if expected was provided.
	if expectedWKT == "" {
		return
	}
	expected := readWKT(t, expectedWKT)
	checkEqualGeom(t, expected, result)
}

func readWKT(t *testing.T, wkt string) *Geom_Geometry {
	t.Helper()
	reader := Io_NewWKTReader()
	geom, err := reader.Read(wkt)
	if err != nil {
		t.Fatalf("failed to parse WKT: %v", err)
	}
	return geom
}

func checkEqualGeom(t *testing.T, expected, actual *Geom_Geometry) {
	t.Helper()
	// Normalize geometries before comparison.
	expected.Normalize()
	actual.Normalize()
	if !expected.EqualsExactWithTolerance(actual, 0.0) {
		t.Errorf("geometries not equal:\nexpected: %v\nactual:   %v",
			expected.ToText(), actual.ToText())
	}
}
