package jts

import "testing"

const snapRoundingTest_snapTolerance = 1.0

func TestSnapRoundingPolyWithCloseNode(t *testing.T) {
	wkts := []string{
		"POLYGON ((20 0, 20 160, 140 1, 160 160, 160 1, 20 0))",
	}
	checkSnapRoundingLines(t, wkts)
}

func TestSnapRoundingPolyWithCloseNodeFrac(t *testing.T) {
	wkts := []string{
		"POLYGON ((20 0, 20 160, 140 0.2, 160 160, 160 0, 20 0))",
	}
	checkSnapRoundingLines(t, wkts)
}

func TestSnapRoundingLineStringLongShort(t *testing.T) {
	wkts := []string{
		"LINESTRING (0 0, 2 0)",
		"LINESTRING (0 0, 10 -1)",
	}
	checkSnapRoundingLines(t, wkts)
}

func TestSnapRoundingBadLines1(t *testing.T) {
	wkts := []string{
		"LINESTRING ( 171 157, 175 154, 170 154, 170 155, 170 156, 170 157, 171 158, 171 159, 172 160, 176 156, 171 156, 171 159, 176 159, 172 155, 170 157, 174 161, 174 156, 173 156, 172 156 )",
	}
	checkSnapRoundingLines(t, wkts)
}

func TestSnapRoundingBadLines2(t *testing.T) {
	wkts := []string{
		"LINESTRING ( 175 222, 176 222, 176 219, 174 221, 175 222, 177 220, 174 220, 174 222, 177 222, 175 220, 174 221 )",
	}
	checkSnapRoundingLines(t, wkts)
}

func TestSnapRoundingCollapse1(t *testing.T) {
	wkts := []string{
		"LINESTRING ( 362 177, 375 164, 374 164, 372 161, 373 163, 372 165, 373 164, 442 58 )",
	}
	checkSnapRoundingLines(t, wkts)
}

func TestSnapRoundingCollapse2(t *testing.T) {
	wkts := []string{
		"LINESTRING ( 393 175, 391 173, 390 175, 391 174, 391 173 )",
	}
	checkSnapRoundingLines(t, wkts)
}

func TestSnapRoundingLineWithManySelfSnaps(t *testing.T) {
	wkts := []string{
		"LINESTRING (0 0, 6 4, 8 11, 13 13, 14 12, 11 12, 7 7, 7 3, 4 2)",
	}
	checkSnapRoundingLines(t, wkts)
}

func TestSnapRoundingBadNoding1(t *testing.T) {
	wkts := []string{
		"LINESTRING ( 76 47, 81 52, 81 53, 85 57, 88 62, 89 64, 57 80, 82 55, 101 74, 76 99, 92 67, 94 68, 99 71, 103 75, 139 111 )",
	}
	checkSnapRoundingLines(t, wkts)
}

func TestSnapRoundingBadNoding1Extract(t *testing.T) {
	wkts := []string{
		"LINESTRING ( 82 55, 101 74 )",
		"LINESTRING ( 94 68, 99 71 )",
		"LINESTRING ( 85 57, 88 62 )",
	}
	checkSnapRoundingLines(t, wkts)
}

func TestSnapRoundingBadNoding1ExtractShift(t *testing.T) {
	wkts := []string{
		"LINESTRING ( 0 0, 19 19 )",
		"LINESTRING ( 12 13, 17 16 )",
		"LINESTRING ( 3 2, 6 7 )",
	}
	checkSnapRoundingLines(t, wkts)
}

func checkSnapRoundingLines(t *testing.T, wkts []string) {
	t.Helper()
	geoms := fromWKTArray(t, wkts)
	pm := Geom_NewPrecisionModelWithScale(snapRoundingTest_snapTolerance)
	noder := NodingSnapround_NewGeometryNoder(pm)
	noder.SetValidate(true)
	nodedLines := noder.Node(geoms)

	if !isSnapped(nodedLines, snapRoundingTest_snapTolerance) {
		t.Errorf("result is not properly snapped")
	}
}

func fromWKTArray(t *testing.T, wkts []string) []*Geom_Geometry {
	t.Helper()
	reader := Io_NewWKTReader()
	result := make([]*Geom_Geometry, 0, len(wkts))
	for _, wkt := range wkts {
		geom, err := reader.Read(wkt)
		if err != nil {
			t.Fatalf("failed to parse WKT: %v", err)
		}
		result = append(result, geom)
	}
	return result
}

func isSnapped(lines []*Geom_LineString, tol float64) bool {
	for _, line := range lines {
		for j := 0; j < line.GetNumPoints(); j++ {
			v := line.GetCoordinateN(j)
			if !isVertexSnapped(v, lines) {
				return false
			}
		}
	}
	return true
}

func isVertexSnapped(v *Geom_Coordinate, lines []*Geom_LineString) bool {
	for _, line := range lines {
		for j := 0; j < line.GetNumPoints()-1; j++ {
			p0 := line.GetCoordinateN(j)
			p1 := line.GetCoordinateN(j + 1)
			if !isSnappedToSegment(v, p0, p1) {
				return false
			}
		}
	}
	return true
}

func isSnappedToSegment(v, p0, p1 *Geom_Coordinate) bool {
	if v.Equals2D(p0) {
		return true
	}
	if v.Equals2D(p1) {
		return true
	}
	seg := Geom_NewLineSegmentFromCoordinates(p0, p1)
	dist := seg.DistanceToPoint(v)
	if dist < snapRoundingTest_snapTolerance/2.05 {
		return false
	}
	return true
}
