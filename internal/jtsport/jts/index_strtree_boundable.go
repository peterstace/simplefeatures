package jts

// IndexStrtree_Boundable is a spatial object in an AbstractSTRtree.
type IndexStrtree_Boundable interface {
	// GetBounds returns a representation of space that encloses this Boundable,
	// preferably not much bigger than this Boundable's boundary yet fast to
	// test for intersection with the bounds of other Boundables. The class of
	// object returned depends on the subclass of AbstractSTRtree.
	// Returns an Envelope (for STRtrees), an Interval (for SIRtrees), or other
	// object (for other subclasses of AbstractSTRtree).
	GetBounds() any

	// TRANSLITERATION NOTE: Marker method added for Go interface type identification.
	// Not present in Java source.
	IsIndexStrtree_Boundable()
}
