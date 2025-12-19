package jts

import "testing"

// TestSegmentStringNodingThinTriangle tests noding with a thin triangle.
func TestSegmentStringNodingThinTriangle(t *testing.T) {
	wkt := "LINESTRING ( 55121.54481117887 42694.49730855581, 55121.54481117887 42694.4973085558, 55121.458748617406 42694.419143944244, 55121.54481117887 42694.49730855581 )"
	pm := Geom_NewPrecisionModelWithScale(1.1131949079327356e11)
	checkNodedStrings(t, wkt, pm)
}

// TestSegmentStringNodingSegmentLength1Failure tests a failure case.
func TestSegmentStringNodingSegmentLength1Failure(t *testing.T) {
	wkt := "LINESTRING ( -1677607.6366504875 -588231.47100446, -1674050.1010869485 -587435.2186255794, -1670493.6527468169 -586636.7948791061, -1424286.3681743187 -525586.1397894835, -1670493.6527468169 -586636.7948791061, -1674050.1010869485 -587435.2186255795, -1677607.6366504875 -588231.47100446)"
	pm := Geom_NewPrecisionModelWithScale(1.11e10)
	checkNodedStrings(t, wkt, pm)
}

func checkNodedStrings(t *testing.T, wkt string, pm *Geom_PrecisionModel) {
	t.Helper()
	reader := Io_NewWKTReader()
	geom, err := reader.Read(wkt)
	if err != nil {
		t.Fatalf("failed to parse WKT: %v", err)
	}

	nss := Noding_NewNodedSegmentString(geom.GetCoordinates(), nil)
	strings := []Noding_SegmentString{nss}
	noder := NodingSnapround_NewSnapRoundingNoder(pm)
	noder.ComputeNodes(strings)

	noded := Noding_NodedSegmentString_GetNodedSubstrings(
		[]*Noding_NodedSegmentString{nss},
	)

	for _, s := range noded {
		if s.Size() < 2 {
			t.Errorf("found a 1-point segmentstring")
		}
		if isCollapsed(s) {
			t.Errorf("found a collapsed edge")
		}
	}
}

// isCollapsed tests if the segmentString is a collapsed edge of the form ABA.
// These should not be returned by noding.
func isCollapsed(s *Noding_NodedSegmentString) bool {
	if s.Size() != 3 {
		return false
	}
	isEndsEqual := s.GetCoordinate(0).Equals2D(s.GetCoordinate(2))
	isMiddleDifferent := !s.GetCoordinate(0).Equals2D(s.GetCoordinate(1))
	return isEndsEqual && isMiddleDifferent
}
