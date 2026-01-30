package jts_test

import (
	"testing"

	"github.com/peterstace/simplefeatures/internal/jtsport/jts"
)

func TestRobustLineIntersector2Lines(t *testing.T) {
	li := jts.Algorithm_NewRobustLineIntersector()
	p1 := jts.Geom_NewCoordinateWithXY(10, 10)
	p2 := jts.Geom_NewCoordinateWithXY(20, 20)
	q1 := jts.Geom_NewCoordinateWithXY(20, 10)
	q2 := jts.Geom_NewCoordinateWithXY(10, 20)
	x := jts.Geom_NewCoordinateWithXY(15, 15)
	li.ComputeIntersection(p1, p2, q1, q2)
	if li.GetIntersectionNum() != jts.Algorithm_LineIntersector_PointIntersection {
		t.Errorf("expected POINT_INTERSECTION, got %d", li.GetIntersectionNum())
	}
	if li.GetIntersectionNum() != 1 {
		t.Errorf("expected 1 intersection, got %d", li.GetIntersectionNum())
	}
	if !x.Equals(li.GetIntersection(0)) {
		t.Errorf("expected %v, got %v", x, li.GetIntersection(0))
	}
	if !li.IsProper() {
		t.Error("expected proper intersection")
	}
	if !li.HasIntersection() {
		t.Error("expected hasIntersection to be true")
	}
}

func TestRobustLineIntersectorCollinear1(t *testing.T) {
	li := jts.Algorithm_NewRobustLineIntersector()
	p1 := jts.Geom_NewCoordinateWithXY(10, 10)
	p2 := jts.Geom_NewCoordinateWithXY(20, 10)
	q1 := jts.Geom_NewCoordinateWithXY(22, 10)
	q2 := jts.Geom_NewCoordinateWithXY(30, 10)
	li.ComputeIntersection(p1, p2, q1, q2)
	if li.GetIntersectionNum() != jts.Algorithm_LineIntersector_NoIntersection {
		t.Errorf("expected NO_INTERSECTION, got %d", li.GetIntersectionNum())
	}
	if li.IsProper() {
		t.Error("expected not proper intersection")
	}
	if li.HasIntersection() {
		t.Error("expected hasIntersection to be false")
	}
}

func TestRobustLineIntersectorCollinear2(t *testing.T) {
	li := jts.Algorithm_NewRobustLineIntersector()
	p1 := jts.Geom_NewCoordinateWithXY(10, 10)
	p2 := jts.Geom_NewCoordinateWithXY(20, 10)
	q1 := jts.Geom_NewCoordinateWithXY(20, 10)
	q2 := jts.Geom_NewCoordinateWithXY(30, 10)
	li.ComputeIntersection(p1, p2, q1, q2)
	if li.GetIntersectionNum() != jts.Algorithm_LineIntersector_PointIntersection {
		t.Errorf("expected POINT_INTERSECTION, got %d", li.GetIntersectionNum())
	}
	if li.IsProper() {
		t.Error("expected not proper intersection")
	}
	if !li.HasIntersection() {
		t.Error("expected hasIntersection to be true")
	}
}

func TestRobustLineIntersectorCollinear3(t *testing.T) {
	li := jts.Algorithm_NewRobustLineIntersector()
	p1 := jts.Geom_NewCoordinateWithXY(10, 10)
	p2 := jts.Geom_NewCoordinateWithXY(20, 10)
	q1 := jts.Geom_NewCoordinateWithXY(15, 10)
	q2 := jts.Geom_NewCoordinateWithXY(30, 10)
	li.ComputeIntersection(p1, p2, q1, q2)
	if li.GetIntersectionNum() != jts.Algorithm_LineIntersector_CollinearIntersection {
		t.Errorf("expected COLLINEAR_INTERSECTION, got %d", li.GetIntersectionNum())
	}
	if li.IsProper() {
		t.Error("expected not proper intersection")
	}
	if !li.HasIntersection() {
		t.Error("expected hasIntersection to be true")
	}
}

func TestRobustLineIntersectorCollinear4(t *testing.T) {
	li := jts.Algorithm_NewRobustLineIntersector()
	p1 := jts.Geom_NewCoordinateWithXY(30, 10)
	p2 := jts.Geom_NewCoordinateWithXY(20, 10)
	q1 := jts.Geom_NewCoordinateWithXY(10, 10)
	q2 := jts.Geom_NewCoordinateWithXY(30, 10)
	li.ComputeIntersection(p1, p2, q1, q2)
	if li.GetIntersectionNum() != jts.Algorithm_LineIntersector_CollinearIntersection {
		t.Errorf("expected COLLINEAR_INTERSECTION, got %d", li.GetIntersectionNum())
	}
	if !li.HasIntersection() {
		t.Error("expected hasIntersection to be true")
	}
}

func TestRobustLineIntersectorEndpointIntersection(t *testing.T) {
	li := jts.Algorithm_NewRobustLineIntersector()
	li.ComputeIntersection(
		jts.Geom_NewCoordinateWithXY(100, 100),
		jts.Geom_NewCoordinateWithXY(10, 100),
		jts.Geom_NewCoordinateWithXY(100, 10),
		jts.Geom_NewCoordinateWithXY(100, 100))
	if !li.HasIntersection() {
		t.Error("expected hasIntersection to be true")
	}
	if li.GetIntersectionNum() != 1 {
		t.Errorf("expected 1 intersection, got %d", li.GetIntersectionNum())
	}
}

