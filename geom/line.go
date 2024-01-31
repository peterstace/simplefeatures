package geom

import (
	"github.com/peterstace/simplefeatures/rtree"
)

// line represents a line segment between two XY locations. It's an invariant
// that a and b are distinct XY values. Do not create a line that has the same
// a and b value.
type line struct {
	a, b XY
}

// less orders provides an ordering on lines.
func (ln line) less(ot line) bool {
	if ln.a != ot.a {
		return ln.a.Less(ot.a)
	}
	return ln.b.Less(ot.b)
}

// uncheckedEnvelope directly constructs an Envelope that bounds the line. It
// skips envelope validation because line coordinates never come directly from
// users. Instead, line coordinates come directly from pre-validated
// LineStrings, or from operations on pre-validated geometries.
func (ln line) uncheckedEnvelope() Envelope {
	ln.a.X, ln.b.X = sortFloat64Pair(ln.a.X, ln.b.X)
	ln.a.Y, ln.b.Y = sortFloat64Pair(ln.a.Y, ln.b.Y)
	return newUncheckedEnvelope(ln.a, ln.b)
}

func (ln line) box() rtree.Box {
	ln.a.X, ln.b.X = sortFloat64Pair(ln.a.X, ln.b.X)
	ln.a.Y, ln.b.Y = sortFloat64Pair(ln.a.Y, ln.b.Y)
	return rtree.Box{
		MinX: ln.a.X,
		MinY: ln.a.Y,
		MaxX: ln.b.X,
		MaxY: ln.b.Y,
	}
}

func (ln line) length() float64 {
	return ln.a.distanceTo(ln.b)
}

func (ln line) centroid() XY {
	return XY{
		0.5 * (ln.a.X + ln.b.X),
		0.5 * (ln.a.Y + ln.b.Y),
	}
}

func (ln line) asLineString() LineString {
	return NewLineString(NewSequence([]float64{
		ln.a.X, ln.a.Y,
		ln.b.X, ln.b.Y,
	}, DimXY))
}

func (ln line) intersectsXY(xy XY) bool {
	// Speed is O(1) using a bounding box check then a point-on-line check.
	env := ln.uncheckedEnvelope()
	if !env.Contains(xy) {
		return false
	}
	lhs := (xy.X - ln.a.X) * (ln.b.Y - ln.a.Y)
	rhs := (xy.Y - ln.a.Y) * (ln.b.X - ln.a.X)
	return lhs == rhs
}

func (ln line) hasEndpoint(xy XY) bool {
	return ln.a == xy || ln.b == xy
}

// canonicalise swaps the endpoints of the line to be in a canonical form. Two
// lines with reversed endpoint order will have the same canonical form.
func (ln line) canonicalise() line {
	if ln.b.Less(ln.a) {
		ln.a, ln.b = ln.b, ln.a
	}
	return ln
}

// lineWithLineIntersection represents the result of intersecting two line
// segments together. It can either be empty (flag set), a single point (both
// points set the same), or a line segment (defined by the two points).
type lineWithLineIntersection struct {
	empty    bool
	ptA, ptB XY
}

// intersectLine calculates the intersection between two line
// segments without performing any heap allocations.
func (ln line) intersectLine(other line) lineWithLineIntersection {
	a := ln.a
	b := ln.b
	c := other.a
	d := other.b

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

// canonicaliseLinePair canonicalises two lines with respect to both the
// endpoint order in each line, and the ordering between the two lines
// themselves.
func canonicaliseLinePair(lnA, lnB line) (line, line) {
	lnA = lnA.canonicalise()
	lnB = lnB.canonicalise()
	if !lnA.less(lnB) {
		return lnB, lnA
	}
	return lnA, lnB
}

// symmetricLineIntersection is a symmetric version of line.intersectLine. The
// result will be the same, no matter the order of the input arguments (either
// the order of the endpoints, or the order of the two lines).
func symmetricLineIntersection(lnA, lnB line) lineWithLineIntersection {
	lnA, lnB = canonicaliseLinePair(lnA, lnB)
	return lnA.intersectLine(lnB)
}

// onSegement checks if point r is on the segment formed by p and q (all 3
// points should be collinear).
func onSegment(p XY, q XY, r XY) bool {
	return r.X <= fastMax(p.X, q.X) &&
		r.X >= fastMin(p.X, q.X) &&
		r.Y <= fastMax(p.Y, q.Y) &&
		r.Y >= fastMin(p.Y, q.Y)
}

// rightmostThenHighestIndex finds the rightmost-then-highest point.
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
		if ps[i].Less(ps[rpi]) {
			rpi = i
		}
	}
	return rpi
}
