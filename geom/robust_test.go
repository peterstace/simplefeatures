package geom

import (
	"math"
	"math/big"
	"strconv"
	"testing"
)

func TestExactSum(t *testing.T) {
	for i, tt := range []struct {
		a, b float64
	}{
		{0, 0},
		{0.1, 0.1},
		{1e0, 1e10},
		{1e0, 1e20},
		{1e0, 1e40},
		{1e-20, 1e20},
		{-200, 200},
		{math.Pi, math.E},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			gotS, gotE := exactSum(tt.a, tt.b)
			gotSum := new(big.Rat).Add(
				new(big.Rat).SetFloat64(gotS),
				new(big.Rat).SetFloat64(gotE),
			)
			wantSum := new(big.Rat).Add(
				new(big.Rat).SetFloat64(tt.a),
				new(big.Rat).SetFloat64(tt.b),
			)
			t.Logf("a: %v", tt.a)
			t.Logf("b: %v", tt.b)
			t.Logf("s: %v", gotS)
			t.Logf("e: %v", gotE)
			if gotSum.Cmp(wantSum) != 0 {
				t.Errorf("got: %v, want: %v", gotSum, wantSum)
			}
		})
	}
}

func TestExactMul(t *testing.T) {
	for i, tt := range []struct {
		a, b float64
	}{
		{0, 0},
		{1, 1},
		{0.1, 0.1},
		{1e0, 1e10},
		{1e0, 1e20},
		{1e0, 1e40},
		{1e-20, 1e20},
		{-200, 200},
		{math.Pi, math.E},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			gotP, gotQ := exactMul(tt.a, tt.b)
			gotMul := new(big.Rat).Add(
				new(big.Rat).SetFloat64(gotP),
				new(big.Rat).SetFloat64(gotQ),
			)
			wantMul := new(big.Rat).Mul(
				new(big.Rat).SetFloat64(tt.a),
				new(big.Rat).SetFloat64(tt.b),
			)
			t.Logf("a: %v", tt.a)
			t.Logf("b: %v", tt.b)
			t.Logf("p: %v", gotP)
			t.Logf("q: %v", gotQ)
			if gotMul.Cmp(wantMul) != 0 {
				t.Errorf("got: %v, want: %v", gotMul, wantMul)
			}
		})
	}
}

func BenchmarkExactMul(b *testing.B) {
	for i := 0; i < b.N; i++ {
		exactMul(math.Pi, math.E)
	}
}