func TestRobustLineIntersectorEndpointIntersection2(t *testing.T) {
	li := jts.Algorithm_NewRobustLineIntersector()
	li.ComputeIntersection(
		jts.Geom_NewCoordinateWithXY(190, 50),
		jts.Geom_NewCoordinateWithXY(120, 100),
		jts.Geom_NewCoordinateWithXY(120, 100),
		jts.Geom_NewCoordinateWithXY(50, 150))
	if !li.HasIntersection() {
		t.Error("expected hasIntersection to be true")
	}
	if li.GetIntersectionNum() != 1 {
		t.Errorf("expected 1 intersection, got %d", li.GetIntersectionNum())
	}
	expected := jts.Geom_NewCoordinateWithXY(120, 100)
	if !expected.Equals(li.GetIntersection(1)) {
		t.Errorf("expected %v, got %v", expected, li.GetIntersection(1))
	}
}

func TestRobustLineIntersectorOverlap(t *testing.T) {
	li := jts.Algorithm_NewRobustLineIntersector()
	li.ComputeIntersection(
		jts.Geom_NewCoordinateWithXY(180, 200),
		jts.Geom_NewCoordinateWithXY(160, 180),
		jts.Geom_NewCoordinateWithXY(220, 240),
		jts.Geom_NewCoordinateWithXY(140, 160))
	if !li.HasIntersection() {
		t.Error("expected hasIntersection to be true")
	}
	if li.GetIntersectionNum() != 2 {
		t.Errorf("expected 2 intersections, got %d", li.GetIntersectionNum())
	}
}

func TestRobustLineIntersectorIsProper1(t *testing.T) {
	li := jts.Algorithm_NewRobustLineIntersector()
	li.ComputeIntersection(
		jts.Geom_NewCoordinateWithXY(30, 10),
		jts.Geom_NewCoordinateWithXY(30, 30),
		jts.Geom_NewCoordinateWithXY(10, 10),
		jts.Geom_NewCoordinateWithXY(90, 11))
	if !li.HasIntersection() {
		t.Error("expected hasIntersection to be true")
	}
	if li.GetIntersectionNum() != 1 {
		t.Errorf("expected 1 intersection, got %d", li.GetIntersectionNum())
	}
	if !li.IsProper() {
		t.Error("expected proper intersection")
	}
}

func TestRobustLineIntersectorIsProper2(t *testing.T) {
	li := jts.Algorithm_NewRobustLineIntersector()
	li.ComputeIntersection(
		jts.Geom_NewCoordinateWithXY(10, 30),
		jts.Geom_NewCoordinateWithXY(10, 0),
		jts.Geom_NewCoordinateWithXY(11, 90),
		jts.Geom_NewCoordinateWithXY(10, 10))
	if !li.HasIntersection() {
		t.Error("expected hasIntersection to be true")
	}
	if li.GetIntersectionNum() != 1 {
		t.Errorf("expected 1 intersection, got %d", li.GetIntersectionNum())
	}
	if li.IsProper() {
		t.Error("expected not proper intersection")
	}
}

func TestRobustLineIntersectorIsCCW(t *testing.T) {
	result := jts.Algorithm_Orientation_Index(
		jts.Geom_NewCoordinateWithXY(-123456789, -40),
		jts.Geom_NewCoordinateWithXY(0, 0),
		jts.Geom_NewCoordinateWithXY(381039468754763, 123456789))
	if result != 1 {
		t.Errorf("expected 1, got %d", result)
	}
}

func TestRobustLineIntersectorIsCCW2(t *testing.T) {
	result := jts.Algorithm_Orientation_Index(
		jts.Geom_NewCoordinateWithXY(10, 10),
		jts.Geom_NewCoordinateWithXY(20, 20),
		jts.Geom_NewCoordinateWithXY(0, 0))
	if result != 0 {
		t.Errorf("expected 0, got %d", result)
	}
}

func TestRobustLineIntersectorA(t *testing.T) {
	p1 := jts.Geom_NewCoordinateWithXY(-123456789, -40)
	p2 := jts.Geom_NewCoordinateWithXY(381039468754763, 123456789)
	q := jts.Geom_NewCoordinateWithXY(0, 0)

	factory := jts.Geom_NewGeometryFactoryDefault()
	l := factory.CreateLineStringFromCoordinates([]*jts.Geom_Coordinate{p1, p2})
	p := factory.CreatePointFromCoordinate(q)

	// Line should NOT intersect point.
	if l.Intersects(p.Geom_Geometry) {
		t.Error("expected line to not intersect point")
	}

	// PointLocation.isOnLine should return false.
	if jts.Algorithm_PointLocation_IsOnLine(q, []*jts.Geom_Coordinate{p1, p2}) {
		t.Error("expected PointLocation.isOnLine to return false")
	}

	// Orientation.index should return -1 (clockwise).
	if jts.Algorithm_Orientation_Index(p1, p2, q) != -1 {
		t.Errorf("expected Orientation.index to return -1, got %d", jts.Algorithm_Orientation_Index(p1, p2, q))
	}
}
