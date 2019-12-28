package geom

import (
	"container/heap"
	"fmt"
	"sort"
)

func hasIntersection(g1, g2 Geometry) bool {
	if g2.IsEmpty() {
		return false
	}
	if g1.IsEmpty() {
		return false
	}

	if rank(g1) > rank(g2) {
		g1, g2 = g2, g1
	}

	if gc, ok := g2.(GeometryCollection); ok {
		n := gc.NumGeometries()
		for i := 0; i < n; i++ {
			g := gc.GeometryN(i)
			if g1.Intersects(g) {
				return true
			}
		}
		return false
	}

	switch g1 := g1.(type) {
	case Point:
		switch g2 := g2.(type) {
		case Point:
			return hasIntersectionPointWithPoint(g1, g2)
		case Line:
			return hasIntersectionPointWithLine(g1, g2)
		case LineString:
			return hasIntersectionPointWithLineString(g1, g2)
		case Polygon:
			return hasIntersectionPointWithPolygon(g1, g2)
		case MultiPoint:
			return hasIntersectionPointWithMultiPoint(g1, g2)
		case MultiLineString:
			return hasIntersectionPointWithMultiLineString(g1, g2)
		case MultiPolygon:
			return hasIntersectionPointWithMultiPolygon(g1, g2)
		}
	case Line:
		switch g2 := g2.(type) {
		case Line:
			return hasIntersectionLineWithLine(g1, g2)
		case LineString:
			return hasIntersectionMultiLineStringWithMultiLineString(
				g1.AsLineString().AsMultiLineString(),
				g2.AsMultiLineString(),
			)
		case Polygon:
			return hasIntersectionMultiLineStringWithMultiPolygon(
				g1.AsLineString().AsMultiLineString(),
				g2.AsMultiPolygon(),
			)
		case MultiPoint:
			return hasIntersectionLineWithMultiPoint(g1, g2)
		case MultiLineString:
			return hasIntersectionMultiLineStringWithMultiLineString(
				g1.AsLineString().AsMultiLineString(), g2,
			)
		case MultiPolygon:
			return hasIntersectionMultiLineStringWithMultiPolygon(
				g1.AsLineString().AsMultiLineString(), g2,
			)
		}
	case LineString:
		switch g2 := g2.(type) {
		case LineString:
			return hasIntersectionMultiLineStringWithMultiLineString(
				g1.AsMultiLineString(),
				g2.AsMultiLineString(),
			)
		case Polygon:
			return hasIntersectionMultiLineStringWithMultiPolygon(
				g1.AsMultiLineString(),
				g2.AsMultiPolygon(),
			)
		case MultiPoint:
			return hasIntersectionMultiPointWithMultiLineString(
				g2, g1.AsMultiLineString(),
			)
		case MultiLineString:
			return hasIntersectionMultiLineStringWithMultiLineString(
				g1.AsMultiLineString(),
				g2,
			)
		case MultiPolygon:
			return hasIntersectionMultiLineStringWithMultiPolygon(
				g1.AsMultiLineString(), g2,
			)
		}
	case Polygon:
		switch g2 := g2.(type) {
		case Polygon:
			return hasIntersectionPolygonWithPolygon(g1, g2)
		case MultiPoint:
			return hasIntersectionMultiPointWithPolygon(g2, g1)
		case MultiLineString:
			return hasIntersectionMultiLineStringWithMultiPolygon(
				g2, g1.AsMultiPolygon(),
			)
		case MultiPolygon:
			return hasIntersectionMultiPolygonWithMultiPolygon(
				g1.AsMultiPolygon(), g2,
			)
		}
	case MultiPoint:
		switch g2 := g2.(type) {
		case MultiPoint:
			return hasIntersectionMultiPointWithMultiPoint(g1, g2)
		case MultiLineString:
			return hasIntersectionMultiPointWithMultiLineString(g1, g2)
		case MultiPolygon:
			return hasIntersectionMultiPointWithMultiPolygon(g1, g2)
		}
	case MultiLineString:
		switch g2 := g2.(type) {
		case MultiLineString:
			return hasIntersectionMultiLineStringWithMultiLineString(g1, g2)
		case MultiPolygon:
			return hasIntersectionMultiLineStringWithMultiPolygon(g1, g2)
		}
	case MultiPolygon:
		switch g2 := g2.(type) {
		case MultiPolygon:
			return hasIntersectionMultiPolygonWithMultiPolygon(g1, g2)
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
		return true
	}

	if o1 == collinear && o2 == collinear {
		if (!onSegment(a, b, c) && !onSegment(a, b, d)) && (!onSegment(c, d, a) && !onSegment(c, d, b)) {
			return false
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
			return true
		}
		//----------------------

		return true
	}

	return false
}

func hasIntersectionLineWithMultiPoint(ln Line, mp MultiPoint) bool {
	// Worst case speed is O(n), n is the number of points.
	n := mp.NumPoints()
	for i := 0; i < n; i++ {
		pt := mp.PointN(i)
		if hasIntersectionPointWithLine(pt, ln) {
			return true
		}
	}
	return false
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

type lineHeap []Line

func (h *lineHeap) Push(x interface{}) {
	*h = append(*h, x.(Line))
}
func (h *lineHeap) Pop() interface{} {
	e := (*h)[len(*h)-1]
	*h = (*h)[:len(*h)-1]
	return e
}
func (h *lineHeap) Len() int {
	return len(*h)
}
func (h *lineHeap) Less(i, j int) bool {
	return (*h)[i].EndPoint().XY().X < (*h)[j].EndPoint().XY().X
}
func (h *lineHeap) Swap(i, j int) {
	(*h)[i], (*h)[j] = (*h)[j], (*h)[i]
}

func hasIntersectionMultiLineStringWithMultiLineString(mls1, mls2 MultiLineString) bool {
	extractSegments := func(mls MultiLineString) []Line {
		var lines []Line
		for _, ls := range mls.lines {
			for _, ln := range ls.lines {
				if ln.StartPoint().XY().X > ln.EndPoint().XY().X {
					// TODO: Use ST_Reverse
					ln.a, ln.b = ln.b, ln.a
				}
				lines = append(lines, ln)
			}
		}
		sort.Slice(lines, func(i, j int) bool {
			return lines[i].StartPoint().XY().X < lines[j].StartPoint().XY().X
		})
		return lines
	}
	unprocessed1 := extractSegments(mls1)
	unprocessed2 := extractSegments(mls2)

	var active1, active2 lineHeap
	var toCheck1, toCheck2 []Line

	for len(unprocessed1) > 0 || len(unprocessed2) > 0 {
		var status float64
		switch {
		case len(unprocessed1) == 0:
			status = unprocessed2[0].StartPoint().XY().X
		case len(unprocessed2) == 0:
			status = unprocessed1[0].StartPoint().XY().X
		case unprocessed1[0].StartPoint().XY().X < unprocessed2[0].StartPoint().XY().X:
			status = unprocessed1[0].StartPoint().XY().X
		default:
			status = unprocessed2[0].StartPoint().XY().X
		}

		for len(active1) > 0 && active1[0].EndPoint().XY().X < status {
			heap.Pop(&active1)
		}
		for len(active2) > 0 && active2[0].EndPoint().XY().X < status {
			heap.Pop(&active2)
		}

		toCheck1 = toCheck1[:0]
		toCheck2 = toCheck2[:0]

		for len(unprocessed1) > 0 && unprocessed1[0].StartPoint().XY().X == status {
			toCheck1 = append(toCheck1, unprocessed1[0])
			heap.Push(&active1, unprocessed1[0])
			unprocessed1 = unprocessed1[1:]
		}
		for len(unprocessed2) > 0 && unprocessed2[0].StartPoint().XY().X == status {
			toCheck2 = append(toCheck2, unprocessed2[0])
			heap.Push(&active2, unprocessed2[0])
			unprocessed2 = unprocessed2[1:]
		}

		for _, checkLine := range toCheck1 {
			for _, ln := range active2 {
				if ln.Intersects(checkLine) {
					return true
				}
			}
		}
		for _, checkLine := range toCheck2 {
			for _, ln := range active1 {
				if ln.Intersects(checkLine) {
					return true
				}
			}
		}
	}
	return false
}

func hasIntersectionMultiLineStringWithMultiPolygon(mls MultiLineString, mp MultiPolygon) bool {
	if hasIntersection(mls, mp.Boundary()) {
		return true
	}

	// Because there is no intersection of the MultiLineString with the
	// boundary of the MultiPolygon, the MultiLineString is either fully
	// contained within the MultiPolygon, or fully outside of it. So we just
	// have to check any control point of the MultiLineString to see if it
	// falls inside or outside of the MultiPolygon.
	for i := 0; i < mls.NumLineStrings(); i++ {
		for j := 0; j < mls.LineStringN(i).NumPoints(); j++ {
			pt := mls.LineStringN(i).PointN(j)
			return hasIntersectionPointWithMultiPolygon(pt, mp)
		}
	}
	return false
}

func hasIntersectionPointWithLine(point Point, line Line) bool {
	// Speed is O(1) using a bounding box check then a point-on-line check.
	env := mustEnvelope(line)
	if !env.Contains(point.coords.XY) {
		return false
	}
	lhs := (point.coords.X - line.a.X) * (line.b.Y - line.a.Y)
	rhs := (point.coords.Y - line.a.Y) * (line.b.X - line.a.X)
	if lhs == rhs {
		return true
	}
	return false
}

func hasIntersectionPointWithLineString(pt Point, ls LineString) bool {
	// Worst case speed is O(n), n is the number of lines.
	for _, ln := range ls.lines {
		if hasIntersectionPointWithLine(pt, ln) {
			return true
		}
	}
	return false
}

func hasIntersectionMultiPointWithMultiPoint(mp1, mp2 MultiPoint) bool {
	// To do: improve the speed efficiency, it's currently O(n1*n2)
	for _, pt := range mp1.pts {
		if hasIntersectionPointWithMultiPoint(pt, mp2) {
			return true // Point and MultiPoint both have dimension 0
		}
	}
	return false
}

func hasIntersectionPointWithMultiPoint(point Point, mp MultiPoint) bool {
	// Worst case speed is O(n) but that's optimal because mp is not sorted.
	for _, pt := range mp.pts {
		if pt.EqualsExact(point) {
			return true
		}
	}
	return false
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
		return true
	}
	return false
}

func hasIntersectionPointWithPolygon(pt Point, p Polygon) bool {
	// Speed is O(m), m is the number of holes in the polygon.
	m := p.NumInteriorRings()

	if pointRingSide(pt.XY(), p.ExteriorRing()) == exterior {
		return false
	}
	for j := 0; j < m; j++ {
		ring := p.InteriorRingN(j)
		if pointRingSide(pt.XY(), ring) == interior {
			return false
		}
	}
	return true
}

func hasIntersectionMultiPointWithPolygon(mp MultiPoint, p Polygon) bool {
	// Speed is O(n*m), n is the number of points, m is the number of holes in the polygon.
	n := mp.NumPoints()

	for i := 0; i < n; i++ {
		pt := mp.PointN(i)
		if hasIntersectionPointWithPolygon(pt, p) {
			return true
		}
	}
	return false
}

func hasIntersectionPolygonWithPolygon(p1, p2 Polygon) bool {
	// Check if the boundaries intersect. If so, then the polygons must
	// intersect.
	b1 := p1.Boundary()
	b2 := p2.Boundary()
	if b1.Intersects(b2) {
		return true
	}

	// Other check to see if an arbitrary point from each polygon is inside the
	// other polygon.
	return p1.ExteriorRing().StartPoint().Intersects(p2) ||
		p2.ExteriorRing().StartPoint().Intersects(p1)
}

func hasIntersectionMultiPolygonWithMultiPolygon(mp1, mp2 MultiPolygon) bool {
	n := mp1.NumPolygons()
	for i := 0; i < n; i++ {
		p1 := mp1.PolygonN(i)
		m := mp2.NumPolygons()
		for j := 0; j < m; j++ {
			p2 := mp2.PolygonN(j)
			if p1.Intersects(p2) {
				return true
			}
		}
	}
	return false
}

func hasIntersectionMultiPointWithMultiPolygon(pts MultiPoint, polys MultiPolygon) bool {
	n := pts.NumPoints()
	for i := 0; i < n; i++ {
		pt := pts.PointN(i)
		if hasIntersectionPointWithMultiPolygon(pt, polys) {
			return true
		}
	}
	return false
}
