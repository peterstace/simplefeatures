package simplefeatures_test

import . "github.com/peterstace/simplefeatures"

func Must(g Geometry, err error) Geometry {
	if err != nil {
		panic(err)
	}
	return g
}

func NewXY(x, y float64) XY {
	return XY{
		NewScalarFromFloat64(x),
		NewScalarFromFloat64(y),
	}
}
