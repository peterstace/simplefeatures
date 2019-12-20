package geom

import (
	"fmt"
	"math"
	"sort"
)

func noImpl(t1, t2 interface{}) error {
	return fmt.Errorf("operation not implemented for type pair %T and %T", t1, t2)
}

func mustIntersection(g1, g2 Geometry) Geometry {
	g, err := intersection(g1, g2)
	if err != nil {
		panic(err)
	}
	return g
}

func intersection(g1, g2 Geometry) (Geometry, error) {
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

func intersectLineWithMultiPoint(ln Line, mp MultiPoint) (Geometry, error) {
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

func intersectMultiLineStringWithMultiLineString(mls1, mls2 MultiLineString) (Geometry, error) {
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

func intersectPointWithLine(point Point, line Line) Geometry {
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

func intersectPointWithLineString(pt Point, ls LineString) Geometry {
	for _, ln := range ls.lines {
		g := intersectPointWithLine(pt, ln)
		if !g.IsEmpty() {
			return g
		}
	}
	return NewGeometryCollection(nil)
}

func intersectMultiPointWithMultiPoint(mp1, mp2 MultiPoint) (Geometry, error) {
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

func intersectMultiPointWithPolygon(mp MultiPoint, p Polygon) (Geometry, error) {
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

func hasIntersection(g1, g2 Geometry) (intersects bool, err error) {
	if g2.IsEmpty() {
		return false, nil // No intersection
	}
	if g1.IsEmpty() {
		return false, nil // No intersection
	}

	if rank(g1) > rank(g2) {
		g1, g2 = g2, g1
	}
	switch g1 := g1.(type) {
	case Point:
		switch g2 := g2.(type) {
		case Point:
			intersects = hasIntersectionPointWithPoint(g1, g2)
			return intersects, nil
		case Line:
			intersects = hasIntersectionPointWithLine(g1, g2)
			return intersects, nil
		case LineString:
			intersects = hasIntersectionPointWithLineString(g1, g2)
			return intersects, nil
		case Polygon:
			intersects = hasIntersectionPointWithPolygon(g1, g2)
			return intersects, nil
		case MultiPoint:
			intersects = hasIntersectionPointWithMultiPoint(g1, g2)
			return intersects, nil
		case MultiLineString:
			intersects = hasIntersectionPointWithMultiLineString(g1, g2)
			return intersects, nil
		case MultiPolygon:
			intersects = hasIntersectionPointWithMultiPolygon(g1, g2)
			return intersects, nil
		case GeometryCollection:
			return false, noImpl(g1, g2)
		}
	case Line:
		switch g2 := g2.(type) {
		case Line:
			intersects = hasIntersectionLineWithLine(g1, g2)
			return intersects, nil
		case LineString:
			ln, err := NewLineStringC(g1.Coordinates())
			if err != nil {
				return false, err
			}
			intersects = hasIntersectionMultiLineStringWithMultiLineString(
				NewMultiLineString([]LineString{ln}),
				NewMultiLineString([]LineString{g2}),
			)
			return intersects, nil
		case Polygon:
			ls, err := NewLineStringC(g1.Coordinates())
			if err != nil {
				return false, err
			}
			mls := NewMultiLineString([]LineString{ls})
			mp, err := NewMultiPolygon([]Polygon{g2})
			if err != nil {
				return false, err
			}
			return hasIntersectionMultiLineStringWithMultiPolygon(mls, mp)
		case MultiPoint:
			intersects = hasIntersectionLineWithMultiPoint(g1, g2)
			return intersects, nil
		case MultiLineString:
			ln, err := NewLineStringC(g1.Coordinates())
			if err != nil {
				return false, err
			}
			intersects = hasIntersectionMultiLineStringWithMultiLineString(
				NewMultiLineString([]LineString{ln}), g2,
			)
			return intersects, nil
		case MultiPolygon:
			ls, err := NewLineStringC(g1.Coordinates())
			if err != nil {
				return false, err
			}
			return hasIntersectionMultiLineStringWithMultiPolygon(
				NewMultiLineString([]LineString{ls}), g2,
			)
		case GeometryCollection:
			return false, noImpl(g1, g2)
		}
	case LineString:
		switch g2 := g2.(type) {
		case LineString:
			intersects = hasIntersectionMultiLineStringWithMultiLineString(
				NewMultiLineString([]LineString{g1}),
				NewMultiLineString([]LineString{g2}),
			)
			return intersects, nil
		case Polygon:
			mp, err := NewMultiPolygon([]Polygon{g2})
			if err != nil {
				return false, err
			}
			return hasIntersectionMultiLineStringWithMultiPolygon(
				NewMultiLineString([]LineString{g1}), mp,
			)
		case MultiPoint:
			return hasIntersectionMultiPointWithMultiLineString(
				g2, NewMultiLineString([]LineString{g1}),
			), nil
		case MultiLineString:
			intersects = hasIntersectionMultiLineStringWithMultiLineString(
				NewMultiLineString([]LineString{g1}),
				g2,
			)
			return intersects, nil
		case MultiPolygon:
			return hasIntersectionMultiLineStringWithMultiPolygon(
				NewMultiLineString([]LineString{g1}), g2,
			)
		case GeometryCollection:
			return false, noImpl(g1, g2)
		}
	case Polygon:
		switch g2 := g2.(type) {
		case Polygon:
			return false, noImpl(g1, g2)
		case MultiPoint:
			intersects = hasIntersectionMultiPointWithPolygon(g2, g1)
			return intersects, nil
		case MultiLineString:
			return false, noImpl(g1, g2)
		case MultiPolygon:
			return false, noImpl(g1, g2)
		case GeometryCollection:
			return false, noImpl(g1, g2)
		}
	case MultiPoint:
		switch g2 := g2.(type) {
		case MultiPoint:
			intersects = hasIntersectionMultiPointWithMultiPoint(g1, g2)
			return intersects, nil
		case MultiLineString:
			return false, noImpl(g1, g2)
		case MultiPolygon:
			return false, noImpl(g1, g2)
		case GeometryCollection:
			return false, noImpl(g1, g2)
		}
	case MultiLineString:
		switch g2 := g2.(type) {
		case MultiLineString:
			intersects = hasIntersectionMultiLineStringWithMultiLineString(g1, g2)
			return intersects, nil
		case MultiPolygon:
			return hasIntersectionMultiLineStringWithMultiPolygon(g1, g2)
		case GeometryCollection:
			return false, noImpl(g1, g2)
		}
	case MultiPolygon:
		switch g2 := g2.(type) {
		case MultiPolygon:
			return false, noImpl(g1, g2)
		case GeometryCollection:
			return false, noImpl(g1, g2)
		}
	case GeometryCollection:
		switch g2 := g2.(type) {
		case GeometryCollection:
			return false, noImpl(g1, g2)
		}
	}

	panic(fmt.Sprintf("implementation error: unhandled geometry types %T and %T", g1, g2))
}

func hasIntersectionLineWithLine(n1, n2 Line) bool {
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
		return true // Point has dimension 0
	}

	if o1 == collinear && o2 == collinear {
		if (!onSegment(a, b, c) && !onSegment(a, b, d)) && (!onSegment(c, d, a) && !onSegment(c, d, b)) {
			return false // No intersection
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
			return true // Point has dimension 0
		}
		//----------------------

		return true // Line has dimension 1
	}

	return false // No intersection
}

func hasIntersectionLineWithMultiPoint(ln Line, mp MultiPoint) bool {
	// Worst case speed is O(n), n is the number of points.
	n := mp.NumPoints()
	for i := 0; i < n; i++ {
		pt := mp.PointN(i)
		if hasIntersectionPointWithLine(pt, ln) {
			return true // Point and MultiPoint both have dimension 0
		}
	}
	return false // No intersection
}

func hasIntersectionMultiPointWithMultiLineString(mp MultiPoint, mls MultiLineString) bool {
	numPts := mp.NumPoints()
	for i := 0; i < numPts; i++ {
		pt := mp.PointN(i)
		numLSs := mls.NumLineStrings()
		for j := 0; j < numLSs; j++ {
			ls := mls.LineStringN(j)
			numLSPts := ls.NumPoints()
			for k := 0; k < numLSPts-1; k++ {
				ln, err := NewLineC(
					ls.PointN(k).Coordinates(),
					ls.PointN(k+1).Coordinates(),
				)
				if err != nil {
					// Should never occur due to construction.
					panic(err)
				}
				if hasIntersectionPointWithLine(pt, ln) {
					return true
				}
			}
		}
	}
	return false
}

func hasIntersectionMultiLineStringWithMultiLineString(mls1, mls2 MultiLineString) bool {
	// Speed is O(n * m) where n, m are the number of lines in each input.
	// This may be the best case, because we must visit all combinations in case
	// any colinear line overlaps exist which would raise the dimensionality.
	for _, ls1 := range mls1.lines {
		for _, ln1 := range ls1.lines {
			for _, ls2 := range mls2.lines {
				for _, ln2 := range ls2.lines {
					if hasIntersectionLineWithLine(ln1, ln2) {
						return true
					}
				}
			}
		}
	}
	return false
}

func hasIntersectionMultiLineStringWithMultiPolygon(mls MultiLineString, mp MultiPolygon) (bool, error) {
	inter, err := hasIntersection(mls, mp.Boundary())
	if err != nil {
		return false, err
	}
	if inter {
		return true, nil
	}

	// Because there is no intersection of the MultiLineString with the
	// boundary of the MultiPolygon, the MultiLineString is either fully
	// contained within the MultiPolygon, or fully outside of it. So we just
	// have to check any control point of the MultiLineString to see if it
	// falls inside or outside of the MultiPolygon.
	for i := 0; i < mls.NumLineStrings(); i++ {
		for j := 0; j < mls.LineStringN(i).NumPoints(); j++ {
			pt := mls.LineStringN(i).PointN(j)
			return hasIntersectionPointWithMultiPolygon(pt, mp), nil
		}
	}
	return false, nil
}

func hasIntersectionPointWithLine(point Point, line Line) bool {
	// Speed is O(1) using a bounding box check then a point-on-line check.
	env := mustEnvelope(line)
	if !env.Contains(point.coords.XY) {
		return false // No intersection
	}
	lhs := (point.coords.X - line.a.X) * (line.b.Y - line.a.Y)
	rhs := (point.coords.Y - line.a.Y) * (line.b.X - line.a.X)
	if lhs == rhs {
		return true // Point has dimension 0
	}
	return false // No intersection
}

func hasIntersectionPointWithLineString(pt Point, ls LineString) bool {
	// Worst case speed is O(n), n is the number of lines.
	for _, ln := range ls.lines {
		if hasIntersectionPointWithLine(pt, ln) {
			return true // Point has dimension 0
		}
	}
	return false // No intersection
}

func hasIntersectionMultiPointWithMultiPoint(mp1, mp2 MultiPoint) bool {
	// To do: improve the speed efficiency, it's currently O(n1*n2)
	for _, pt := range mp1.pts {
		if hasIntersectionPointWithMultiPoint(pt, mp2) {
			return true // Point and MultiPoint both have dimension 0
		}
	}
	return false // No intersection
}

func hasIntersectionPointWithMultiPoint(point Point, mp MultiPoint) bool {
	// Worst case speed is O(n) but that's optimal because mp is not sorted.
	for _, pt := range mp.pts {
		if pt.EqualsExact(point) {
			return true // Point and MultiPoint both have dimension 0
		}
	}
	return false // No intersection
}

func hasIntersectionPointWithMultiLineString(point Point, mls MultiLineString) bool {
	n := mls.NumLineStrings()
	for i := 0; i < n; i++ {
		if hasIntersectionPointWithLineString(point, mls.LineStringN(i)) {
			// There will never be higher dimensionality, so no point in
			// continuing to check other line strings.
			return true
		}
	}
	return false
}

func hasIntersectionPointWithMultiPolygon(pt Point, mp MultiPolygon) bool {
	n := mp.NumPolygons()
	for i := 0; i < n; i++ {
		if hasIntersectionPointWithPolygon(pt, mp.PolygonN(i)) {
			// There will never be higher dimensionality, so no point in
			// continuing to check other line strings.
			return true
		}
	}
	return false
}

func hasIntersectionPointWithPoint(pt1, pt2 Point) bool {
	// Speed is O(1).
	if pt1.EqualsExact(pt2) {
		return true // Point has dimension 0
	}
	return false // No intersection
}

func hasIntersectionPointWithPolygon(pt Point, p Polygon) bool {
	// Speed is O(m), m is the number of holes in the polygon.
	m := p.NumInteriorRings()

	if pointRingSide(pt.XY(), p.ExteriorRing()) == exterior {
		return false // No intersection (outside the exterior)
	}
	for j := 0; j < m; j++ {
		ring := p.InteriorRingN(j)
		if pointRingSide(pt.XY(), ring) == interior {
			return false // No intersection (inside a hole)
		}
	}
	return true // Point has dimension 0
}

func hasIntersectionMultiPointWithPolygon(mp MultiPoint, p Polygon) bool {
	// Speed is O(n*m), n is the number of points, m is the number of holes in the polygon.
	n := mp.NumPoints()

	for i := 0; i < n; i++ {
		pt := mp.PointN(i)
		if hasIntersectionPointWithPolygon(pt, p) {
			return true // Point and MultiPoint have dimension 0
		}
	}
	return false // No intersection
}
