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
			p:        XY{0, 0},
			q:        XY{1, 0},
			s:        XY{0, 1},
			expected: leftTurn,
		},
		{
			name:     "when the s is on right hand side of line of p and q",
			p:        XY{0, 0},
			q:        XY{0, 1},
			s:        XY{1, 0},
			expected: rightTurn,
		},
		{
			name:     "when the s, q and p are collinear",
			p:        XY{1, 1},
			q:        XY{2, 2},
			s:        XY{3, 3},
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
