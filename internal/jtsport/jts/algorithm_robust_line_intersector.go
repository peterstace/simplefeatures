package jts

import (
	"math"

	"github.com/peterstace/simplefeatures/internal/jtsport/java"
)

// Algorithm_RobustLineIntersector is a robust version of LineIntersector.
type Algorithm_RobustLineIntersector struct {
	*Algorithm_LineIntersector
	child java.Polymorphic
}

// GetChild returns the immediate child in the type hierarchy chain.
func (rli *Algorithm_RobustLineIntersector) GetChild() java.Polymorphic {
	return rli.child
}

// GetParent returns the immediate parent in the type hierarchy chain.
func (rli *Algorithm_RobustLineIntersector) GetParent() java.Polymorphic {
	return rli.Algorithm_LineIntersector
}

// Algorithm_NewRobustLineIntersector creates a new RobustLineIntersector.
func Algorithm_NewRobustLineIntersector() *Algorithm_RobustLineIntersector {
	li := Algorithm_NewLineIntersector()
	rli := &Algorithm_RobustLineIntersector{
		Algorithm_LineIntersector: li,
	}
	li.child = rli
	return rli
}

// ComputeIntersectionPointLine_BODY computes the intersection of a point p and
// the line p1-p2. This function computes the boolean value of the
// hasIntersection test. The actual value of the intersection (if there is one)
// is equal to the value of p.
func (rli *Algorithm_RobustLineIntersector) ComputeIntersectionPointLine_BODY(p, p1, p2 *Geom_Coordinate) {
	rli.isProper = false
	// Do between check first, since it is faster than the orientation test.
	if Geom_Envelope_IntersectsPointEnvelope(p1, p2, p) {
		if Algorithm_Orientation_Index(p1, p2, p) == 0 &&
			Algorithm_Orientation_Index(p2, p1, p) == 0 {
			rli.isProper = true
			if p.Equals(p1) || p.Equals(p2) {
				rli.isProper = false
			}
			rli.result = Algorithm_LineIntersector_PointIntersection
			return
		}
	}
	rli.result = Algorithm_LineIntersector_NoIntersection
}

