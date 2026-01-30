package jts_test

import (
	"testing"

	"github.com/peterstace/simplefeatures/internal/jtsport/jts"
)

func checkIntersectsDisjoint(t *testing.T, wkta, wktb string, expectedValue bool) {
	t.Helper()
	checkPredicate(t, jts.OperationRelateng_RelatePredicate_Intersects(), wkta, wktb, expectedValue)
	checkPredicate(t, jts.OperationRelateng_RelatePredicate_Intersects(), wktb, wkta, expectedValue)
	checkPredicate(t, jts.OperationRelateng_RelatePredicate_Disjoint(), wkta, wktb, !expectedValue)
	checkPredicate(t, jts.OperationRelateng_RelatePredicate_Disjoint(), wktb, wkta, !expectedValue)
}

func checkContainsWithin(t *testing.T, wkta, wktb string, expectedValue bool) {
	t.Helper()
	checkPredicate(t, jts.OperationRelateng_RelatePredicate_Contains(), wkta, wktb, expectedValue)
	checkPredicate(t, jts.OperationRelateng_RelatePredicate_Within(), wktb, wkta, expectedValue)
}

func checkCoversCoveredBy(t *testing.T, wkta, wktb string, expectedValue bool) {
	t.Helper()
	checkPredicate(t, jts.OperationRelateng_RelatePredicate_Covers(), wkta, wktb, expectedValue)
	checkPredicate(t, jts.OperationRelateng_RelatePredicate_CoveredBy(), wktb, wkta, expectedValue)
}

func checkCrosses(t *testing.T, wkta, wktb string, expectedValue bool) {
	t.Helper()
	checkPredicate(t, jts.OperationRelateng_RelatePredicate_Crosses(), wkta, wktb, expectedValue)
	checkPredicate(t, jts.OperationRelateng_RelatePredicate_Crosses(), wktb, wkta, expectedValue)
}

func checkOverlaps(t *testing.T, wkta, wktb string, expectedValue bool) {
	t.Helper()
	checkPredicate(t, jts.OperationRelateng_RelatePredicate_Overlaps(), wkta, wktb, expectedValue)
	checkPredicate(t, jts.OperationRelateng_RelatePredicate_Overlaps(), wktb, wkta, expectedValue)
}

func checkTouches(t *testing.T, wkta, wktb string, expectedValue bool) {
	t.Helper()
	checkPredicate(t, jts.OperationRelateng_RelatePredicate_Touches(), wkta, wktb, expectedValue)
	checkPredicate(t, jts.OperationRelateng_RelatePredicate_Touches(), wktb, wkta, expectedValue)
}

func checkEquals(t *testing.T, wkta, wktb string, expectedValue bool) {
	t.Helper()
	checkPredicate(t, jts.OperationRelateng_RelatePredicate_EqualsTopo(), wkta, wktb, expectedValue)
	checkPredicate(t, jts.OperationRelateng_RelatePredicate_EqualsTopo(), wktb, wkta, expectedValue)
}

func checkRelate(t *testing.T, wkta, wktb string, expectedValue string) {
	t.Helper()
	a := readWKT(t, wkta)
	b := readWKT(t, wktb)
	im := jts.OperationRelateng_RelateNG_RelateMatrix(a, b)
	actualVal := im.String()
	if expectedValue != actualVal {
		t.Errorf("relate(%s, %s): expected %q, got %q", wkta, wktb, expectedValue, actualVal)
	}
}

func checkRelateMatches(t *testing.T, wkta, wktb, pattern string, expectedValue bool) {
	t.Helper()
	pred := jts.OperationRelateng_RelatePredicate_Matches(pattern)
	checkPredicate(t, pred, wkta, wktb, expectedValue)
}

func checkPredicate(t *testing.T, pred jts.OperationRelateng_TopologyPredicate, wkta, wktb string, expectedValue bool) {
	t.Helper()
	a := readWKT(t, wkta)
	b := readWKT(t, wktb)
	actualVal := jts.OperationRelateng_RelateNG_Relate(a, b, pred)
	if expectedValue != actualVal {
		t.Errorf("%s(%s, %s): expected %v, got %v", pred.Name(), wkta, wktb, expectedValue, actualVal)
	}
}

