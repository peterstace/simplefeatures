package geom

import (
	"strconv"
	"testing"
)

func TestLERP(t *testing.T) {
	for i, tc := range []struct {
		a, b, t, want float64
	}{
		// Interpolation between a and b.
		{a: -1, b: 3, t: 0.00, want: -1},
		{a: -1, b: 3, t: 0.25, want: +0},
		{a: -1, b: 3, t: 0.50, want: +1},
		{a: -1, b: 3, t: 0.75, want: +2},
		{a: -1, b: 3, t: 1.00, want: +3},

		// Extrapolation outside a and b:
		{a: -1, b: 3, t: -0.5, want: -3},
		{a: -1, b: 3, t: +1.5, want: +5},

		// Reproduces a bug when lerp implemented as: return a + t*(b-a)
		{
			a:    0.4295025244660839,
			b:    0.11201266333061713,
			t:    1,
			want: 0.11201266333061713, // same as 'a'
		},

		// Reproduces a bug when lerp implemented as: return t*a + (1-t)*b
		{
			a:    0.9202968672544602,
			b:    0.9202968672544602,
			t:    0.3251482131256554,
			want: 0.9202968672544602,
		},

		// Reproduces other bugs:
		{a: 1, b: 3, t: 0.5, want: 2},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			got := lerp(tc.a, tc.b, tc.t)
			if got != tc.want {
				t.Logf("a=%v, b=%v, t=%v", tc.a, tc.b, tc.t)
				t.Errorf("got %v, want %v", got, tc.want)
			}
		})
	}
}
