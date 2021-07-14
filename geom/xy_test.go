package geom_test

import (
	"math"
	"testing"

	"github.com/peterstace/simplefeatures/geom"
)

func TestXYUnit(t *testing.T) {
	sqrt2 := math.Sqrt(2)
	sqrt5 := math.Sqrt(5)
	for _, tc := range []struct {
		description string
		input       geom.XY
		output      geom.XY
	}{
		{
			description: "+ve unit in X",
			input:       geom.XY{X: 1},
			output:      geom.XY{X: 1},
		},
		{
			description: "+ve unit in Y",
			input:       geom.XY{Y: 1},
			output:      geom.XY{Y: 1},
		},
		{
			description: "-ve unit in X",
			input:       geom.XY{X: -1},
			output:      geom.XY{X: -1},
		},
		{
			description: "-ve unit in Y",
			input:       geom.XY{Y: -1},
			output:      geom.XY{Y: -1},
		},
		{
			description: "non-aligned unit",
			input:       geom.XY{X: -1 / sqrt2, Y: 1 / sqrt2},
			output:      geom.XY{X: -1 / sqrt2, Y: 1 / sqrt2},
		},
		{
			description: "non-unit",
			input:       geom.XY{X: 1, Y: -2},
			output:      geom.XY{X: 1 / sqrt5, Y: -2 / sqrt5},
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
			got := tc.input.Unit()
			expectXYWithinTolerance(t, got, tc.output, 0.0000001)
		})
	}
}
