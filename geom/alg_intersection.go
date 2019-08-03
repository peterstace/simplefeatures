package geom

import (
	"fmt"
	"sort"
)

func intersection(g1, g2 Geometry) Geometry {
	if rank(g1) > rank(g2) {
		g1, g2 = g2, g1
	}
	switch g1 := g1.(type) {
	case Point:
		switch g2 := g2.(type) {
		case Point:
			return intersectPointWithPoint(g1, g2)
		case Line:
			return intersectPointWithLine(g1, g2)
		case MultiPoint:
			return intersectPointWithMultiPoint(g1, g2)
		}
	case Line:
		switch g2 := g2.(type) {
		case Line:
			return intersectLineWithLine(g1, g2)
		}
	case LineString:
		switch g2 := g2.(type) {
		case LineString:
			return intersectMultiLineStringWithMultiLineString(
				NewMultiLineString([]LineString{g1}),
				NewMultiLineString([]LineString{g2}),
			)
		case LinearRing:
			return intersectMultiLineStringWithMultiLineString(
				NewMultiLineString([]LineString{g1}),
				NewMultiLineString([]LineString{g2.ls}),
			)
		case MultiLineString:
			return intersectMultiLineStringWithMultiLineString(
				NewMultiLineString([]LineString{g1}),
				g2,
			)
		}
	case LinearRing:
		switch g2 := g2.(type) {
		case LinearRing:
			return intersectMultiLineStringWithMultiLineString(
				NewMultiLineString([]LineString{g1.ls}),
				NewMultiLineString([]LineString{g2.ls}),
			)
		case MultiLineString:
			return intersectMultiLineStringWithMultiLineString(
				NewMultiLineString([]LineString{g1.ls}),
				g2,
			)
		}
	case MultiPoint:
		switch g2 := g2.(type) {
		case MultiPoint:
			return intersectMultiPointWithMultiPoint(g1, g2)
		}
	case MultiLineString:
		switch g2 := g2.(type) {
		case MultiLineString:
			return intersectMultiLineStringWithMultiLineString(g1, g2)
		}
	}

	panic(fmt.Sprintf("not implemented: intersection with %T and %T", g1, g2))
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
		return NewPointXY(b.Sub(a).Scale(p).Add(a))
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
			return NewPointS(abBB.max.X, abBB.min.Y)
		}

		if cdBB.max.X.Equals(abBB.min.X) && cdBB.min.Y.Equals(abBB.max.Y) {
			// Line segments overlap at a point.
			return NewPointS(cdBB.max.X, cdBB.min.Y)
		}

		if abBB.max.Equals(cdBB.min) {
			// Line segments overlap at a point.
			return NewPointXY(abBB.max)
		}
		if cdBB.max.Equals(abBB.min) {
			// Line segments overlap at a point.
			return NewPointXY(cdBB.max)
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

		return must(NewLineC(Coordinates{u}, Coordinates{v}))
	}

	// Parrallel but not colinear, so cannot intersect anywhere.
	return NewGeometryCollection(nil)
}

type bbox struct {
	min, max XY
}

func intersectMultiLineStringWithMultiLineString(mls1, mls2 MultiLineString) Geometry {
	var collection []Geometry
	for _, ls1 := range mls1.lines {
		for _, ln1 := range ls1.lines {
			for _, ls2 := range mls2.lines {
				for _, ln2 := range ls2.lines {
					inter := ln1.Intersection(ln2)
					if !inter.IsEmpty() {
						collection = append(collection, inter)
					}
				}
			}
		}
	}
	return canonicalise(collection)
}

func intersectPointWithLine(point Point, line Line) Geometry {
	env, ok := line.Envelope()
	if !ok {
		panic("line must have envelope")
	}
	if !env.IntersectsPoint(point.coords.XY) {
		return NewEmptyPoint()
	}
	lhs := point.coords.X.Sub(line.a.X).Mul(line.b.Y.Sub(line.a.Y))
	rhs := point.coords.Y.Sub(line.a.Y).Mul(line.b.X.Sub(line.a.X))
	if lhs.Equals(rhs) {
		return point
	}
	return NewEmptyPoint()
}

func intersectMultiPointWithMultiPoint(mp1, mp2 MultiPoint) Geometry {
	mp1Set := newXYSet()
	for _, pt := range mp1.pts {
		mp1Set.add(pt.coords.XY)
	}
	mp2Set := newXYSet()
	for _, pt := range mp2.pts {
		mp2Set.add(pt.coords.XY)
	}

	allSet := newXYSet()
	for _, pt := range mp1Set {
		if mp2Set.contains(pt) {
			allSet.add(pt)
		}
	}
	for _, pt := range mp2Set {
		if mp1Set.contains(pt) {
			allSet.add(pt)
		}
	}

	intersection := make([]Point, 0, len(allSet))
	for _, pt := range allSet {
		intersection = append(intersection, NewPointXY(pt))
	}
	sort.Slice(intersection, func(i, j int) bool {
		return intersection[i].coords.XY.Less(intersection[j].coords.XY)
	})

	if len(intersection) == 1 {
		return intersection[0]
	}
	return NewMultiPoint(intersection)
}

func intersectPointWithMultiPoint(point Point, mp MultiPoint) Geometry {
	if mp.IsEmpty() {
		return mp
	}
	for _, pt := range mp.pts {
		if pt.Equals(point) {
			return NewPointXY(point.coords.XY)
		}
	}
	return NewGeometryCollection(nil)
}

func intersectPointWithPoint(pt1, pt2 Point) Geometry {
	if pt1.Equals(pt2) {
		return NewPointXY(pt1.coords.XY)
	}
	return NewGeometryCollection(nil)
}
