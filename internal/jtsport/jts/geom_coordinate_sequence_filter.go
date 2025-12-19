package jts

// Geom_CoordinateSequenceFilter is an interface for classes which process the
// coordinates in a Geom_CoordinateSequence. A filter can either record information
// about each coordinate, or change the value of the coordinate. Filters can be
// used to implement operations such as coordinate transformations, centroid and
// envelope computation, and many other functions. Geometry classes support the
// concept of applying a Geom_CoordinateSequenceFilter to each Geom_CoordinateSequence
// they contain.
//
// For maximum efficiency, the execution of filters can be short-circuited by
// using the IsDone method.
//
// Geom_CoordinateSequenceFilter is an example of the Gang-of-Four Visitor pattern.
//
// Note: In general, it is preferable to treat Geometries as immutable. Mutation
// should be performed by creating a new Geometry object (see GeometryEditor and
// GeometryTransformer for convenient ways to do this). An exception to this
// rule is when a new Geometry has been created via Geometry.Copy(). In this
// case mutating the Geometry will not cause aliasing issues, and a filter is a
// convenient way to implement coordinate transformation.
type Geom_CoordinateSequenceFilter interface {
	// Filter performs an operation on a coordinate in a Geom_CoordinateSequence.
	// seq is the Geom_CoordinateSequence to which the filter is applied.
	// i is the index of the coordinate to apply the filter to.
	Filter(seq Geom_CoordinateSequence, i int)

	// IsDone reports whether the application of this filter can be terminated.
	// Once this method returns true, it must continue to return true on every
	// subsequent call.
	IsDone() bool

	// IsGeometryChanged reports whether the execution of this filter has modified
	// the coordinates of the geometry. If so, Geometry.GeometryChanged will be
	// executed after this filter has finished being executed.
	//
	// Most filters can simply return a constant value reflecting whether they are
	// able to change the coordinates.
	IsGeometryChanged() bool

	// IsGeom_CoordinateSequenceFilter is a marker method for interface identification.
	IsGeom_CoordinateSequenceFilter()
}