// computeIntersect_BODY is the core intersection computation.
func (rli *Algorithm_RobustLineIntersector) computeIntersect_BODY(p1, p2, q1, q2 *Geom_Coordinate) int {
	rli.isProper = false

	// First try a fast test to see if the envelopes of the lines intersect.
	if !Geom_Envelope_IntersectsEnvelopeEnvelope(p1, p2, q1, q2) {
		return Algorithm_LineIntersector_NoIntersection
	}

	// For each endpoint, compute which side of the other segment it lies.
	// If both endpoints lie on the same side of the other segment,
	// the segments do not intersect.
	pq1 := Algorithm_Orientation_Index(p1, p2, q1)
	pq2 := Algorithm_Orientation_Index(p1, p2, q2)

	if (pq1 > 0 && pq2 > 0) || (pq1 < 0 && pq2 < 0) {
		return Algorithm_LineIntersector_NoIntersection
	}

	qp1 := Algorithm_Orientation_Index(q1, q2, p1)
	qp2 := Algorithm_Orientation_Index(q1, q2, p2)

	if (qp1 > 0 && qp2 > 0) || (qp1 < 0 && qp2 < 0) {
		return Algorithm_LineIntersector_NoIntersection
	}

	// Intersection is collinear if each endpoint lies on the other line.
	collinear := pq1 == 0 && pq2 == 0 && qp1 == 0 && qp2 == 0
	if collinear {
		return rli.computeCollinearIntersection(p1, p2, q1, q2)
	}

	// At this point we know that there is a single intersection point
	// (since the lines are not collinear).

	// Check if the intersection is an endpoint. If it is, copy the endpoint as
	// the intersection point. Copying the point rather than computing it
	// ensures the point has the exact value, which is important for robustness.
	// It is sufficient to simply check for an endpoint which is on the other
	// line, since at this point we know that the inputLines must intersect.
	var p *Geom_Coordinate
	z := math.NaN()
	if pq1 == 0 || pq2 == 0 || qp1 == 0 || qp2 == 0 {
		rli.isProper = false

		// Check for two equal endpoints. This is done explicitly rather than by
		// the orientation tests below in order to improve robustness.
		if p1.Equals2D(q1) {
			p = p1
			z = algorithm_RobustLineIntersector_zGet(p1, q1)
		} else if p1.Equals2D(q2) {
			p = p1
			z = algorithm_RobustLineIntersector_zGet(p1, q2)
		} else if p2.Equals2D(q1) {
			p = p2
			z = algorithm_RobustLineIntersector_zGet(p2, q1)
		} else if p2.Equals2D(q2) {
			p = p2
			z = algorithm_RobustLineIntersector_zGet(p2, q2)
		} else if pq1 == 0 {
			// Now check to see if any endpoint lies on the interior of the other
			// segment.
			p = q1
			z = algorithm_RobustLineIntersector_zGetOrInterpolate(q1, p1, p2)
		} else if pq2 == 0 {
			p = q2
			z = algorithm_RobustLineIntersector_zGetOrInterpolate(q2, p1, p2)
		} else if qp1 == 0 {
			p = p1
			z = algorithm_RobustLineIntersector_zGetOrInterpolate(p1, q1, q2)
		} else if qp2 == 0 {
			p = p2
			z = algorithm_RobustLineIntersector_zGetOrInterpolate(p2, q1, q2)
		}
	} else {
		rli.isProper = true
		p = rli.intersection(p1, p2, q1, q2)
		z = algorithm_RobustLineIntersector_zInterpolate4(p, p1, p2, q1, q2)
	}
	rli.intPt[0] = algorithm_RobustLineIntersector_copyWithZ(p, z)
	return Algorithm_LineIntersector_PointIntersection
}

func (rli *Algorithm_RobustLineIntersector) computeCollinearIntersection(p1, p2, q1, q2 *Geom_Coordinate) int {
	q1inP := Geom_Envelope_IntersectsPointEnvelope(p1, p2, q1)
	q2inP := Geom_Envelope_IntersectsPointEnvelope(p1, p2, q2)
	p1inQ := Geom_Envelope_IntersectsPointEnvelope(q1, q2, p1)
	p2inQ := Geom_Envelope_IntersectsPointEnvelope(q1, q2, p2)

	if q1inP && q2inP {
		rli.intPt[0] = algorithm_RobustLineIntersector_copyWithZInterpolate(q1, p1, p2)
		rli.intPt[1] = algorithm_RobustLineIntersector_copyWithZInterpolate(q2, p1, p2)
		return Algorithm_LineIntersector_CollinearIntersection
	}
	if p1inQ && p2inQ {
		rli.intPt[0] = algorithm_RobustLineIntersector_copyWithZInterpolate(p1, q1, q2)
		rli.intPt[1] = algorithm_RobustLineIntersector_copyWithZInterpolate(p2, q1, q2)
		return Algorithm_LineIntersector_CollinearIntersection
	}
	if q1inP && p1inQ {
		// If pts are equal Z is chosen arbitrarily.
		rli.intPt[0] = algorithm_RobustLineIntersector_copyWithZInterpolate(q1, p1, p2)
		rli.intPt[1] = algorithm_RobustLineIntersector_copyWithZInterpolate(p1, q1, q2)
		if q1.Equals(p1) && !q2inP && !p2inQ {
			return Algorithm_LineIntersector_PointIntersection
		}
		return Algorithm_LineIntersector_CollinearIntersection
	}
	if q1inP && p2inQ {
		// If pts are equal Z is chosen arbitrarily.
		rli.intPt[0] = algorithm_RobustLineIntersector_copyWithZInterpolate(q1, p1, p2)
		rli.intPt[1] = algorithm_RobustLineIntersector_copyWithZInterpolate(p2, q1, q2)
		if q1.Equals(p2) && !q2inP && !p1inQ {
			return Algorithm_LineIntersector_PointIntersection
		}
		return Algorithm_LineIntersector_CollinearIntersection
	}
	if q2inP && p1inQ {
		// If pts are equal Z is chosen arbitrarily.
		rli.intPt[0] = algorithm_RobustLineIntersector_copyWithZInterpolate(q2, p1, p2)
		rli.intPt[1] = algorithm_RobustLineIntersector_copyWithZInterpolate(p1, q1, q2)
		if q2.Equals(p1) && !q1inP && !p2inQ {
			return Algorithm_LineIntersector_PointIntersection
		}
		return Algorithm_LineIntersector_CollinearIntersection
	}
	if q2inP && p2inQ {
		// If pts are equal Z is chosen arbitrarily.
		rli.intPt[0] = algorithm_RobustLineIntersector_copyWithZInterpolate(q2, p1, p2)
		rli.intPt[1] = algorithm_RobustLineIntersector_copyWithZInterpolate(p2, q1, q2)
		if q2.Equals(p2) && !q1inP && !p1inQ {
			return Algorithm_LineIntersector_PointIntersection
		}
		return Algorithm_LineIntersector_CollinearIntersection
	}
	return Algorithm_LineIntersector_NoIntersection
}

