package geom

import (
	"errors"
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

func TestCatch(t *testing.T) {
	for _, tc := range []struct {
		name    string
		fn      func() error
		wantErr string
	}{
		{
			name:    "function returns nil",
			fn:      func() error { return nil },
			wantErr: "",
		},
		{
			name:    "function returns error",
			fn:      func() error { return errors.New("test error") },
			wantErr: "test error",
		},
		{
			name:    "function panics with string",
			fn:      func() error { panic("something went wrong") },
			wantErr: "panic: something went wrong",
		},
		{
			name:    "function panics with error",
			fn:      func() error { panic(errors.New("panic error")) },
			wantErr: "panic: panic error",
		},
		{
			name:    "function panics with int",
			fn:      func() error { panic(42) },
			wantErr: "panic: 42",
		},
		{
			name:    "function panics with nil",
			fn:      func() error { panic(nil) },
			wantErr: "panic: panic called with nil argument",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := catch(tc.fn)
			if tc.wantErr == "" {
				if err != nil {
					t.Errorf("got %v, want nil", err)
				}
			} else {
				if err == nil {
					t.Fatalf("got nil, want %q", tc.wantErr)
				}
				if err.Error() != tc.wantErr {
					t.Errorf("got %q, want %q", err.Error(), tc.wantErr)
				}
			}
		})
	}
}
