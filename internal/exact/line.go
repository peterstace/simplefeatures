package exact

type Line struct {
	A, B XY64
}

type Intersection struct {
	Empty bool
	A, B  XY64
}

func LineIntersection(lineA, lineB Line) Intersection {
	// Algorithm from https://en.wikipedia.org/wiki/Line%E2%80%93line_intersection

	if lineA.A == lineA.B || lineB.A == lineB.B {
		panic("invalid line")
	}

	v1 := lineA.A.ToRat()
	v2 := lineA.B.ToRat()
	v3 := lineB.A.ToRat()
	v4 := lineB.B.ToRat()

	// d := (x1-x2)*(y3-y4)-(y1-y2)*(x3-x4)
	sub12 := v1.Sub(v2)
	sub34 := v3.Sub(v4)
	d := sub12.Cross(sub34)

	if d.Sign() == 0 {
		sub23 := v2.Sub(v3)
		if sub12.Cross(sub23).Sign() != 0 {
			return Intersection{Empty: true}
		}

		if v2.Less(v1) {
			v2, v1 = v1, v2
		}
		if v4.Less(v3) {
			v4, v3 = v3, v4
		}

		if v2.Less(v3) || v4.Less(v1) {
			return Intersection{Empty: true}
		}
		min := v2.Min(v4)
		max := v1.Max(v3)
		return Intersection{
			A: max.ToXY64(),
			B: min.ToXY64(),
		}
	}

	// t := [(x1-x3)*(y3-y4)-(y1-y3)*(x3-x4)] / d
	sub13 := v1.Sub(v3)
	t := div(sub13.Cross(sub34), d)

	// u := [(x2-x1)*(y1-y3)-(y2-y1)*(x1-x3)] / d
	sub21 := sub12.Neg()
	u := div(sub21.Cross(sub13), d)

	if inUnitInterval(t) && inUnitInterval(u) {
		pt := sub21.Scale(t).Add(v1).ToXY64()
		return Intersection{A: pt, B: pt}
	}

	return Intersection{Empty: true}
}
