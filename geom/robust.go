package geom

import "math"

func exactSum(a, b float64) (s, e float64) {
	if exponentFloat64(b) < exponentFloat64(a) {
		a, b = b, a
	}
	s = a + b
	e = (b - s) + a
	return
}

func exponentFloat64(f float64) uint64 {
	u := math.Float64bits(f)
	return (u & 0x7ff0000000000000) >> 52
}
