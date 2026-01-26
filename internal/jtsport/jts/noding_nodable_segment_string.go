package jts

// Noding_NodableSegmentString is an interface for classes which support adding
// nodes to a segment string.
type Noding_NodableSegmentString interface {
	Noding_SegmentString

	// AddIntersection adds an intersection node for a given point and segment
	// to this segment string.
	AddIntersection(intPt *Geom_Coordinate, segmentIndex int)
}
