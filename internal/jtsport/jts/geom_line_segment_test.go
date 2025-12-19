package jts_test

import (
	"math"
	"testing"

	"github.com/peterstace/simplefeatures/internal/jtsport/jts"
)

var root2 = math.Sqrt(2)

func TestLineSegmentHashCode(t *testing.T) {
	checkHashcode(t, jts.Geom_NewLineSegmentFromXY(0, 0, 10, 0), jts.Geom_NewLineSegmentFromXY(0, 10, 10, 10))
	checkHashcode(t, jts.Geom_NewLineSegmentFromXY(580.0, 1330.0, 590.0, 1330.0), jts.Geom_NewLineSegmentFromXY(580.0, 1340.0, 590.0, 1340.0))
}

func checkHashcode(t *testing.T, seg, seg2 *jts.Geom_LineSegment) {
	t.Helper()
	if seg.HashCode() == seg2.HashCode() {
		t.Errorf("expected different hash codes for %v and %v", seg, seg2)
	}
}

func TestLineSegmentProjectionFactor(t *testing.T) {
	// Zero-length line.
	seg := jts.Geom_NewLineSegmentFromXY(10, 0, 10, 0)
	pf := seg.ProjectionFactor(jts.Geom_NewCoordinateWithXY(11, 0))
	if !math.IsNaN(pf) {
		t.Errorf("expected NaN for zero-length segment, got %v", pf)
	}

	seg2 := jts.Geom_NewLineSegmentFromXY(10, 0, 20, 0)
	pf2 := seg2.ProjectionFactor(jts.Geom_NewCoordinateWithXY(11, 0))
	if pf2 != 0.1 {
		t.Errorf("expected 0.1, got %v", pf2)
	}
}

func TestLineSegmentLineIntersection(t *testing.T) {
	// Simple case.
	checkLineIntersection(t,
		0, 0, 10, 10,
		0, 10, 10, 0,
		5, 5)

	// Almost collinear - See JTS GitHub issue #464.
	checkLineIntersection(t,
		35613471.6165017, 4257145.306132293, 35613477.7705378, 4257160.528222711,
		35613477.77505724, 4257160.539653536, 35613479.85607389, 4257165.92369170,
		35613477.772841461, 4257160.5339209242)
}

const maxAbsErrorIntersection = 1e-5

func checkLineIntersection(t *testing.T, p1x, p1y, p2x, p2y, q1x, q1y, q2x, q2y, expectedx, expectedy float64) {
	t.Helper()
	seg1 := jts.Geom_NewLineSegmentFromXY(p1x, p1y, p2x, p2y)
	seg2 := jts.Geom_NewLineSegmentFromXY(q1x, q1y, q2x, q2y)

	actual := seg1.LineIntersection(seg2)
	expected := jts.Geom_NewCoordinateWithXY(expectedx, expectedy)
	dist := actual.Distance(expected)
	if dist > maxAbsErrorIntersection {
		t.Errorf("expected %v, got %v, dist=%v", expected, actual, dist)
	}
}

func TestLineSegmentDistancePerpendicular(t *testing.T) {
	checkDistancePerpendicular(t, 1, 1, 1, 3, 2, 4, 1)
	checkDistancePerpendicular(t, 1, 1, 1, 3, 0, 4, 1)
	checkDistancePerpendicular(t, 1, 1, 1, 3, 1, 4, 0)
	checkDistancePerpendicular(t, 1, 1, 2, 2, 4, 4, 0)
	// Zero-length line segment.
	checkDistancePerpendicular(t, 1, 1, 1, 1, 1, 2, 1)
}

func TestLineSegmentDistancePerpendicularOriented(t *testing.T) {
	// Right of line.
	checkDistancePerpendicularOriented(t, 1, 1, 1, 3, 2, 4, -1)
	// Left of line.
	checkDistancePerpendicularOriented(t, 1, 1, 1, 3, 0, 4, 1)
	// On line.
	checkDistancePerpendicularOriented(t, 1, 1, 1, 3, 1, 4, 0)
	checkDistancePerpendicularOriented(t, 1, 1, 2, 2, 4, 4, 0)
	// Zero-length segment.
	checkDistancePerpendicularOriented(t, 1, 1, 1, 1, 1, 2, 1)
}

func checkDistancePerpendicular(t *testing.T, x0, y0, x1, y1, px, py, expected float64) {
	t.Helper()
	seg := jts.Geom_NewLineSegmentFromXY(x0, y0, x1, y1)
	dist := seg.DistancePerpendicular(jts.Geom_NewCoordinateWithXY(px, py))
	if math.Abs(dist-expected) > 0.000001 {
		t.Errorf("expected %v, got %v", expected, dist)
	}
}

func checkDistancePerpendicularOriented(t *testing.T, x0, y0, x1, y1, px, py, expected float64) {
	t.Helper()
	seg := jts.Geom_NewLineSegmentFromXY(x0, y0, x1, y1)
	dist := seg.DistancePerpendicularOriented(jts.Geom_NewCoordinateWithXY(px, py))
	if math.Abs(dist-expected) > 0.000001 {
		t.Errorf("expected %v, got %v", expected, dist)
	}
}

