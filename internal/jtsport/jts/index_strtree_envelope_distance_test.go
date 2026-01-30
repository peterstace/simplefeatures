package jts_test

import (
	"testing"

	"github.com/peterstace/simplefeatures/internal/jtsport/jts"
	"github.com/peterstace/simplefeatures/internal/jtsport/junit"
)

func TestEnvelopeDistance_Disjoint(t *testing.T) {
	checkEnvelopeDistance(t, jts.Geom_NewEnvelopeFromXY(0, 10, 0, 10), jts.Geom_NewEnvelopeFromXY(20, 30, 20, 40), 50)
}

func TestEnvelopeDistance_Overlapping(t *testing.T) {
	checkEnvelopeDistance(t, jts.Geom_NewEnvelopeFromXY(0, 30, 0, 30), jts.Geom_NewEnvelopeFromXY(20, 30, 20, 40), 50)
}

func TestEnvelopeDistance_Crossing(t *testing.T) {
	checkEnvelopeDistance(t, jts.Geom_NewEnvelopeFromXY(0, 40, 10, 20), jts.Geom_NewEnvelopeFromXY(20, 30, 0, 30), 50)
}

func TestEnvelopeDistance_Crossing2(t *testing.T) {
	checkEnvelopeDistance(t, jts.Geom_NewEnvelopeFromXY(0, 10, 4, 6), jts.Geom_NewEnvelopeFromXY(4, 6, 0, 10), 14.142135623730951)
}

func checkEnvelopeDistance(t *testing.T, env1, env2 *jts.Geom_Envelope, expected float64) {
	t.Helper()
	result := jts.IndexStrtree_EnvelopeDistance_MaximumDistance(env1, env2)
	junit.AssertEquals(t, expected, result)
}
