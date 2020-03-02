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
		if g2.IsGeometryCollection() {
			return NewEmptyGeometryCollection(XYOnly).AsGeometry(), nil
		}
		return g2, nil
	}
	if g1.IsEmpty() {
		if g1.IsGeometryCollection() {
			return NewEmptyGeometryCollection(XYOnly).AsGeometry(), nil
		}
		return g1, nil
	}

	if rank(g1) > rank(g2) {
		g1, g2 = g2, g1
	}
	switch {
	case g1.IsPoint():
		switch {
		case g2.IsPoint():
			return intersectPointWithPoint(g1.AsPoint(), g2.AsPoint()), nil
		case g2.IsLine():
			return intersectPointWithLine(g1.AsPoint(), g2.AsLine()), nil
		case g2.IsLineString():
			return intersectPointWithLineString(g1.AsPoint(), g2.AsLineString()), nil
		case g2.IsPolygon():
			return intersectMultiPointWithPolygon(g1.AsPoint().AsMultiPoint(), g2.AsPolygon())
		case g2.IsMultiPoint():
			return intersectPointWithMultiPoint(g1.AsPoint(), g2.AsMultiPoint()), nil
		case g2.IsMultiLineString():
			return Geometry{}, noImpl(g1.AsPoint(), g2.AsMultiLineString())
		case g2.IsMultiPolygon():
			return Geometry{}, noImpl(g1.AsPoint(), g2.AsMultiPolygon())
		case g2.IsGeometryCollection():
			return Geometry{}, noImpl(g1.AsPoint(), g2.AsGeometryCollection())
		}
	case g1.IsLine():
		switch {
		case g2.IsLine():
			return intersectLineWithLine(g1.AsLine(), g2.AsLine()), nil
		case g2.IsLineString():
			ls, err := NewLineStringFromSequence(g1.AsLine().Coordinates())
			if err != nil {
				return Geometry{}, err
			}
			return intersectMultiLineStringWithMultiLineString(
				ls.AsMultiLineString(),
				g2.AsLineString().AsMultiLineString(),
			)
		case g2.IsPolygon():
			return Geometry{}, noImpl(g1.AsLine(), g2.AsPolygon)
		case g2.IsMultiPoint():
			return intersectLineWithMultiPoint(g1.AsLine(), g2.AsMultiPoint())
		case g2.IsMultiLineString():
			return Geometry{}, noImpl(g1.AsLine(), g2.AsMultiLineString())
		case g2.IsMultiPolygon():
			return Geometry{}, noImpl(g1.AsLine(), g2.AsMultiPolygon())
		case g2.IsGeometryCollection():
			return Geometry{}, noImpl(g1.AsLine(), g2.AsGeometryCollection())
		}
	case g1.IsLineString():
		switch {
		case g2.IsLineString():
			return intersectMultiLineStringWithMultiLineString(
				g1.AsLineString().AsMultiLineString(),
				g2.AsLineString().AsMultiLineString(),
			)
		case g2.IsPolygon():
			return Geometry{}, noImpl(g1.AsLineString(), g2.AsPolygon())
		case g2.IsMultiPoint():
			return Geometry{}, noImpl(g1.AsLineString(), g2.AsMultiPoint())
		case g2.IsMultiLineString():
			return intersectMultiLineStringWithMultiLineString(
				g1.AsLineString().AsMultiLineString(),
				g2.AsMultiLineString(),
			)
		case g2.IsMultiPolygon():
			return Geometry{}, noImpl(g1.AsLineString(), g2.AsMultiPolygon())
		case g2.IsGeometryCollection():
			return Geometry{}, noImpl(g1.AsLineString(), g2.AsGeometryCollection())
		}
	case g1.IsPolygon():
		switch {
		case g2.IsPolygon():
			return Geometry{}, noImpl(g1.AsPolygon(), g2.AsPolygon())
		case g2.IsMultiPoint():
			return intersectMultiPointWithPolygon(g2.AsMultiPoint(), g1.AsPolygon())
		case g2.IsMultiLineString():
			return Geometry{}, noImpl(g1.AsPolygon(), g2.AsMultiLineString())
		case g2.IsMultiPolygon():
			return Geometry{}, noImpl(g1.AsPolygon(), g2.AsMultiPolygon())
		case g2.IsGeometryCollection():
			return Geometry{}, noImpl(g1.AsPolygon(), g2.AsGeometryCollection())
		}
	case g1.IsMultiPoint():
		switch {
		case g2.IsMultiPoint():
			return intersectMultiPointWithMultiPoint(g1.AsMultiPoint(), g2.AsMultiPoint())
		case g2.IsMultiLineString():
			return Geometry{}, noImpl(g1.AsMultiPoint(), g2.AsMultiLineString())
		case g2.IsMultiPolygon():
			return Geometry{}, noImpl(g1.AsMultiPoint(), g2.AsMultiPolygon())
		case g2.IsGeometryCollection():
			return Geometry{}, noImpl(g1.AsMultiPoint(), g2.AsGeometryCollection())
		}
	case g1.IsMultiLineString():
		switch {
		case g2.IsMultiLineString():
			return intersectMultiLineStringWithMultiLineString(g1.AsMultiLineString(), g2.AsMultiLineString())
		case g2.IsMultiPolygon():
			return Geometry{}, noImpl(g1.AsMultiLineString(), g2.AsMultiPolygon())
		case g2.IsGeometryCollection():
			return Geometry{}, noImpl(g1.AsMultiLineString(), g2.AsGeometryCollection())
		}
	case g1.IsMultiPolygon():
		switch {
		case g2.IsMultiPolygon():
			return Geometry{}, noImpl(g1.AsMultiPolygon(), g2.AsMultiPolygon())
		case g2.IsGeometryCollection():
			return Geometry{}, noImpl(g1.AsMultiPolygon(), g2.AsGeometryCollection())
		}
	case g1.IsGeometryCollection():
		switch {
		case g2.IsGeometryCollection():
			return Geometry{}, noImpl(g1.AsGeometryCollection(), g2.AsGeometryCollection())
		}
	}

	panic(fmt.Sprintf("implementation error: unhandled geometry types %T and %T", g1, g2))
}

