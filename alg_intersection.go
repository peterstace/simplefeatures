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

	if parallel := b.Sub(a).Cross(d.Sub(c)).Equals(zero); !parallel {
		e := c.Y.Sub(d.Y).Mul(a.X.Sub(c.X)).Add(d.X.Sub(c.X).Mul(a.Y.Sub(c.Y)))
		f := d.X.Sub(c.X).Mul(a.Y.Sub(b.Y)).Sub(a.X.Sub(b.X).Mul(d.Y.Sub(c.Y)))
		g := a.Y.Sub(b.Y).Mul(a.X.Sub(c.X)).Add(b.X.Sub(a.X).Mul(a.Y.Sub(c.Y)))
		h := d.X.Sub(c.X).Mul(a.Y.Sub(b.Y)).Sub(a.X.Sub(b.X).Mul(d.Y.Sub(c.Y)))
		// Division by zero is not possible, since the lines are not parallel.
		p := e.Div(f)
		q := g.Div(h)
		if p.LT(zero) || p.GT(one) || q.LT(zero) || q.GT(one) {
			// Intersection between lines occurs beyond line endpoints.
			return NewGeometryCollection(nil)
		}
		pt, err := NewPoint(
			a.X.Add(p.Mul(b.X.Sub(a.X))),
			a.Y.Add(p.Mul(b.Y.Sub(a.Y))),
		)
		if err != nil {
			panic(err)
		}
		return pt
	}

	// TODO: invert if to un-indent flow.
	if colinear := b.Sub(a).Cross(d.Sub(a)).Equals(zero); colinear {
		// TODO: use a proper bbox type
		abBB := bbox{
			min: XY{a.X.Min(b.X), a.Y.Min(b.Y)},
			max: XY{a.X.Max(b.X), a.Y.Max(b.Y)},
		}
		cdBB := bbox{
			min: XY{c.X.Min(d.X), c.Y.Min(d.Y)},
			max: XY{c.X.Max(d.X), c.Y.Max(d.Y)},
		}
		if abBB.min.X.GT(cdBB.max.X) || abBB.max.X.LT(cdBB.min.X) ||
			abBB.min.Y.GT(cdBB.max.Y) || abBB.max.Y.LT(cdBB.min.Y) {
			// Line segments don't overlap at all.
			return NewGeometryCollection(nil)
		}

		// TODO: the checks for intersecting at a point could go above the
		// overlap case. They don't need to use the bounding box, because we
		// can just do a pairwise check on the endpoints for each 4
		// combinations.

		if abBB.max.X.Equals(cdBB.min.X) && abBB.min.Y.Equals(cdBB.max.Y) {
			// Line segments overlap at a point.
			pt, err := NewPoint(abBB.max.X, abBB.min.Y)
			if err != nil {
				panic(err)
			}
			return pt
		}

		if cdBB.max.X.Equals(abBB.min.X) && cdBB.min.Y.Equals(abBB.max.Y) {
			// Line segments overlap at a point.
			pt, err := NewPoint(cdBB.max.X, cdBB.min.Y)
			if err != nil {
				panic(err)
			}
			return pt
		}

		if abBB.max.Equals(cdBB.min) {
			// Line segments overlap at a point.
			pt, err := NewPoint(abBB.max.X, abBB.max.Y)
			if err != nil {
				panic(err)
			}
			return pt
		}
		if cdBB.max.Equals(abBB.min) {
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
				abBB.min.X.Max(cdBB.min.X),
				abBB.min.Y.Max(cdBB.min.Y),
			},
			max: XY{
				abBB.max.X.Min(cdBB.max.X),
				abBB.max.Y.Min(cdBB.max.Y),
			},
		}
		var (
			u    = XY{bb.min.X, bb.min.Y}
			v    = XY{bb.max.X, bb.max.Y}
			rise = b.Y.Sub(a.Y)
			run  = b.X.Sub(a.X)
		)
		if rise.GT(zero) && run.LT(zero) || rise.LT(zero) && run.GT(zero) {
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
	if point.coords.X.LT(line.a.X.Min(line.b.X)) ||
		point.coords.X.GT(line.a.X.Max(line.b.X)) ||
		point.coords.Y.LT(line.a.Y.Min(line.b.Y)) ||
		point.coords.Y.GT(line.a.Y.Max(line.b.Y)) {
		return NewEmptyPoint()
	}
	lhs := point.coords.X.Sub(line.a.X).Mul(line.b.Y.Sub(line.a.Y))
	rhs := point.coords.Y.Sub(line.a.Y).Mul(line.b.X.Sub(line.a.X))
	if lhs.Equals(rhs) {
		return point
	}
	return NewEmptyPoint()
}
