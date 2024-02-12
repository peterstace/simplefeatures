package geom

import (
	"strconv"
	"testing"
)

func TestNodeSetNoCrash(t *testing.T) {
	for i, tc := range []struct {
		maxULP float64
		pts    []XY
	}{
		// Reproduces a crash.
		{
			maxULP: 2.220446049250313e-16,
			pts: []XY{
				{0, 1},
				{4.440892098500626e-16, 0.9999999999999997},
				{0, 0.9999999999999997},
			},
		},
	} {
		t.Run(strconv.Itoa(i), func(*testing.T) {
			ns := newNodeSet(tc.maxULP, len(tc.pts))
			for _, pt := range tc.pts {
				ns.insertOrGet(pt)
			}
		})
	}
}