func intersectLineWithLine(n1, n2 Line) Geometry {
	result := intersectLineWithLineNoAlloc(n1, n2)
	switch {
	case result.empty:
		return NewEmptyGeometryCollection(XYOnly).AsGeometry()
	case result.ptA == result.ptB:
		return NewPointXY(result.ptA).AsGeometry()
	default:
		ln, err := NewLineXY(result.ptA, result.ptB)
		if err != nil {
			// Cannot occur because the case where ptA and ptB are equal is
			// already handled.
			panic(err)
		}
		return ln.AsGeometry()
	}
}

// lineWithLineIntersection represents the result of intersecting two line
// segments together. It can either be empty (flag set), a single point (both
// points set the same), or a line segment (defined by the two points).
type lineWithLineIntersection struct {
	empty    bool
	ptA, ptB XY
}

// intersectLineWithLine calculates the intersection between two line segments
// without performing any heap allocations.
func intersectLineWithLineNoAlloc(n1, n2 Line) lineWithLineIntersection {
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
			return lineWithLineIntersection{false, c, c}
		}
		if o2 == collinear {
			return lineWithLineIntersection{false, d, d}
		}
		if o3 == collinear {
			return lineWithLineIntersection{false, a, a}
		}
		if o4 == collinear {
			return lineWithLineIntersection{false, b, b}
		}

		e := (c.Y-d.Y)*(a.X-c.X) + (d.X-c.X)*(a.Y-c.Y)
		f := (d.X-c.X)*(a.Y-b.Y) - (a.X-b.X)*(d.Y-c.Y)
		// Division by zero is not possible, since the lines are not parallel.
		p := e / f

		pt := b.Sub(a).Scale(p).Add(a)
		return lineWithLineIntersection{false, pt, pt}
	}

	if o1 == collinear && o2 == collinear {
		if (!onSegment(a, b, c) && !onSegment(a, b, d)) && (!onSegment(c, d, a) && !onSegment(c, d, b)) {
			return lineWithLineIntersection{empty: true}
		}

		// ---------------------
		// This block is to remove the collinear points in between the two endpoints
		pts := make([]XY, 0, 4)
		pts = append(pts, a, b, c, d)
		rth := rightmostThenHighestIndex(pts)
		pts = append(pts[:rth], pts[rth+1:]...)
		ltl := leftmostThenLowestIndex(pts)
		pts = append(pts[:ltl], pts[ltl+1:]...)
		// pts[0] and pts[1] _may_ be coincident, but that's ok.
		return lineWithLineIntersection{false, pts[0], pts[1]}
		//----------------------
	}

	return lineWithLineIntersection{empty: true}
}

func intersectLineWithMultiPoint(ln Line, mp MultiPoint) (Geometry, error) {
	var pts []Point
	n := mp.NumPoints()
	for i := 0; i < n; i++ {
		pt := mp.PointN(i)
		if hasIntersectionPointWithLine(pt, ln) {
			pts = append(pts, pt)
		}
	}
	return canonicalPointsAndLines(pts, nil)
}

