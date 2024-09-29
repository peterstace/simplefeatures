package carto

import "math"

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
	tan  = math.Tan
	ln   = math.Log
	sin  = math.Sin
	cos  = math.Cos
	atan = math.Atan
	asin = math.Asin
	exp  = math.Exp
)
