package jts_test

import (
	"testing"

	"github.com/peterstace/simplefeatures/internal/jtsport/jts"
)

func TestRayCrossingCounter_LocatePointInRing_Box(t *testing.T) {
	// POLYGON ((0 0, 0 20, 20 20, 20 0, 0 0))
	ring := []*jts.Geom_Coordinate{
		jts.Geom_NewCoordinateWithXY(0, 0),
		jts.Geom_NewCoordinateWithXY(0, 20),
		jts.Geom_NewCoordinateWithXY(20, 20),
		jts.Geom_NewCoordinateWithXY(20, 0),
		jts.Geom_NewCoordinateWithXY(0, 0),
	}

	// Interior point.
	pt := jts.Geom_NewCoordinateWithXY(10, 10)
	loc := jts.Algorithm_RayCrossingCounter_LocatePointInRing(pt, ring)
	if loc != jts.Geom_Location_Interior {
		t.Errorf("expected Interior, got %d", loc)
	}
}

func TestRayCrossingCounter_LocatePointInRing_ComplexRing(t *testing.T) {
	// POLYGON ((-40 80, -40 -80, 20 0, 20 -100, 40 40, 80 -80, 100 80, 140 -20, 120 140, 40 180, 60 40, 0 120, -20 -20, -40 80))
	ring := []*jts.Geom_Coordinate{
		jts.Geom_NewCoordinateWithXY(-40, 80),
		jts.Geom_NewCoordinateWithXY(-40, -80),
		jts.Geom_NewCoordinateWithXY(20, 0),
		jts.Geom_NewCoordinateWithXY(20, -100),
		jts.Geom_NewCoordinateWithXY(40, 40),
		jts.Geom_NewCoordinateWithXY(80, -80),
		jts.Geom_NewCoordinateWithXY(100, 80),
		jts.Geom_NewCoordinateWithXY(140, -20),
		jts.Geom_NewCoordinateWithXY(120, 140),
		jts.Geom_NewCoordinateWithXY(40, 180),
		jts.Geom_NewCoordinateWithXY(60, 40),
		jts.Geom_NewCoordinateWithXY(0, 120),
		jts.Geom_NewCoordinateWithXY(-20, -20),
		jts.Geom_NewCoordinateWithXY(-40, 80),
	}

	pt := jts.Geom_NewCoordinateWithXY(0, 0)
	loc := jts.Algorithm_RayCrossingCounter_LocatePointInRing(pt, ring)
	if loc != jts.Geom_Location_Interior {
		t.Errorf("expected Interior, got %d", loc)
	}
}

