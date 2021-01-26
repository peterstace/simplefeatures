package geom

import (
	"strconv"
	"testing"
)

func TestSpanningTree(t *testing.T) {
	for i, tc := range []struct {
		xys     []XY
		wantWKT string
	}{
		{
			xys:     nil,
			wantWKT: "MULTILINESTRING EMPTY",
		},
		{
			xys:     []XY{{1, 1}},
			wantWKT: "MULTILINESTRING EMPTY",
		},
		{
			xys:     []XY{{2, 1}, {1, 2}},
			wantWKT: "MULTILINESTRING((2 1,1 2))",
		},
		{
			xys:     []XY{{2, 0}, {2, 2}, {0, 0}, {1.5, 1.5}},
			wantWKT: "MULTILINESTRING((0 0,2 0),(1.5 1.5,2 2),(2 0,1.5 1.5))",
		},
		{
			xys:     []XY{{-0.5, 0.5}, {0, 0}, {0, 1}, {1, 0}},
			wantWKT: "MULTILINESTRING((-0.5 0.5,0 0),(0 0,0 1),(0 1,1 0))",
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			want, err := UnmarshalWKT(tc.wantWKT)
			if err != nil {
				t.Fatal(err)
			}
			got := spanningTree(tc.xys)
			if !ExactEquals(want, got.AsGeometry(), IgnoreOrder) {
				t.Logf("got:  %v", got.AsText())
				t.Logf("want: %v", want.AsText())
				t.Fatal("mismatch")
			}
		})
	}
}
