package jts

// Algorithm_RayCrossingCounter counts the number of segments crossed by a
// horizontal ray extending to the right from a given point, in an incremental
// fashion. This can be used to determine whether a point lies in a Polygonal
// geometry. The class determines the situation where the point lies exactly on
// a segment. When being used for Point-In-Polygon determination, this case
// allows short-circuiting the evaluation.
//
// This class handles polygonal geometries with any number of shells and holes.
// The orientation of the shell and hole rings is unimportant. In order to
// compute a correct location for a given polygonal geometry, it is essential
// that all segments are counted which:
//   - touch the ray
//   - lie in any ring which may contain the point
//
// The only exception is when the point-on-segment situation is detected, in
// which case no further processing is required. The implication of the above
// rule is that segments which can be a priori determined to not touch the ray
// (i.e. by a test of their bounding box or Y-extent) do not need to be counted.
// This allows for optimization by indexing.
//
// This implementation uses the extended-precision orientation test, to provide
// maximum robustness and consistency within other algorithms.
type Algorithm_RayCrossingCounter struct {
	p                *Geom_Coordinate
	crossingCount    int
	isPointOnSegment bool
}

// Algorithm_NewRayCrossingCounter creates a new RayCrossingCounter for the given
// point.
func Algorithm_NewRayCrossingCounter(p *Geom_Coordinate) *Algorithm_RayCrossingCounter {
	return &Algorithm_RayCrossingCounter{
		p:                p,
		crossingCount:    0,
		isPointOnSegment: false,
	}
}

// Algorithm_RayCrossingCounter_LocatePointInRing determines the Location of a
// point in a ring. This method is an exemplar of how to use this class.
func Algorithm_RayCrossingCounter_LocatePointInRing(p *Geom_Coordinate, ring []*Geom_Coordinate) int {
	counter := Algorithm_NewRayCrossingCounter(p)
	for i := 1; i < len(ring); i++ {
		p1 := ring[i]
		p2 := ring[i-1]
		counter.CountSegment(p1, p2)
		if counter.IsOnSegment() {
			return counter.GetLocation()
		}
	}
	return counter.GetLocation()
}

// Algorithm_RayCrossingCounter_LocatePointInRingSeq determines the Location of a
// point in a ring defined by a CoordinateSequence.
func Algorithm_RayCrossingCounter_LocatePointInRingSeq(p *Geom_Coordinate, ring Geom_CoordinateSequence) int {
	counter := Algorithm_NewRayCrossingCounter(p)
	p1 := Geom_NewCoordinate()
	p2 := Geom_NewCoordinate()
	for i := 1; i < ring.Size(); i++ {
		p1.X = ring.GetOrdinate(i, Geom_CoordinateSequence_X)
		p1.Y = ring.GetOrdinate(i, Geom_CoordinateSequence_Y)
		p2.X = ring.GetOrdinate(i-1, Geom_CoordinateSequence_X)
		p2.Y = ring.GetOrdinate(i-1, Geom_CoordinateSequence_Y)
		counter.CountSegment(p1, p2)
		if counter.IsOnSegment() {
			return counter.GetLocation()
		}
	}
	return counter.GetLocation()
}

// CountSegment counts a segment.
func (rcc *Algorithm_RayCrossingCounter) CountSegment(p1, p2 *Geom_Coordinate) {
	// For each segment, check if it crosses a horizontal ray running from the
	// test point in the positive x direction.

	// Check if the segment is strictly to the left of the test point.
	if p1.X < rcc.p.X && p2.X < rcc.p.X {
		return
	}

	// Check if the point is equal to the current ring vertex.
	if rcc.p.X == p2.X && rcc.p.Y == p2.Y {
		rcc.isPointOnSegment = true
		return
	}

	// For horizontal segments, check if the point is on the segment. Otherwise,
	// horizontal segments are not counted.
	if p1.Y == rcc.p.Y && p2.Y == rcc.p.Y {
		minx := p1.X
		maxx := p2.X
		if minx > maxx {
			minx = p2.X
			maxx = p1.X
		}
		if rcc.p.X >= minx && rcc.p.X <= maxx {
			rcc.isPointOnSegment = true
		}
		return
	}

	// Evaluate all non-horizontal segments which cross a horizontal ray to the
	// right of the test pt. To avoid double-counting shared vertices, we use the
	// convention that:
	//   - an upward edge includes its starting endpoint, and excludes its final
	//     endpoint
	//   - a downward edge excludes its starting endpoint, and includes its final
	//     endpoint
	if (p1.Y > rcc.p.Y && p2.Y <= rcc.p.Y) ||
		(p2.Y > rcc.p.Y && p1.Y <= rcc.p.Y) {
		orient := Algorithm_Orientation_Index(p1, p2, rcc.p)
		if orient == Algorithm_Orientation_Collinear {
			rcc.isPointOnSegment = true
			return
		}
		// Re-orient the result if needed to ensure effective segment direction is
		// upwards.
		if p2.Y < p1.Y {
			orient = -orient
		}
		// The upward segment crosses the ray if the test point lies to the left
		// (CCW) of the segment.
		if orient == Algorithm_Orientation_Left {
			rcc.crossingCount++
		}
	}
}

// GetCount gets the count of crossings.
func (rcc *Algorithm_RayCrossingCounter) GetCount() int {
	return rcc.crossingCount
}

// IsOnSegment reports whether the point lies exactly on one of the supplied
// segments. This method may be called at any time as segments are processed. If
// the result of this method is true, no further segments need be supplied,
// since the result will never change again.
func (rcc *Algorithm_RayCrossingCounter) IsOnSegment() bool {
	return rcc.isPointOnSegment
}

// GetLocation gets the Location of the point relative to the ring, polygon or
// multipolygon from which the processed segments were provided.
//
// This method only determines the correct location if all relevant segments
// have been processed.
func (rcc *Algorithm_RayCrossingCounter) GetLocation() int {
	if rcc.isPointOnSegment {
		return Geom_Location_Boundary
	}
	// The point is in the interior of the ring if the number of X-crossings is
	// odd.
	if (rcc.crossingCount % 2) == 1 {
		return Geom_Location_Interior
	}
	return Geom_Location_Exterior
}

// IsPointInPolygon tests whether the point lies in or on the ring, polygon or
// multipolygon from which the processed segments were provided.
//
// This method only determines the correct location if all relevant segments
// have been processed.
func (rcc *Algorithm_RayCrossingCounter) IsPointInPolygon() bool {
	return rcc.GetLocation() != Geom_Location_Exterior
}
