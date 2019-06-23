package simplefeatures

func intersection(g1, g2 Geometry) Geometry {
	if rank(g1) > rank(g2) {
		g1, g2 = g2, g1
	}
	switch g1 := g1.(type) {
	case Point:
		switch g2 := g2.(type) {
		case Line:
			return intersectPointWithLine(g1, g2)
		}
	case Line:
		switch g2 := g2.(type) {
		case Line:
			return intersectLineWithLine(g1, g2)
		}
	case LinearRing:
		switch g2 := g2.(type) {
		case LinearRing:
			return intersectLinearRingWithLinearRing(g1, g2)
		}
	}
	panic("not implemented")
}

func intersectLineWithLine(n1, n2 Line) Geometry {
	a := n1.a.XY
	b := n1.b.XY
	c := n2.a.XY
	d := n2.b.XY

	if parallel := seq(xycross(xysub(b, a), xysub(d, c)), zero); !parallel {
		e := sadd(smul(ssub(c.Y, d.Y), ssub(a.X, c.X)), smul(ssub(d.X, c.X), ssub(a.Y, c.Y)))
		//e := (c.Y-d.Y)*(a.X-c.X) + (d.X-c.X)*(a.Y-c.Y)
		f := ssub(smul(ssub(d.X, c.X), ssub(a.Y, b.Y)), smul(ssub(a.X, b.X), ssub(d.Y, c.Y)))
		//f := (d.X-c.X)*(a.Y-b.Y) - (a.X-b.X)*(d.Y-c.Y)
		g := sadd(smul(ssub(a.Y, b.Y), ssub(a.X, c.X)), smul(ssub(b.X, a.X), ssub(a.Y, c.Y)))
		//g := (a.Y-b.Y)*(a.X-c.X) + (b.X-a.X)*(a.Y-c.Y)
		h := ssub(smul(ssub(d.X, c.X), ssub(a.Y, b.Y)), smul(ssub(a.X, b.X), ssub(d.Y, c.Y)))
		//h := (d.X-c.X)*(a.Y-b.Y) - (a.X-b.X)*(d.Y-c.Y)
		// Division by zero is not possible, since the lines are not parallel.
		p := sdiv(e, f)
		//p := e / f
		q := sdiv(g, h)
		//q := g / h
		if slt(p, zero) || sgt(p, one) || slt(q, zero) || sgt(q, one) {
			//if p < 0 || p > 1 || q < 0 || q > 1 {
			// Intersection between lines occurs beyond line endpoints.
			return NewGeometryCollection(nil)
		}
		pt, err := NewPoint(
			sadd(a.X, smul(p, ssub(b.X, a.X))),
			//a.X+p*(b.X-a.X),
			sadd(a.Y, smul(p, ssub(b.Y, a.Y))),
			//a.Y+p*(b.Y-a.Y),
		)
		if err != nil {
			panic(err)
		}
		return pt
	}

	// TODO: invert if to un-indent flow.
	if colinear := seq(xycross(xysub(b, a), xysub(d, a)), zero); colinear {
		// TODO: use a proper bbox type
		abBB := bbox{
			min: XY{smin(a.X, b.X), smin(a.Y, b.Y)},
			//min: XY{math.Min(a.X, b.X), math.Min(a.Y, b.Y)},
			max: XY{smax(a.X, b.X), smax(a.Y, b.Y)},
			//max: XY{math.Max(a.X, b.X), math.Max(a.Y, b.Y)},
		}
		cdBB := bbox{
			min: XY{smin(c.X, d.X), smin(c.Y, d.Y)},
			//min: XY{math.Min(c.X, d.X), math.Min(c.Y, d.Y)},
			max: XY{smax(c.X, d.X), smax(c.Y, d.Y)},
			//max: XY{math.Max(c.X, d.X), math.Max(c.Y, d.Y)},
		}
		if sgt(abBB.min.X, cdBB.max.X) || slt(abBB.max.X, cdBB.min.X) ||
			sgt(abBB.min.Y, cdBB.max.Y) || slt(abBB.max.Y, cdBB.min.Y) {
			//if abBB.min.X > cdBB.max.X || abBB.max.X < cdBB.min.X ||
			//abBB.min.Y > cdBB.max.Y || abBB.max.Y < cdBB.min.Y {
			// Line segments don't overlap at all.
			return NewGeometryCollection(nil)
		}

		// TODO: the checks for intersecting at a point could go above the
		// overlap case. They don't need to use the bounding box, because we
		// can just do a pairwise check on the endpoints for each 4
		// combinations.

		if seq(abBB.max.X, cdBB.min.X) && seq(abBB.min.Y, cdBB.max.Y) {
			//if abBB.max.X == cdBB.min.X && abBB.min.Y == cdBB.max.Y {
			// Line segments overlap at a point.
			pt, err := NewPoint(abBB.max.X, abBB.min.Y)
			if err != nil {
				panic(err)
			}
			return pt
		}

		if seq(cdBB.max.X, abBB.min.X) && seq(cdBB.min.Y, abBB.max.Y) {
			//if cdBB.max.X == abBB.min.X && cdBB.min.Y == abBB.max.Y {
			// Line segments overlap at a point.
			pt, err := NewPoint(cdBB.max.X, cdBB.min.Y)
			if err != nil {
				panic(err)
			}
			return pt
		}

		if xyeq(abBB.max, cdBB.min) {
			//if abBB.max == cdBB.min {
			// Line segments overlap at a point.
			pt, err := NewPoint(abBB.max.X, abBB.max.Y)
			if err != nil {
				panic(err)
			}
			return pt
		}
		if xyeq(cdBB.max, abBB.min) {
			//if cdBB.max == abBB.min {
			// Line segments overlap at a point.
			pt, err := NewPoint(cdBB.max.X, cdBB.max.Y)
			if err != nil {
				panic(err)
			}
			return pt
		}

		// Line segments overlap over a line segment.
		bb := bbox{
			min: XY{
				smax(abBB.min.X, cdBB.min.X),
				//math.Max(abBB.min.X, cdBB.min.X),
				smax(abBB.min.Y, cdBB.min.Y),
				//math.Max(abBB.min.Y, cdBB.min.Y),
			},
			max: XY{
				smin(abBB.max.X, cdBB.max.X),
				//math.Min(abBB.max.X, cdBB.max.X),
				smin(abBB.max.Y, cdBB.max.Y),
				//math.Min(abBB.max.Y, cdBB.max.Y),
			},
		}
		var (
			u    = XY{bb.min.X, bb.min.Y}
			v    = XY{bb.max.X, bb.max.Y}
			rise = ssub(b.Y, a.Y)
			run  = ssub(b.X, a.X)
		)
		if sgt(rise, zero) && slt(run, zero) || slt(rise, zero) && sgt(run, zero) {
			//if rise > 0 && run < 0 || rise < 0 && run > 0 {
			u.X, v.X = v.X, u.X
		}

		ln, err := NewLine(Coordinates{u}, Coordinates{v})
		if err != nil {
			panic(err)
		}
		return ln
	}

	// Parrallel but not colinear, so cannot intersect anywhere.
	return NewGeometryCollection(nil)
}

