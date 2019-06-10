package simplefeatures_test

import (
	"strconv"
	"testing"

	. "github.com/peterstace/simplefeatures"
)

func TestLineStringValidation(t *testing.T) {
	pt := MustNewPoint
	for i, pts := range [][]Point{
		{pt(0, 0)},
		{pt(1, 1)},
		{pt(0, 0), pt(0, 0)},
		{pt(1, 1), pt(1, 1)},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			_, err := NewLineString(pts)
			if err == nil {
				t.Error("expected error")
			}
		})
	}
}
