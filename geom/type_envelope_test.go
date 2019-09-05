package geom_test

import (
	"fmt"
	"strconv"
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

func TestEnvelopeIntersects(t *testing.T) {
	for i, tt := range []struct {
		e1, e2 Envelope
		want   bool
	}{
		{
			NewEnvelope(XY{0, 0}, XY{1, 1}),
			NewEnvelope(XY{2, 2}, XY{3, 3}),
			false,
		},
		{
			NewEnvelope(XY{0, 2}, XY{1, 3}),
			NewEnvelope(XY{2, 0}, XY{3, 1}),
			false,
		},
		{
			NewEnvelope(XY{0, 0}, XY{1, 1}),
			NewEnvelope(XY{1, 1}, XY{2, 2}),
			true,
		},
		{
			NewEnvelope(XY{0, 1}, XY{1, 2}),
			NewEnvelope(XY{1, 0}, XY{2, 1}),
			true,
		},
		{
			NewEnvelope(XY{0, 0}, XY{2, 2}),
			NewEnvelope(XY{1, 1}, XY{3, 3}),
			true,
		},
		{
			NewEnvelope(XY{0, 1}, XY{2, 3}),
			NewEnvelope(XY{1, 0}, XY{3, 2}),
			true,
		},
		{
			NewEnvelope(XY{0, 0}, XY{2, 1}),
			NewEnvelope(XY{1, 0}, XY{3, 1}),
			true,
		},
		{
			NewEnvelope(XY{0, 0}, XY{1, 2}),
			NewEnvelope(XY{0, 1}, XY{1, 3}),
			true,
		},
		{
			NewEnvelope(XY{0, 0}, XY{2, 2}),
			NewEnvelope(XY{1, -1}, XY{3, 3}),
			true,
		},
		{
			NewEnvelope(XY{0, 0}, XY{2, 2}),
			NewEnvelope(XY{1, -1}, XY{3, 3}),
			true,
		},
		{
			NewEnvelope(XY{-1, 0}, XY{2, 1}),
			NewEnvelope(XY{0, -1}, XY{1, 2}),
			true,
		},
		{
			NewEnvelope(XY{0, 0}, XY{1, 1}),
			NewEnvelope(XY{-1, -1}, XY{2, 2}),
			true,
		},
		{
			NewEnvelope(XY{0, 0}, XY{1, 1}),
			NewEnvelope(XY{1, 0}, XY{2, 1}),
			true,
		},
		{
			NewEnvelope(XY{0, 0}, XY{1, 1}),
			NewEnvelope(XY{0, 1}, XY{1, 2}),
			true,
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			got1 := tt.e1.Intersects(tt.e2)
			got2 := tt.e2.Intersects(tt.e1)
			if got1 != tt.want || got2 != tt.want {
				t.Logf("env1: %v", tt.e1)
				t.Logf("env2: %v", tt.e2)
				t.Errorf("want=%v got1=%v", tt.want, got1)
				t.Errorf("want=%v got2=%v", tt.want, got2)
			}
		})
	}
}
