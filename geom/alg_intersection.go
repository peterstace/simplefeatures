package geom

import (
	"fmt"
	"math"
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
		case LineString:
			return intersectPointWithLineString(g1, g2)
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

	o1 := orientation(a, b, c)
	o2 := orientation(a, b, d)
	o3 := orientation(c, d, a)
	o4 := orientation(c, d, b)

	if o1 != o2 &&
		(o1 == leftTurn || o1 == rightTurn) &&
		(o2 == leftTurn || o2 == rightTurn) &&
		o3 != o4 &&
		(o3 == leftTurn || o3 == rightTurn) &&
		(o4 == leftTurn || o4 == rightTurn) {
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

	if o1 != o2 && o3 != o4 {
		if o1 == collinear {
			return NewPointXY(c)
		}
		if o2 == collinear {
			return NewPointXY(d)
		}
		if o3 == collinear {
			return NewPointXY(a)
		}
		if o4 == collinear {
			return NewPointXY(b)
		}
	}

	if o1 == collinear && o2 == collinear && o3 == collinear && o4 == collinear {
		if (!onSegment(a, b, c) && !onSegment(a, b, d)) && (!onSegment(c, d, a) && !onSegment(c, d, b)) {
			return NewGeometryCollection(nil)
		}
		pts := make([]XY, 0, 4)
		pts = append(pts, a, b, c, d)
		rth := rightmostThenHighestIndex(pts)
		pts = append(pts[:rth], pts[rth+1:]...)
		ltl := leftmostThenLowestIndex(pts)
		pts = append(pts[:ltl], pts[ltl+1:]...)
		if pts[0].Equals(pts[1]) {
			return NewPointXY(pts[0])
		}

		return must(NewLineC(Coordinates{pts[leftmostThenLowestIndex(pts)]}, Coordinates{pts[rightmostThenHighestIndex(pts)]}))
	}

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

func intersectPointWithLineString(pt Point, ls LineString) Geometry {
	for _, ln := range ls.lines {
		g := intersectPointWithLine(pt, ln)
		if !g.IsEmpty() {
			return g
		}
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

// rightmostThenHighest finds the rightmost-then-highest point
func rightmostThenHighest(ps []XY) XY {
	return ps[rightmostThenHighestIndex(ps)]
}

// rightmostThenHighestIndex finds the rightmost-then-highest point
func rightmostThenHighestIndex(ps []XY) int {
	rpi := 0
	for i := 1; i < len(ps); i++ {
		if ps[i].X.GT(ps[rpi].X) ||
			(ps[i].X.Equals(ps[rpi].X) &&
				ps[i].Y.GT(ps[rpi].Y)) {
			rpi = i
		}
	}
	return rpi
}

// leftmostThenLowestIndex finds the index of the leftmost-then-lowest point.
func leftmostThenLowestIndex(ps []XY) int {
	rpi := 0
	for i := 1; i < len(ps); i++ {
		if ps[i].X.LT(ps[rpi].X) ||
			(ps[i].X.Equals(ps[rpi].X) &&
				ps[i].Y.LT(ps[rpi].Y)) {
			rpi = i
		}
	}
	return rpi
}

// leftmostThenLowest finds the lowest-then-leftmost point
func leftmostThenLowest(ps []XY) XY {
	return ps[leftmostThenLowestIndex(ps)]
}

// onSegement check if point r on the segment formed by p and q.
// p, q and r should be collinear
func onSegment(p XY, q XY, r XY) bool {
	if r.X.AsFloat() <= math.Max(p.X.AsFloat(), q.X.AsFloat()) &&
		r.X.AsFloat() >= math.Min(p.X.AsFloat(), q.X.AsFloat()) &&
		r.Y.AsFloat() <= math.Max(p.Y.AsFloat(), q.Y.AsFloat()) &&
		r.Y.AsFloat() >= math.Min(p.Y.AsFloat(), q.Y.AsFloat()) {
		return true
	}

	return false
}
