package jts

// IndexStrtree_ItemDistance is a function method which computes the distance
// between two ItemBoundables in an STRtree. Used for Nearest Neighbour searches.
//
// To make a distance function suitable for querying a single index tree via
// STRtree.NearestNeighbour(ItemDistance), the function should have a non-zero
// reflexive distance. That is, if the two arguments are the same object, the
// distance returned should be non-zero. If it is required that only pairs of
// distinct items be returned, the distance function must be anti-reflexive, and
// must return math.MaxFloat64 for identical arguments.
type IndexStrtree_ItemDistance interface {
	// Distance computes the distance between two items.
	Distance(item1, item2 *IndexStrtree_ItemBoundable) float64

	// IsIndexStrtree_ItemDistance is a marker method for interface identification.
	IsIndexStrtree_ItemDistance()
}
