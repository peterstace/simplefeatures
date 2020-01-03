package geom

import (
	"fmt"
	"math"
	"sort"
)

func noImpl(t1, t2 interface{}) error {
	return fmt.Errorf("operation not implemented for type pair %T and %T", t1, t2)
}

func mustIntersection(g1, g2 GeometryX) GeometryX {
	g, err := intersection(g1, g2)
	if err != nil {
		panic(err)
	}
	return g
}

func intersection(g1, g2 GeometryX) (GeometryX, error) {
	// Matches PostGIS behaviour for empty geometries.
	if g2.IsEmpty() {
		if _, ok := g2.(GeometryCollection); ok {
			return NewGeometryCollection(nil), nil
		}
		return g2, nil
	}
	if g1.IsEmpty() {
		if _, ok := g1.(GeometryCollection); ok {
			return NewGeometryCollection(nil), nil
		}
		return g1, nil
	}

	if rank(ToGeometry(g1)) > rank(ToGeometry(g2)) {
		g1, g2 = g2, g1
	}
	switch g1 := g1.(type) {
	case Point:
		switch g2 := g2.(type) {
		case Point:
			return intersectPointWithPoint(g1, g2), nil
		case Line:
			return intersectPointWithLine(g1, g2), nil
		case LineString:
			return intersectPointWithLineString(g1, g2), nil
		case Polygon:
			return intersectMultiPointWithPolygon(NewMultiPoint([]Point{g1}), g2)
		case MultiPoint:
			return intersectPointWithMultiPoint(g1, g2), nil
		case MultiLineString:
			return nil, noImpl(g1, g2)
		case MultiPolygon:
			return nil, noImpl(g1, g2)
		case GeometryCollection:
			return nil, noImpl(g1, g2)
		}
	case Line:
		switch g2 := g2.(type) {
		case Line:
			return intersectLineWithLine(g1, g2), nil
		case LineString:
			ls, err := NewLineStringC(g1.Coordinates())
			if err != nil {
				return nil, err
			}
			return intersectMultiLineStringWithMultiLineString(
				NewMultiLineString([]LineString{ls}),
				NewMultiLineString([]LineString{g2}),
			)
		case Polygon:
			return nil, noImpl(g1, g2)
		case MultiPoint:
			return intersectLineWithMultiPoint(g1, g2)
		case MultiLineString:
			return nil, noImpl(g1, g2)
		case MultiPolygon:
			return nil, noImpl(g1, g2)
		case GeometryCollection:
			return nil, noImpl(g1, g2)
		}
	case LineString:
		switch g2 := g2.(type) {
		case LineString:
			return intersectMultiLineStringWithMultiLineString(
				NewMultiLineString([]LineString{g1}),
				NewMultiLineString([]LineString{g2}),
			)
		case Polygon:
			return nil, noImpl(g1, g2)
		case MultiPoint:
			return nil, noImpl(g1, g2)
		case MultiLineString:
			return intersectMultiLineStringWithMultiLineString(
				NewMultiLineString([]LineString{g1}),
				g2,
			)
		case MultiPolygon:
			return nil, noImpl(g1, g2)
		case GeometryCollection:
			return nil, noImpl(g1, g2)
		}
	case Polygon:
		switch g2 := g2.(type) {
		case Polygon:
			return nil, noImpl(g1, g2)
		case MultiPoint:
			return intersectMultiPointWithPolygon(g2, g1)
		case MultiLineString:
			return nil, noImpl(g1, g2)
		case MultiPolygon:
			return nil, noImpl(g1, g2)
		case GeometryCollection:
			return nil, noImpl(g1, g2)
		}
	case MultiPoint:
		switch g2 := g2.(type) {
		case MultiPoint:
			return intersectMultiPointWithMultiPoint(g1, g2)
		case MultiLineString:
			return nil, noImpl(g1, g2)
		case MultiPolygon:
			return nil, noImpl(g1, g2)
		case GeometryCollection:
			return nil, noImpl(g1, g2)
		}
	case MultiLineString:
		switch g2 := g2.(type) {
		case MultiLineString:
			return intersectMultiLineStringWithMultiLineString(g1, g2)
		case MultiPolygon:
			return nil, noImpl(g1, g2)
		case GeometryCollection:
			return nil, noImpl(g1, g2)
		}
	case MultiPolygon:
		switch g2 := g2.(type) {
		case MultiPolygon:
			return nil, noImpl(g1, g2)
		case GeometryCollection:
			return nil, noImpl(g1, g2)
		}
	case GeometryCollection:
		switch g2 := g2.(type) {
		case GeometryCollection:
			return nil, noImpl(g1, g2)
		}
	}

	panic(fmt.Sprintf("implementation error: unhandled geometry types %T and %T", g1, g2))
}