func intersectMultiLineStringWithMultiLineString(mls1, mls2 MultiLineString) (Geometry, error) {
	var points []Point
	var lines []Line
	for _, ls1 := range mls1.lines {
		iter1 := newLineStringIterator(ls1)
		for iter1.next() {
			ln1 := iter1.line()
			for _, ls2 := range mls2.lines {
				iter2 := newLineStringIterator(ls2)
				for iter2.next() {
					ln2 := iter2.line()
					inter := intersectLineWithLineNoAlloc(ln1, ln2)
					switch {
					case inter.empty:
						continue
					case inter.ptA == inter.ptB:
						points = append(points, NewPointXY(inter.ptA))
					default:
						ln, err := NewLineXY(inter.ptA, inter.ptB)
						if err != nil {
							// The case where ptA and ptB are coincident
							// has already been handled.
							panic(err)
						}
						lines = append(lines, ln)
					}
				}
			}
		}
	}
	return canonicalPointsAndLines(points, lines)
}

func intersectPointWithLine(pt Point, ln Line) Geometry {
	env := ln.Envelope()
	ptXY, ok := pt.XY()
	if !ok || !env.Contains(ptXY) {
		return NewEmptyGeometryCollection(XYOnly).AsGeometry()
	}
	lhs := (ptXY.X - ln.StartPoint().X) * (ln.EndPoint().Y - ln.StartPoint().Y)
	rhs := (ptXY.Y - ln.StartPoint().Y) * (ln.EndPoint().X - ln.StartPoint().X)
	if lhs == rhs {
		return pt.AsGeometry()
	}
	return NewEmptyGeometryCollection(XYOnly).AsGeometry()
}

func intersectPointWithLineString(pt Point, ls LineString) Geometry {
	iter := newLineStringIterator(ls)
	for iter.next() {
		ln := iter.line()
		g := intersectPointWithLine(pt, ln)
		if !g.IsEmpty() {
			return g
		}
	}
	return NewEmptyGeometryCollection(XYOnly).AsGeometry()
}

func intersectMultiPointWithMultiPoint(mp1, mp2 MultiPoint) (Geometry, error) {
	mp1Set := make(map[XY]struct{})
	for i := 0; i < mp1.NumPoints(); i++ {
		xy, ok := mp1.PointN(i).XY()
		if ok {
			mp1Set[xy] = struct{}{}
		}
	}
	mp2Set := make(map[XY]struct{})
	for i := 0; i < mp2.NumPoints(); i++ {
		xy, ok := mp2.PointN(i).XY()
		if ok {
			mp2Set[xy] = struct{}{}
		}
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
		// Because only non-empty Points are added to the intersection slice,
		// we don't need to check the flags returned from XY().
		xyi, _ := intersection[i].XY()
		xyj, _ := intersection[j].XY()
		return xyi.Less(xyj)
	})

	return canonicalPointsAndLines(intersection, nil)
}

func intersectPointWithMultiPoint(point Point, mp MultiPoint) Geometry {
	if mp.IsEmpty() {
		return mp.AsGeometry()
	}
	for i := 0; i < mp.NumPoints(); i++ {
		pt := mp.PointN(i)
		if pt.EqualsExact(point.AsGeometry()) {
			xy, ok := point.XY()
			if !ok {
				return NewEmptyPoint(XYOnly).AsGeometry()
			}
			return NewPointXY(xy).AsGeometry()
		}
	}
	return NewEmptyGeometryCollection(XYOnly).AsGeometry()
}

func intersectPointWithPoint(pt1, pt2 Point) Geometry {
	if pt1.EqualsExact(pt2.AsGeometry()) {
		xy, ok := pt1.XY()
		if !ok {
			return NewEmptyPoint(XYOnly).AsGeometry()
		}
		return NewPointXY(xy).AsGeometry()
	}
	return NewEmptyGeometryCollection(XYOnly).AsGeometry()
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
	if p.IsEmpty() {
		return p.AsGeometry(), nil
	}
	pts := make(map[XY]Point)
	n := mp.NumPoints()
outer:
	for i := 0; i < n; i++ {
		pt := mp.PointN(i)
		xy, ok := pt.XY()
		if !ok {
			continue
		}
		if pointRingSide(xy, p.ExteriorRing()) == exterior {
			continue
		}
		m := p.NumInteriorRings()
		for j := 0; j < m; j++ {
			ring := p.InteriorRingN(j)
			if pointRingSide(xy, ring) == interior {
				continue outer
			}
		}
		pts[xy] = pt
	}

	ptsSlice := make([]Point, 0, len(pts))
	for _, pt := range pts {
		ptsSlice = append(ptsSlice, pt)
	}
	return canonicalPointsAndLines(ptsSlice, nil)
}
