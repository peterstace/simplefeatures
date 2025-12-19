package jts_test

import (
	"testing"

	"github.com/peterstace/simplefeatures/internal/jtsport/jts"
)

const intersectionMaxAbsError = 1e-5

func TestIntersectionSimple(t *testing.T) {
	checkIntersection(t,
		0, 0, 10, 10,
		0, 10, 10, 0,
		5, 5)
}

func TestIntersectionCollinear(t *testing.T) {
	checkIntersectionNull(t,
		0, 0, 10, 10,
		20, 20, 30, 30)
}

func TestIntersectionParallel(t *testing.T) {
	checkIntersectionNull(t,
		0, 0, 10, 10,
		10, 0, 20, 10)
}

// See JTS GitHub issue #464.
func TestIntersectionAlmostCollinear(t *testing.T) {
	checkIntersection(t,
		35613471.6165017, 4257145.306132293, 35613477.7705378, 4257160.528222711,
		35613477.77505724, 4257160.539653536, 35613479.85607389, 4257165.92369170,
		35613477.772841461, 4257160.5339209242)
}

// Same as above but conditioned manually.
func TestIntersectionAlmostCollinearCond(t *testing.T) {
	checkIntersection(t,
		1.6165017, 45.306132293, 7.7705378, 60.528222711,
		7.77505724, 60.539653536, 9.85607389, 65.92369170,
		7.772841461, 60.5339209242)
}

func TestIntersectionLineSegCross(t *testing.T) {
	checkIntersectionLineSegment(t, 0, 0, 0, 1, -1, 9, 1, 9, 0, 9)
	checkIntersectionLineSegment(t, 0, 0, 0, 1, -1, 2, 1, 4, 0, 3)
}

func TestIntersectionLineSegTouch(t *testing.T) {
	checkIntersectionLineSegment(t, 0, 0, 0, 1, -1, 9, 0, 9, 0, 9)
	checkIntersectionLineSegment(t, 0, 0, 0, 1, 0, 2, 1, 4, 0, 2)
}

func TestIntersectionLineSegCollinear(t *testing.T) {
	checkIntersectionLineSegment(t, 0, 0, 0, 1, 0, 9, 0, 8, 0, 9)
}

func TestIntersectionLineSegNone(t *testing.T) {
	checkIntersectionLineSegmentNull(t, 0, 0, 0, 1, 2, 9, 1, 9)
	checkIntersectionLineSegmentNull(t, 0, 0, 0, 1, -2, 9, -1, 9)
	checkIntersectionLineSegmentNull(t, 0, 0, 0, 1, 2, 9, 1, 9)
}

func TestIntersectionXY(t *testing.T) {
	reader := jts.Io_NewWKTReader()

	// Intersection with dim 3 x dim3.
	poly1, err := reader.Read("POLYGON((0 0 0, 0 10000 2, 10000 10000 2, 10000 0 0, 0 0 0))")
	if err != nil {
		t.Fatalf("failed to read poly1: %v", err)
	}
	clipArea, err := reader.Read("POLYGON((0 0, 0 2500, 2500 2500, 2500 0, 0 0))")
	if err != nil {
		t.Fatalf("failed to read clipArea: %v", err)
	}
	clipped1 := poly1.Intersection(clipArea)

	// Intersection with dim 3 x dim 2.
	gf := poly1.GetFactory()
	csf := gf.GetCoordinateSequenceFactory()
	xmin := 0.0
	xmax := 2500.0
	ymin := 0.0
	ymax := 2500.0

	cs := csf.CreateWithSizeAndDimension(5, 2)
	cs.SetOrdinate(0, 0, xmin)
	cs.SetOrdinate(0, 1, ymin)
	cs.SetOrdinate(1, 0, xmin)
	cs.SetOrdinate(1, 1, ymax)
	cs.SetOrdinate(2, 0, xmax)
	cs.SetOrdinate(2, 1, ymax)
	cs.SetOrdinate(3, 0, xmax)
	cs.SetOrdinate(3, 1, ymin)
	cs.SetOrdinate(4, 0, xmin)
	cs.SetOrdinate(4, 1, ymin)

	bounds := gf.CreateLinearRingFromCoordinateSequence(cs)
	fence := gf.CreatePolygonFromLinearRing(bounds)
	clipped2 := poly1.Intersection(fence.Geom_Geometry)

	// Use EqualsExact since EqualsTopo depends on Relate operations not yet ported.
	// The results should be coordinatewise equal for this test case.
	if !clipped1.EqualsExact(clipped2) {
		t.Errorf("clipped1 and clipped2 should be equal.\nclipped1: %v\nclipped2: %v", clipped1, clipped2)
	}
}

