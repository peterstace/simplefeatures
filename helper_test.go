package simplefeatures_test

import . "github.com/peterstace/simplefeatures"

func NewXY(x, y float64) XY {
	return XY{
		NewScalarFromFloat64(x),
		NewScalarFromFloat64(y),
	}
}
