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
			p:        XY{NewScalarFromFloat64(0), NewScalarFromFloat64(0)},
			q:        XY{NewScalarFromFloat64(1), NewScalarFromFloat64(0)},
			s:        XY{NewScalarFromFloat64(0), NewScalarFromFloat64(1)},
			expected: NewScalarFromFloat64(1),
		},
		{
			name:     "happy path 2",
			p:        XY{NewScalarFromFloat64(0), NewScalarFromFloat64(0)},
			q:        XY{NewScalarFromFloat64(0.5), NewScalarFromFloat64(0)},
			s:        XY{NewScalarFromFloat64(0), NewScalarFromFloat64(0.5)},
			expected: NewScalarFromFloat64(0.25),
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
			p:        XY{NewScalarFromFloat64(0), NewScalarFromFloat64(0)},
			q:        XY{NewScalarFromFloat64(1), NewScalarFromFloat64(0)},
			s:        XY{NewScalarFromFloat64(0), NewScalarFromFloat64(1)},
			expected: leftTurn,
		},
		{
			name:     "when the s is on right hand side of line of p and q",
			p:        XY{NewScalarFromFloat64(0), NewScalarFromFloat64(0)},
			q:        XY{NewScalarFromFloat64(0), NewScalarFromFloat64(1)},
			s:        XY{NewScalarFromFloat64(1), NewScalarFromFloat64(0)},
			expected: rightTurn,
		},
		{
			name:     "when the s, q and p are collinear",
			p:        XY{NewScalarFromFloat64(1), NewScalarFromFloat64(1)},
			q:        XY{NewScalarFromFloat64(2), NewScalarFromFloat64(2)},
			s:        XY{NewScalarFromFloat64(3), NewScalarFromFloat64(3)},
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
			p:        XY{NewScalarFromFloat64(1), NewScalarFromFloat64(1)},
			q:        XY{NewScalarFromFloat64(1), NewScalarFromFloat64(1)},
			expected: NewScalarFromFloat64(0),
		},
		{
			name:     "when the points are different",
			p:        XY{NewScalarFromFloat64(0), NewScalarFromFloat64(0)},
			q:        XY{NewScalarFromFloat64(3), NewScalarFromFloat64(4)},
			expected: NewScalarFromFloat64(25),
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
				XY{NewScalarFromFloat64(1), NewScalarFromFloat64(2)},
				XY{NewScalarFromFloat64(2), NewScalarFromFloat64(4)},
			},
			expected: nil,
		},
		{
			name: "when there are no collinear points",
			points: []XY{
				XY{NewScalarFromFloat64(7), NewScalarFromFloat64(7)},
				XY{NewScalarFromFloat64(8), NewScalarFromFloat64(5)},
				XY{NewScalarFromFloat64(7), NewScalarFromFloat64(2)},
				XY{NewScalarFromFloat64(6), NewScalarFromFloat64(5)},
				XY{NewScalarFromFloat64(5), NewScalarFromFloat64(5)},
				XY{NewScalarFromFloat64(4), NewScalarFromFloat64(6)},
				XY{NewScalarFromFloat64(4), NewScalarFromFloat64(2)},
				XY{NewScalarFromFloat64(3), NewScalarFromFloat64(7)},
				XY{NewScalarFromFloat64(2), NewScalarFromFloat64(1)},
			},
			expected: []XY{
				XY{NewScalarFromFloat64(2), NewScalarFromFloat64(1)},
				XY{NewScalarFromFloat64(7), NewScalarFromFloat64(2)},
				XY{NewScalarFromFloat64(8), NewScalarFromFloat64(5)},
				XY{NewScalarFromFloat64(7), NewScalarFromFloat64(7)},
				XY{NewScalarFromFloat64(3), NewScalarFromFloat64(7)},
				XY{NewScalarFromFloat64(2), NewScalarFromFloat64(1)},
			},
		},
		{
			name: "when there are collinear points",
			points: []XY{
				XY{NewScalarFromFloat64(7), NewScalarFromFloat64(7)},
				XY{NewScalarFromFloat64(8), NewScalarFromFloat64(5)},
				XY{NewScalarFromFloat64(7), NewScalarFromFloat64(2)},
				XY{NewScalarFromFloat64(6), NewScalarFromFloat64(5)},
				XY{NewScalarFromFloat64(5), NewScalarFromFloat64(5)},
				XY{NewScalarFromFloat64(4), NewScalarFromFloat64(6)},
				XY{NewScalarFromFloat64(4), NewScalarFromFloat64(2)},
				XY{NewScalarFromFloat64(3), NewScalarFromFloat64(7)},
				XY{NewScalarFromFloat64(2), NewScalarFromFloat64(1)},
				XY{NewScalarFromFloat64(14), NewScalarFromFloat64(9)},
			},
			expected: []XY{
				XY{NewScalarFromFloat64(2), NewScalarFromFloat64(1)},
				XY{NewScalarFromFloat64(7), NewScalarFromFloat64(2)},
				XY{NewScalarFromFloat64(14), NewScalarFromFloat64(9)},
				XY{NewScalarFromFloat64(7), NewScalarFromFloat64(7)},
				XY{NewScalarFromFloat64(3), NewScalarFromFloat64(7)},
				XY{NewScalarFromFloat64(2), NewScalarFromFloat64(1)},
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
