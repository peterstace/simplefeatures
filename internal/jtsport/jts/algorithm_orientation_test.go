package jts

import "testing"

func TestOrientationIndexRobust(t *testing.T) {
	p0 := Geom_NewCoordinateWithXY(219.3649559090992, 140.84159161824724)
	p1 := Geom_NewCoordinateWithXY(168.9018919682399, -5.713787599646864)
	p := Geom_NewCoordinateWithXY(186.80814046338352, 46.28973405831556)
	orient := Algorithm_Orientation_Index(p0, p1, p)
	orientInv := Algorithm_Orientation_Index(p1, p0, p)
	if orient == orientInv {
		t.Errorf("Expected orient != orientInv, but got orient=%d, orientInv=%d", orient, orientInv)
	}
}

func TestOrientationCCW(t *testing.T) {
	pts := []*Geom_Coordinate{
		Geom_NewCoordinateWithXY(0, 0),
		Geom_NewCoordinateWithXY(0, 1),
		Geom_NewCoordinateWithXY(1, 1),
	}
	if !isAllOrientationsEqual(pts) {
		t.Errorf("Expected all orientations equal for CCW triangle")
	}
}

func TestOrientationCCW2(t *testing.T) {
	pts := []*Geom_Coordinate{
		Geom_NewCoordinateWithXY(1.0000000000004998, -7.989685402102996),
		Geom_NewCoordinateWithXY(10.0, -7.004368924503866),
		Geom_NewCoordinateWithXY(1.0000000000005, -7.989685402102996),
	}
	if !isAllOrientationsEqual(pts) {
		t.Errorf("Expected all orientations equal for CCW2 triangle")
	}
}

func TestOrientationIsCCWTooFewPoints(t *testing.T) {
	pts := []*Geom_Coordinate{
		Geom_NewCoordinateWithXY(0, 0),
		Geom_NewCoordinateWithXY(1, 1),
		Geom_NewCoordinateWithXY(2, 2),
	}
	Algorithm_Orientation_IsCCW(pts)
}

func TestOrientationIsCCWCCW(t *testing.T) {
	ring := []*Geom_Coordinate{
		Geom_NewCoordinateWithXY(60, 180),
		Geom_NewCoordinateWithXY(140, 120),
		Geom_NewCoordinateWithXY(100, 180),
		Geom_NewCoordinateWithXY(140, 240),
		Geom_NewCoordinateWithXY(60, 180),
	}
	if !Algorithm_Orientation_IsCCW(ring) {
		t.Errorf("Expected ring to be CCW")
	}
}

func TestOrientationIsCCWCW(t *testing.T) {
	ring := []*Geom_Coordinate{
		Geom_NewCoordinateWithXY(60, 180),
		Geom_NewCoordinateWithXY(140, 240),
		Geom_NewCoordinateWithXY(100, 180),
		Geom_NewCoordinateWithXY(140, 120),
		Geom_NewCoordinateWithXY(60, 180),
	}
	if Algorithm_Orientation_IsCCW(ring) {
		t.Errorf("Expected ring to be CW")
	}
}

func TestOrientationIsCCWFlatTopSegment(t *testing.T) {
	ring := []*Geom_Coordinate{
		Geom_NewCoordinateWithXY(100, 200),
		Geom_NewCoordinateWithXY(200, 200),
		Geom_NewCoordinateWithXY(200, 100),
		Geom_NewCoordinateWithXY(100, 100),
		Geom_NewCoordinateWithXY(100, 200),
	}
	if Algorithm_Orientation_IsCCW(ring) {
		t.Errorf("Expected ring to be CW")
	}
}

func TestOrientationIsCCWAreaBowTie(t *testing.T) {
	ring := []*Geom_Coordinate{
		Geom_NewCoordinateWithXY(10, 10),
		Geom_NewCoordinateWithXY(50, 10),
		Geom_NewCoordinateWithXY(25, 35),
		Geom_NewCoordinateWithXY(35, 35),
		Geom_NewCoordinateWithXY(10, 10),
	}
	if !Algorithm_Orientation_IsCCWArea(ring) {
		t.Errorf("Expected ring to be CCW by area")
	}
}

func isAllOrientationsEqual(pts []*Geom_Coordinate) bool {
	orient := make([]int, 3)
	orient[0] = Algorithm_Orientation_Index(pts[0], pts[1], pts[2])
	orient[1] = Algorithm_Orientation_Index(pts[1], pts[2], pts[0])
	orient[2] = Algorithm_Orientation_Index(pts[2], pts[0], pts[1])
	return orient[0] == orient[1] && orient[0] == orient[2]
}
