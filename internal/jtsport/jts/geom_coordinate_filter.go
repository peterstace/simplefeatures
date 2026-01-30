package jts

// Geom_CoordinateFilter is an interface for classes which use the values of the
// coordinates in a Geometry. Coordinate filters can be used to implement
// centroid and envelope computation, and many other functions.
//
// Geom_CoordinateFilter is an example of the Gang-of-Four Visitor pattern.
//
// Note: it is not recommended to use these filters to mutate the coordinates.
// There is no guarantee that the coordinate is the actual object stored in the
// source geometry. In particular, modified values may not be preserved if the
// source Geometry uses a non-default Geom_CoordinateSequence. If in-place mutation
// is required, use Geom_CoordinateSequenceFilter.
type Geom_CoordinateFilter interface {
	// Filter performs an operation with the provided coord. Note that there is no
	// guarantee that the input coordinate is the actual object stored in the
	// source geometry, so changes to the coordinate object may not be persistent.
	Filter(coord *Geom_Coordinate)

	// Marker method for type identification.
	IsGeom_CoordinateFilter()
}
