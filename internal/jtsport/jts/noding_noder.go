package jts

// Noding_Noder computes all intersections between segments in a set of
// SegmentStrings. Intersections found are represented as SegmentNodes and
// added to the SegmentStrings in which they occur. As a final step in the
// noding a new set of segment strings split at the nodes may be returned.
type Noding_Noder interface {
	// ComputeNodes computes the noding for a collection of SegmentStrings. Some
	// Noders may add all these nodes to the input SegmentStrings; others may
	// only add some or none at all.
	ComputeNodes(segStrings []Noding_SegmentString)

	// GetNodedSubstrings returns a collection of fully noded SegmentStrings.
	// The SegmentStrings have the same context as their parent.
	GetNodedSubstrings() []Noding_SegmentString

	// IsNoding_Noder is a marker method for interface identification.
	IsNoding_Noder()
}