func TestLineSegmentOffsetPoint(t *testing.T) {
	checkOffsetPoint(t, 0, 0, 10, 10, 0.0, root2, -1, 1)
	checkOffsetPoint(t, 0, 0, 10, 10, 0.0, -root2, 1, -1)

	checkOffsetPoint(t, 0, 0, 10, 10, 1.0, root2, 9, 11)
	checkOffsetPoint(t, 0, 0, 10, 10, 0.5, root2, 4, 6)

	checkOffsetPoint(t, 0, 0, 10, 10, 0.5, -root2, 6, 4)
	checkOffsetPoint(t, 0, 0, 10, 10, 0.5, -root2, 6, 4)

	checkOffsetPoint(t, 0, 0, 10, 10, 2.0, root2, 19, 21)
	checkOffsetPoint(t, 0, 0, 10, 10, 2.0, -root2, 21, 19)

	checkOffsetPoint(t, 0, 0, 10, 10, 2.0, 5*root2, 15, 25)
	checkOffsetPoint(t, 0, 0, 10, 10, -2.0, 5*root2, -25, -15)
}

func TestLineSegmentOffsetLine(t *testing.T) {
	checkOffsetLine(t, 0, 0, 10, 10, 0, 0, 0, 10, 10)

	checkOffsetLine(t, 0, 0, 10, 10, root2, -1, 1, 9, 11)
	checkOffsetLine(t, 0, 0, 10, 10, -root2, 1, -1, 11, 9)
}

func checkOffsetPoint(t *testing.T, x0, y0, x1, y1, segFrac, offset, expectedX, expectedY float64) {
	t.Helper()
	seg := jts.Geom_NewLineSegmentFromXY(x0, y0, x1, y1)
	p := seg.PointAlongOffset(segFrac, offset)

	if !equalsTolerance(jts.Geom_NewCoordinateWithXY(expectedX, expectedY), p, 0.000001) {
		t.Errorf("expected (%v, %v), got (%v, %v)", expectedX, expectedY, p.X, p.Y)
	}
}

func checkOffsetLine(t *testing.T, x0, y0, x1, y1, offset, expectedX0, expectedY0, expectedX1, expectedY1 float64) {
	t.Helper()
	seg := jts.Geom_NewLineSegmentFromXY(x0, y0, x1, y1)
	actual := seg.Offset(offset)

	if !equalsTolerance(jts.Geom_NewCoordinateWithXY(expectedX0, expectedY0), actual.P0, 0.000001) {
		t.Errorf("P0: expected (%v, %v), got (%v, %v)", expectedX0, expectedY0, actual.P0.X, actual.P0.Y)
	}
	if !equalsTolerance(jts.Geom_NewCoordinateWithXY(expectedX1, expectedY1), actual.P1, 0.000001) {
		t.Errorf("P1: expected (%v, %v), got (%v, %v)", expectedX1, expectedY1, actual.P1.X, actual.P1.Y)
	}
}

func equalsTolerance(p0, p1 *jts.Geom_Coordinate, tolerance float64) bool {
	if math.Abs(p0.X-p1.X) > tolerance {
		return false
	}
	if math.Abs(p0.Y-p1.Y) > tolerance {
		return false
	}
	return true
}

func TestLineSegmentReflect(t *testing.T) {
	checkReflect(t, 0, 0, 10, 10, 1, 2, 2, 1)
	checkReflect(t, 0, 1, 10, 1, 1, 2, 1, 0)
}

func checkReflect(t *testing.T, x0, y0, x1, y1, x, y, expectedX, expectedY float64) {
	t.Helper()
	seg := jts.Geom_NewLineSegmentFromXY(x0, y0, x1, y1)
	p := seg.Reflect(jts.Geom_NewCoordinateWithXY(x, y))
	if !equalsTolerance(jts.Geom_NewCoordinateWithXY(expectedX, expectedY), p, 0.000001) {
		t.Errorf("expected (%v, %v), got (%v, %v)", expectedX, expectedY, p.X, p.Y)
	}
}

func TestLineSegmentOrientationIndexCoordinate(t *testing.T) {
	seg := jts.Geom_NewLineSegmentFromXY(0, 0, 10, 10)
	checkOrientationIndex(t, seg, 10, 11, 1)
	checkOrientationIndex(t, seg, 10, 9, -1)

	checkOrientationIndex(t, seg, 11, 11, 0)

	checkOrientationIndex(t, seg, 11, 11.0000001, 1)
	checkOrientationIndex(t, seg, 11, 10.9999999, -1)

	checkOrientationIndex(t, seg, -2, -1.9999999, 1)
	checkOrientationIndex(t, seg, -2, -2.0000001, -1)
}

func TestLineSegmentOrientationIndexSegment(t *testing.T) {
	seg := jts.Geom_NewLineSegmentFromXY(100, 100, 110, 110)

	checkOrientationIndexSegment(t, seg, 100, 101, 105, 106, 1)
	checkOrientationIndexSegment(t, seg, 100, 99, 105, 96, -1)

	checkOrientationIndexSegment(t, seg, 200, 200, 210, 210, 0)

	checkOrientationIndexSegment(t, seg, 105, 105, 110, 100, -1)
}

func checkOrientationIndex(t *testing.T, seg *jts.Geom_LineSegment, px, py float64, expectedOrient int) {
	t.Helper()
	p := jts.Geom_NewCoordinateWithXY(px, py)
	orient := seg.OrientationIndex(p)
	if orient != expectedOrient {
		t.Errorf("expected %v, got %v", expectedOrient, orient)
	}
}

func checkOrientationIndexSegment(t *testing.T, seg *jts.Geom_LineSegment, s0x, s0y, s1x, s1y float64, expectedOrient int) {
	t.Helper()
	seg2 := jts.Geom_NewLineSegmentFromXY(s0x, s0y, s1x, s1y)
	orient := seg.OrientationIndexSegment(seg2)
	if orient != expectedOrient {
		t.Errorf("orientationIndex of %v and %v: expected %v, got %v", seg, seg2, expectedOrient, orient)
	}
}
