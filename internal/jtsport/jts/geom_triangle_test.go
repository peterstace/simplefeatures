package jts_test

import (
	"math"
	"testing"

	"github.com/peterstace/simplefeatures/internal/jtsport/jts"
	"github.com/peterstace/simplefeatures/internal/jtsport/junit"
)

const triangleTestTolerance = 1e-5

func TestTriangleInterpolateZ(t *testing.T) {
	tests := []struct {
		name     string
		v0       *jts.Geom_Coordinate
		v1       *jts.Geom_Coordinate
		v2       *jts.Geom_Coordinate
		p        *jts.Geom_Coordinate
		expected float64
	}{
		{
			// LINESTRING(1 1 0, 2 1 0, 1 2 10)
			name:     "midpoint interpolation",
			v0:       jts.Geom_NewCoordinateWithXYZ(1, 1, 0),
			v1:       jts.Geom_NewCoordinateWithXYZ(2, 1, 0),
			v2:       jts.Geom_NewCoordinateWithXYZ(1, 2, 10),
			p:        jts.Geom_NewCoordinateWithXY(1.5, 1.5),
			expected: 5,
		},
		{
			// LINESTRING(1 1 0, 2 1 0, 1 2 10)
			name:     "near vertex interpolation",
			v0:       jts.Geom_NewCoordinateWithXYZ(1, 1, 0),
			v1:       jts.Geom_NewCoordinateWithXYZ(2, 1, 0),
			v2:       jts.Geom_NewCoordinateWithXYZ(1, 2, 10),
			p:        jts.Geom_NewCoordinateWithXY(1.2, 1.2),
			expected: 2,
		},
		{
			// LINESTRING(1 1 0, 2 1 0, 1 2 10)
			name:     "exterior point interpolation",
			v0:       jts.Geom_NewCoordinateWithXYZ(1, 1, 0),
			v1:       jts.Geom_NewCoordinateWithXYZ(2, 1, 0),
			v2:       jts.Geom_NewCoordinateWithXYZ(1, 2, 10),
			p:        jts.Geom_NewCoordinateWithXY(0, 0),
			expected: -10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tri := jts.Geom_NewTriangle(tt.v0, tt.v1, tt.v2)
			z := tri.InterpolateZ(tt.p)
			junit.AssertEqualsFloat64(t, tt.expected, z, 0.000001)
		})
	}
}

func TestTriangleArea3D(t *testing.T) {
	tests := []struct {
		name     string
		v0       *jts.Geom_Coordinate
		v1       *jts.Geom_Coordinate
		v2       *jts.Geom_Coordinate
		expected float64
	}{
		{
			// POLYGON((0 0 10, 100 0 110, 100 100 110, 0 0 10))
			name:     "3D triangle 1",
			v0:       jts.Geom_NewCoordinateWithXYZ(0, 0, 10),
			v1:       jts.Geom_NewCoordinateWithXYZ(100, 0, 110),
			v2:       jts.Geom_NewCoordinateWithXYZ(100, 100, 110),
			expected: 7071.067811865475,
		},
		{
			// POLYGON((0 0 10, 100 0 10, 50 100 110, 0 0 10))
			name:     "3D triangle 2",
			v0:       jts.Geom_NewCoordinateWithXYZ(0, 0, 10),
			v1:       jts.Geom_NewCoordinateWithXYZ(100, 0, 10),
			v2:       jts.Geom_NewCoordinateWithXYZ(50, 100, 110),
			expected: 7071.067811865475,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tri := jts.Geom_NewTriangle(tt.v0, tt.v1, tt.v2)
			area3D := tri.Area3D()
			junit.AssertEqualsFloat64(t, tt.expected, area3D, triangleTestTolerance)
		})
	}
}