func readWKT(t *testing.T, wkt string) *jts.Geom_Geometry {
	t.Helper()
	reader := jts.Io_NewWKTReaderWithFactory(jts.Geom_NewGeometryFactoryDefault())
	geom, err := reader.Read(wkt)
	if err != nil {
		t.Fatalf("failed to parse WKT %q: %v", wkt, err)
	}
	return geom
}

// P/P tests.

func TestRelateNGPointsDisjoint(t *testing.T) {
	a := "POINT (0 0)"
	b := "POINT (1 1)"
	checkIntersectsDisjoint(t, a, b, false)
	checkContainsWithin(t, a, b, false)
	checkEquals(t, a, b, false)
	checkRelate(t, a, b, "FF0FFF0F2")
}

func TestRelateNGPointsContained(t *testing.T) {
	a := "MULTIPOINT (0 0, 1 1, 2 2)"
	b := "MULTIPOINT (1 1, 2 2)"
	checkIntersectsDisjoint(t, a, b, true)
	checkContainsWithin(t, a, b, true)
	checkEquals(t, a, b, false)
	checkRelate(t, a, b, "0F0FFFFF2")
}

func TestRelateNGPointsEqual(t *testing.T) {
	a := "MULTIPOINT (0 0, 1 1, 2 2)"
	b := "MULTIPOINT (0 0, 1 1, 2 2)"
	checkIntersectsDisjoint(t, a, b, true)
	checkContainsWithin(t, a, b, true)
	checkEquals(t, a, b, true)
}

func TestRelateNGValidateRelatePP13(t *testing.T) {
	a := "MULTIPOINT ((80 70), (140 120), (20 20), (200 170))"
	b := "MULTIPOINT ((80 70), (140 120), (80 170), (200 80))"
	checkIntersectsDisjoint(t, a, b, true)
	checkContainsWithin(t, a, b, false)
	checkContainsWithin(t, b, a, false)
	checkCoversCoveredBy(t, a, b, false)
	checkOverlaps(t, a, b, true)
	checkTouches(t, a, b, false)
}

// L/P tests.

func TestRelateNGLinePointContains(t *testing.T) {
	a := "LINESTRING (0 0, 1 1, 2 2)"
	b := "MULTIPOINT (0 0, 1 1, 2 2)"
	checkRelate(t, a, b, "0F10FFFF2")
	checkIntersectsDisjoint(t, a, b, true)
	checkContainsWithin(t, a, b, true)
	checkContainsWithin(t, b, a, false)
	checkCoversCoveredBy(t, a, b, true)
	checkCoversCoveredBy(t, b, a, false)
}

func TestRelateNGLinePointOverlaps(t *testing.T) {
	a := "LINESTRING (0 0, 1 1)"
	b := "MULTIPOINT (0 0, 1 1, 2 2)"
	checkIntersectsDisjoint(t, a, b, true)
	checkContainsWithin(t, a, b, false)
	checkContainsWithin(t, b, a, false)
	checkCoversCoveredBy(t, a, b, false)
	checkCoversCoveredBy(t, b, a, false)
}

func TestRelateNGZeroLengthLinePoint(t *testing.T) {
	a := "LINESTRING (0 0, 0 0)"
	b := "POINT (0 0)"
	checkRelate(t, a, b, "0FFFFFFF2")
	checkIntersectsDisjoint(t, a, b, true)
	checkContainsWithin(t, a, b, true)
	checkContainsWithin(t, b, a, true)
	checkCoversCoveredBy(t, a, b, true)
	checkCoversCoveredBy(t, b, a, true)
	checkEquals(t, a, b, true)
}

func TestRelateNGZeroLengthLineLine(t *testing.T) {
	a := "LINESTRING (10 10, 10 10, 10 10)"
	b := "LINESTRING (10 10, 10 10)"
	checkRelate(t, a, b, "0FFFFFFF2")
	checkIntersectsDisjoint(t, a, b, true)
	checkContainsWithin(t, a, b, true)
	checkContainsWithin(t, b, a, true)
	checkCoversCoveredBy(t, a, b, true)
	checkCoversCoveredBy(t, b, a, true)
	checkEquals(t, a, b, true)
}