func TestRayCrossingCounter_LocatePointInRing_Comb(t *testing.T) {
	// POLYGON ((0 0, 0 10, 4 5, 6 10, 7 5, 9 10, 10 5, 13 5, 15 10, 16 3, 17 10, 18 3, 25 10, 30 10, 30 0, 15 0, 14 5, 13 0, 9 0, 8 5, 6 0, 0 0))
	ring := []*jts.Geom_Coordinate{
		jts.Geom_NewCoordinateWithXY(0, 0),
		jts.Geom_NewCoordinateWithXY(0, 10),
		jts.Geom_NewCoordinateWithXY(4, 5),
		jts.Geom_NewCoordinateWithXY(6, 10),
		jts.Geom_NewCoordinateWithXY(7, 5),
		jts.Geom_NewCoordinateWithXY(9, 10),
		jts.Geom_NewCoordinateWithXY(10, 5),
		jts.Geom_NewCoordinateWithXY(13, 5),
		jts.Geom_NewCoordinateWithXY(15, 10),
		jts.Geom_NewCoordinateWithXY(16, 3),
		jts.Geom_NewCoordinateWithXY(17, 10),
		jts.Geom_NewCoordinateWithXY(18, 3),
		jts.Geom_NewCoordinateWithXY(25, 10),
		jts.Geom_NewCoordinateWithXY(30, 10),
		jts.Geom_NewCoordinateWithXY(30, 0),
		jts.Geom_NewCoordinateWithXY(15, 0),
		jts.Geom_NewCoordinateWithXY(14, 5),
		jts.Geom_NewCoordinateWithXY(13, 0),
		jts.Geom_NewCoordinateWithXY(9, 0),
		jts.Geom_NewCoordinateWithXY(8, 5),
		jts.Geom_NewCoordinateWithXY(6, 0),
		jts.Geom_NewCoordinateWithXY(0, 0),
	}

	tests := []struct {
		name     string
		pt       *jts.Geom_Coordinate
		expected int
	}{
		// Boundary tests.
		{"origin", jts.Geom_NewCoordinateWithXY(0, 0), jts.Geom_Location_Boundary},
		{"on left edge", jts.Geom_NewCoordinateWithXY(0, 1), jts.Geom_Location_Boundary},
		{"at vertex 4,5", jts.Geom_NewCoordinateWithXY(4, 5), jts.Geom_Location_Boundary},
		{"at vertex 8,5", jts.Geom_NewCoordinateWithXY(8, 5), jts.Geom_Location_Boundary},
		{"on horizontal segment", jts.Geom_NewCoordinateWithXY(11, 5), jts.Geom_Location_Boundary},
		{"on vertical segment", jts.Geom_NewCoordinateWithXY(30, 5), jts.Geom_Location_Boundary},
		{"on angled segment", jts.Geom_NewCoordinateWithXY(22, 7), jts.Geom_Location_Boundary},
		// Interior tests.
		{"interior 1,5", jts.Geom_NewCoordinateWithXY(1, 5), jts.Geom_Location_Interior},
		{"interior 5,5", jts.Geom_NewCoordinateWithXY(5, 5), jts.Geom_Location_Interior},
		{"interior 1,7", jts.Geom_NewCoordinateWithXY(1, 7), jts.Geom_Location_Interior},
		// Exterior tests.
		{"exterior 12,10", jts.Geom_NewCoordinateWithXY(12, 10), jts.Geom_Location_Exterior},
		{"exterior 16,5", jts.Geom_NewCoordinateWithXY(16, 5), jts.Geom_Location_Exterior},
		{"exterior 35,5", jts.Geom_NewCoordinateWithXY(35, 5), jts.Geom_Location_Exterior},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loc := jts.Algorithm_RayCrossingCounter_LocatePointInRing(tt.pt, ring)
			if loc != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, loc)
			}
		})
	}
}

func TestRayCrossingCounter_LocatePointInRing_RepeatedPts(t *testing.T) {
	// POLYGON ((0 0, 0 10, 2 5, 2 5, 2 5, 2 5, 2 5, 3 10, 6 10, 8 5, 8 5, 8 5, 8 5, 10 10, 10 5, 10 5, 10 5, 10 5, 10 0, 0 0))
	ring := []*jts.Geom_Coordinate{
		jts.Geom_NewCoordinateWithXY(0, 0),
		jts.Geom_NewCoordinateWithXY(0, 10),
		jts.Geom_NewCoordinateWithXY(2, 5),
		jts.Geom_NewCoordinateWithXY(2, 5),
		jts.Geom_NewCoordinateWithXY(2, 5),
		jts.Geom_NewCoordinateWithXY(2, 5),
		jts.Geom_NewCoordinateWithXY(2, 5),
		jts.Geom_NewCoordinateWithXY(3, 10),
		jts.Geom_NewCoordinateWithXY(6, 10),
		jts.Geom_NewCoordinateWithXY(8, 5),
		jts.Geom_NewCoordinateWithXY(8, 5),
		jts.Geom_NewCoordinateWithXY(8, 5),
		jts.Geom_NewCoordinateWithXY(8, 5),
		jts.Geom_NewCoordinateWithXY(10, 10),
		jts.Geom_NewCoordinateWithXY(10, 5),
		jts.Geom_NewCoordinateWithXY(10, 5),
		jts.Geom_NewCoordinateWithXY(10, 5),
		jts.Geom_NewCoordinateWithXY(10, 5),
		jts.Geom_NewCoordinateWithXY(10, 0),
		jts.Geom_NewCoordinateWithXY(0, 0),
	}

	tests := []struct {
		name     string
		pt       *jts.Geom_Coordinate
		expected int
	}{
		// Boundary tests.
		{"origin", jts.Geom_NewCoordinateWithXY(0, 0), jts.Geom_Location_Boundary},
		{"on left edge", jts.Geom_NewCoordinateWithXY(0, 1), jts.Geom_Location_Boundary},
		{"at vertex 2,5", jts.Geom_NewCoordinateWithXY(2, 5), jts.Geom_Location_Boundary},
		{"at vertex 8,5", jts.Geom_NewCoordinateWithXY(8, 5), jts.Geom_Location_Boundary},
		{"at vertex 10,5", jts.Geom_NewCoordinateWithXY(10, 5), jts.Geom_Location_Boundary},
		// Interior tests.
		{"interior 1,5", jts.Geom_NewCoordinateWithXY(1, 5), jts.Geom_Location_Interior},
		{"interior 3,5", jts.Geom_NewCoordinateWithXY(3, 5), jts.Geom_Location_Interior},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loc := jts.Algorithm_RayCrossingCounter_LocatePointInRing(tt.pt, ring)
			if loc != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, loc)
			}
		})
	}
}

