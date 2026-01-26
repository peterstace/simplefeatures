package jts

import "math"

// Functions to compute intersection points between lines and line segments.
//
// In general it is not possible to compute the intersection point of two lines
// exactly, due to numerical roundoff. This is particularly true when the lines
// are nearly parallel. These routines use numerical conditioning on the input
// values to ensure that the computed value is very close to the correct value.
//
// The Z-ordinate is ignored, and not populated.

// Algorithm_Intersection_Intersection computes the intersection point of two
// lines. If the lines are parallel or collinear this case is detected and nil
// is returned.
func Algorithm_Intersection_Intersection(p1, p2, q1, q2 *Geom_Coordinate) *Geom_Coordinate {
	return Algorithm_CGAlgorithmsDD_Intersection(p1, p2, q1, q2)
}

// algorithm_Intersection_intersectionFP computes intersection of two lines,
// using a floating-point algorithm. This is less accurate than
// Algorithm_CGAlgorithmsDD_Intersection. It has caused spatial predicate
// failures in some cases. This is kept for testing purposes.
func algorithm_Intersection_intersectionFP(p1, p2, q1, q2 *Geom_Coordinate) *Geom_Coordinate {
	// Compute midpoint of "kernel envelope".
	var minX0, minY0, maxX0, maxY0 float64
	if p1.X < p2.X {
		minX0 = p1.X
	} else {
		minX0 = p2.X
	}
	if p1.Y < p2.Y {
		minY0 = p1.Y
	} else {
		minY0 = p2.Y
	}
	if p1.X > p2.X {
		maxX0 = p1.X
	} else {
		maxX0 = p2.X
	}
	if p1.Y > p2.Y {
		maxY0 = p1.Y
	} else {
		maxY0 = p2.Y
	}

	var minX1, minY1, maxX1, maxY1 float64
	if q1.X < q2.X {
		minX1 = q1.X
	} else {
		minX1 = q2.X
	}
	if q1.Y < q2.Y {
		minY1 = q1.Y
	} else {
		minY1 = q2.Y
	}
	if q1.X > q2.X {
		maxX1 = q1.X
	} else {
		maxX1 = q2.X
	}
	if q1.Y > q2.Y {
		maxY1 = q1.Y
	} else {
		maxY1 = q2.Y
	}

	var intMinX, intMaxX, intMinY, intMaxY float64
	if minX0 > minX1 {
		intMinX = minX0
	} else {
		intMinX = minX1
	}
	if maxX0 < maxX1 {
		intMaxX = maxX0
	} else {
		intMaxX = maxX1
	}
	if minY0 > minY1 {
		intMinY = minY0
	} else {
		intMinY = minY1
	}
	if maxY0 < maxY1 {
		intMaxY = maxY0
	} else {
		intMaxY = maxY1
	}

	midx := (intMinX + intMaxX) / 2.0
	midy := (intMinY + intMaxY) / 2.0

	// Condition ordinate values by subtracting midpoint.
	p1x := p1.X - midx
	p1y := p1.Y - midy
	p2x := p2.X - midx
	p2y := p2.Y - midy
	q1x := q1.X - midx
	q1y := q1.Y - midy
	q2x := q2.X - midx
	q2y := q2.Y - midy

	// Unrolled computation using homogeneous coordinates eqn.
	px := p1y - p2y
	py := p2x - p1x
	pw := p1x*p2y - p2x*p1y

	qx := q1y - q2y
	qy := q2x - q1x
	qw := q1x*q2y - q2x*q1y

	x := py*qw - qy*pw
	y := qx*pw - px*qw
	w := px*qy - qx*py

	xInt := x / w
	yInt := y / w

	// Check for parallel lines.
	if math.IsNaN(xInt) || math.IsInf(xInt, 0) || math.IsNaN(yInt) || math.IsInf(yInt, 0) {
		return nil
	}
	// De-condition intersection point.
	return Geom_NewCoordinateWithXY(xInt+midx, yInt+midy)
}

// Algorithm_Intersection_LineSegment computes the intersection point of a line
// and a line segment (if any). There will be no intersection point if:
//   - the segment does not intersect the line
//   - the line or the segment are degenerate (have zero length)
//
// If the segment is collinear with the line the first segment endpoint is
// returned.
func Algorithm_Intersection_LineSegment(line1, line2, seg1, seg2 *Geom_Coordinate) *Geom_Coordinate {
	orientS1 := Algorithm_Orientation_Index(line1, line2, seg1)
	if orientS1 == 0 {
		return seg1.Copy()
	}

	orientS2 := Algorithm_Orientation_Index(line1, line2, seg2)
	if orientS2 == 0 {
		return seg2.Copy()
	}

	// If segment lies completely on one side of the line, it does not intersect.
	if (orientS1 > 0 && orientS2 > 0) || (orientS1 < 0 && orientS2 < 0) {
		return nil
	}

	// The segment intersects the line. The full line-line intersection is used
	// to compute the intersection point.
	intPt := Algorithm_Intersection_Intersection(line1, line2, seg1, seg2)
	if intPt != nil {
		return intPt
	}

	// Due to robustness failure it is possible the intersection computation will
	// return nil. In this case choose the closest point.
	dist1 := Algorithm_Distance_PointToLinePerpendicular(seg1, line1, line2)
	dist2 := Algorithm_Distance_PointToLinePerpendicular(seg2, line1, line2)
	if dist1 < dist2 {
		return seg1.Copy()
	}
	return seg2
}