func TestRelateNGNonZeroLengthLinePoint(t *testing.T) {
	a := "LINESTRING (0 0, 0 0, 9 9)"
	b := "POINT (1 1)"
	checkRelate(t, a, b, "0F1FF0FF2")
	checkIntersectsDisjoint(t, a, b, true)
	checkContainsWithin(t, a, b, true)
	checkContainsWithin(t, b, a, false)
	checkCoversCoveredBy(t, a, b, true)
	checkCoversCoveredBy(t, b, a, false)
	checkEquals(t, a, b, false)
}

func TestRelateNGLinePointIntAndExt(t *testing.T) {
	a := "MULTIPOINT((60 60), (100 100))"
	b := "LINESTRING(40 40, 80 80)"
	checkRelate(t, a, b, "0F0FFF102")
}

// L/L tests.

func TestRelateNGLinesCrossProper(t *testing.T) {
	a := "LINESTRING (0 0, 9 9)"
	b := "LINESTRING(0 9, 9 0)"
	checkIntersectsDisjoint(t, a, b, true)
	checkContainsWithin(t, a, b, false)
}

func TestRelateNGLinesOverlap(t *testing.T) {
	a := "LINESTRING (0 0, 5 5)"
	b := "LINESTRING(3 3, 9 9)"
	checkIntersectsDisjoint(t, a, b, true)
	checkTouches(t, a, b, false)
	checkOverlaps(t, a, b, true)
}

func TestRelateNGLinesCrossVertex(t *testing.T) {
	a := "LINESTRING (0 0, 8 8)"
	b := "LINESTRING(0 8, 4 4, 8 0)"
	checkIntersectsDisjoint(t, a, b, true)
}

func TestRelateNGLinesTouchVertex(t *testing.T) {
	a := "LINESTRING (0 0, 8 0)"
	b := "LINESTRING(0 8, 4 0, 8 8)"
	checkIntersectsDisjoint(t, a, b, true)
}

func TestRelateNGLinesDisjointByEnvelope(t *testing.T) {
	a := "LINESTRING (0 0, 9 9)"
	b := "LINESTRING(10 19, 19 10)"
	checkIntersectsDisjoint(t, a, b, false)
	checkContainsWithin(t, a, b, false)
}

func TestRelateNGLinesDisjoint(t *testing.T) {
	a := "LINESTRING (0 0, 9 9)"
	b := "LINESTRING (4 2, 8 6)"
	checkIntersectsDisjoint(t, a, b, false)
	checkContainsWithin(t, a, b, false)
}

func TestRelateNGLinesClosedEmpty(t *testing.T) {
	a := "MULTILINESTRING ((0 0, 0 1), (0 1, 1 1, 1 0, 0 0))"
	b := "LINESTRING EMPTY"
	checkRelate(t, a, b, "FF1FFFFF2")
	checkIntersectsDisjoint(t, a, b, false)
	checkContainsWithin(t, a, b, false)
}

func TestRelateNGLinesRingTouchAtNode(t *testing.T) {
	a := "LINESTRING (5 5, 1 8, 1 1, 5 5)"
	b := "LINESTRING (5 5, 9 5)"
	checkRelate(t, a, b, "F01FFF102")
	checkIntersectsDisjoint(t, a, b, true)
	checkContainsWithin(t, a, b, false)
	checkTouches(t, a, b, true)
}

func TestRelateNGLinesTouchAtBdy(t *testing.T) {
	a := "LINESTRING (5 5, 1 8)"
	b := "LINESTRING (5 5, 9 5)"
	checkRelate(t, a, b, "FF1F00102")
	checkIntersectsDisjoint(t, a, b, true)
	checkContainsWithin(t, a, b, false)
	checkTouches(t, a, b, true)
}

func TestRelateNGLinesOverlapWithDisjointLine(t *testing.T) {
	a := "LINESTRING (1 1, 9 9)"
	b := "MULTILINESTRING ((2 2, 8 8), (6 2, 8 4))"
	checkRelate(t, a, b, "101FF0102")
	checkIntersectsDisjoint(t, a, b, true)
	checkContainsWithin(t, a, b, false)
	checkOverlaps(t, a, b, true)
}

