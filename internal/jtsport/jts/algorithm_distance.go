package jts

import "math"

// Functions to compute distance between basic geometric structures.

// Algorithm_Distance_SegmentToSegment computes the distance from a line segment
// AB to a line segment CD.
//
// Note: NON-ROBUST!
func Algorithm_Distance_SegmentToSegment(a, b, c, d *Geom_Coordinate) float64 {
	// Check for zero-length segments.
	if a.Equals(b) {
		return Algorithm_Distance_PointToSegment(a, c, d)
	}
	if c.Equals(d) {
		return Algorithm_Distance_PointToSegment(d, a, b)
	}

	// AB and CD are line segments.
	noIntersection := false
	if !Geom_Envelope_IntersectsEnvelopeEnvelope(a, b, c, d) {
		noIntersection = true
	} else {
		denom := (b.GetX()-a.GetX())*(d.GetY()-c.GetY()) - (b.GetY()-a.GetY())*(d.GetX()-c.GetX())

		if denom == 0 {
			noIntersection = true
		} else {
			rNum := (a.GetY()-c.GetY())*(d.GetX()-c.GetX()) - (a.GetX()-c.GetX())*(d.GetY()-c.GetY())
			sNum := (a.GetY()-c.GetY())*(b.GetX()-a.GetX()) - (a.GetX()-c.GetX())*(b.GetY()-a.GetY())

			s := sNum / denom
			r := rNum / denom

			if r < 0 || r > 1 || s < 0 || s > 1 {
				noIntersection = true
			}
		}
	}
	if noIntersection {
		return Math_MathUtil_Min4(
			Algorithm_Distance_PointToSegment(a, c, d),
			Algorithm_Distance_PointToSegment(b, c, d),
			Algorithm_Distance_PointToSegment(c, a, b),
			Algorithm_Distance_PointToSegment(d, a, b))
	}
	// Segments intersect.
	return 0.0
}

// Algorithm_Distance_PointToSegmentString computes the distance from a point to
// a sequence of line segments.
func Algorithm_Distance_PointToSegmentString(p *Geom_Coordinate, line []*Geom_Coordinate) float64 {
	if len(line) == 0 {
		panic("line array must contain at least one vertex")
	}
	// This handles the case of length = 1.
	minDistance := p.Distance(line[0])
	for i := 0; i < len(line)-1; i++ {
		dist := Algorithm_Distance_PointToSegment(p, line[i], line[i+1])
		if dist < minDistance {
			minDistance = dist
		}
	}
	return minDistance
}

// Algorithm_Distance_PointToSegment computes the distance from a point p to a
// line segment AB.
//
// Note: NON-ROBUST!
func Algorithm_Distance_PointToSegment(p, a, b *Geom_Coordinate) float64 {
	// If start = end, then just compute distance to one of the endpoints.
	if a.GetX() == b.GetX() && a.GetY() == b.GetY() {
		return p.Distance(a)
	}

	// Otherwise use comp.graphics.algorithms Frequently Asked Questions method.
	len2 := (b.GetX()-a.GetX())*(b.GetX()-a.GetX()) + (b.GetY()-a.GetY())*(b.GetY()-a.GetY())
	r := ((p.GetX()-a.GetX())*(b.GetX()-a.GetX()) + (p.GetY()-a.GetY())*(b.GetY()-a.GetY())) / len2

	if r <= 0.0 {
		return p.Distance(a)
	}
	if r >= 1.0 {
		return p.Distance(b)
	}

	s := ((a.GetY()-p.GetY())*(b.GetX()-a.GetX()) - (a.GetX()-p.GetX())*(b.GetY()-a.GetY())) / len2
	return math.Abs(s) * math.Sqrt(len2)
}

// Algorithm_Distance_PointToLinePerpendicular computes the perpendicular
// distance from a point p to the (infinite) line containing the points AB.
func Algorithm_Distance_PointToLinePerpendicular(p, a, b *Geom_Coordinate) float64 {
	len2 := (b.GetX()-a.GetX())*(b.GetX()-a.GetX()) + (b.GetY()-a.GetY())*(b.GetY()-a.GetY())
	s := ((a.GetY()-p.GetY())*(b.GetX()-a.GetX()) - (a.GetX()-p.GetX())*(b.GetY()-a.GetY())) / len2
	return math.Abs(s) * math.Sqrt(len2)
}

// Algorithm_Distance_PointToLinePerpendicularSigned computes the signed
// perpendicular distance from a point p to the (infinite) line containing
// the points AB.
func Algorithm_Distance_PointToLinePerpendicularSigned(p, a, b *Geom_Coordinate) float64 {
	len2 := (b.GetX()-a.GetX())*(b.GetX()-a.GetX()) + (b.GetY()-a.GetY())*(b.GetY()-a.GetY())
	s := ((a.GetY()-p.GetY())*(b.GetX()-a.GetX()) - (a.GetX()-p.GetX())*(b.GetY()-a.GetY())) / len2
	return s * math.Sqrt(len2)
}