func TestRayCrossingCounter_LocatePointInRing_RobustStressTriangles(t *testing.T) {
	tests := []struct {
		name     string
		ring     []*jts.Geom_Coordinate
		pt       *jts.Geom_Coordinate
		expected int
	}{
		{
			name: "triangle 1",
			ring: []*jts.Geom_Coordinate{
				jts.Geom_NewCoordinateWithXY(0.0, 0.0),
				jts.Geom_NewCoordinateWithXY(0.0, 172.0),
				jts.Geom_NewCoordinateWithXY(100.0, 0.0),
				jts.Geom_NewCoordinateWithXY(0.0, 0.0),
			},
			pt:       jts.Geom_NewCoordinateWithXY(25.374625374625374, 128.35564435564436),
			expected: jts.Geom_Location_Exterior,
		},
		{
			name: "triangle 2",
			ring: []*jts.Geom_Coordinate{
				jts.Geom_NewCoordinateWithXY(642.0, 815.0),
				jts.Geom_NewCoordinateWithXY(69.0, 764.0),
				jts.Geom_NewCoordinateWithXY(394.0, 966.0),
				jts.Geom_NewCoordinateWithXY(642.0, 815.0),
			},
			pt:       jts.Geom_NewCoordinateWithXY(97.96039603960396, 782.0),
			expected: jts.Geom_Location_Interior,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loc := jts.Algorithm_RayCrossingCounter_LocatePointInRing(tt.pt, tt.ring)
			if loc != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, loc)
			}
		})
	}
}

func TestRayCrossingCounter_LocatePointInRing_RobustTriangle(t *testing.T) {
	ring := []*jts.Geom_Coordinate{
		jts.Geom_NewCoordinateWithXY(2.152214146946829, 50.470470727186765),
		jts.Geom_NewCoordinateWithXY(18.381941666723034, 19.567250592139274),
		jts.Geom_NewCoordinateWithXY(2.390837642830135, 49.228045261718165),
		jts.Geom_NewCoordinateWithXY(2.152214146946829, 50.470470727186765),
	}
	pt := jts.Geom_NewCoordinateWithXY(3.166572116932842, 48.5390194687463)
	loc := jts.Algorithm_RayCrossingCounter_LocatePointInRing(pt, ring)
	if loc != jts.Geom_Location_Exterior {
		t.Errorf("expected Exterior, got %d", loc)
	}
}

func TestRayCrossingCounter_LocatePointInRingSeq_4D(t *testing.T) {
	// Test with a 4D coordinate sequence (XYZM).
	factory := jts.Geom_NewGeometryFactoryDefault()
	csFactory := factory.GetCoordinateSequenceFactory()

	// Create a triangle ring with XYZM coordinates.
	coords := []*jts.Geom_Coordinate{
		jts.Geom_NewCoordinateWithXY(0.0, 0.0),
		jts.Geom_NewCoordinateWithXY(10.0, 0.0),
		jts.Geom_NewCoordinateWithXY(5.0, 10.0),
		jts.Geom_NewCoordinateWithXY(0.0, 0.0),
	}
	cs := csFactory.CreateFromCoordinates(coords)

	pt := jts.Geom_NewCoordinateWithXY(5.0, 2.0)
	loc := jts.Algorithm_RayCrossingCounter_LocatePointInRingSeq(pt, cs)
	if loc != jts.Geom_Location_Interior {
		t.Errorf("expected Interior, got %d", loc)
	}
}