func intersectLineWithLine(n1, n2 Line) GeometryX {
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
		// Division by zero is not possible, since the lines are not parallel.
		p := e / f

		return NewPointXY(b.Sub(a).Scale(p).Add(a))
	}

	if o1 == collinear && o2 == collinear {
		if (!onSegment(a, b, c) && !onSegment(a, b, d)) && (!onSegment(c, d, a) && !onSegment(c, d, b)) {
			return NewGeometryCollection(nil)
		}

		// ---------------------
		// This block is to remove the collinear points in between the two endpoints
		pts := make([]XY, 0, 4)
		pts = append(pts, a, b, c, d)
		rth := rightmostThenHighestIndex(pts)
		pts = append(pts[:rth], pts[rth+1:]...)
		ltl := leftmostThenLowestIndex(pts)
		pts = append(pts[:ltl], pts[ltl+1:]...)
		if pts[0].Equals(pts[1]) {
			return NewPointXY(pts[0])
		}
		//----------------------

		return must(NewLineC(Coordinates{pts[0]}, Coordinates{pts[1]}))
	}

	return NewGeometryCollection(nil)
}

func intersectLineWithMultiPoint(ln Line, mp MultiPoint) (GeometryX, error) {
	var pts []Point
	n := mp.NumPoints()
	for i := 0; i < n; i++ {
		pt := mp.PointN(i)
		if !mustIntersection(pt, ln).IsEmpty() {
			pts = append(pts, pt)
		}
	}
	return canonicalPointsAndLines(pts, nil)
}

func intersectMultiLineStringWithMultiLineString(mls1, mls2 MultiLineString) (GeometryX, error) {
	var points []Point
	var lines []Line
	for _, ls1 := range mls1.lines {
		for _, ln1 := range ls1.lines {
			for _, ls2 := range mls2.lines {
				for _, ln2 := range ls2.lines {
					inter, err := ln1.Intersection(ln2)
					if err != nil {
						return nil, err
					}
					if inter.IsEmpty() {
						continue
					}
					switch inter := inter.(type) {
					case Point:
						points = append(points, inter)
					case Line:
						lines = append(lines, inter)
					default:
						return nil, fmt.Errorf("unhandled intersection result type: %T", inter)
					}
				}
			}
		}
	}
	return canonicalPointsAndLines(points, lines)
}

func intersectPointWithLine(point Point, line Line) GeometryX {
	env := mustEnvelope(line)
	if !env.Contains(point.coords.XY) {
		return NewGeometryCollection(nil)
	}
	lhs := (point.coords.X - line.a.X) * (line.b.Y - line.a.Y)
	rhs := (point.coords.Y - line.a.Y) * (line.b.X - line.a.X)
	if lhs == rhs {
		return point
	}
	return NewGeometryCollection(nil)
}

func intersectPointWithLineString(pt Point, ls LineString) GeometryX {
	for _, ln := range ls.lines {
		g := intersectPointWithLine(pt, ln)
		if !g.IsEmpty() {
			return g
		}
	}
	return NewGeometryCollection(nil)
}

func intersectMultiPointWithMultiPoint(mp1, mp2 MultiPoint) (GeometryX, error) {
	mp1Set := make(map[XY]struct{})
	for _, pt := range mp1.pts {
		mp1Set[pt.Coordinates().XY] = struct{}{}
	}
	mp2Set := make(map[XY]struct{})
	for _, pt := range mp2.pts {
		mp2Set[pt.Coordinates().XY] = struct{}{}
	}

	seen := make(map[XY]bool)
	var intersection []Point
	for pt := range mp1Set {
		if _, ok := mp2Set[pt]; ok && !seen[pt] {
			intersection = append(intersection, NewPointXY(pt))
			seen[pt] = true
		}
	}
	for pt := range mp2Set {
		if _, ok := mp1Set[pt]; ok && !seen[pt] {
			intersection = append(intersection, NewPointXY(pt))
			seen[pt] = true
		}
	}

	// Sort in order to give deterministic output.
	sort.Slice(intersection, func(i, j int) bool {
		return intersection[i].coords.XY.Less(intersection[j].coords.XY)
	})

	return canonicalPointsAndLines(intersection, nil)
}

func intersectPointWithMultiPoint(point Point, mp MultiPoint) GeometryX {
	if mp.IsEmpty() {
		return mp
	}
	for _, pt := range mp.pts {
		if pt.EqualsExact(point) {
			return NewPointXY(point.coords.XY)
		}
	}
	return NewGeometryCollection(nil)
}

func intersectPointWithPoint(pt1, pt2 Point) GeometryX {
	if pt1.EqualsExact(pt2) {
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
	return r.X <= math.Max(p.X, q.X) &&
		r.X >= math.Min(p.X, q.X) &&
		r.Y <= math.Max(p.Y, q.Y) &&
		r.Y >= math.Min(p.Y, q.Y)
}

func intersectMultiPointWithPolygon(mp MultiPoint, p Polygon) (GeometryX, error) {
	pts := make(map[XY]Point)
	n := mp.NumPoints()
outer:
	for i := 0; i < n; i++ {
		pt := mp.PointN(i)
		if pointRingSide(pt.XY(), p.ExteriorRing()) == exterior {
			continue
		}
		m := p.NumInteriorRings()
		for j := 0; j < m; j++ {
			ring := p.InteriorRingN(j)
			if pointRingSide(pt.XY(), ring) == interior {
				continue outer
			}
		}
		pts[pt.XY()] = pt
	}

	ptsSlice := make([]Point, 0, len(pts))
	for _, pt := range pts {
		ptsSlice = append(ptsSlice, pt)
	}
	return canonicalPointsAndLines(ptsSlice, nil)
}