func TestTriangleArea(t *testing.T) {
	tests := []struct {
		name           string
		v0             *jts.Geom_Coordinate
		v1             *jts.Geom_Coordinate
		v2             *jts.Geom_Coordinate
		expectedSigned float64
	}{
		{
			// POLYGON((10 10, 20 20, 20 10, 10 10)) - CW
			name:           "CW triangle",
			v0:             jts.Geom_NewCoordinateWithXY(10, 10),
			v1:             jts.Geom_NewCoordinateWithXY(20, 20),
			v2:             jts.Geom_NewCoordinateWithXY(20, 10),
			expectedSigned: 50,
		},
		{
			// POLYGON((10 10, 20 10, 20 20, 10 10)) - CCW
			name:           "CCW triangle",
			v0:             jts.Geom_NewCoordinateWithXY(10, 10),
			v1:             jts.Geom_NewCoordinateWithXY(20, 10),
			v2:             jts.Geom_NewCoordinateWithXY(20, 20),
			expectedSigned: -50,
		},
		{
			// POLYGON((10 10, 10 10, 10 10, 10 10)) - degenerate point triangle
			name:           "degenerate point triangle",
			v0:             jts.Geom_NewCoordinateWithXY(10, 10),
			v1:             jts.Geom_NewCoordinateWithXY(10, 10),
			v2:             jts.Geom_NewCoordinateWithXY(10, 10),
			expectedSigned: 0,
		},
		{
			// POLYGON((10 10, 20 10, 15 10, 10 10)) - degenerate line triangle
			name:           "degenerate line triangle",
			v0:             jts.Geom_NewCoordinateWithXY(10, 10),
			v1:             jts.Geom_NewCoordinateWithXY(20, 10),
			v2:             jts.Geom_NewCoordinateWithXY(15, 10),
			expectedSigned: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tri := jts.Geom_NewTriangle(tt.v0, tt.v1, tt.v2)
			signedArea := tri.SignedArea()
			junit.AssertEqualsFloat64(t, tt.expectedSigned, signedArea, triangleTestTolerance)

			area := tri.Area()
			junit.AssertEqualsFloat64(t, math.Abs(tt.expectedSigned), area, triangleTestTolerance)
		})
	}
}

func TestTriangleAcute(t *testing.T) {
	tests := []struct {
		name     string
		v0       *jts.Geom_Coordinate
		v1       *jts.Geom_Coordinate
		v2       *jts.Geom_Coordinate
		expected bool
	}{
		{
			// POLYGON((10 10, 20 20, 20 10, 10 10)) - right triangle
			name:     "right triangle",
			v0:       jts.Geom_NewCoordinateWithXY(10, 10),
			v1:       jts.Geom_NewCoordinateWithXY(20, 20),
			v2:       jts.Geom_NewCoordinateWithXY(20, 10),
			expected: false,
		},
		{
			// POLYGON((10 10, 20 10, 20 20, 10 10)) - CCW right tri
			name:     "CCW right triangle",
			v0:       jts.Geom_NewCoordinateWithXY(10, 10),
			v1:       jts.Geom_NewCoordinateWithXY(20, 10),
			v2:       jts.Geom_NewCoordinateWithXY(20, 20),
			expected: false,
		},
		{
			// POLYGON((10 10, 20 10, 15 20, 10 10)) - acute
			name:     "acute triangle",
			v0:       jts.Geom_NewCoordinateWithXY(10, 10),
			v1:       jts.Geom_NewCoordinateWithXY(20, 10),
			v2:       jts.Geom_NewCoordinateWithXY(15, 20),
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tri := jts.Geom_NewTriangle(tt.v0, tt.v1, tt.v2)
			isAcute := tri.IsAcute()
			junit.AssertEquals(t, tt.expected, isAcute)
		})
	}
}

func TestTriangleCircumCentre(t *testing.T) {
	tests := []struct {
		name     string
		v0       *jts.Geom_Coordinate
		v1       *jts.Geom_Coordinate
		v2       *jts.Geom_Coordinate
		expected *jts.Geom_Coordinate
	}{
		{
			// POLYGON((10 10, 20 20, 20 10, 10 10)) - right triangle
			name:     "right triangle",
			v0:       jts.Geom_NewCoordinateWithXY(10, 10),
			v1:       jts.Geom_NewCoordinateWithXY(20, 20),
			v2:       jts.Geom_NewCoordinateWithXY(20, 10),
			expected: jts.Geom_NewCoordinateWithXY(15.0, 15.0),
		},
		{
			// POLYGON((10 10, 20 10, 20 20, 10 10)) - CCW right tri
			name:     "CCW right triangle",
			v0:       jts.Geom_NewCoordinateWithXY(10, 10),
			v1:       jts.Geom_NewCoordinateWithXY(20, 10),
			v2:       jts.Geom_NewCoordinateWithXY(20, 20),
			expected: jts.Geom_NewCoordinateWithXY(15.0, 15.0),
		},
		{
			// POLYGON((10 10, 20 10, 15 20, 10 10)) - acute
			name:     "acute triangle",
			v0:       jts.Geom_NewCoordinateWithXY(10, 10),
			v1:       jts.Geom_NewCoordinateWithXY(20, 10),
			v2:       jts.Geom_NewCoordinateWithXY(15, 20),
			expected: jts.Geom_NewCoordinateWithXY(15.0, 13.75),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test static version.
			circumcentre := jts.Geom_Triangle_Circumcentre(tt.v0, tt.v1, tt.v2)
			junit.AssertEquals(t, tt.expected.String(), circumcentre.String())

			// Test instance version.
			tri := jts.Geom_NewTriangle(tt.v0, tt.v1, tt.v2)
			circumcentre = tri.Circumcentre()
			junit.AssertEquals(t, tt.expected.String(), circumcentre.String())
		})
	}
}

