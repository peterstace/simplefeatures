package geom

import (
	"math"
	"strconv"
)

// appendFloat appends the decimal representation of f to dst and returns it.
func appendFloat(dst []byte, f float64) []byte {
	return strconv.AppendFloat(dst, f, 'f', -1, 64)
}

// ulpSize returns the distance from f to the float64 after f.
func ulpSize(f float64) float64 {
	u := math.Float64bits(f) + 1
	next := math.Float64frombits(u)
	return next - f
}