func TestRelateNGLinesDisjointOverlappingEnvelopes(t *testing.T) {
	a := "LINESTRING (60 0, 20 80, 100 80, 80 120, 40 140)"
	b := "LINESTRING (60 40, 140 40, 140 160, 0 160)"
	checkRelate(t, a, b, "FF1FF0102")
	checkIntersectsDisjoint(t, a, b, false)
	checkContainsWithin(t, a, b, false)
	checkTouches(t, a, b, false)
}

func TestRelateNGLinesCrossJTS270(t *testing.T) {
	a := "LINESTRING (0 0, -10 0.0000000000000012)"
	b := "LINESTRING (-9.999143275740073 -0.1308959557133398, -10 0.0000000000001054)"
	checkIntersectsDisjoint(t, a, b, true)
	checkContainsWithin(t, a, b, false)
	checkCoversCoveredBy(t, a, b, false)
	checkCrosses(t, a, b, false)
	checkOverlaps(t, a, b, false)
	checkTouches(t, a, b, true)
}

func TestRelateNGLinesContainedJTS396(t *testing.T) {
	a := "LINESTRING (1 0, 0 2, 0 0, 2 2)"
	b := "LINESTRING (0 0, 2 2)"
	checkIntersectsDisjoint(t, a, b, true)
	checkContainsWithin(t, a, b, true)
	checkCoversCoveredBy(t, a, b, true)
	checkCrosses(t, a, b, false)
	checkOverlaps(t, a, b, false)
	checkTouches(t, a, b, false)
}

func TestRelateNGLinesContainedWithSelfIntersection(t *testing.T) {
	a := "LINESTRING (2 0, 0 2, 0 0, 2 2)"
	b := "LINESTRING (0 0, 2 2)"
	checkContainsWithin(t, a, b, true)
	checkCoversCoveredBy(t, a, b, true)
	checkCrosses(t, a, b, false)
	checkOverlaps(t, a, b, false)
	checkTouches(t, a, b, false)
}

func TestRelateNGLineContainedInRing(t *testing.T) {
	a := "LINESTRING(60 60, 100 100, 140 60)"
	b := "LINESTRING(100 100, 180 20, 20 20, 100 100)"
	checkIntersectsDisjoint(t, a, b, true)
	checkContainsWithin(t, b, a, true)
	checkCoversCoveredBy(t, b, a, true)
	checkCrosses(t, a, b, false)
	checkOverlaps(t, a, b, false)
	checkTouches(t, a, b, false)
}

func TestRelateNGLineLineProperIntersection(t *testing.T) {
	a := "MULTILINESTRING ((0 0, 1 1), (0.5 0.5, 1 0.1, -1 0.1))"
	b := "LINESTRING (0 0, 1 1)"
	checkContainsWithin(t, a, b, true)
	checkCoversCoveredBy(t, a, b, true)
	checkCrosses(t, a, b, false)
	checkOverlaps(t, a, b, false)
	checkTouches(t, a, b, false)
}

func TestRelateNGLineSelfIntersectionCollinear(t *testing.T) {
	a := "LINESTRING (9 6, 1 6, 1 0, 5 6, 9 6)"
	b := "LINESTRING (9 9, 3 1)"
	checkRelate(t, a, b, "0F1FFF102")
}

// A/P tests.

func TestRelateNGPolygonPointInside(t *testing.T) {
	a := "POLYGON ((0 10, 10 10, 10 0, 0 0, 0 10))"
	b := "POINT (1 1)"
	checkIntersectsDisjoint(t, a, b, true)
	checkContainsWithin(t, a, b, true)
}

func TestRelateNGPolygonPointOutside(t *testing.T) {
	a := "POLYGON ((10 0, 0 0, 0 10, 10 0))"
	b := "POINT (8 8)"
	checkIntersectsDisjoint(t, a, b, false)
	checkContainsWithin(t, a, b, false)
}

func TestRelateNGPolygonPointInBoundary(t *testing.T) {
	a := "POLYGON ((10 0, 0 0, 0 10, 10 0))"
	b := "POINT (1 0)"
	checkIntersectsDisjoint(t, a, b, true)
	checkContainsWithin(t, a, b, false)
	checkCoversCoveredBy(t, a, b, true)
}

