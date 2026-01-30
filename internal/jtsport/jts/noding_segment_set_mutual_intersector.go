package jts

// Noding_SegmentSetMutualIntersector is an intersector for the red-blue
// intersection problem. In this class of line arrangement problem, two
// disjoint sets of linestrings are intersected.
//
// Implementing types must provide a way of supplying the base set of segment
// strings to test against (e.g. in the constructor, for straightforward
// thread-safety).
//
// In order to allow optimizing processing, the following condition is assumed
// to hold for each set: the only intersection between any two linestrings
// occurs at their endpoints.
//
// Implementations can take advantage of this fact to optimize processing (i.e.
// by avoiding testing for intersections between linestrings belonging to the
// same set).
type Noding_SegmentSetMutualIntersector interface {
	// Process computes the intersections with a given set of SegmentStrings,
	// using the supplied SegmentIntersector.
	Process(segStrings []Noding_SegmentString, segInt Noding_SegmentIntersector)
}
