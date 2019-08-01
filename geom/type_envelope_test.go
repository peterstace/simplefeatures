package simplefeatures_test

import (
	"fmt"
	"testing"

	. "github.com/peterstace/simplefeatures/geom"
)

func TestEnvelopeIntersectsPoint(t *testing.T) {
	env := NewEnvelope(
		XY{MustNewScalarF(12), MustNewScalarF(4)},
		XY{MustNewScalarF(14), MustNewScalarF(2)},
	)
	for x := 11; x <= 15; x++ {
		for y := 1; y <= 5; y++ {
			t.Run(fmt.Sprintf("%d_%d", x, y), func(t *testing.T) {
				want := x >= 12 && x <= 14 && y >= 2 && y <= 4
				pt := XY{
					MustNewScalarF(float64(x)),
					MustNewScalarF(float64(y)),
				}
				got := env.IntersectsPoint(pt)
				if got != want {
					t.Errorf("want=%v got=%v", want, got)
				}
			})
		}
	}
}