func TestRelateNGAreaPointInExterior(t *testing.T) {
	a := "POLYGON ((1 5, 5 5, 5 1, 1 1, 1 5))"
	b := "POINT (7 7)"
	checkRelate(t, a, b, "FF2FF10F2")
	checkIntersectsDisjoint(t, a, b, false)
	checkContainsWithin(t, a, b, false)
	checkCoversCoveredBy(t, a, b, false)
	checkTouches(t, a, b, false)
	checkOverlaps(t, a, b, false)
}

// A/L tests.

func TestRelateNGAreaLineContainedAtLineVertex(t *testing.T) {
	a := "POLYGON ((1 5, 5 5, 5 1, 1 1, 1 5))"
	b := "LINESTRING (2 3, 3 5, 4 3)"
	checkIntersectsDisjoint(t, a, b, true)
	checkTouches(t, a, b, false)
	checkOverlaps(t, a, b, false)
}

func TestRelateNGAreaLineTouchAtLineVertex(t *testing.T) {
	a := "POLYGON ((1 5, 5 5, 5 1, 1 1, 1 5))"
	b := "LINESTRING (1 8, 3 5, 5 8)"
	checkIntersectsDisjoint(t, a, b, true)
	checkContainsWithin(t, a, b, false)
	checkCoversCoveredBy(t, a, b, false)
	checkTouches(t, a, b, true)
	checkOverlaps(t, a, b, false)
}

func TestRelateNGPolygonLineInside(t *testing.T) {
	a := "POLYGON ((0 10, 10 10, 10 0, 0 0, 0 10))"
	b := "LINESTRING (1 8, 3 5, 5 8)"
	checkRelate(t, a, b, "102FF1FF2")
	checkIntersectsDisjoint(t, a, b, true)
	checkContainsWithin(t, a, b, true)
}

func TestRelateNGPolygonLineOutside(t *testing.T) {
	a := "POLYGON ((10 0, 0 0, 0 10, 10 0))"
	b := "LINESTRING (4 8, 9 3)"
	checkIntersectsDisjoint(t, a, b, false)
	checkContainsWithin(t, a, b, false)
}

func TestRelateNGPolygonLineInBoundary(t *testing.T) {
	a := "POLYGON ((10 0, 0 0, 0 10, 10 0))"
	b := "LINESTRING (1 0, 9 0)"
	checkIntersectsDisjoint(t, a, b, true)
	checkContainsWithin(t, a, b, false)
	checkCoversCoveredBy(t, a, b, true)
	checkTouches(t, a, b, true)
	checkOverlaps(t, a, b, false)
}

func TestRelateNGPolygonLineCrossingContained(t *testing.T) {
	a := "MULTIPOLYGON (((20 80, 180 80, 100 0, 20 80)), ((20 160, 180 160, 100 80, 20 160)))"
	b := "LINESTRING (100 140, 100 40)"
	checkRelate(t, a, b, "1020F1FF2")
	checkIntersectsDisjoint(t, a, b, true)
	checkContainsWithin(t, a, b, true)
	checkCoversCoveredBy(t, a, b, true)
	checkTouches(t, a, b, false)
	checkOverlaps(t, a, b, false)
}

func TestRelateNGValidateRelateLA220(t *testing.T) {
	a := "LINESTRING (90 210, 210 90)"
	b := "POLYGON ((150 150, 410 150, 280 20, 20 20, 150 150))"
	checkIntersectsDisjoint(t, a, b, true)
	checkContainsWithin(t, a, b, false)
	checkCoversCoveredBy(t, a, b, false)
	checkTouches(t, a, b, false)
	checkOverlaps(t, a, b, false)
}

func TestRelateNGLineCrossingPolygonAtShellHolePoint(t *testing.T) {
	a := "LINESTRING (60 160, 150 70)"
	b := "POLYGON ((190 190, 360 20, 20 20, 190 190), (110 110, 250 100, 140 30, 110 110))"
	checkRelate(t, a, b, "F01FF0212")
	checkTouches(t, a, b, true)
	checkIntersectsDisjoint(t, a, b, true)
	checkContainsWithin(t, a, b, false)
	checkCoversCoveredBy(t, a, b, false)
	checkOverlaps(t, a, b, false)
}

