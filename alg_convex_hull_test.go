package simplefeatures

import (
	"reflect"
	"testing"
)

func TestCrossProduct(t *testing.T) {
	testCases := []struct {
		name     string
		p        Point
		q        Point
		s        Point
		expected float64
	}{
		{
			name:     "happy path 1",
			p:        NewPointXY(NewScalarFromFloat64(0), NewScalarFromFloat64(0)),
			q:        NewPointXY(NewScalarFromFloat64(1), NewScalarFromFloat64(0)),
			s:        NewPointXY(NewScalarFromFloat64(0), NewScalarFromFloat64(1)),
			expected: 1,
		},
		{
			name:     "happy path 2",
			p:        NewPointXY(NewScalarFromFloat64(0), NewScalarFromFloat64(0)),
			q:        NewPointXY(NewScalarFromFloat64(0.5), NewScalarFromFloat64(0)),
			s:        NewPointXY(NewScalarFromFloat64(0), NewScalarFromFloat64(0.5)),
			expected: 0.25,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := crossProduct(tc.p, tc.q, tc.s)
			if actual != tc.expected {
				t.Errorf("expected: %f, got: %f", tc.expected, actual)
			}
		})
	}
}

func TestOrientation(t *testing.T) {
	testCases := []struct {
		name     string
		p        Point
		q        Point
		s        Point
		expected int
	}{
		{
			name:     "when the s is on left hand side of line of p and q",
			p:        NewPointXY(NewScalarFromFloat64(0), NewScalarFromFloat64(0)),
			q:        NewPointXY(NewScalarFromFloat64(1), NewScalarFromFloat64(0)),
			s:        NewPointXY(NewScalarFromFloat64(0), NewScalarFromFloat64(1)),
			expected: Left,
		},
		{
			name:     "when the s is on right hand side of line of p and q",
			p:        NewPointXY(NewScalarFromFloat64(0), NewScalarFromFloat64(0)),
			q:        NewPointXY(NewScalarFromFloat64(0), NewScalarFromFloat64(1)),
			s:        NewPointXY(NewScalarFromFloat64(1), NewScalarFromFloat64(0)),
			expected: Right,
		},
		{
			name:     "when the s, q and p are collinear",
			p:        NewPointXY(NewScalarFromFloat64(1), NewScalarFromFloat64(1)),
			q:        NewPointXY(NewScalarFromFloat64(2), NewScalarFromFloat64(2)),
			s:        NewPointXY(NewScalarFromFloat64(3), NewScalarFromFloat64(3)),
			expected: Collinear,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := orientation(tc.p, tc.q, tc.s)
			if actual != tc.expected {
				t.Errorf("expected: %d, got: %d", tc.expected, actual)
			}
		})
	}
}

func TestDistance(t *testing.T) {
	testCases := []struct {
		name     string
		p        Point
		q        Point
		expected float64
	}{
		{
			name:     "when the points are the same",
			p:        NewPointXY(NewScalarFromFloat64(1), NewScalarFromFloat64(1)),
			q:        NewPointXY(NewScalarFromFloat64(1), NewScalarFromFloat64(1)),
			expected: 0,
		},
		{
			name:     "when the points are different",
			p:        NewPointXY(NewScalarFromFloat64(0), NewScalarFromFloat64(0)),
			q:        NewPointXY(NewScalarFromFloat64(3), NewScalarFromFloat64(4)),
			expected: 5,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := distance(tc.p, tc.q)
			if actual != tc.expected {
				t.Errorf("expected %f, got %f", actual, tc.expected)
			}
		})
	}
}