func TestTriangleCircumradius(t *testing.T) {
	tests := []struct {
		name string
		v0   *jts.Geom_Coordinate
		v1   *jts.Geom_Coordinate
		v2   *jts.Geom_Coordinate
	}{
		{
			// POLYGON((10 10, 20 20, 20 10, 10 10)) - right triangle
			name: "right triangle",
			v0:   jts.Geom_NewCoordinateWithXY(10, 10),
			v1:   jts.Geom_NewCoordinateWithXY(20, 20),
			v2:   jts.Geom_NewCoordinateWithXY(20, 10),
		},
		{
			// POLYGON((10 10, 20 10, 20 20, 10 10)) - CCW right tri
			name: "CCW right triangle",
			v0:   jts.Geom_NewCoordinateWithXY(10, 10),
			v1:   jts.Geom_NewCoordinateWithXY(20, 10),
			v2:   jts.Geom_NewCoordinateWithXY(20, 20),
		},
		{
			// POLYGON((10 10, 20 10, 15 20, 10 10)) - acute
			name: "acute triangle",
			v0:   jts.Geom_NewCoordinateWithXY(10, 10),
			v1:   jts.Geom_NewCoordinateWithXY(20, 10),
			v2:   jts.Geom_NewCoordinateWithXY(15, 20),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			circumcentre := jts.Geom_Triangle_Circumcentre(tt.v0, tt.v1, tt.v2)
			circumradius := jts.Geom_Triangle_Circumradius(tt.v0, tt.v1, tt.v2)
			rad0 := tt.v0.Distance(circumcentre)
			rad1 := tt.v1.Distance(circumcentre)
			rad2 := tt.v2.Distance(circumcentre)
			junit.AssertEqualsFloat64(t, rad0, circumradius, 0.00001)
			junit.AssertEqualsFloat64(t, rad1, circumradius, 0.00001)
			junit.AssertEqualsFloat64(t, rad2, circumradius, 0.00001)
		})
	}
}

func TestTriangleCentroid(t *testing.T) {
	tests := []struct {
		name     string
		v0       *jts.Geom_Coordinate
		v1       *jts.Geom_Coordinate
		v2       *jts.Geom_Coordinate
		expected *jts.Geom_Coordinate
	}{
		{
			// POLYGON((10 10, 20 20, 20 10, 10 10)) - right triangle
			name:     "right triangle",
			v0:       jts.Geom_NewCoordinateWithXY(10, 10),
			v1:       jts.Geom_NewCoordinateWithXY(20, 20),
			v2:       jts.Geom_NewCoordinateWithXY(20, 10),
			expected: jts.Geom_NewCoordinateWithXY((10.0+20.0+20.0)/3.0, (10.0+20.0+10.0)/3.0),
		},
		{
			// POLYGON((10 10, 20 10, 20 20, 10 10)) - CCW right tri
			name:     "CCW right triangle",
			v0:       jts.Geom_NewCoordinateWithXY(10, 10),
			v1:       jts.Geom_NewCoordinateWithXY(20, 10),
			v2:       jts.Geom_NewCoordinateWithXY(20, 20),
			expected: jts.Geom_NewCoordinateWithXY((10.0+20.0+20.0)/3.0, (10.0+10.0+20.0)/3.0),
		},
		{
			// POLYGON((10 10, 20 10, 15 20, 10 10)) - acute
			name:     "acute triangle",
			v0:       jts.Geom_NewCoordinateWithXY(10, 10),
			v1:       jts.Geom_NewCoordinateWithXY(20, 10),
			v2:       jts.Geom_NewCoordinateWithXY(15, 20),
			expected: jts.Geom_NewCoordinateWithXY((10.0+20.0+15.0)/3.0, (10.0+10.0+20.0)/3.0),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test static version.
			centroid := jts.Geom_Triangle_Centroid(tt.v0, tt.v1, tt.v2)
			junit.AssertEquals(t, tt.expected.String(), centroid.String())

			// Test instance version.
			tri := jts.Geom_NewTriangle(tt.v0, tt.v1, tt.v2)
			centroid = tri.Centroid()
			junit.AssertEquals(t, tt.expected.String(), centroid.String())
		})
	}
}

