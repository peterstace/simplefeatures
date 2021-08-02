package exact_test

import (
	"testing"

	"github.com/peterstace/simplefeatures/internal/exact"
)

func TestBentlyOttmann(t *testing.T) {
	for _, tc := range []struct {
		description string
		segments    []exact.Segment
		want        []exact.IntersectionReport
	}{
		{
			description: "no segments",
			segments:    nil,
			want:        nil,
		},
		{
			description: "single segments",
			segments: []exact.Segment{
				{
					A: exact.XY64{
						X: 0,
						Y: 0,
					},
					B: exact.XY64{
						X: 1,
						Y: 1,
					},
				},
			},
			want: nil,
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
			var got []exact.IntersectionReport
			exact.BentlyOttmann(
				tc.segments,
				func(ir exact.IntersectionReport) bool {
					got = append(got, ir)
					return true
				},
			)

			show := func() {
				t.Logf("want: len=%d", len(tc.want))
				for i, w := range tc.want {
					t.Logf(" w[%d]: %v", i, w)
				}
				t.Logf("got: len=%d", len(got))
				for i, g := range got {
					t.Logf(" g[%d]: %v", i, g)
				}
			}
			if len(tc.want) != len(got) {
				t.Fatal("length mismatch")
				show()
			}
			for i := range tc.want {
				var any bool
				if tc.want[i] != got[i] {
					t.Errorf("mismatch at %d", i)
					any = true
				}
				if any {
					show()
				}
			}
		})
	}
}
