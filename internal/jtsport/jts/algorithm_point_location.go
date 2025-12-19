package jts

// Functions for locating points within basic geometric structures such as line
// segments, lines and rings.

// Algorithm_PointLocation_IsOnSegment tests whether a point lies on a line
// segment.
func Algorithm_PointLocation_IsOnSegment(p, p0, p1 *Geom_Coordinate) bool {
	// Test envelope first since it's faster.
	if !Geom_Envelope_IntersectsPointEnvelope(p0, p1, p) {
		return false
	}
	// Handle zero-length segments.
	if p.Equals2D(p0) {
		return true
	}
	isOnLine := Algorithm_Orientation_Collinear == Algorithm_Orientation_Index(p0, p1, p)
	return isOnLine
}

// Algorithm_PointLocation_IsOnLine tests whether a point lies on the line
// defined by a list of coordinates.
func Algorithm_PointLocation_IsOnLine(p *Geom_Coordinate, line []*Geom_Coordinate) bool {
	for i := 1; i < len(line); i++ {
		p0 := line[i-1]
		p1 := line[i]
		if Algorithm_PointLocation_IsOnSegment(p, p0, p1) {
			return true
		}
	}
	return false
}

// Algorithm_PointLocation_IsOnLineSeq tests whether a point lies on the line
// defined by a CoordinateSequence.
func Algorithm_PointLocation_IsOnLineSeq(p *Geom_Coordinate, line Geom_CoordinateSequence) bool {
	p0 := Geom_NewCoordinate()
	p1 := Geom_NewCoordinate()
	n := line.Size()
	for i := 1; i < n; i++ {
		line.GetCoordinateInto(i-1, p0)
		line.GetCoordinateInto(i, p1)
		if Algorithm_PointLocation_IsOnSegment(p, p0, p1) {
			return true
		}
	}
	return false
}

// Algorithm_PointLocation_IsInRing tests whether a point lies inside or on a
// ring. The ring may be oriented in either direction. A point lying exactly on
// the ring boundary is considered to be inside the ring.
//
// This method does not first check the point against the envelope of the ring.
func Algorithm_PointLocation_IsInRing(p *Geom_Coordinate, ring []*Geom_Coordinate) bool {
	return Algorithm_PointLocation_LocateInRing(p, ring) != Geom_Location_Exterior
}

// Algorithm_PointLocation_LocateInRing determines whether a point lies in the
// interior, on the boundary, or in the exterior of a ring. The ring may be
// oriented in either direction.
//
// This method does not first check the point against the envelope of the ring.
func Algorithm_PointLocation_LocateInRing(p *Geom_Coordinate, ring []*Geom_Coordinate) int {
	return Algorithm_RayCrossingCounter_LocatePointInRing(p, ring)
}
