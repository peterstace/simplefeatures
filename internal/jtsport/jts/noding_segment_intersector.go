package jts

// Noding_SegmentIntersector processes possible intersections detected by a
// Noder. The SegmentIntersector is passed to a Noder. The
// ProcessIntersections method is called whenever the Noder detects that two
// SegmentStrings might intersect. This class may be used either to find all
// intersections, or to detect the presence of an intersection. In the latter
// case, Noders may choose to short-circuit their computation by calling the
// IsDone method.
type Noding_SegmentIntersector interface {
	// ProcessIntersections is called by clients to process intersections for
	// two segments of the SegmentStrings being intersected.
	ProcessIntersections(e0 Noding_SegmentString, segIndex0 int, e1 Noding_SegmentString, segIndex1 int)

	// IsDone reports whether the client of this class needs to continue testing
	// all intersections in an arrangement.
	IsDone() bool

	// IsNoding_SegmentIntersector is a marker method for interface identification.
	IsNoding_SegmentIntersector()
}
