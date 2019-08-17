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

	if parallel := b.Sub(a).Cross(d.Sub(c)) == 0; !parallel {
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

	// TODO: invert if to un-indent flow.
	if colinear := b.Sub(a).Cross(d.Sub(a)) == 0; colinear {
		// TODO: use a proper bbox type
		abBB := bbox{
			min: XY{math.Min(a.X, b.X), math.Min(a.Y, b.Y)},
			max: XY{math.Max(a.X, b.X), math.Max(a.Y, b.Y)},
		}
		cdBB := bbox{
			min: XY{math.Min(c.X, d.X), math.Min(c.Y, d.Y)},
			max: XY{math.Max(c.X, d.X), math.Max(c.Y, d.Y)},
		}
		if abBB.min.X > cdBB.max.X || abBB.max.X < cdBB.min.X ||
			abBB.min.Y > cdBB.max.Y || abBB.max.Y < cdBB.min.Y {
			// Line segments don't overlap at all.
			return NewGeometryCollection(nil)
		}

		// TODO: the checks for intersecting at a point could go above the
		// overlap case. They don't need to use the bounding box, because we
		// can just do a pairwise check on the endpoints for each 4
		// combinations.

		if abBB.max.X == cdBB.min.X && abBB.min.Y == cdBB.max.Y {
			// Line segments overlap at a point.
			return NewPointS(abBB.max.X, abBB.min.Y)
		}

		if cdBB.max.X == abBB.min.X && cdBB.min.Y == abBB.max.Y {
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
				math.Max(abBB.min.X, cdBB.min.X),
				math.Max(abBB.min.Y, cdBB.min.Y),
			},
			max: XY{
				math.Min(abBB.max.X, cdBB.max.X),
				math.Min(abBB.max.Y, cdBB.max.Y),
			},
		}
		var (
			u    = XY{bb.min.X, bb.min.Y}
			v    = XY{bb.max.X, bb.max.Y}
			rise = b.Y - a.Y
			run  = b.X - a.X
		)
		if rise > 0 && run < 0 || rise < 0 && run > 0 {
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
