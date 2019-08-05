package geom

import "testing"

func TestOrientation(t *testing.T) {
	testCases := []struct {
		name     string
		p        XY
		q        XY
		s        XY
		expected threePointOrientation
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