func TestRelateNGLineCrossingPolygonAtNonVertex(t *testing.T) {
	a := "LINESTRING (20 60, 150 60)"
	b := "POLYGON ((150 150, 410 150, 280 20, 20 20, 150 150))"
	checkIntersectsDisjoint(t, a, b, true)
	checkContainsWithin(t, a, b, false)
	checkCoversCoveredBy(t, a, b, false)
	checkTouches(t, a, b, false)
	checkOverlaps(t, a, b, false)
}

func TestRelateNGPolygonLinesContainedCollinearEdge(t *testing.T) {
	a := "POLYGON ((110 110, 200 20, 20 20, 110 110))"
	b := "MULTILINESTRING ((110 110, 60 40, 70 20, 150 20, 170 40), (180 30, 40 30, 110 80))"
	checkRelate(t, a, b, "102101FF2")
}

// A/A tests.

func TestRelateNGPolygonsEdgeAdjacent(t *testing.T) {
	a := "POLYGON ((1 3, 3 3, 3 1, 1 1, 1 3))"
	b := "POLYGON ((5 3, 5 1, 3 1, 3 3, 5 3))"
	checkOverlaps(t, a, b, false)
	checkTouches(t, a, b, true)
}

func TestRelateNGPolygonsEdgeAdjacent2(t *testing.T) {
	a := "POLYGON ((1 3, 4 3, 3 0, 1 1, 1 3))"
	b := "POLYGON ((5 3, 5 1, 3 0, 4 3, 5 3))"
	checkOverlaps(t, a, b, false)
	checkTouches(t, a, b, true)
}

func TestRelateNGPolygonsNested(t *testing.T) {
	a := "POLYGON ((1 9, 9 9, 9 1, 1 1, 1 9))"
	b := "POLYGON ((2 8, 8 8, 8 2, 2 2, 2 8))"
	checkIntersectsDisjoint(t, a, b, true)
	checkContainsWithin(t, a, b, true)
	checkCoversCoveredBy(t, a, b, true)
	checkOverlaps(t, a, b, false)
	checkTouches(t, a, b, false)
}

func TestRelateNGPolygonsOverlapProper(t *testing.T) {
	a := "POLYGON ((1 1, 1 7, 7 7, 7 1, 1 1))"
	b := "POLYGON ((2 8, 8 8, 8 2, 2 2, 2 8))"
	checkIntersectsDisjoint(t, a, b, true)
	checkContainsWithin(t, a, b, false)
	checkCoversCoveredBy(t, a, b, false)
	checkOverlaps(t, a, b, true)
	checkTouches(t, a, b, false)
}

func TestRelateNGPolygonsOverlapAtNodes(t *testing.T) {
	a := "POLYGON ((1 5, 5 5, 5 1, 1 1, 1 5))"
	b := "POLYGON ((7 3, 5 1, 3 3, 5 5, 7 3))"
	checkIntersectsDisjoint(t, a, b, true)
	checkContainsWithin(t, a, b, false)
	checkCoversCoveredBy(t, a, b, false)
	checkOverlaps(t, a, b, true)
	checkTouches(t, a, b, false)
}

func TestRelateNGPolygonsContainedAtNodes(t *testing.T) {
	a := "POLYGON ((1 5, 5 5, 6 2, 1 1, 1 5))"
	b := "POLYGON ((1 1, 5 5, 6 2, 1 1))"
	checkContainsWithin(t, a, b, true)
	checkCoversCoveredBy(t, a, b, true)
	checkOverlaps(t, a, b, false)
	checkTouches(t, a, b, false)
}

func TestRelateNGPolygonsNestedWithHole(t *testing.T) {
	a := "POLYGON ((40 60, 420 60, 420 320, 40 320, 40 60), (200 140, 160 220, 260 200, 200 140))"
	b := "POLYGON ((80 100, 360 100, 360 280, 80 280, 80 100))"
	checkContainsWithin(t, a, b, false)
	checkContainsWithin(t, b, a, false)
	checkPredicate(t, jts.OperationRelateng_RelatePredicate_Contains(), a, b, false)
}

