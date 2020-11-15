package geom

import (
	"math"
)

// exactSum finds the exact sum (s) of a and b and its error (e). It uses the
// algorithm found here: https://core.ac.uk/download/pdf/191319061.pdf
func exactSum(a, b float64) (s, e float64) {
	if exponentFloat64(b) < exponentFloat64(a) {
		a, b = b, a
	}
	s = a + b
	e = (b - s) + a
	return
}

func exponentFloat64(f float64) int {
	u := math.Float64bits(f)
	biased := (u & 0x7ff0000000000000) >> 52
	return int(biased) - 1023
}

// exactMul finds the exact product (p) of a and b and its error (q). It uses
// the algorithm found here: https://core.ac.uk/download/pdf/191319061.pdf
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

func accurateDotProduct(u, v XY) float64 {
	px, qx := exactMul(u.X, v.X)
	py, qy := exactMul(u.Y, v.Y)
	return sumFloat64s(px, qx, py, qy)
}

// sumFloat64s sums together 4 floats using the algorithm described here:
// https://hal.inria.fr/inria-00517618/PDF/Yong-KangZhu2005b.pdf
func sumFloat64s(a, b, c, d float64) float64 {
	var pos, neg []float64
	addToSet := func(f float64) {
		if f > 0 {
			pos = append(pos, f)
		} else if f < 0 {
			neg = append(neg, f)
		}
	}
	addToSet(a)
	addToSet(b)
	addToSet(c)
	addToSet(d)

	var s, e1, e2 float64
	for {
		addToSet(e1)
		addToSet(e2)

		sPrime := s
		n1 := len(pos)
		var sPos float64
		for i := 0; i < n1; i++ {
			var e float64
			sPos, e = exactSum(pos[0], sPos)
			pos = pos[1:]
			addToSet(e)
		}
		s, e1 = exactSum(s, sPos)

		sDoublePrime := s
		n2 := len(neg)
		var sNeg float64
		for i := 0; i < n2; i++ {
			var e float64
			sNeg, e = exactSum(neg[0], sNeg)
			neg = neg[1:]
			addToSet(e)
		}
		s, e2 = exactSum(s, sNeg)

		if s == sPrime && s == sDoublePrime {
			break
		}
	}
	return s + (e1 + e2)
}
