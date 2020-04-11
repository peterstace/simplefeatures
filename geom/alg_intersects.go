package geom

import (
	"fmt"
	"math"
	"sort"
)

func hasIntersection(g1, g2 Geometry) bool {
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
	case g1.IsLineString():
		switch {
		case g2.IsLineString():
			has, _ := hasIntersectionMultiLineStringWithMultiLineString(
				g1.AsLineString().AsMultiLineString(),
				g2.AsLineString().AsMultiLineString(),
				false,
			)
			return has
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
			has, _ := hasIntersectionMultiLineStringWithMultiLineString(
				g1.AsLineString().AsMultiLineString(),
				g2.AsMultiLineString(),
				false,
			)
			return has
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
			has, _ := hasIntersectionMultiLineStringWithMultiLineString(
				g1.AsMultiLineString(),
				g2.AsMultiLineString(),
				false,
			)
			return has
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

func hasIntersectionMultiPointWithMultiLineString(mp MultiPoint, mls MultiLineString) bool {
	for i := 0; i < mp.NumPoints(); i++ {
		pt := mp.PointN(i)
		ptXY, ok := pt.XY()
		if !ok {
			continue
		}
		for j := 0; j < mls.NumLineStrings(); j++ {
			seq := mls.LineStringN(j).Coordinates()
			for k := 0; k < seq.Length(); k++ {
				ln, ok := getLine(seq, k)
				if ok && ln.intersectsXY(ptXY) {
					return true
				}
			}
		}
	}
	return false
}

type mlsWithMLSIntersectsExtension struct {
	// set to true iff the intersection covers multiple points (e.g. multiple 0
	// dimension points, or at least one line segment).
	multiplePoints bool

	// If an intersection occurs, singlePoint is set to one of the intersection
	// points.
	singlePoint XY
}

func hasIntersectionMultiLineStringWithMultiLineString(
	mls1, mls2 MultiLineString, populateExtension bool,
) (
	bool, mlsWithMLSIntersectsExtension,
) {
	// A Sweep-Line-Algorithm approach is used to reduce the number of raw line
	// segment intersection tests that must be performed. A vertical sweep line
	// is swept across the plane from left to right. Two 'active' sets of
	// segments are maintained for each multi line string, corresponding to the
	// segments that intersect with the sweep line. Only segments in the active
	// sets need to be considered when checking to see if the multi line
	// strings intersect with each other.

	type side struct {
		mls         MultiLineString
		lines       []line  // all lines from the MLS
		next        int     // index into lines
		active      intHeap // indexes into lines
		newSegments []int   // indexes into lines
	}
	var sides [2]*side
	sides[0] = &side{mls: mls1}
	sides[1] = &side{mls: mls2}

	for _, side := range sides {
		var n int
		for _, ls := range side.mls.lines {
			n += ls.Coordinates().Length()
		}
		side.lines = make([]line, 0, n)
		for _, ls := range side.mls.lines {
			seq := ls.Coordinates()
			for i := 0; i < seq.Length(); i++ {
				ln, ok := getLine(seq, i)
				if !ok {
					continue
				}
				if ln.a.X > ln.b.X {
					ln = ln.reverse()
				}
				side.lines = append(side.lines, ln)
			}
		}
		sort.Slice(side.lines, func(i, j int) bool {
			return side.lines[i].a.X < side.lines[j].a.X
		})
		sideCopy := side // copy because we're using anon func
		side.active.less = func(i, j int) bool {
			ix := sideCopy.lines[i].b.X
			jx := sideCopy.lines[j].b.X
			return ix < jx
		}
	}

	var env Envelope
	var envPopulated bool
	for sides[0].next < len(sides[0].lines) || sides[1].next < len(sides[1].lines) {
		// Calculate the X coordinate of the next line segment(s) that will be
		// processed when sweeping left to right.
		sweepX := math.Inf(+1)
		for _, side := range sides {
			if side.next < len(side.lines) {
				sweepX = math.Min(sweepX, side.lines[side.next].a.X)
			}
		}

		// Update the active line segment sets by throwing away any line
		// segments that can no longer possibly intersect with any unprocessed
		// line segments, and adding any new line segments to the active sets.
		for _, side := range sides {
			for len(side.active.data) != 0 && side.lines[side.active.data[0]].b.X < sweepX {
				side.active.pop()
			}
			side.newSegments = side.newSegments[:0]
			for side.next < len(side.lines) && side.lines[side.next].a.X == sweepX {
				side.newSegments = append(side.newSegments, side.next)
				side.active.push(side.next)
				side.next++
			}
		}

		// Check for intersection between any new line segments, and segments
		// in the opposing active set.
		for i, side := range sides {
			other := sides[1-i]
			for _, lnIdxA := range side.newSegments {
				lnA := side.lines[lnIdxA]
				for _, lnIdxB := range other.active.data {
					lnB := other.lines[lnIdxB]
					inter := lnA.intersectsLine(lnB)
					if inter.empty {
						continue
					}
					if !populateExtension {
						return true, mlsWithMLSIntersectsExtension{}
					}
					if inter.ptA != inter.ptB {
						return true, mlsWithMLSIntersectsExtension{
							multiplePoints: true,
							singlePoint:    inter.ptA,
						}
					}
					if !envPopulated {
						env = NewEnvelope(inter.ptA)
						envPopulated = true
					} else {
						env = env.ExtendToIncludePoint(inter.ptA)
						if env.Min() != env.Max() {
							return true, mlsWithMLSIntersectsExtension{
								multiplePoints: true,
								singlePoint:    env.Min(),
							}
						}
					}
				}
			}
		}
	}
	return envPopulated, mlsWithMLSIntersectsExtension{
		multiplePoints: false,
		singlePoint:    env.Min(),
	}
}

func hasIntersectionMultiLineStringWithMultiPolygon(mls MultiLineString, mp MultiPolygon) bool {
	if has, _ := hasIntersectionMultiLineStringWithMultiLineString(mls, mp.Boundary(), false); has {
		return true
	}

	// Because there is no intersection of the MultiLineString with the
	// boundary of the MultiPolygon, then each LineString inside the
	// MultiLineString is either fully contained within the MultiPolygon, or
	// fully outside of it. So we just have to check any control point of each
	// LineString to see if it falls inside or outside of the MultiPolygon.
	for i := 0; i < mls.NumLineStrings(); i++ {
		ls := mls.LineStringN(i)
		if hasIntersectionPointWithMultiPolygon(ls.StartPoint(), mp) {
			return true
		}
	}
	return false
}

func hasIntersectionPointWithLineString(pt Point, ls LineString) bool {
	// Worst case speed is O(n), n is the number of lines.
	ptXY, ok := pt.XY()
	if !ok {
		return false
	}
	seq := ls.Coordinates()
	for i := 0; i < seq.Length(); i++ {
		ln, ok := getLine(seq, i)
		if ok && ln.intersectsXY(ptXY) {
			return true
		}
	}
	return false
}

func hasIntersectionMultiPointWithMultiPoint(mp1, mp2 MultiPoint) bool {
	// To do: improve the speed efficiency, it's currently O(n1*n2)
	for i := 0; i < mp1.NumPoints(); i++ {
		pt := mp1.PointN(i)
		if hasIntersectionPointWithMultiPoint(pt, mp2) {
			return true // Point and MultiPoint both have dimension 0
		}
	}
	return false
}

func hasIntersectionPointWithMultiPoint(point Point, mp MultiPoint) bool {
	// Worst case speed is O(n) but that's optimal because mp is not sorted.
	for i := 0; i < mp.NumPoints(); i++ {
		pt := mp.PointN(i)
		if hasIntersectionPointWithPoint(point, pt) {
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
	xy1, ok1 := pt1.XY()
	xy2, ok2 := pt2.XY()
	return ok1 && ok2 && xy1 == xy2
}

func hasIntersectionPointWithPolygon(pt Point, p Polygon) bool {
	// Speed is O(m), m is the number of holes in the polygon.
	xy, ok := pt.XY()
	if !ok {
		return false
	}

	if p.IsEmpty() {
		return false
	}
	if pointRingSide(xy, p.ExteriorRing()) == exterior {
		return false
	}
	m := p.NumInteriorRings()
	for i := 0; i < m; i++ {
		ring := p.InteriorRingN(i)
		if pointRingSide(xy, ring) == interior {
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
	if has, _ := hasIntersectionMultiLineStringWithMultiLineString(b1, b2, false); has {
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
