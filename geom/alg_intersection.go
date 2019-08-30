package geom

import (
	"fmt"
	"math"
	"sort"
)

func intersection(g1, g2 Geometry) Geometry {
	if g2.IsEmpty() {
		return g2
	}
	if g1.IsEmpty() {
		return g1
	}

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
		case MultiPoint:
			return intersectLineWithMultiPoint(g1, g2)
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

		e := (c.Y-d.Y)*(a.X-c.X) + (d.X-c.X)*(a.Y-c.Y)
		f := (d.X-c.X)*(a.Y-b.Y) - (a.X-b.X)*(d.Y-c.Y)
		g := (a.Y-b.Y)*(a.X-c.X) + (b.X-a.X)*(a.Y-c.Y)
		h := (d.X-c.X)*(a.Y-b.Y) - (a.X-b.X)*(d.Y-c.Y)
		// Division by zero is not possible, since the lines are not parallel.
		p := e / f
		q := g / h
		if p < 0 || p > 1 || q < 0 || q > 1 {
			// Intersection between lines occurs beyond line endpoints.
			return NewGeometryCollection(nil)
		}
		return NewPointXY(b.Sub(a).Scale(p).Add(a))
	}

	if o1 == collinear && o2 == collinear {
		if (!onSegment(a, b, c) && !onSegment(a, b, d)) && (!onSegment(c, d, a) && !onSegment(c, d, b)) {
			return NewGeometryCollection(nil)
		}

		// --------------------------------------------------
		// This block is to pass tests
		// for exmaple, LINESTRING(1 0, 0, 1) and (0 1,1 0) are same but it will fail test
		pts := make([]XY, 0, 4)
		pts = append(pts, a, b, c, d)
		rth := rightmostThenHighestIndex(pts)
		pts = append(pts[:rth], pts[rth+1:]...)
		ltl := leftmostThenLowestIndex(pts)
		pts = append(pts[:ltl], pts[ltl+1:]...)
		if pts[0].Equals(pts[1]) {
			return NewPointXY(pts[0])
		}
		// --------------------------------------------------

		return must(NewLineC(Coordinates{pts[leftmostThenLowestIndex(pts)]}, Coordinates{pts[rightmostThenHighestIndex(pts)]}))
	}

	return NewGeometryCollection(nil)
}

func intersectLineWithMultiPoint(ln Line, mp MultiPoint) Geometry {
	var pts []Point
	n := mp.NumPoints()
	for i := 0; i < n; i++ {
		pt := mp.PointN(i)
		if !pt.Intersection(ln).IsEmpty() {
			pts = append(pts, pt)
		}
	}
	if len(pts) == 1 {
		return pts[0]
	}
	return NewMultiPoint(pts)
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
	lhs := (point.coords.X - line.a.X) * (line.b.Y - line.a.Y)
	rhs := (point.coords.Y - line.a.Y) * (line.b.X - line.a.X)
	if lhs == rhs {
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
	mp1Set := make(map[XY]struct{})
	for _, pt := range mp1.pts {
		mp1Set[pt.Coordinates().XY] = struct{}{}
	}
	mp2Set := make(map[XY]struct{})
	for _, pt := range mp2.pts {
		mp2Set[pt.Coordinates().XY] = struct{}{}
	}

	interSet := make(map[XY]struct{})
	for pt := range mp1Set {
		if _, ok := mp2Set[pt]; ok {
			interSet[pt] = struct{}{}
		}
	}
	for pt := range mp2Set {
		if _, ok := mp1Set[pt]; ok {
			interSet[pt] = struct{}{}
		}
	}

	intersection := make([]Point, 0, len(interSet))
	for pt := range interSet {
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
		if ps[i].X > ps[rpi].X ||
			(ps[i].X == ps[rpi].X &&
				ps[i].Y > ps[rpi].Y) {
			rpi = i
		}
	}
	return rpi
}

// leftmostThenLowestIndex finds the index of the leftmost-then-lowest point.
func leftmostThenLowestIndex(ps []XY) int {
	rpi := 0
	for i := 1; i < len(ps); i++ {
		if ps[i].X < ps[rpi].X ||
			(ps[i].X == ps[rpi].X &&
				ps[i].Y < ps[rpi].Y) {
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
	if r.X <= math.Max(p.X, q.X) &&
		r.X >= math.Min(p.X, q.X) &&
		r.Y <= math.Max(p.Y, q.Y) &&
		r.Y >= math.Min(p.Y, q.Y) {
		return true
	}

	return false
}
