package jts

import "testing"

func TestSnappingNoderOverlappingLinesWithNearVertex(t *testing.T) {
	wkt1 := "LINESTRING (100 100, 300 100)"
	wkt2 := "LINESTRING (200 100.1, 400 100)"
	expected := "MULTILINESTRING ((100 100, 200 100.1), (200 100.1, 300 100), (200 100.1, 300 100), (300 100, 400 100))"
	checkSnappingNoder(t, wkt1, wkt2, 1, expected)
}

func TestSnappingNoderSnappedVertex(t *testing.T) {
	wkt1 := "LINESTRING (100 100, 200 100, 300 100)"
	wkt2 := "LINESTRING (200 100.3, 400 110)"
	expected := "MULTILINESTRING ((100 100, 200 100), (200 100, 300 100), (200 100, 400 110))"
	checkSnappingNoder(t, wkt1, wkt2, 1, expected)
}

func TestSnappingNoderSelfSnap(t *testing.T) {
	wkt1 := "LINESTRING (100 200, 100 100, 300 100, 200 99.3, 200 0)"
	expected := "MULTILINESTRING ((100 200, 100 100, 200 99.3), (200 99.3, 300 100), (300 100, 200 99.3), (200 99.3, 200 0))"
	checkSnappingNoder(t, wkt1, "", 1, expected)
}

func TestSnappingNoderLineCondensePoints(t *testing.T) {
	wkt1 := "LINESTRING (1 1, 1.3 1, 1.6 1, 1.9 1, 2.2 1, 2.5 1, 2.8 1, 3.1 1, 3.5 1, 4 1)"
	expected := "LINESTRING (1 1, 2.2 1, 3.5 1)"
	checkSnappingNoder(t, wkt1, "", 1, expected)
}

func TestSnappingNoderLineDensePointsSelfSnap(t *testing.T) {
	wkt1 := "LINESTRING (1 1, 1.3 1, 1.6 1, 1.9 1, 2.2 1, 2.5 1, 2.8 1, 3.1 1, 3.5 1, 4.8 1, 3.8 3.1, 2.5 1.1, 0.5 3.1)"
	expected := "MULTILINESTRING ((1 1, 2.2 1), (2.2 1, 3.5 1, 4.8 1, 3.8 3.1, 2.2 1), (2.2 1, 1 1), (1 1, 0.5 3.1))"
	checkSnappingNoder(t, wkt1, "", 1, expected)
}

func TestSnappingNoderAlmostCoincidentEdge(t *testing.T) {
	// Two rings with edges which are almost coincident. Edges are snapped to
	// produce the same segment.
	wkt1 := "MULTILINESTRING ((698400.5682737827 2388494.3828697307, 698402.3209180075 2388497.0819257903, 698415.3598714538 2388498.764371397, 698413.5003455497 2388495.90071853, 698400.5682737827 2388494.3828697307), (698231.847335025 2388474.57994264, 698440.416211779 2388499.05985776, 698432.582638943 2388300.28294705, 698386.666515791 2388303.40346027, 698328.29462841 2388312.88889197, 698231.847335025 2388474.57994264))"
	expected := "MULTILINESTRING ((698231.847335025 2388474.57994264, 698328.29462841 2388312.88889197, 698386.666515791 2388303.40346027, 698432.582638943 2388300.28294705, 698440.416211779 2388499.05985776, 698413.5003455497 2388495.90071853), (698231.847335025 2388474.57994264, 698400.5682737827 2388494.3828697307), (698400.5682737827 2388494.3828697307, 698402.3209180075 2388497.0819257903, 698415.3598714538 2388498.764371397, 698413.5003455497 2388495.90071853), (698400.5682737827 2388494.3828697307, 698413.5003455497 2388495.90071853), (698400.5682737827 2388494.3828697307, 698413.5003455497 2388495.90071853))"
	checkSnappingNoder(t, wkt1, "", 1, expected)
}

func TestSnappingNoderAlmostCoincidentLines(t *testing.T) {
	// Extract from previous test.
	wkt1 := "MULTILINESTRING ((698413.5003455497 2388495.90071853, 698400.5682737827 2388494.3828697307), (698231.847335025 2388474.57994264, 698440.416211779 2388499.05985776))"
	expected := "MULTILINESTRING ((698231.847335025 2388474.57994264, 698400.5682737827 2388494.3828697307), (698400.5682737827 2388494.3828697307, 698413.5003455497 2388495.90071853), (698400.5682737827 2388494.3828697307, 698413.5003455497 2388495.90071853), (698413.5003455497 2388495.90071853, 698440.416211779 2388499.05985776))"
	checkSnappingNoder(t, wkt1, "", 1, expected)
}

func checkSnappingNoder(t *testing.T, wkt1, wkt2 string, snapDist float64, expectedWKT string) {
	t.Helper()
	geom1 := readWKT(t, wkt1)
	var geom2 *Geom_Geometry
	if wkt2 != "" {
		geom2 = readWKT(t, wkt2)
	}

	noder := NodingSnap_NewSnappingNoder(snapDist)
	result := Noding_TestUtil_NodeValidated(geom1, geom2, noder)

	// Only check if expected was provided.
	if expectedWKT == "" {
		return
	}
	expected := readWKT(t, expectedWKT)
	checkEqualGeom(t, expected, result)
}
