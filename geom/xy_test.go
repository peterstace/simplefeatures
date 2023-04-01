package geom_test

import (
	"math"
	"strconv"
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

func TestXYCross(t *testing.T) {
	xy := func(x, y float64) geom.XY {
		return geom.XY{X: x, Y: y}
	}
	for i, tc := range []struct {
		u, v geom.XY
		want float64
	}{
		// Contains zero:
		{u: xy(1, 1), v: xy(0, 0), want: 0},
		{u: xy(0.1, 0.1), v: xy(0, 0), want: 0},

		// Not linearly independent:
		{u: xy(1, 1), v: xy(2, 2), want: 0},
		{u: xy(-1, -1), v: xy(2, 2), want: 0},

		// Reproduces numeric precision bug that occurred on aarch64 but *not* x86_64.
		{u: xy(0.2, 0.2), v: xy(0.1, 0.1), want: 0},

		// Linearly independent:
		{u: xy(1, 0), v: xy(0, 1), want: 1},
		{u: xy(0, 2), v: xy(1, 0), want: -2},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Run("fwd", func(t *testing.T) {
				got := tc.u.Cross(tc.v)
				expectFloat64Eq(t, got, tc.want)
			})
			t.Run("rev", func(t *testing.T) {
				got := tc.v.Cross(tc.u)
				want := -tc.want // Cross product is anti-commutative.
				expectFloat64Eq(t, got, want)
			})
		})
	}
}