func TestRelateNGPolygonsOverlappingWithBoundaryInside(t *testing.T) {
	a := "POLYGON ((100 60, 140 100, 100 140, 60 100, 100 60))"
	b := "MULTIPOLYGON (((80 40, 120 40, 120 80, 80 80, 80 40)), ((120 80, 160 80, 160 120, 120 120, 120 80)), ((80 120, 120 120, 120 160, 80 160, 80 120)), ((40 80, 80 80, 80 120, 40 120, 40 80)))"
	checkRelate(t, a, b, "21210F212")
	checkIntersectsDisjoint(t, a, b, true)
	checkContainsWithin(t, a, b, false)
	checkContainsWithin(t, b, a, false)
	checkCoversCoveredBy(t, a, b, false)
	checkOverlaps(t, a, b, true)
	checkTouches(t, a, b, false)
}

func TestRelateNGPolygonsOverlapVeryNarrow(t *testing.T) {
	a := "POLYGON ((120 100, 120 200, 200 200, 200 100, 120 100))"
	b := "POLYGON ((100 100, 100000 110, 100000 100, 100 100))"
	checkRelate(t, a, b, "212111212")
	checkIntersectsDisjoint(t, a, b, true)
	checkContainsWithin(t, a, b, false)
	checkContainsWithin(t, b, a, false)
}

func TestRelateNGValidateRelateAA86(t *testing.T) {
	a := "POLYGON ((170 120, 300 120, 250 70, 120 70, 170 120))"
	b := "POLYGON ((150 150, 410 150, 280 20, 20 20, 150 150), (170 120, 330 120, 260 50, 100 50, 170 120))"
	checkIntersectsDisjoint(t, a, b, true)
	checkContainsWithin(t, a, b, false)
	checkCoversCoveredBy(t, a, b, false)
	checkOverlaps(t, a, b, false)
	checkPredicate(t, jts.OperationRelateng_RelatePredicate_Within(), a, b, false)
	checkTouches(t, a, b, true)
}

func TestRelateNGValidateRelateAA97(t *testing.T) {
	a := "POLYGON ((330 150, 200 110, 150 150, 280 190, 330 150))"
	b := "MULTIPOLYGON (((140 110, 260 110, 170 20, 50 20, 140 110)), ((300 270, 420 270, 340 190, 220 190, 300 270)))"
	checkIntersectsDisjoint(t, a, b, true)
	checkContainsWithin(t, a, b, false)
	checkCoversCoveredBy(t, a, b, false)
	checkOverlaps(t, a, b, false)
	checkPredicate(t, jts.OperationRelateng_RelatePredicate_Within(), a, b, false)
	checkTouches(t, a, b, true)
}

func TestRelateNGAdjacentPolygons(t *testing.T) {
	a := "POLYGON ((1 9, 6 9, 6 1, 1 1, 1 9))"
	b := "POLYGON ((9 9, 9 4, 6 4, 6 9, 9 9))"
	checkRelateMatches(t, a, b, jts.OperationRelateng_IntersectionMatrixPattern_ADJACENT, true)
}

func TestRelateNGAdjacentPolygonsTouchingAtPoint(t *testing.T) {
	a := "POLYGON ((1 9, 6 9, 6 1, 1 1, 1 9))"
	b := "POLYGON ((9 9, 9 4, 6 4, 7 9, 9 9))"
	checkRelateMatches(t, a, b, jts.OperationRelateng_IntersectionMatrixPattern_ADJACENT, false)
}

func TestRelateNGAdjacentPolygonsOverlapping(t *testing.T) {
	a := "POLYGON ((1 9, 6 9, 6 1, 1 1, 1 9))"
	b := "POLYGON ((9 9, 9 4, 6 4, 5 9, 9 9))"
	checkRelateMatches(t, a, b, jts.OperationRelateng_IntersectionMatrixPattern_ADJACENT, false)
}

func TestRelateNGContainsProperlyPolygonContained(t *testing.T) {
	a := "POLYGON ((1 9, 9 9, 9 1, 1 1, 1 9))"
	b := "POLYGON ((2 8, 5 8, 5 5, 2 5, 2 8))"
	checkRelateMatches(t, a, b, jts.OperationRelateng_IntersectionMatrixPattern_CONTAINS_PROPERLY, true)
}