func TestTriangleLongestSideLength(t *testing.T) {
	tests := []struct {
		name     string
		v0       *jts.Geom_Coordinate
		v1       *jts.Geom_Coordinate
		v2       *jts.Geom_Coordinate
		expected float64
	}{
		{
			// POLYGON((10 10 1, 20 20 2, 20 10 3, 10 10 1)) - right triangle
			name:     "right triangle",
			v0:       jts.Geom_NewCoordinateWithXYZ(10, 10, 1),
			v1:       jts.Geom_NewCoordinateWithXYZ(20, 20, 2),
			v2:       jts.Geom_NewCoordinateWithXYZ(20, 10, 3),
			expected: 14.142135623730951,
		},
		{
			// POLYGON((10 10 1, 20 10 2, 20 20 3, 10 10 1)) - CCW right tri
			name:     "CCW right triangle",
			v0:       jts.Geom_NewCoordinateWithXYZ(10, 10, 1),
			v1:       jts.Geom_NewCoordinateWithXYZ(20, 10, 2),
			v2:       jts.Geom_NewCoordinateWithXYZ(20, 20, 3),
			expected: 14.142135623730951,
		},
		{
			// POLYGON((10 10 1, 20 10 2, 15 20 3, 10 10 1)) - acute
			name:     "acute triangle",
			v0:       jts.Geom_NewCoordinateWithXYZ(10, 10, 1),
			v1:       jts.Geom_NewCoordinateWithXYZ(20, 10, 2),
			v2:       jts.Geom_NewCoordinateWithXYZ(15, 20, 3),
			expected: 11.180339887498949,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test static version.
			length := jts.Geom_Triangle_LongestSideLength(tt.v0, tt.v1, tt.v2)
			junit.AssertEqualsFloat64(t, tt.expected, length, 0.00000001)

			// Test instance version.
			tri := jts.Geom_NewTriangle(tt.v0, tt.v1, tt.v2)
			length = tri.LongestSideLength()
			junit.AssertEqualsFloat64(t, tt.expected, length, 0.00000001)
		})
	}
}

func TestTriangleIsCCW(t *testing.T) {
	tests := []struct {
		name     string
		v0       *jts.Geom_Coordinate
		v1       *jts.Geom_Coordinate
		v2       *jts.Geom_Coordinate
		expected bool
	}{
		{
			// POLYGON ((30 90, 80 50, 20 20, 30 90))
			name:     "CW triangle",
			v0:       jts.Geom_NewCoordinateWithXY(30, 90),
			v1:       jts.Geom_NewCoordinateWithXY(80, 50),
			v2:       jts.Geom_NewCoordinateWithXY(20, 20),
			expected: false,
		},
		{
			// POLYGON ((90 90, 20 40, 10 10, 90 90))
			name:     "CCW triangle",
			v0:       jts.Geom_NewCoordinateWithXY(90, 90),
			v1:       jts.Geom_NewCoordinateWithXY(20, 40),
			v2:       jts.Geom_NewCoordinateWithXY(10, 10),
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := jts.Geom_Triangle_IsCCW(tt.v0, tt.v1, tt.v2)
			junit.AssertEquals(t, tt.expected, actual)
		})
	}
}

func TestTriangleIntersects(t *testing.T) {
	tests := []struct {
		name     string
		v0       *jts.Geom_Coordinate
		v1       *jts.Geom_Coordinate
		v2       *jts.Geom_Coordinate
		p        *jts.Geom_Coordinate
		expected bool
	}{
		{
			// POLYGON ((30 90, 80 50, 20 20, 30 90)), POINT (70 20)
			name:     "point outside triangle",
			v0:       jts.Geom_NewCoordinateWithXY(30, 90),
			v1:       jts.Geom_NewCoordinateWithXY(80, 50),
			v2:       jts.Geom_NewCoordinateWithXY(20, 20),
			p:        jts.Geom_NewCoordinateWithXY(70, 20),
			expected: false,
		},
		{
			// POLYGON ((30 90, 80 50, 20 20, 30 90)), POINT (30 90) - triangle vertex
			name:     "point at triangle vertex",
			v0:       jts.Geom_NewCoordinateWithXY(30, 90),
			v1:       jts.Geom_NewCoordinateWithXY(80, 50),
			v2:       jts.Geom_NewCoordinateWithXY(20, 20),
			p:        jts.Geom_NewCoordinateWithXY(30, 90),
			expected: true,
		},
		{
			// POLYGON ((30 90, 80 50, 20 20, 30 90)), POINT (40 40)
			name:     "point inside triangle",
			v0:       jts.Geom_NewCoordinateWithXY(30, 90),
			v1:       jts.Geom_NewCoordinateWithXY(80, 50),
			v2:       jts.Geom_NewCoordinateWithXY(20, 20),
			p:        jts.Geom_NewCoordinateWithXY(40, 40),
			expected: true,
		},
		{
			// POLYGON ((30 90, 70 50, 71.5 16.5, 30 90)), POINT (50 70) - on an edge
			name:     "point on edge",
			v0:       jts.Geom_NewCoordinateWithXY(30, 90),
			v1:       jts.Geom_NewCoordinateWithXY(70, 50),
			v2:       jts.Geom_NewCoordinateWithXY(71.5, 16.5),
			p:        jts.Geom_NewCoordinateWithXY(50, 70),
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := jts.Geom_Triangle_Intersects(tt.v0, tt.v1, tt.v2, tt.p)
			junit.AssertEquals(t, tt.expected, actual)
		})
	}
}
