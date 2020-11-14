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

func exponentFloat64(f float64) int {
	u := math.Float64bits(f)
	biased := (u & 0x7ff0000000000000) >> 52
	return int(biased) - 1023
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

func accurateDotProduct(u, v XY) float64 {
	// Step 1
	var pSet, nSet, qSet []float64

	// Step 2
	for i := 0; i < 2; i++ {
		var ai, bi float64
		switch i {
		case 0:
			ai = u.X
			bi = v.X
		case 1:
			ai = u.Y
			bi = v.Y
		default:
			panic(i)
		}

		// Step 2.a
		pi, qi := exactMul(ai, bi)

		// Step 2.b
		if pi > 0 {
			pSet = append(pSet, pi)
		} else if pi < 0 {
			nSet = append(nSet, pi)
		}

		// Step 2.c
		if qi != 0 {
			qSet = append(qSet, qi)
		}
	}

	// Step 3
	s := 0.0
	e1 := 0.0
	e2 := 0.0
	first := true

	// Step 4
step4:
	if e1 > 0 {
		pSet = append(pSet, e1)
	} else if e1 < 0 {
		nSet = append(nSet, e1)
	}

	// Step 5
	if e2 > 0 {
		pSet = append(pSet, e2)
	} else if e2 < 0 {
		nSet = append(nSet, e2)
	}

	// Step 6
	n1 := len(pSet)
	sPos := 0.0

	// Step 7
	for i := 0; i < n1; i++ {
		// Step 7.a
		a := pSet[0]
		pSet = pSet[1:]
		b := sPos

		// Step 7.b
		var e float64
		sPos, e = exactSum(a, b)
		if e > 0 {
			pSet = append(pSet, e)
		} else if e < 0 {
			nSet = append(nSet, e)
		}
	}

	// Step 8
	s, e1 = exactSum(s, sPos)

	// Step 9
	n2 := len(nSet)
	sNeg := 0.0

	// Step 10
	for i := 0; i < n2; i++ {
		// Step 10.a
		a := nSet[0]
		nSet = nSet[1:]
		b := sNeg

		// Step 10.b
		var e float64
		sNeg, e = exactSum(a, b)
		if e > 0 {
			pSet = append(pSet, e)
		} else if e < 0 {
			nSet = append(nSet, e)
		}
	}

	// Step 11
	s, e2 = exactSum(s, sNeg)

	// Step 12
	if first {
		// Step 12.a
		for _, q := range qSet {
			if q > 0 {
				pSet = append(pSet, q)
			} else if q < 0 {
				nSet = append(nSet, q)
			}
		}
		first = false

		// Step 12.b
		goto step4
	}

	// Step 13 and 14 (this is a bit different compared to the paper)
	if len(pSet)+len(nSet) > 0 {
		goto step4
	}

	// Step 15
	s = s + (e1 + e2)
	return s
}
