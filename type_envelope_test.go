package simplefeatures_test

import (
	"fmt"
	"testing"

	. "github.com/peterstace/simplefeatures"
)

func TestEnvelopeIntersectsPoint(t *testing.T) {
	env := NewEnvelope(
		XY{NewScalarFromFloat64(12), NewScalarFromFloat64(4)},
		XY{NewScalarFromFloat64(14), NewScalarFromFloat64(2)},
	)
	for x := 11; x <= 15; x++ {
		for y := 1; y <= 5; y++ {
			t.Run(fmt.Sprintf("%d_%d", x, y), func(t *testing.T) {
				want := x >= 12 && x <= 14 && y >= 2 && y <= 4
				pt := XY{
					NewScalarFromFloat64(float64(x)),
					NewScalarFromFloat64(float64(y)),
				}
				got := env.IntersectsPoint(pt)
				if got != want {
					t.Errorf("want=%v got=%v", want, got)
				}
			})
		}
	}
}