func algorithm_RobustLineIntersector_copyWithZInterpolate(p, p1, p2 *Geom_Coordinate) *Geom_Coordinate {
	return algorithm_RobustLineIntersector_copyWithZ(p, algorithm_RobustLineIntersector_zGetOrInterpolate(p, p1, p2))
}

func algorithm_RobustLineIntersector_copyWithZ(p *Geom_Coordinate, z float64) *Geom_Coordinate {
	pCopy := p.Copy()
	if !math.IsNaN(z) && Geom_Coordinates_HasZ(pCopy) {
		pCopy.SetZ(z)
	}
	return pCopy
}

// intersection computes the actual value of the intersection point. It is
// rounded to the precision model if being used.
func (rli *Algorithm_RobustLineIntersector) intersection(p1, p2, q1, q2 *Geom_Coordinate) *Geom_Coordinate {
	intPt := rli.intersectionSafe(p1, p2, q1, q2)

	if !rli.isInSegmentEnvelopes(intPt) {
		// Compute a safer result. Copy the coordinate, since it may be rounded later.
		intPt = algorithm_RobustLineIntersector_nearestEndpoint(p1, p2, q1, q2).Copy()
	}
	if rli.precisionModel != nil {
		rli.precisionModel.MakePreciseCoordinate(intPt)
	}
	return intPt
}

// intersectionSafe computes a segment intersection. Round-off error can cause
// the raw computation to fail, (usually due to the segments being approximately
// parallel). If this happens, a reasonable approximation is computed instead.
func (rli *Algorithm_RobustLineIntersector) intersectionSafe(p1, p2, q1, q2 *Geom_Coordinate) *Geom_Coordinate {
	intPt := Algorithm_Intersection_Intersection(p1, p2, q1, q2)
	if intPt == nil {
		intPt = algorithm_RobustLineIntersector_nearestEndpoint(p1, p2, q1, q2)
	}
	return intPt
}

// isInSegmentEnvelopes tests whether a point lies in the envelopes of both
// input segments. A correctly computed intersection point should return true
// for this test. Since this test is for debugging purposes only, no attempt is
// made to optimize the envelope test.
func (rli *Algorithm_RobustLineIntersector) isInSegmentEnvelopes(intPt *Geom_Coordinate) bool {
	env0 := Geom_NewEnvelopeFromCoordinates(rli.inputLines[0][0], rli.inputLines[0][1])
	env1 := Geom_NewEnvelopeFromCoordinates(rli.inputLines[1][0], rli.inputLines[1][1])
	return env0.ContainsCoordinate(intPt) && env1.ContainsCoordinate(intPt)
}

