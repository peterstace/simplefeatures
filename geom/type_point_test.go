package geom

import (
	"testing"
)

func TestPointFlipCoordinates(t *testing.T) {
	p := NewPointXY(1, 2)
	flipped := p.FlipCoordinates()
	xy, _ := flipped.XY()
	if xy != (XY{2, 1}) {
		t.Errorf("expected (2,1), got %v", xy)
	}
}
