package simplefeatures

import (
	"strconv"
	"testing"
)

func TestLinearRingValidation(t *testing.T) {
	xy := func(x, y float64) Coordinates {
		return Coordinates{XY: XY{x, y}}
	}
	for i, pts := range [][]Coordinates{
		{xy(0, 0), xy(1, 1), xy(0, 1)},                     // not closed
		{xy(0, 0), xy(1, 1), xy(0, 1), xy(1, 0), xy(0, 0)}, // not simple
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			_, err := NewLinearRing(pts)
			if err == nil {
				t.Error("expected error")
			}
			t.Log(err)
		})
	}
}