func TestGrahamScan(t *testing.T) {
	testCases := []struct {
		name     string
		points   []Point
		expected []Point
	}{
		{
			name: "when the number of points is less than 3",
			points: []Point{
				NewPointXY(
					NewScalarFromFloat64(1),
					NewScalarFromFloat64(2),
				),
				NewPointXY(
					NewScalarFromFloat64(2),
					NewScalarFromFloat64(4),
				),
			},
			expected: nil,
		},
		{
			name: "when there are no collinear points",
			points: []Point{
				NewPointXY(
					NewScalarFromFloat64(7),
					NewScalarFromFloat64(7),
				),
				NewPointXY(
					NewScalarFromFloat64(8),
					NewScalarFromFloat64(5),
				),
				NewPointXY(
					NewScalarFromFloat64(7),
					NewScalarFromFloat64(2),
				),
				NewPointXY(
					NewScalarFromFloat64(6),
					NewScalarFromFloat64(5),
				),
				NewPointXY(
					NewScalarFromFloat64(5),
					NewScalarFromFloat64(5),
				),
				NewPointXY(
					NewScalarFromFloat64(4),
					NewScalarFromFloat64(6),
				),
				NewPointXY(
					NewScalarFromFloat64(4),
					NewScalarFromFloat64(2),
				),
				NewPointXY(
					NewScalarFromFloat64(3),
					NewScalarFromFloat64(7),
				),
				NewPointXY(
					NewScalarFromFloat64(2),
					NewScalarFromFloat64(1),
				),
			},
			expected: []Point{
				NewPointXY(
					NewScalarFromFloat64(2),
					NewScalarFromFloat64(1),
				),
				NewPointXY(
					NewScalarFromFloat64(7),
					NewScalarFromFloat64(2),
				),
				NewPointXY(
					NewScalarFromFloat64(8),
					NewScalarFromFloat64(5),
				),
				NewPointXY(
					NewScalarFromFloat64(7),
					NewScalarFromFloat64(7),
				),
				NewPointXY(
					NewScalarFromFloat64(3),
					NewScalarFromFloat64(7),
				),
				NewPointXY(
					NewScalarFromFloat64(2),
					NewScalarFromFloat64(1),
				),
			},
		},
		{
			name: "when there are collinear points",
			points: []Point{
				NewPointXY(
					NewScalarFromFloat64(7),
					NewScalarFromFloat64(7),
				),
				NewPointXY(
					NewScalarFromFloat64(8),
					NewScalarFromFloat64(5),
				),
				NewPointXY(
					NewScalarFromFloat64(7),
					NewScalarFromFloat64(2),
				),
				NewPointXY(
					NewScalarFromFloat64(6),
					NewScalarFromFloat64(5),
				),
				NewPointXY(
					NewScalarFromFloat64(5),
					NewScalarFromFloat64(5),
				),
				NewPointXY(
					NewScalarFromFloat64(4),
					NewScalarFromFloat64(6),
				),
				NewPointXY(
					NewScalarFromFloat64(4),
					NewScalarFromFloat64(2),
				),
				NewPointXY(
					NewScalarFromFloat64(3),
					NewScalarFromFloat64(7),
				),
				NewPointXY(
					NewScalarFromFloat64(2),
					NewScalarFromFloat64(1),
				),
				NewPointXY(
					NewScalarFromFloat64(14),
					NewScalarFromFloat64(9),
				),
			},
			expected: []Point{
				NewPointXY(
					NewScalarFromFloat64(2),
					NewScalarFromFloat64(1),
				),
				NewPointXY(
					NewScalarFromFloat64(7),
					NewScalarFromFloat64(2),
				),
				NewPointXY(
					NewScalarFromFloat64(14),
					NewScalarFromFloat64(9),
				),
				NewPointXY(
					NewScalarFromFloat64(7),
					NewScalarFromFloat64(7),
				),
				NewPointXY(
					NewScalarFromFloat64(3),
					NewScalarFromFloat64(7),
				),
				NewPointXY(
					NewScalarFromFloat64(2),
					NewScalarFromFloat64(1),
				),
			},
		},
	}

	for _, tc := range testCases {
		actual := grahamScan(tc.points)
		if !reflect.DeepEqual(actual, tc.expected) {
			t.Errorf("\nexpected %#v, \ngot %#v", tc.expected, actual)
		}
	}
}