func checkIntersection(t *testing.T, p1x, p1y, p2x, p2y, q1x, q1y, q2x, q2y, expectedx, expectedy float64) {
	t.Helper()
	p1 := jts.Geom_NewCoordinateWithXY(p1x, p1y)
	p2 := jts.Geom_NewCoordinateWithXY(p2x, p2y)
	q1 := jts.Geom_NewCoordinateWithXY(q1x, q1y)
	q2 := jts.Geom_NewCoordinateWithXY(q2x, q2y)
	actual := jts.Algorithm_Intersection_Intersection(p1, p2, q1, q2)
	expected := jts.Geom_NewCoordinateWithXY(expectedx, expectedy)
	dist := actual.Distance(expected)
	if dist > intersectionMaxAbsError {
		t.Errorf("expected %v, got %v (dist=%v)", expected, actual, dist)
	}
}

func checkIntersectionNull(t *testing.T, p1x, p1y, p2x, p2y, q1x, q1y, q2x, q2y float64) {
	t.Helper()
	p1 := jts.Geom_NewCoordinateWithXY(p1x, p1y)
	p2 := jts.Geom_NewCoordinateWithXY(p2x, p2y)
	q1 := jts.Geom_NewCoordinateWithXY(q1x, q1y)
	q2 := jts.Geom_NewCoordinateWithXY(q2x, q2y)
	actual := jts.Algorithm_Intersection_Intersection(p1, p2, q1, q2)
	if actual != nil {
		t.Errorf("expected nil, got %v", actual)
	}
}

func checkIntersectionLineSegment(t *testing.T, p1x, p1y, p2x, p2y, q1x, q1y, q2x, q2y, expectedx, expectedy float64) {
	t.Helper()
	p1 := jts.Geom_NewCoordinateWithXY(p1x, p1y)
	p2 := jts.Geom_NewCoordinateWithXY(p2x, p2y)
	q1 := jts.Geom_NewCoordinateWithXY(q1x, q1y)
	q2 := jts.Geom_NewCoordinateWithXY(q2x, q2y)
	actual := jts.Algorithm_Intersection_LineSegment(p1, p2, q1, q2)
	expected := jts.Geom_NewCoordinateWithXY(expectedx, expectedy)
	dist := actual.Distance(expected)
	if dist > intersectionMaxAbsError {
		t.Errorf("expected %v, got %v (dist=%v)", expected, actual, dist)
	}
}

func checkIntersectionLineSegmentNull(t *testing.T, p1x, p1y, p2x, p2y, q1x, q1y, q2x, q2y float64) {
	t.Helper()
	p1 := jts.Geom_NewCoordinateWithXY(p1x, p1y)
	p2 := jts.Geom_NewCoordinateWithXY(p2x, p2y)
	q1 := jts.Geom_NewCoordinateWithXY(q1x, q1y)
	q2 := jts.Geom_NewCoordinateWithXY(q2x, q2y)
	actual := jts.Algorithm_Intersection_LineSegment(p1, p2, q1, q2)
	if actual != nil {
		t.Errorf("expected nil, got %v", actual)
	}
}
