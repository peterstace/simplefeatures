package simplefeatures

import (
	"fmt"
	"math"
)

func rank(g Geometry) int {
	switch g.(type) {
	case EmptySet:
		return 1
	case Point:
		return 2
	case Line:
		return 3
	case LineString:
		return 4
	case LinearRing:
		return 5
	case Polygon:
		return 6
	case MultiPoint:
		return 7
	case MultiLineString:
		return 8
	case MultiPolygon:
		return 9
	case GeometryCollection:
		return 10
	default:
		panic(fmt.Sprintf("unknown geometry type: %T", g))
	}
}

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

	if parallel := cross(sub(b, a), sub(d, c)) == 0; !parallel {
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
		pt, err := NewPoint(
			a.X+p*(b.X-a.X),
			a.Y+p*(b.Y-a.Y),
		)
		if err != nil {
			panic(err)
		}
		return pt
	}

	if colinear := cross(sub(b, a), sub(d, a)) == 0; colinear {
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

		if abBB.max == cdBB.min {
			// Line segments overlap at a point.
			pt, err := NewPoint(abBB.max.X, abBB.max.Y)
			if err != nil {
				panic(err)
			}
			return pt
		}
		if cdBB.max == abBB.min {
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

func dot(a, b XY) float64 {
	return a.X*b.X + a.Y*b.Y
}

func sub(a, b XY) XY {
	return XY{a.X - b.X, a.Y - b.Y}
}

func cross(a, b XY) float64 {
	return a.X*b.Y - a.Y*b.X
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
	return NewEmptyPoint()
}
