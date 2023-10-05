package geom

import (
	"math"
	"strconv"
	"testing"
)

func TestFastMinFastMax(t *testing.T) {
	var (
		nan = math.NaN()
		inf = math.Inf(1)
		eq  = func(a, b float64) bool {
			return (math.IsNaN(a) && math.IsNaN(b)) || a == b
		}
	)
	for i, tc := range []struct {
		a, b     float64
		min, max float64
	}{
		{0, 0, 0, 0},
		{1, 2, 1, 2},
		{2, 1, 1, 2},
		{0, nan, nan, nan},
		{nan, 0, nan, nan},
		{nan, nan, nan, nan},
		{0, inf, 0, inf},
		{inf, 0, 0, inf},
		{inf, inf, inf, inf},
		{0, -inf, -inf, 0},
		{-inf, 0, -inf, 0},
		{-inf, -inf, -inf, -inf},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			gotMin := fastMin(tc.a, tc.b)
			gotMax := fastMax(tc.a, tc.b)
			if !eq(gotMin, tc.min) {
				t.Errorf("fastMin(%v, %v) = %v, want %v", tc.a, tc.b, gotMin, tc.min)
			}
			if !eq(gotMax, tc.max) {
				t.Errorf("fastMax(%v, %v) = %v, want %v", tc.a, tc.b, gotMax, tc.max)
			}
		})
	}
}

var global float64

func BenchmarkFastMin(b *testing.B) {
	for i := 0; i < b.N; i++ {
		global = fastMin(global, 2)
	}
}

func BenchmarkFastMax(b *testing.B) {
	for i := 0; i < b.N; i++ {
		global = fastMax(global, 2)
	}
}

func BenchmarkMathMin(b *testing.B) {
	for i := 0; i < b.N; i++ {
		global = math.Min(global, 2)
	}
}

func BenchmarkMathMax(b *testing.B) {
	for i := 0; i < b.N; i++ {
		global = math.Max(global, 2)
	}
}
