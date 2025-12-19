package jts

// Functions to compute topological information about nodes (ring
// intersections) in polygonal geometry.

// Algorithm_PolygonNodeTopology_IsCrossing checks if four segments at a node
// cross. Typically the segments lie in two different rings, or different
// sections of one ring. The node is topologically valid if the rings do not
// cross. If any segments are collinear, the test returns false.
//
// Parameters:
//   - nodePt: the node location
//   - a0: the previous segment endpoint in a ring
//   - a1: the next segment endpoint in a ring
//   - b0: the previous segment endpoint in the other ring
//   - b1: the next segment endpoint in the other ring
//
// Returns true if the rings cross at the node.
func Algorithm_PolygonNodeTopology_IsCrossing(nodePt, a0, a1, b0, b1 *Geom_Coordinate) bool {
	aLo := a0
	aHi := a1
	if algorithm_PolygonNodeTopology_isAngleGreater(nodePt, aLo, aHi) {
		aLo = a1
		aHi = a0
	}

	// Find positions of b0 and b1. The edges cross if the positions are
	// different. If any edge is collinear they are reported as not crossing.
	compBetween0 := algorithm_PolygonNodeTopology_compareBetween(nodePt, b0, aLo, aHi)
	if compBetween0 == 0 {
		return false
	}
	compBetween1 := algorithm_PolygonNodeTopology_compareBetween(nodePt, b1, aLo, aHi)
	if compBetween1 == 0 {
		return false
	}

	return compBetween0 != compBetween1
}

// Algorithm_PolygonNodeTopology_IsInteriorSegment tests whether a segment
// node-b lies in the interior or exterior of a corner of a ring formed by the
// two segments a0-node-a1. The ring interior is assumed to be on the right of
// the corner (i.e. a CW shell or CCW hole). The test segment must not be
// collinear with the corner segments.
//
// Parameters:
//   - nodePt: the node location
//   - a0: the first vertex of the corner
//   - a1: the second vertex of the corner
//   - b: the other vertex of the test segment
//
// Returns true if the segment is interior to the ring corner.
func Algorithm_PolygonNodeTopology_IsInteriorSegment(nodePt, a0, a1, b *Geom_Coordinate) bool {
	aLo := a0
	aHi := a1
	isInteriorBetween := true
	if algorithm_PolygonNodeTopology_isAngleGreater(nodePt, aLo, aHi) {
		aLo = a1
		aHi = a0
		isInteriorBetween = false
	}
	isBetween := algorithm_PolygonNodeTopology_isBetween(nodePt, b, aLo, aHi)
	isInterior := (isBetween && isInteriorBetween) ||
		(!isBetween && !isInteriorBetween)
	return isInterior
}

// algorithm_PolygonNodeTopology_isBetween tests if an edge p is between edges
// e0 and e1, where the edges all originate at a common origin. The "inside" of
// e0 and e1 is the arc which does not include the origin. The edges are assumed
// to be distinct (non-collinear).
func algorithm_PolygonNodeTopology_isBetween(origin, p, e0, e1 *Geom_Coordinate) bool {
	isGreater0 := algorithm_PolygonNodeTopology_isAngleGreater(origin, p, e0)
	if !isGreater0 {
		return false
	}
	isGreater1 := algorithm_PolygonNodeTopology_isAngleGreater(origin, p, e1)
	return !isGreater1
}

// algorithm_PolygonNodeTopology_compareBetween compares whether an edge p is
// between or outside the edges e0 and e1, where the edges all originate at a
// common origin. The "inside" of e0 and e1 is the arc which does not include
// the positive X-axis at the origin. If p is collinear with an edge 0 is
// returned.
//
// Returns a negative integer, zero or positive integer as the vector P lies
// outside, collinear with, or inside the vectors E0 and E1.
func algorithm_PolygonNodeTopology_compareBetween(origin, p, e0, e1 *Geom_Coordinate) int {
	comp0 := Algorithm_PolygonNodeTopology_CompareAngle(origin, p, e0)
	if comp0 == 0 {
		return 0
	}
	comp1 := Algorithm_PolygonNodeTopology_CompareAngle(origin, p, e1)
	if comp1 == 0 {
		return 0
	}
	if comp0 > 0 && comp1 < 0 {
		return 1
	}
	return -1
}

// algorithm_PolygonNodeTopology_isAngleGreater tests if the angle with the
// origin of a vector P is greater than that of the vector Q.
func algorithm_PolygonNodeTopology_isAngleGreater(origin, p, q *Geom_Coordinate) bool {
	quadrantP := algorithm_PolygonNodeTopology_quadrant(origin, p)
	quadrantQ := algorithm_PolygonNodeTopology_quadrant(origin, q)

	// If the vectors are in different quadrants, that determines the ordering.
	if quadrantP > quadrantQ {
		return true
	}
	if quadrantP < quadrantQ {
		return false
	}

	// Vectors are in the same quadrant. Check relative orientation of vectors.
	// P > Q if it is CCW of Q.
	orient := Algorithm_Orientation_Index(origin, q, p)
	return orient == Algorithm_Orientation_Counterclockwise
}

// Algorithm_PolygonNodeTopology_CompareAngle compares the angles of two vectors
// relative to the positive X-axis at their origin. Angles increase CCW from the
// X-axis.
//
// Returns a negative integer, zero, or a positive integer as this vector P has
// angle less than, equal to, or greater than vector Q.
func Algorithm_PolygonNodeTopology_CompareAngle(origin, p, q *Geom_Coordinate) int {
	quadrantP := algorithm_PolygonNodeTopology_quadrant(origin, p)
	quadrantQ := algorithm_PolygonNodeTopology_quadrant(origin, q)

	// If the vectors are in different quadrants, that determines the ordering.
	if quadrantP > quadrantQ {
		return 1
	}
	if quadrantP < quadrantQ {
		return -1
	}

	// Vectors are in the same quadrant. Check relative orientation of vectors.
	// P > Q if it is CCW of Q.
	orient := Algorithm_Orientation_Index(origin, q, p)
	switch orient {
	case Algorithm_Orientation_Counterclockwise:
		return 1
	case Algorithm_Orientation_Clockwise:
		return -1
	default:
		return 0
	}
}

func algorithm_PolygonNodeTopology_quadrant(origin, p *Geom_Coordinate) int {
	dx := p.GetX() - origin.GetX()
	dy := p.GetY() - origin.GetY()
	return Geom_Quadrant_QuadrantFromDeltas(dx, dy)
}
