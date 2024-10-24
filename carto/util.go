package carto

import (
	"math"

	"github.com/peterstace/simplefeatures/geom"
)

func dtor(d float64) float64 {
	return d * π / 180
}

func rtod(r float64) float64 {
	return r * 180 / π
}

const (
	π = math.Pi
)

// TODO: Make these funcs rather than vars.
var (
	sqrt  = math.Sqrt
	tan   = math.Tan
	ln    = math.Log
	sin   = math.Sin
	cos   = math.Cos
	atan  = math.Atan
	atan2 = math.Atan2
	asin  = math.Asin
	acos  = math.Acos
	exp   = math.Exp
)

func xy(x, y float64) geom.XY {
	return geom.XY{X: x, Y: y}
}

func sign(x float64) float64 {
	return math.Copysign(1, x)
}
