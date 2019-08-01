package simplefeatures

import (
	"reflect"
	"testing"
)

func TestCrossProduct(t *testing.T) {
	testCases := []struct {
		name     string
		p        XY
		q        XY
		s        XY
		expected Scalar
	}{
		{
			name:     "happy path 1",
			p:        XY{MustNewScalarF(0), MustNewScalarF(0)},
			q:        XY{MustNewScalarF(1), MustNewScalarF(0)},
			s:        XY{MustNewScalarF(0), MustNewScalarF(1)},
			expected: MustNewScalarF(1),
		},
		{
			name:     "happy path 2",
			p:        XY{MustNewScalarF(0), MustNewScalarF(0)},
			q:        XY{MustNewScalarF(0.5), MustNewScalarF(0)},
			s:        XY{MustNewScalarF(0), MustNewScalarF(0.5)},
			expected: MustNewScalarF(0.25),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := crossProduct(tc.p, tc.q, tc.s)
			if !actual.Equals(tc.expected) {
				t.Errorf("expected: %v, got: %v", tc.expected, actual)
			}
		})
	}
}

func TestOrientation(t *testing.T) {
	testCases := []struct {
		name     string
		p        XY
		q        XY
		s        XY
		expected int
	}{
		{
			name:     "when the s is on left hand side of line of p and q",
			p:        XY{MustNewScalarF(0), MustNewScalarF(0)},
			q:        XY{MustNewScalarF(1), MustNewScalarF(0)},
			s:        XY{MustNewScalarF(0), MustNewScalarF(1)},
			expected: leftTurn,
		},
		{
			name:     "when the s is on right hand side of line of p and q",
			p:        XY{MustNewScalarF(0), MustNewScalarF(0)},
			q:        XY{MustNewScalarF(0), MustNewScalarF(1)},
			s:        XY{MustNewScalarF(1), MustNewScalarF(0)},
			expected: rightTurn,
		},
		{
			name:     "when the s, q and p are collinear",
			p:        XY{MustNewScalarF(1), MustNewScalarF(1)},
			q:        XY{MustNewScalarF(2), MustNewScalarF(2)},
			s:        XY{MustNewScalarF(3), MustNewScalarF(3)},
			expected: collinear,
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

func TestDistanceSq(t *testing.T) {
	testCases := []struct {
		name     string
		p        XY
		q        XY
		expected Scalar
	}{
		{
			name:     "when the points are the same",
			p:        XY{MustNewScalarF(1), MustNewScalarF(1)},
			q:        XY{MustNewScalarF(1), MustNewScalarF(1)},
			expected: MustNewScalarF(0),
		},
		{
			name:     "when the points are different",
			p:        XY{MustNewScalarF(0), MustNewScalarF(0)},
			q:        XY{MustNewScalarF(3), MustNewScalarF(4)},
			expected: MustNewScalarF(25),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := distanceSq(tc.p, tc.q)
			if !actual.Equals(tc.expected) {
				t.Errorf("expected %v, got %v", actual, tc.expected)
			}
		})
	}
}

func TestGrahamScan(t *testing.T) {
	testCases := []struct {
		name     string
		points   []XY
		expected []XY
	}{
		{
			name: "when the number of points is less than 3",
			points: []XY{
				XY{MustNewScalarF(1), MustNewScalarF(2)},
				XY{MustNewScalarF(2), MustNewScalarF(4)},
			},
			expected: nil,
		},
		{
			name: "when there are no collinear points",
			points: []XY{
				XY{MustNewScalarF(7), MustNewScalarF(7)},
				XY{MustNewScalarF(8), MustNewScalarF(5)},
				XY{MustNewScalarF(7), MustNewScalarF(2)},
				XY{MustNewScalarF(6), MustNewScalarF(5)},
				XY{MustNewScalarF(5), MustNewScalarF(5)},
				XY{MustNewScalarF(4), MustNewScalarF(6)},
				XY{MustNewScalarF(4), MustNewScalarF(2)},
				XY{MustNewScalarF(3), MustNewScalarF(7)},
				XY{MustNewScalarF(2), MustNewScalarF(1)},
			},
			expected: []XY{
				XY{MustNewScalarF(2), MustNewScalarF(1)},
				XY{MustNewScalarF(7), MustNewScalarF(2)},
				XY{MustNewScalarF(8), MustNewScalarF(5)},
				XY{MustNewScalarF(7), MustNewScalarF(7)},
				XY{MustNewScalarF(3), MustNewScalarF(7)},
				XY{MustNewScalarF(2), MustNewScalarF(1)},
			},
		},
		{
			name: "when there are collinear points",
			points: []XY{
				XY{MustNewScalarF(7), MustNewScalarF(7)},
				XY{MustNewScalarF(8), MustNewScalarF(5)},
				XY{MustNewScalarF(7), MustNewScalarF(2)},
				XY{MustNewScalarF(6), MustNewScalarF(5)},
				XY{MustNewScalarF(5), MustNewScalarF(5)},
				XY{MustNewScalarF(4), MustNewScalarF(6)},
				XY{MustNewScalarF(4), MustNewScalarF(2)},
				XY{MustNewScalarF(3), MustNewScalarF(7)},
				XY{MustNewScalarF(2), MustNewScalarF(1)},
				XY{MustNewScalarF(14), MustNewScalarF(9)},
			},
			expected: []XY{
				XY{MustNewScalarF(2), MustNewScalarF(1)},
				XY{MustNewScalarF(7), MustNewScalarF(2)},
				XY{MustNewScalarF(14), MustNewScalarF(9)},
				XY{MustNewScalarF(7), MustNewScalarF(7)},
				XY{MustNewScalarF(3), MustNewScalarF(7)},
				XY{MustNewScalarF(2), MustNewScalarF(1)},
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
