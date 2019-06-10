package simplefeatures_test

import (
	"fmt"
	"math"
	"testing"

	. "github.com/peterstace/simplefeatures"
)

func TestPointValidation(t *testing.T) {
	for _, tt := range []struct {
		x, y float64
	}{
		{0, math.Inf(-1)},
		{0, math.Inf(+1)},
		{0, math.NaN()},
		{math.Inf(-1), 0},
		{math.Inf(+1), 0},
		{math.NaN(), 0},
		{math.Inf(-1), math.Inf(-1)},
		{math.Inf(+1), math.Inf(+1)},
		{math.NaN(), math.NaN()},
	} {
		t.Run(fmt.Sprintf("%f_%f", tt.x, tt.y), func(t *testing.T) {
			_, err := NewPoint(tt.x, tt.y)
			if err == nil {
				t.Error("expected error")
			}
		})
	}
}
