package geom

import (
	"testing"
)

func TestPolygonFlipCoordinates(t *testing.T) {
	poly := NewPolygonXY([]float64{1, 2, 3, 4, 1, 2})
	flipped := poly.FlipCoordinates()
	seq := flipped.ExteriorRing().Coordinates()
	xy0 := seq.GetXY(0)
	xy1 := seq.GetXY(1)
	if xy0 != (XY{2, 1}) || xy1 != (XY{4, 3}) {
		t.Errorf("expected [(2,1),(4,3)], got [%v,%v]", xy0, xy1)
	}
}
