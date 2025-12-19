package jts_test

import (
	"testing"

	"github.com/peterstace/simplefeatures/internal/jtsport/jts"
)

// TestSegmentNodeOrderingForSnappedNodes tests a case which involves nodes
// added when using the SnappingNoder. In this case one of the added nodes is
// relatively "far" from its segment, and "near" the start vertex of the
// segment. Computing the noding correctly requires the fix to
// SegmentNode.CompareTo added in https://github.com/locationtech/jts/pull/399
//
// See https://trac.osgeo.org/geos/ticket/1051
func TestSegmentNodeOrderingForSnappedNodes(t *testing.T) {
	checkNoding(t,
		"LINESTRING (655103.6628454948 1794805.456674405, 655016.20226 1794940.10998, 655014.8317182435 1794941.5196832407)",
		"MULTIPOINT((655016.29615051334 1794939.965427252), (655016.20226531825 1794940.1099718122), (655016.20226 1794940.10998), (655016.20225819293 1794940.1099794197))",
		[]int{0, 0, 1, 1},
		"MULTILINESTRING ((655014.8317182435 1794941.5196832407, 655016.2022581929 1794940.1099794197), (655016.2022581929 1794940.1099794197, 655016.20226 1794940.10998), (655016.20226 1794940.10998, 655016.2022653183 1794940.1099718122), (655016.2022653183 1794940.1099718122, 655016.2961505133 1794939.965427252), (655016.2961505133 1794939.965427252, 655103.6628454948 1794805.456674405))",
	)
}

func checkNoding(t *testing.T, wktLine, wktNodes string, segmentIndex []int, wktExpected string) {
	t.Helper()
	reader := jts.Io_NewWKTReader()
	line, err := reader.Read(wktLine)
	if err != nil {
		t.Fatalf("failed to parse line: %v", err)
	}
	pts, err := reader.Read(wktNodes)
	if err != nil {
		t.Fatalf("failed to parse nodes: %v", err)
	}

	nss := jts.Noding_NewNodedSegmentString(line.GetCoordinates(), nil)
	nodes := pts.GetCoordinates()

	for i, node := range nodes {
		nss.AddIntersection(node, segmentIndex[i])
	}

	nodedSS := nodingTestUtilGetNodedSubstrings(nss)
	result := nodingTestUtilToLines(nodedSS, line.GetFactory())

	expected, err := reader.Read(wktExpected)
	if err != nil {
		t.Fatalf("failed to parse expected: %v", err)
	}

	if !result.EqualsNorm(expected) {
		t.Errorf("result does not match expected\nexpected: %s\ngot: %s", wktExpected, result.String())
	}
}

func nodingTestUtilGetNodedSubstrings(nss *jts.Noding_NodedSegmentString) []*jts.Noding_NodedSegmentString {
	var resultEdgelist []*jts.Noding_NodedSegmentString
	nss.GetNodeList().AddSplitEdges(&resultEdgelist)
	return resultEdgelist
}

func nodingTestUtilToLines(nodedList []*jts.Noding_NodedSegmentString, geomFact *jts.Geom_GeometryFactory) *jts.Geom_Geometry {
	lines := make([]*jts.Geom_LineString, len(nodedList))
	for i, nss := range nodedList {
		pts := nss.GetCoordinates()
		line := geomFact.CreateLineStringFromCoordinates(pts)
		lines[i] = line
	}
	if len(lines) == 1 {
		return lines[0].Geom_Geometry
	}
	return geomFact.CreateMultiLineStringFromLineStrings(lines).Geom_Geometry
}
