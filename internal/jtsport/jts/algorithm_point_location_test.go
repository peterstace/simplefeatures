package jts_test

import (
	"testing"

	"github.com/peterstace/simplefeatures/internal/jtsport/jts"
	"github.com/peterstace/simplefeatures/internal/jtsport/junit"
)

func TestPointLocation_IsOnSegment(t *testing.T) {
	tests := []struct {
		name     string
		p        *jts.Geom_Coordinate
		p0       *jts.Geom_Coordinate
		p1       *jts.Geom_Coordinate
		expected bool
	}{
		{
			name:     "midpoint on segment",
			p:        jts.Geom_NewCoordinateWithXY(5, 5),
			p0:       jts.Geom_NewCoordinateWithXY(0, 0),
			p1:       jts.Geom_NewCoordinateWithXY(9, 9),
			expected: true,
		},
		{
			name:     "start point",
			p:        jts.Geom_NewCoordinateWithXY(0, 0),
			p0:       jts.Geom_NewCoordinateWithXY(0, 0),
			p1:       jts.Geom_NewCoordinateWithXY(9, 9),
			expected: true,
		},
		{
			name:     "end point",
			p:        jts.Geom_NewCoordinateWithXY(9, 9),
			p0:       jts.Geom_NewCoordinateWithXY(0, 0),
			p1:       jts.Geom_NewCoordinateWithXY(9, 9),
			expected: true,
		},
		{
			name:     "not on segment - off line",
			p:        jts.Geom_NewCoordinateWithXY(5, 6),
			p0:       jts.Geom_NewCoordinateWithXY(0, 0),
			p1:       jts.Geom_NewCoordinateWithXY(9, 9),
			expected: false,
		},
		{
			name:     "not on segment - beyond endpoint",
			p:        jts.Geom_NewCoordinateWithXY(10, 10),
			p0:       jts.Geom_NewCoordinateWithXY(0, 0),
			p1:       jts.Geom_NewCoordinateWithXY(9, 9),
			expected: false,
		},
		{
			name:     "not on segment - barely off",
			p:        jts.Geom_NewCoordinateWithXY(9, 9.00001),
			p0:       jts.Geom_NewCoordinateWithXY(0, 0),
			p1:       jts.Geom_NewCoordinateWithXY(9, 9),
			expected: false,
		},
		{
			name:     "zero length segment - point on",
			p:        jts.Geom_NewCoordinateWithXY(1, 1),
			p0:       jts.Geom_NewCoordinateWithXY(1, 1),
			p1:       jts.Geom_NewCoordinateWithXY(1, 1),
			expected: true,
		},
		{
			name:     "zero length segment - point off",
			p:        jts.Geom_NewCoordinateWithXY(1, 2),
			p0:       jts.Geom_NewCoordinateWithXY(1, 1),
			p1:       jts.Geom_NewCoordinateWithXY(1, 1),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := jts.Algorithm_PointLocation_IsOnSegment(tt.p, tt.p0, tt.p1)
			junit.AssertEquals(t, tt.expected, result)
		})
	}
}

func TestPointLocation_IsOnLine(t *testing.T) {
	// LINESTRING (0 0, 20 20, 30 30)
	line := []*jts.Geom_Coordinate{
		jts.Geom_NewCoordinateWithXY(0, 0),
		jts.Geom_NewCoordinateWithXY(20, 20),
		jts.Geom_NewCoordinateWithXY(30, 30),
	}

	tests := []struct {
		name     string
		p        *jts.Geom_Coordinate
		expected bool
	}{
		{"on vertex", jts.Geom_NewCoordinateWithXY(20, 20), true},
		{"in segment 1", jts.Geom_NewCoordinateWithXY(10, 10), true},
		{"not on line", jts.Geom_NewCoordinateWithXY(0, 100), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := jts.Algorithm_PointLocation_IsOnLine(tt.p, line)
			junit.AssertEquals(t, tt.expected, result)
		})
	}
}

func TestPointLocation_IsOnLineSeq(t *testing.T) {
	factory := jts.Geom_NewGeometryFactoryDefault()
	csFactory := factory.GetCoordinateSequenceFactory()

	// LINESTRING (0 0, 20 20, 0 40)
	coords := []*jts.Geom_Coordinate{
		jts.Geom_NewCoordinateWithXY(0, 0),
		jts.Geom_NewCoordinateWithXY(20, 20),
		jts.Geom_NewCoordinateWithXY(0, 40),
	}
	cs := csFactory.CreateFromCoordinates(coords)

	tests := []struct {
		name     string
		p        *jts.Geom_Coordinate
		expected bool
	}{
		{"in first segment", jts.Geom_NewCoordinateWithXY(10, 10), true},
		{"in second segment", jts.Geom_NewCoordinateWithXY(10, 30), true},
		{"not on line", jts.Geom_NewCoordinateWithXY(0, 100), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := jts.Algorithm_PointLocation_IsOnLineSeq(tt.p, cs)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestPointLocation_IsInRing(t *testing.T) {
	// POLYGON ((0 0, 0 20, 20 20, 20 0, 0 0))
	ring := []*jts.Geom_Coordinate{
		jts.Geom_NewCoordinateWithXY(0, 0),
		jts.Geom_NewCoordinateWithXY(0, 20),
		jts.Geom_NewCoordinateWithXY(20, 20),
		jts.Geom_NewCoordinateWithXY(20, 0),
		jts.Geom_NewCoordinateWithXY(0, 0),
	}

	tests := []struct {
		name     string
		p        *jts.Geom_Coordinate
		expected bool
	}{
		{"interior", jts.Geom_NewCoordinateWithXY(10, 10), true},
		{"on boundary", jts.Geom_NewCoordinateWithXY(0, 10), true},
		{"exterior", jts.Geom_NewCoordinateWithXY(30, 10), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := jts.Algorithm_PointLocation_IsInRing(tt.p, ring)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestPointLocation_LocateInRing(t *testing.T) {
	// POLYGON ((0 0, 0 20, 20 20, 20 0, 0 0))
	ring := []*jts.Geom_Coordinate{
		jts.Geom_NewCoordinateWithXY(0, 0),
		jts.Geom_NewCoordinateWithXY(0, 20),
		jts.Geom_NewCoordinateWithXY(20, 20),
		jts.Geom_NewCoordinateWithXY(20, 0),
		jts.Geom_NewCoordinateWithXY(0, 0),
	}

	tests := []struct {
		name     string
		p        *jts.Geom_Coordinate
		expected int
	}{
		{"interior", jts.Geom_NewCoordinateWithXY(10, 10), jts.Geom_Location_Interior},
		{"on boundary", jts.Geom_NewCoordinateWithXY(0, 10), jts.Geom_Location_Boundary},
		{"exterior", jts.Geom_NewCoordinateWithXY(30, 10), jts.Geom_Location_Exterior},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := jts.Algorithm_PointLocation_LocateInRing(tt.p, ring)
			if result != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, result)
			}
		})
	}
}