// nearestEndpoint finds the endpoint of the segments P and Q which is closest
// to the other segment. This is a reasonable surrogate for the true
// intersection points in ill-conditioned cases (e.g. where two segments are
// nearly coincident, or where the endpoint of one segment lies almost on the
// other segment).
func algorithm_RobustLineIntersector_nearestEndpoint(p1, p2, q1, q2 *Geom_Coordinate) *Geom_Coordinate {
	nearestPt := p1
	minDist := Algorithm_Distance_PointToSegment(p1, q1, q2)

	dist := Algorithm_Distance_PointToSegment(p2, q1, q2)
	if dist < minDist {
		minDist = dist
		nearestPt = p2
	}
	dist = Algorithm_Distance_PointToSegment(q1, p1, p2)
	if dist < minDist {
		minDist = dist
		nearestPt = q1
	}
	dist = Algorithm_Distance_PointToSegment(q2, p1, p2)
	if dist < minDist {
		nearestPt = q2
	}
	return nearestPt
}

// zGet gets the Z value of the first argument if present, otherwise the value
// of the second argument.
func algorithm_RobustLineIntersector_zGet(p, q *Geom_Coordinate) float64 {
	z := p.GetZ()
	if math.IsNaN(z) {
		z = q.GetZ() // may be NaN
	}
	return z
}

// zGetOrInterpolate gets the Z value of a coordinate if present, or
// interpolates it from the segment it lies on. If the segment Z values are not
// fully populated NaN is returned.
func algorithm_RobustLineIntersector_zGetOrInterpolate(p, p1, p2 *Geom_Coordinate) float64 {
	z := p.GetZ()
	if !math.IsNaN(z) {
		return z
	}
	return algorithm_RobustLineIntersector_zInterpolate(p, p1, p2) // may be NaN
}

// zInterpolate interpolates a Z value for a point along a line segment between
// two points. The Z value of the interpolation point (if any) is ignored. If
// either segment point is missing Z, returns NaN.
func algorithm_RobustLineIntersector_zInterpolate(p, p1, p2 *Geom_Coordinate) float64 {
	p1z := p1.GetZ()
	p2z := p2.GetZ()
	if math.IsNaN(p1z) {
		return p2z // may be NaN
	}
	if math.IsNaN(p2z) {
		return p1z // may be NaN
	}
	if p.Equals2D(p1) {
		return p1z // not NaN
	}
	if p.Equals2D(p2) {
		return p2z // not NaN
	}
	dz := p2z - p1z
	if dz == 0.0 {
		return p1z
	}
	// Interpolate Z from distance of p along p1-p2.
	dx := p2.X - p1.X
	dy := p2.Y - p1.Y
	// Seg has non-zero length since p1 < p < p2.
	seglen := dx*dx + dy*dy
	xoff := p.X - p1.X
	yoff := p.Y - p1.Y
	plen := xoff*xoff + yoff*yoff
	frac := math.Sqrt(plen / seglen)
	zoff := dz * frac
	zInterpolated := p1z + zoff
	return zInterpolated
}

// zInterpolate4 interpolates a Z value for a point along two line segments and
// computes their average. The Z value of the interpolation point (if any) is
// ignored. If one segment point is missing Z that segment is ignored; if both
// segments are missing Z, returns NaN.
func algorithm_RobustLineIntersector_zInterpolate4(p, p1, p2, q1, q2 *Geom_Coordinate) float64 {
	zp := algorithm_RobustLineIntersector_zInterpolate(p, p1, p2)
	zq := algorithm_RobustLineIntersector_zInterpolate(p, q1, q2)
	if math.IsNaN(zp) {
		return zq // may be NaN
	}
	if math.IsNaN(zq) {
		return zp // may be NaN
	}
	// Both Zs have values, so average them.
	return (zp + zq) / 2.0
}
