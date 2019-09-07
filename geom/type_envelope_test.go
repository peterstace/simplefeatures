package geom_test

import (
	"fmt"
	"strconv"
	"testing"

	. "github.com/peterstace/simplefeatures/geom"
)

func TestEnvelopeContains(t *testing.T) {
	env := NewEnvelope(
		XY{12, 4},
		XY{14, 2},
	)
	for x := 11; x <= 15; x++ {
		for y := 1; y <= 5; y++ {
			t.Run(fmt.Sprintf("%d_%d", x, y), func(t *testing.T) {
				want := x >= 12 && x <= 14 && y >= 2 && y <= 4
				pt := XY{float64(x), float64(y)}
				got := env.Contains(pt)
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

// env is a helper to create an envelope in a compact way.
func env(x1, y1, x2, y2 float64) Envelope {
	return NewEnvelope(XY{x1, y1}, XY{x2, y2})
}

func TestEnvelopeIntersects(t *testing.T) {
	for i, tt := range []struct {
		e1, e2 Envelope
		want   bool
	}{
		{env(0, 0, 1, 1), env(2, 2, 3, 3), false},
		{env(0, 2, 1, 3), env(2, 0, 3, 1), false},
		{env(0, 0, 1, 1), env(1, 1, 2, 2), true},
		{env(0, 1, 1, 2), env(1, 0, 2, 1), true},
		{env(0, 0, 2, 2), env(1, 1, 3, 3), true},
		{env(0, 1, 2, 3), env(1, 0, 3, 2), true},
		{env(0, 0, 2, 1), env(1, 0, 3, 1), true},
		{env(0, 0, 1, 2), env(0, 1, 1, 3), true},
		{env(0, 0, 2, 2), env(1, -1, 3, 3), true},
		{env(0, 0, 2, 2), env(1, -1, 3, 3), true},
		{env(-1, 0, 2, 1), env(0, -1, 1, 2), true},
		{env(0, 0, 1, 1), env(-1, -1, 2, 2), true},
		{env(0, 0, 1, 1), env(1, 0, 2, 1), true},
		{env(0, 0, 1, 1), env(0, 1, 1, 2), true},
		{env(0, 0, 1, 1), env(2, 0, 3, 1), false},
		{env(0, 0, 1, 1), env(0, 2, 1, 3), false},
		{env(0, 0, 1, 1), env(2, -1, 3, 2), false},
		{env(0, 0, 1, 1), env(-1, -2, 2, -1), false},
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

func TestEnvelopeCenter(t *testing.T) {
	for i, tt := range []struct {
		env  Envelope
		want XY
	}{
		{env(2, 6, 1, 5), XY{1.5, 5.5}},
		{env(4, 1, 4, -2), XY{4, -0.5}},
		{env(-3, 10, -3, 10), XY{-3, 10}},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			got := tt.env.Center()
			if got != tt.want {
				t.Errorf("got=%v want=%v", got, tt.want)
			}
		})
	}
}