func TestRelateNGContainsProperlyPolygonTouching(t *testing.T) {
	a := "POLYGON ((1 9, 9 9, 9 1, 1 1, 1 9))"
	b := "POLYGON ((9 1, 5 1, 5 5, 9 5, 9 1))"
	checkRelateMatches(t, a, b, jts.OperationRelateng_IntersectionMatrixPattern_CONTAINS_PROPERLY, false)
}

func TestRelateNGContainsProperlyPolygonsOverlapping(t *testing.T) {
	a := "GEOMETRYCOLLECTION (POLYGON ((1 9, 6 9, 6 4, 1 4, 1 9)), POLYGON ((2 4, 6 7, 9 1, 2 4)))"
	b := "POLYGON ((5 5, 6 5, 6 4, 5 4, 5 5))"
	checkRelateMatches(t, a, b, jts.OperationRelateng_IntersectionMatrixPattern_CONTAINS_PROPERLY, true)
}

// Repeated Points.

func TestRelateNGRepeatedPointLL(t *testing.T) {
	a := "LINESTRING(0 0, 5 5, 5 5, 5 5, 9 9)"
	b := "LINESTRING(0 9, 5 5, 5 5, 5 5, 9 0)"
	checkRelate(t, a, b, "0F1FF0102")
	checkIntersectsDisjoint(t, a, b, true)
}

func TestRelateNGRepeatedPointAA(t *testing.T) {
	a := "POLYGON ((1 9, 9 7, 9 1, 1 3, 1 9))"
	b := "POLYGON ((1 3, 1 3, 1 3, 3 7, 9 7, 9 7, 1 3))"
	checkRelate(t, a, b, "212F01FF2")
}

// Empty.

func TestRelateNGEmptyEquals(t *testing.T) {
	empties := []string{
		"POINT EMPTY",
		"LINESTRING EMPTY",
		"POLYGON EMPTY",
		"MULTIPOINT EMPTY",
		"MULTILINESTRING EMPTY",
		"MULTIPOLYGON EMPTY",
		"GEOMETRYCOLLECTION EMPTY",
	}
	for _, a := range empties {
		for _, b := range empties {
			checkRelate(t, a, b, "FFFFFFFF2")
			checkEquals(t, a, b, false)
		}
	}
}

// Prepared.

func TestRelateNGPreparedAA(t *testing.T) {
	a := "POLYGON((0 0, 1 0, 1 1, 0 1, 0 0))"
	b := "POLYGON((0.5 0.5, 1.5 0.5, 1.5 1.5, 0.5 1.5, 0.5 0.5))"
	checkPrepared(t, a, b)
}

func checkPrepared(t *testing.T, wkta, wktb string) {
	t.Helper()
	a := readWKT(t, wkta)
	b := readWKT(t, wktb)
	prepA := jts.OperationRelateng_RelateNG_Prepare(a)

	// Test various predicates.
	predicates := []struct {
		name string
		pred jts.OperationRelateng_TopologyPredicate
	}{
		{"equalsTopo", jts.OperationRelateng_RelatePredicate_EqualsTopo()},
		{"intersects", jts.OperationRelateng_RelatePredicate_Intersects()},
		{"disjoint", jts.OperationRelateng_RelatePredicate_Disjoint()},
		{"covers", jts.OperationRelateng_RelatePredicate_Covers()},
		{"coveredBy", jts.OperationRelateng_RelatePredicate_CoveredBy()},
		{"within", jts.OperationRelateng_RelatePredicate_Within()},
		{"contains", jts.OperationRelateng_RelatePredicate_Contains()},
		{"crosses", jts.OperationRelateng_RelatePredicate_Crosses()},
		{"touches", jts.OperationRelateng_RelatePredicate_Touches()},
	}

	for _, p := range predicates {
		prepResult := prepA.EvaluatePredicate(b, p.pred)
		directResult := jts.OperationRelateng_RelateNG_Relate(a, b, p.pred)
		if prepResult != directResult {
			t.Errorf("%s: prepared=%v, direct=%v", p.name, prepResult, directResult)
		}
	}

	// Test relate matrix.
	prepIM := prepA.Evaluate(b).String()
	directIM := jts.OperationRelateng_RelateNG_RelateMatrix(a, b).String()
	if prepIM != directIM {
		t.Errorf("relate: prepared=%s, direct=%s", prepIM, directIM)
	}
}
