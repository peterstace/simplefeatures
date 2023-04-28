package main

import (
	"fmt"
	"math"
	"testing"

	"github.com/peterstace/simplefeatures/geom"
)

func TestMantissaTerminatesQuickly(t *testing.T) {
	// Test mantissaTerminatesQuickly function, since it's fairly complicated
	// internally.
	for _, tt := range []struct {
		f    float64
		want bool
	}{
		{1.0, true},
		{1.5, true},
		{53, true},
		{100, true},
		{0.1, false},
		{-0.1, false},
		{0.7, false},
		{math.Pi, false},
	} {
		t.Run(fmt.Sprintf("%v", tt.f), func(t *testing.T) {
			pt, err := geom.XY{X: tt.f, Y: tt.f}.AsPoint()
			if err != nil {
				t.Fatal(err)
			}
			got := mantissaTerminatesQuickly(pt.AsGeometry())
			if got != tt.want {
				t.Errorf("got=%v want=%v", got, tt.want)
			}
		})
	}
}
