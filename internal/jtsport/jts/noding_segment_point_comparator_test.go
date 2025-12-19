package jts_test

import (
	"testing"

	"github.com/peterstace/simplefeatures/internal/jtsport/jts"
)

func TestSegmentPointComparatorOctant0(t *testing.T) {
	checkNodePosition(t, 0, 1, 1, 2, 2, -1)
	checkNodePosition(t, 0, 1, 0, 1, 1, -1)
}

func checkNodePosition(t *testing.T, octant int, x0, y0, x1, y1 float64, expectedPositionValue int) {
	t.Helper()
	posValue := jts.Noding_SegmentPointComparator_Compare(
		octant,
		jts.Geom_NewCoordinateWithXY(x0, y0),
		jts.Geom_NewCoordinateWithXY(x1, y1),
	)
	if posValue != expectedPositionValue {
		t.Errorf("expected %d, got %d", expectedPositionValue, posValue)
	}
}
