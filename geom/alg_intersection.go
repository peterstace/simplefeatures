package geom

import (
	"fmt"
	"math"
	"sort"
)

func mustIntersection(g1, g2 Geometry) Geometry {
	g, err := intersection(g1, g2)
	if err != nil {
		panic(err)
	}
	return g
}

func intersection(g1, g2 Geometry) (Geometry, error) {
	if g2.IsEmpty() {
		return g2, nil
	}
	if g1.IsEmpty() {
		return g1, nil
	}

	if rank(g1) > rank(g2) {
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
			return intersectMultiPointWithPolygon(NewMultiPoint([]Point{g1}), g2), nil
		case MultiPoint:
			return intersectPointWithMultiPoint(g1, g2), nil
		}
	case Line:
		switch g2 := g2.(type) {
		case Line:
			return intersectLineWithLine(g1, g2), nil
		case MultiPoint:
			return intersectLineWithMultiPoint(g1, g2), nil
		}
	case LineString:
		switch g2 := g2.(type) {
		case LineString:
			return intersectMultiLineStringWithMultiLineString(
				NewMultiLineString([]LineString{g1}),
				NewMultiLineString([]LineString{g2}),
			), nil
		case MultiLineString:
			return intersectMultiLineStringWithMultiLineString(
				NewMultiLineString([]LineString{g1}),
				g2,
			), nil
		}
	case Polygon:
		switch g2 := g2.(type) {
		case MultiPoint:
			return intersectMultiPointWithPolygon(g2, g1), nil
		}
	case MultiPoint:
		switch g2 := g2.(type) {
		case MultiPoint:
			return intersectMultiPointWithMultiPoint(g1, g2), nil
		}
	case MultiLineString:
		switch g2 := g2.(type) {
		case MultiLineString:
			return intersectMultiLineStringWithMultiLineString(g1, g2), nil
		}
	}

	return nil, fmt.Errorf("not implemented: intersection with %T and %T", g1, g2)
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

func intersectLineWithMultiPoint(ln Line, mp MultiPoint) Geometry {
	var pts []Point
	n := mp.NumPoints()
	for i := 0; i < n; i++ {
		pt := mp.PointN(i)
		if !mustIntersection(pt, ln).IsEmpty() {
			pts = append(pts, pt)
		}
	}
	if len(pts) == 1 {
		return pts[0]
	}
	return NewMultiPoint(pts)
}

