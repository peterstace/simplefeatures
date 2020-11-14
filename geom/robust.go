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

func exactMul(a, b float64) (p, q float64) {
	a1 := clearLast27BitsOfMantissa(a)
	b1 := clearLast27BitsOfMantissa(b)
	a2 := a - a1
	b2 := b - b1
	p1 := a1 * b1
	p2 := a1 * b2
	p3 := a2 * b1
	p4 := a2 * b2
	var p5 float64
	if bit27OfMantissaSet(a2) && bit27OfMantissaSet(b2) {
		a3 := clear27thBitOfMantissa(a2)
		a4 := a2 - a3
		p4 = a3 * b2
		p5 := a4 * b2
		p4, p5 = exactSum(p4, p5)
	}
	p2, p3 = exactSum(p2, p3)
	p3, p4 = exactSum(p3, p4)
	p2, p3 = exactSum(p2, p3)
	p, p2 = exactSum(p1, p2)
	q = (p2 + p3) + (p4 + p5)
	return
}

func clearLast27BitsOfMantissa(f float64) float64 {
	u := math.Float64bits(f)
	u = u &^ 0x7ffffff
	return math.Float64frombits(u)
}

func clear27thBitOfMantissa(f float64) float64 {
	u := math.Float64bits(f)
	u = u &^ 0x2000000
	return math.Float64frombits(u)
}

func bit27OfMantissaSet(f float64) bool {
	u := math.Float64bits(f)
	return (u & 0x2000000) != 0
}
