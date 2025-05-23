package carto

import (
	"math"

	"github.com/peterstace/simplefeatures/geom"
)

// This file contains utility functions that make projection formulas terser.
// While terse code is usually _harder_ to read, the opposite is true for
// mathematical formulas.

func dtor(d float64) float64 {
	return d * π / 180
}

func rtod(r float64) float64 {
	return r * 180 / π
}

func rtodxy(λ, φ float64) geom.XY {
	return geom.XY{X: rtod(λ), Y: rtod(φ)}
}

const (
	π = math.Pi
)

func sqrt(x float64) float64 {
	return math.Sqrt(x)
}

func tan(x float64) float64 {
	return math.Tan(x)
}

func ln(x float64) float64 {
	return math.Log(x)
}

func sin(x float64) float64 {
	return math.Sin(x)
}

func sinh(x float64) float64 {
	return math.Sinh(x)
}

func cos(x float64) float64 {
	return math.Cos(x)
}

func cosh(x float64) float64 {
	return math.Cosh(x)
}

func atan(x float64) float64 {
	return math.Atan(x)
}

func atan2(y, x float64) float64 {
	return math.Atan2(y, x)
}

func atanh(x float64) float64 {
	return math.Atanh(x)
}

func asin(x float64) float64 {
	return math.Asin(x)
}

func acos(x float64) float64 {
	return math.Acos(x)
}

func exp(x float64) float64 {
	return math.Exp(x)
}

func pow(x, y float64) float64 {
	return math.Pow(x, y)
}

func sign(x float64) float64 {
	return math.Copysign(1, x)
}

func sec(x float64) float64 {
	return 1 / cos(x)
}

func cot(x float64) float64 {
	return 1 / tan(x)
}

func sq(x float64) float64 {
	return x * x
}

func abs(x float64) float64 {
	return math.Abs(x)
}