func intersectMultiLineStringWithMultiLineString(mls1, mls2 MultiLineString) Geometry {
	var collection []Geometry
	for _, ls1 := range mls1.lines {
		for _, ln1 := range ls1.lines {
			for _, ls2 := range mls2.lines {
				for _, ln2 := range ls2.lines {
					inter := mustIntersection(ln1, ln2)
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
	env := mustEnvelope(line)
	if !env.Contains(point.coords.XY) {
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
		if pt.EqualsExact(point) {
			return NewPointXY(point.coords.XY)
		}
	}
	return NewGeometryCollection(nil)
}

func intersectPointWithPoint(pt1, pt2 Point) Geometry {
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

func intersectMultiPointWithPolygon(mp MultiPoint, p Polygon) Geometry {
	var pts []Point
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
		pts = append(pts, pt)
	}
	switch len(pts) {
	case 0:
		return NewGeometryCollection(nil)
	case 1:
		return pts[0]
	default:
		return NewMultiPoint(pts)
	}
}

func hasIntersection(g1, g2 Geometry) (intersects bool, dimension int, err error) {
	if g2.IsEmpty() {
		return false, 0, nil // No intersection
	}
	if g1.IsEmpty() {
		return false, 0, nil // No intersection
	}

	if rank(g1) > rank(g2) {
		g1, g2 = g2, g1
	}
	switch g1 := g1.(type) {
	case Point:
		switch g2 := g2.(type) {
		case Point:
			intersects, dimension = hasIntersectionPointWithPoint(g1, g2)
			return intersects, dimension, nil
		case Line:
			intersects, dimension = hasIntersectionPointWithLine(g1, g2)
			return intersects, dimension, nil
		case LineString:
			intersects, dimension = hasIntersectionPointWithLineString(g1, g2)
			return intersects, dimension, nil
		case Polygon:
			intersects, dimension = hasIntersectionPointWithPolygon(g1, g2)
			return intersects, dimension, nil
		case MultiPoint:
			intersects, dimension = hasIntersectionPointWithMultiPoint(g1, g2)
			return intersects, dimension, nil
		}
	case Line:
		switch g2 := g2.(type) {
		case Line:
			intersects, dimension = hasIntersectionLineWithLine(g1, g2)
			return intersects, dimension, nil
		case MultiPoint:
			intersects, dimension = hasIntersectionLineWithMultiPoint(g1, g2)
			return intersects, dimension, nil
		}
	case LineString:
		switch g2 := g2.(type) {
		case LineString:
			intersects, dimension = hasIntersectionMultiLineStringWithMultiLineString(
				NewMultiLineString([]LineString{g1}),
				NewMultiLineString([]LineString{g2}),
			)
			return intersects, dimension, nil
		case MultiLineString:
			intersects, dimension = hasIntersectionMultiLineStringWithMultiLineString(
				NewMultiLineString([]LineString{g1}),
				g2,
			)
			return intersects, dimension, nil
		}
	case Polygon:
		switch g2 := g2.(type) {
		case MultiPoint:
			intersects, dimension = hasIntersectionMultiPointWithPolygon(g2, g1)
			return intersects, dimension, nil
		}
	case MultiPoint:
		switch g2 := g2.(type) {
		case MultiPoint:
			intersects, dimension = hasIntersectionMultiPointWithMultiPoint(g1, g2)
			return intersects, dimension, nil
		}
	case MultiLineString:
		switch g2 := g2.(type) {
		case MultiLineString:
			intersects, dimension = hasIntersectionMultiLineStringWithMultiLineString(g1, g2)
			return intersects, dimension, nil
		}
	}

	return false, 0, fmt.Errorf("not implemented: hasIntersection with %T and %T", g1, g2)
}

func hasIntersectionLineWithLine(n1, n2 Line) (intersects bool, dimension int) {
	// Speed is O(1), but there are multiplications involved.
	a := n1.a.XY
	b := n1.b.XY
	c := n2.a.XY
	d := n2.b.XY

	o1 := orientation(a, b, c)
	o2 := orientation(a, b, d)
	o3 := orientation(c, d, a)
	o4 := orientation(c, d, b)

	if o1 != o2 && o3 != o4 {
		return true, 0 // Point has dimension 0
	}

	if o1 == collinear && o2 == collinear {
		if (!onSegment(a, b, c) && !onSegment(a, b, d)) && (!onSegment(c, d, a) && !onSegment(c, d, b)) {
			return false, 0 // No intersection
		}

		// ---------------------
		// This block is to remove the collinear points in between the two endpoints
		abcd := [4]XY{a, b, c, d}
		pts := abcd[:]
		rth := rightmostThenHighestIndex(pts)
		pts = append(pts[:rth], pts[rth+1:]...)
		ltl := leftmostThenLowestIndex(pts)
		pts = append(pts[:ltl], pts[ltl+1:]...)
		if pts[0].Equals(pts[1]) {
			return true, 0 // Point has dimension 0
		}
		//----------------------

		return true, 1 // Line has dimension 1
	}

	return false, 0 // No intersection
}

func hasIntersectionLineWithMultiPoint(ln Line, mp MultiPoint) (intersects bool, dimension int) {
	// Worst case speed is O(n), n is the number of points.
	n := mp.NumPoints()
	for i := 0; i < n; i++ {
		pt := mp.PointN(i)
		intersects, _ = hasIntersectionPointWithLine(pt, ln)
		if intersects {
			return true, 0 // Point and MultiPoint both have dimension 0
		}
	}
	return false, 0 // No intersection
}

func hasIntersectionMultiLineStringWithMultiLineString(mls1, mls2 MultiLineString) (intersects bool, dimension int) {
	// Speed is O(n * m) where n, m are the number of lines in each input.
	// This may be the best case, because we must visit all combinations in case
	// any colinear line overlaps exist which would raise the dimensionality.
	for _, ls1 := range mls1.lines {
		for _, ln1 := range ls1.lines {
			for _, ls2 := range mls2.lines {
				for _, ln2 := range ls2.lines {
					inter, dim := hasIntersectionLineWithLine(ln1, ln2)
					if inter {
						intersects = true
						if dim > dimension {
							dimension = dim
						}
					}
				}
			}
		}
	}
	return intersects, dimension
}

func hasIntersectionPointWithLine(point Point, line Line) (intersects bool, dimension int) {
	// Speed is O(1) using a bounding box check then a point-on-line check.
	env := mustEnvelope(line)
	if !env.Contains(point.coords.XY) {
		return false, 0 // No intersection
	}
	lhs := (point.coords.X - line.a.X) * (line.b.Y - line.a.Y)
	rhs := (point.coords.Y - line.a.Y) * (line.b.X - line.a.X)
	if lhs == rhs {
		return true, 0 // Point has dimension 0
	}
	return false, 0 // No intersection
}

func hasIntersectionPointWithLineString(pt Point, ls LineString) (intersects bool, dimension int) {
	// Worst case speed is O(n), n is the number of lines.
	for _, ln := range ls.lines {
		intersects, _ = hasIntersectionPointWithLine(pt, ln)
		if intersects {
			return true, 0 // Point has dimension 0
		}
	}
	return false, 0 // No intersection
}

func hasIntersectionMultiPointWithMultiPoint(mp1, mp2 MultiPoint) (intersects bool, dimension int) {
	// To do: improve the speed efficiency, it's currently O(n1*n2)
	for _, pt := range mp1.pts {
		intersects, _ = hasIntersectionPointWithMultiPoint(pt, mp2)
		if intersects {
			return true, 0 // Point and MultiPoint both have dimension 0
		}
	}
	return false, 0 // No intersection
}

func hasIntersectionPointWithMultiPoint(point Point, mp MultiPoint) (intersects bool, dimension int) {
	// Worst case speed is O(n) but that's optimal because mp is not sorted.
	for _, pt := range mp.pts {
		if pt.EqualsExact(point) {
			return true, 0 // Point and MultiPoint both have dimension 0
		}
	}
	return false, 0 // No intersection
}

func hasIntersectionPointWithPoint(pt1, pt2 Point) (intersects bool, dimension int) {
	// Speed is O(1).
	if pt1.EqualsExact(pt2) {
		return true, 0 // Point has dimension 0
	}
	return false, 0 // No intersection
}

func hasIntersectionPointWithPolygon(pt Point, p Polygon) (intersects bool, dimension int) {
	// Speed is O(m), m is the number of holes in the polygon.
	m := p.NumInteriorRings()

	if pointRingSide(pt.XY(), p.ExteriorRing()) == exterior {
		return false, 0 // No intersection (outside the exterior)
	}
	for j := 0; j < m; j++ {
		ring := p.InteriorRingN(j)
		if pointRingSide(pt.XY(), ring) == interior {
			return false, 0 // No intersection (inside a hole)
		}
	}
	return true, 0 // Point has dimension 0
}

func hasIntersectionMultiPointWithPolygon(mp MultiPoint, p Polygon) (intersects bool, dimension int) {
	// Speed is O(n*m), n is the number of points, m is the number of holes in the polygon.
	n := mp.NumPoints()

	for i := 0; i < n; i++ {
		pt := mp.PointN(i)
		intersects, _ = hasIntersectionPointWithPolygon(pt, p)
		if intersects {
			return true, 0 // Point and MultiPoint have dimension 0
		}
	}
	return false, 0 // No intersection
}
