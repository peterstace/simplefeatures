package geom

import (
	"testing"
)

func TestLineStringFlipCoordinates(t *testing.T) {
	ls := NewLineStringXY(1, 2, 3, 4)
	flipped := ls.FlipCoordinates()
	seq := flipped.Coordinates()
	xy0 := seq.GetXY(0)
	xy1 := seq.GetXY(1)
	if xy0 != (XY{2, 1}) || xy1 != (XY{4, 3}) {
		t.Errorf("expected [(2,1),(4,3)], got [%v,%v]", xy0, xy1)
	}
}
