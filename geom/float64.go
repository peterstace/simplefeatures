package geom

import (
	"math"
	"strconv"
)

// appendFloat appends the decimal representation of f to dst and returns it.
func appendFloat(dst []byte, f float64) []byte {
	return strconv.AppendFloat(dst, f, 'f', -1, 64)
}

// truncateMantissa returns the input float64 but with the least significant
// byte of its mantissa set to zero.
func truncateMantissa(f float64) float64 {
	u := math.Float64bits(f)
	u &= ^uint64(0xff) // zero out last byte
	return math.Float64frombits(u)
}

// truncateMantissaXY returns the input XY but with the least significant byte
// of the mantissa of each of X and Y set to zero.
func truncateMantissaXY(xy XY) XY {
	return XY{
		truncateMantissa(xy.X),
		truncateMantissa(xy.Y),
	}
}

// addULPs returns the float64 that is delta ULPs away from f. Delta may be negative.
func addULPs(f float64, delta int64) float64 {
	u := math.Float64bits(f)
	u += uint64(delta)
	return math.Float64frombits(u)
}

// ulpSize returns the distance to the float64 after f.
func ulpSize(f float64) float64 {
	u := math.Float64bits(f) + 1
	next := math.Float64frombits(u)
	return next - f
}
