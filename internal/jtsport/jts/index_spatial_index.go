package jts

// Index_SpatialIndex defines the basic operations supported by classes
// implementing spatial index algorithms.
//
// A spatial index typically provides a primary filter for range rectangle
// queries. A secondary filter is required to test for exact intersection. The
// secondary filter may consist of other kinds of tests, such as testing other
// spatial relationships.
type Index_SpatialIndex interface {
	// Insert adds a spatial item with an extent specified by the given Envelope
	// to the index.
	Insert(itemEnv *Geom_Envelope, item any)

	// Query queries the index for all items whose extents intersect the given
	// search Envelope. Note that some kinds of indexes may also return objects
	// which do not in fact intersect the query envelope.
	Query(searchEnv *Geom_Envelope) []any

	// QueryWithVisitor queries the index for all items whose extents intersect
	// the given search Envelope, and applies an ItemVisitor to them. Note that
	// some kinds of indexes may also return objects which do not in fact
	// intersect the query envelope.
	QueryWithVisitor(searchEnv *Geom_Envelope, visitor Index_ItemVisitor)

	// Remove removes a single item from the tree.
	Remove(itemEnv *Geom_Envelope, item any) bool

	// IsIndex_SpatialIndex is a marker method for interface identification.
	IsIndex_SpatialIndex()
}
