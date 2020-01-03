package geom

import (
	"container/heap"
	"fmt"
	"math"
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

	if g2.IsGeometryCollection() {
		gc := g2.AsGeometryCollection()
		n := gc.NumGeometries()
		for i := 0; i < n; i++ {
			g := gc.GeometryN(i)
			if g1.Intersects(g) {
				return true
			}
		}
		return false
	}

	switch {
	case g1.IsPoint():
		switch {
		case g2.IsPoint():
			return hasIntersectionPointWithPoint(g1.AsPoint(), g2.AsPoint())
		case g2.IsLine():
			return hasIntersectionPointWithLine(g1.AsPoint(), g2.AsLine())
		case g2.IsLineString():
			return hasIntersectionPointWithLineString(g1.AsPoint(), g2.AsLineString())
		case g2.IsPolygon():
			return hasIntersectionPointWithPolygon(g1.AsPoint(), g2.AsPolygon())
		case g2.IsMultiPoint():
			return hasIntersectionPointWithMultiPoint(g1.AsPoint(), g2.AsMultiPoint())
		case g2.IsMultiLineString():
			return hasIntersectionPointWithMultiLineString(g1.AsPoint(), g2.AsMultiLineString())
		case g2.IsMultiPolygon():
			return hasIntersectionPointWithMultiPolygon(g1.AsPoint(), g2.AsMultiPolygon())
		}
	case g1.IsLine():
		switch {
		case g2.IsLine():
			return hasIntersectionLineWithLine(g1.AsLine(), g2.AsLine())
		case g2.IsLineString():
			return hasIntersectionMultiLineStringWithMultiLineString(
				g1.AsLine().AsLineString().AsMultiLineString(),
				g2.AsLineString().AsMultiLineString(),
			)
		case g2.IsPolygon():
			return hasIntersectionMultiLineStringWithMultiPolygon(
				g1.AsLine().AsLineString().AsMultiLineString(),
				g2.AsPolygon().AsMultiPolygon(),
			)
		case g2.IsMultiPoint():
			return hasIntersectionLineWithMultiPoint(g1.AsLine(), g2.AsMultiPoint())
		case g2.IsMultiLineString():
			return hasIntersectionMultiLineStringWithMultiLineString(
				g1.AsLine().AsLineString().AsMultiLineString(), g2.AsMultiLineString(),
			)
		case g2.IsMultiPolygon():
			return hasIntersectionMultiLineStringWithMultiPolygon(
				g1.AsLine().AsLineString().AsMultiLineString(), g2.AsMultiPolygon(),
			)
		}
	case g1.IsLineString():
		switch {
		case g2.IsLineString():
			return hasIntersectionMultiLineStringWithMultiLineString(
				g1.AsLineString().AsMultiLineString(),
				g2.AsLineString().AsMultiLineString(),
			)
		case g2.IsPolygon():
			return hasIntersectionMultiLineStringWithMultiPolygon(
				g1.AsLineString().AsMultiLineString(),
				g2.AsPolygon().AsMultiPolygon(),
			)
		case g2.IsMultiPoint():
			return hasIntersectionMultiPointWithMultiLineString(
				g2.AsMultiPoint(),
				g1.AsLineString().AsMultiLineString(),
			)
		case g2.IsMultiLineString():
			return hasIntersectionMultiLineStringWithMultiLineString(
				g1.AsLineString().AsMultiLineString(),
				g2.AsMultiLineString(),
			)
		case g2.IsMultiPolygon():
			return hasIntersectionMultiLineStringWithMultiPolygon(
				g1.AsLineString().AsMultiLineString(),
				g2.AsMultiPolygon(),
			)
		}
	case g1.IsPolygon():
		switch {
		case g2.IsPolygon():
			return hasIntersectionPolygonWithPolygon(
				g1.AsPolygon(),
				g2.AsPolygon(),
			)
		case g2.IsMultiPoint():
			return hasIntersectionMultiPointWithPolygon(
				g2.AsMultiPoint(),
				g1.AsPolygon(),
			)
		case g2.IsMultiLineString():
			return hasIntersectionMultiLineStringWithMultiPolygon(
				g2.AsMultiLineString(),
				g1.AsPolygon().AsMultiPolygon(),
			)
		case g2.IsMultiPolygon():
			return hasIntersectionMultiPolygonWithMultiPolygon(
				g1.AsPolygon().AsMultiPolygon(),
				g2.AsMultiPolygon(),
			)
		}
	case g1.IsMultiPoint():
		switch {
		case g2.IsMultiPoint():
			return hasIntersectionMultiPointWithMultiPoint(
				g1.AsMultiPoint(),
				g2.AsMultiPoint(),
			)
		case g2.IsMultiLineString():
			return hasIntersectionMultiPointWithMultiLineString(
				g1.AsMultiPoint(),
				g2.AsMultiLineString(),
			)
		case g2.IsMultiPolygon():
			return hasIntersectionMultiPointWithMultiPolygon(
				g1.AsMultiPoint(),
				g2.AsMultiPolygon(),
			)
		}
	case g1.IsMultiLineString():
		switch {
		case g2.IsMultiLineString():
			return hasIntersectionMultiLineStringWithMultiLineString(
				g1.AsMultiLineString(),
				g2.AsMultiLineString(),
			)
		case g2.IsMultiPolygon():
			return hasIntersectionMultiLineStringWithMultiPolygon(
				g1.AsMultiLineString(),
				g2.AsMultiPolygon(),
			)
		}
	case g1.IsMultiPolygon():
		switch {
		case g2.IsMultiPolygon():
			return hasIntersectionMultiPolygonWithMultiPolygon(
				g1.AsMultiPolygon(),
				g2.AsMultiPolygon(),
			)
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
	// A Sweep-Line-Algorithm approach is used to reduce the number of raw line
	// segment intersection tests that must be performed. A vertical sweep line
	// is swept across the plane from left to right. Two 'active' sets of
	// segments are maintained for each multi line string, corresponding to the
	// segments that intersect with the sweep line. Only segments in the active
	// sets need to be considered when checking to see if the multi line
	// strings intersect with each other.

	type side struct {
		mls         MultiLineString
		unprocessed []Line
		active      lineHeap
		newSegments []Line
	}
	var sides [2]*side
	sides[0] = &side{mls: mls1}
	sides[1] = &side{mls: mls2}

	// Create a list of line segments from each MultiLineString, in ascending
	// order by X coordinate.
	for _, side := range sides {
		for _, ls := range side.mls.lines {
			for _, ln := range ls.lines {
				if ln.StartPoint().XY().X > ln.EndPoint().XY().X {
					// TODO: Use ST_Reverse
					ln.a, ln.b = ln.b, ln.a
				}
				side.unprocessed = append(side.unprocessed, ln)
			}
		}
		sort.Slice(side.unprocessed, func(i, j int) bool {
			ix := side.unprocessed[i].StartPoint().XY().X
			jx := side.unprocessed[j].StartPoint().XY().X
			return ix < jx
		})
	}

	for len(sides[0].unprocessed)+len(sides[1].unprocessed) > 0 {
		// Calculate the X coordinate of the next line segment(s) that will be
		// processed when sweeping left to right.
		sweepX := math.Inf(+1)
		for _, side := range sides {
			if len(side.unprocessed) > 0 {
				sweepX = math.Min(sweepX, side.unprocessed[0].StartPoint().XY().X)
			}
		}

		// Update the active line segment sets by throwing away any line
		// segments that can no longer possibly intersect with any unprocessed
		// line segments, and adding any new line segments to the active sets.
		for _, side := range sides {
			for len(side.active) > 0 && side.active[0].EndPoint().XY().X < sweepX {
				heap.Pop(&side.active)
			}
			side.newSegments = side.newSegments[:0]
			for len(side.unprocessed) > 0 && side.unprocessed[0].StartPoint().XY().X == sweepX {
				side.newSegments = append(side.newSegments, side.unprocessed[0])
				heap.Push(&side.active, side.unprocessed[0])
				side.unprocessed = side.unprocessed[1:]
			}
		}

		// Check for intersection between any new line segments, and segments
		// in the opposing active set.
		for i, side := range sides {
			other := sides[1-i]
			for _, checkLine := range side.newSegments {
				for _, ln := range other.active {
					if hasIntersectionLineWithLine(ln, checkLine) {
						return true
					}
				}
			}
		}
	}
	return false
}

func hasIntersectionMultiLineStringWithMultiPolygon(mls MultiLineString, mp MultiPolygon) bool {
	if mls.Intersects(mp.Boundary()) {
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
	env := mustEnv(line.Envelope())
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
		if pt.EqualsExact(point.AsGeometry()) {
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
	if pt1.EqualsExact(pt2.AsGeometry()) {
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
	return hasIntersectionPointWithPolygon(p1.ExteriorRing().StartPoint(), p2) ||
		hasIntersectionPointWithPolygon(p2.ExteriorRing().StartPoint(), p1)
}

func hasIntersectionMultiPolygonWithMultiPolygon(mp1, mp2 MultiPolygon) bool {
	n := mp1.NumPolygons()
	for i := 0; i < n; i++ {
		p1 := mp1.PolygonN(i)
		m := mp2.NumPolygons()
		for j := 0; j < m; j++ {
			p2 := mp2.PolygonN(j)
			if hasIntersectionPolygonWithPolygon(p1, p2) {
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