type bbox struct {
	min, max XY
}

func intersectLinearRingWithLinearRing(r1, r2 LinearRing) Geometry {
	// TODO: This should be able to be a bit more generic, e.g. apply to line
	// strings instead.
	var collection []Geometry
	for _, ln1 := range r1.ls.lines {
		for _, ln2 := range r2.ls.lines {
			inter := ln1.Intersection(ln2)
			if !inter.IsEmpty() {
				collection = append(collection, inter)
			}
		}
	}
	return canonicalise(collection)
}

func intersectPointWithLine(point Point, line Line) Geometry {
	// TODO: use envelope instead
	if slt(point.coords.X, smin(line.a.X, line.b.X)) ||
		sgt(point.coords.X, smax(line.a.X, line.b.X)) ||
		slt(point.coords.Y, smin(line.a.Y, line.b.Y)) ||
		sgt(point.coords.Y, smax(line.a.Y, line.b.Y)) {
		//if point.coords.X < math.Min(line.a.X, line.b.X) ||
		//point.coords.X > math.Max(line.a.X, line.b.X) ||
		//point.coords.Y < math.Min(line.a.Y, line.b.Y) ||
		//point.coords.Y > math.Max(line.a.Y, line.b.Y) {
		return NewEmptyPoint()
	}
	lhs := smul(ssub(point.coords.X, line.a.X), ssub(line.b.Y, line.a.Y))
	//lhs := (point.coords.X - line.a.X) * (line.b.Y - line.a.Y)
	rhs := smul(ssub(point.coords.Y, line.a.Y), ssub(line.b.X, line.a.X))
	//rhs := (point.coords.Y - line.a.Y) * (line.b.X - line.a.X)
	if seq(lhs, rhs) {
		//if lhs == rhs {
		return point
	}
	return NewEmptyPoint()
}
