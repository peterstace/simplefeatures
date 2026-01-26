package jts

// Algorithm_RectangleLineIntersector computes whether a rectangle intersects
// line segments.
//
// Rectangles contain a large amount of inherent symmetry (or to put it another
// way, although they contain four coordinates they only actually contain 4
// ordinates worth of information). The algorithm used takes advantage of the
// symmetry of the geometric situation to optimize performance by minimizing
// the number of line intersection tests.
type Algorithm_RectangleLineIntersector struct {
	li        *Algorithm_RobustLineIntersector
	rectEnv   *Geom_Envelope
	diagUp0   *Geom_Coordinate
	diagUp1   *Geom_Coordinate
	diagDown0 *Geom_Coordinate
	diagDown1 *Geom_Coordinate
}

// Algorithm_NewRectangleLineIntersector creates a new intersector for the
// given query rectangle, specified as an Envelope.
func Algorithm_NewRectangleLineIntersector(rectEnv *Geom_Envelope) *Algorithm_RectangleLineIntersector {
	// Up and Down are the diagonal orientations relative to the Left side of
	// the rectangle. Index 0 is the left side, 1 is the right side.
	return &Algorithm_RectangleLineIntersector{
		li:        Algorithm_NewRobustLineIntersector(),
		rectEnv:   rectEnv,
		diagUp0:   Geom_NewCoordinateWithXY(rectEnv.GetMinX(), rectEnv.GetMinY()),
		diagUp1:   Geom_NewCoordinateWithXY(rectEnv.GetMaxX(), rectEnv.GetMaxY()),
		diagDown0: Geom_NewCoordinateWithXY(rectEnv.GetMinX(), rectEnv.GetMaxY()),
		diagDown1: Geom_NewCoordinateWithXY(rectEnv.GetMaxX(), rectEnv.GetMinY()),
	}
}

// Intersects tests whether the query rectangle intersects a given line segment.
func (r *Algorithm_RectangleLineIntersector) Intersects(p0, p1 *Geom_Coordinate) bool {
	// If the segment envelope is disjoint from the rectangle envelope, there
	// is no intersection.
	segEnv := Geom_NewEnvelopeFromCoordinates(p0, p1)
	if !r.rectEnv.IntersectsEnvelope(segEnv) {
		return false
	}

	// If either segment endpoint lies in the rectangle, there is an intersection.
	if r.rectEnv.IntersectsCoordinate(p0) {
		return true
	}
	if r.rectEnv.IntersectsCoordinate(p1) {
		return true
	}

	// Normalize segment. This makes p0 less than p1, so that the segment runs
	// to the right, or vertically upwards.
	if p0.CompareTo(p1) > 0 {
		p0, p1 = p1, p0
	}

	// Compute angle of segment. Since the segment is normalized to run left to
	// right, it is sufficient to simply test the Y ordinate. "Upwards" means
	// relative to the left end of the segment.
	isSegUpwards := p1.GetY() > p0.GetY()

	// Since we now know that neither segment endpoint lies in the rectangle,
	// there are two possible situations:
	// 1) the segment is disjoint to the rectangle
	// 2) the segment crosses the rectangle completely.
	//
	// In the case of a crossing, the segment must intersect a diagonal of the
	// rectangle.
	//
	// To distinguish these two cases, it is sufficient to test intersection
	// with a single diagonal of the rectangle, namely the one with slope
	// "opposite" to the slope of the segment. (Note that if the segment is
	// axis-parallel, it must intersect both diagonals, so this is still
	// sufficient.)
	if isSegUpwards {
		r.li.ComputeIntersection(p0, p1, r.diagDown0, r.diagDown1)
	} else {
		r.li.ComputeIntersection(p0, p1, r.diagUp0, r.diagUp1)
	}
	return r.li.HasIntersection()
}
