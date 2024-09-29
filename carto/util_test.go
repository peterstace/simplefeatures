package carto_test

import (
	"math"
	"testing"

	"github.com/peterstace/simplefeatures/geom"
)

func expectXYWithinTolerance(tb testing.TB, got, want geom.XY, tolerance float64) {
	tb.Helper()
	if delta := math.Abs(got.Sub(want).Length()); delta > tolerance {
		tb.Errorf("\ngot:  %v\nwant: %v\n", got, want)
	}
}
