package simplefeatures

import "testing"

func TestOrientation(t *testing.T) {
	testCases := []struct {
		name     string
		p        XY
		q        XY
		s        XY
		expected orientation
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
			actual := orient(tc.p, tc.q, tc.s)
			if actual != tc.expected {
				t.Errorf("expected: %d, got: %d", tc.expected, actual)
			}
		})
	}
}
