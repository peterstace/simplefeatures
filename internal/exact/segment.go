package exact

type Segment struct {
	A, B XYRat
}

type Intersection struct {
	Empty bool
	A, B  XYRat
}

func SegmentIntersection(segA, segB Segment) Intersection {
	// Algorithm from https://en.wikipedia.org/wiki/Line%E2%80%93line_intersection

	if segA.A == segA.B || segB.A == segB.B {
		panic("invalid segment")
	}

	v1 := segA.A
	v2 := segA.B
	v3 := segB.A
	v4 := segB.B

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
		return Intersection{
			A: v1.Max(v3),
			B: v2.Min(v4),
		}
	}

	// t := [(x1-x3)*(y3-y4)-(y1-y3)*(x3-x4)] / d
	sub13 := v1.Sub(v3)
	t := div(sub13.Cross(sub34), d)

	// u := [(x2-x1)*(y1-y3)-(y2-y1)*(x1-x3)] / d
	sub21 := sub12.Neg()
	u := div(sub21.Cross(sub13), d)

	if inUnitInterval(t) && inUnitInterval(u) {
		pt := sub21.Scale(t).Add(v1)
		return Intersection{A: pt, B: pt}
	}

	return Intersection{Empty: true}
}
