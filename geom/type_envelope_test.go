package geom_test

import (
	"fmt"
	"testing"

	. "github.com/peterstace/simplefeatures/geom"
)

func TestEnvelopeIntersectsPoint(t *testing.T) {
	env := NewEnvelope(
		XY{12, 4},
		XY{14, 2},
	)
	for x := 11; x <= 15; x++ {
		for y := 1; y <= 5; y++ {
			t.Run(fmt.Sprintf("%d_%d", x, y), func(t *testing.T) {
				want := x >= 12 && x <= 14 && y >= 2 && y <= 4
				pt := XY{float64(x), float64(y)}
				got := env.IntersectsPoint(pt)
				if got != want {
					t.Errorf("want=%v got=%v", want, got)
				}
			})
		}
	}
}

func TestEnvelopeAsGeometry(t *testing.T) {
	for _, tt := range []struct {
		env     Envelope
		wantWKT string
	}{
		{NewEnvelope(XY{5, 8}), "POINT(5 8)"},
		{NewEnvelope(XY{1, 2}, XY{5, 2}), "LINESTRING(1 2,5 2)"},
		{NewEnvelope(XY{1, 2}, XY{1, 7}), "LINESTRING(1 2,1 7)"},
		{NewEnvelope(XY{3, 4}, XY{8, 0}), "POLYGON((3 0,3 4,8 4,8 0,3 0))"},
	} {
		got := tt.env.AsGeometry()
		expectDeepEqual(t, got, geomFromWKT(t, tt.wantWKT))
	}
}
