package jts

// Geom_GeometryComponentFilter is an interface for classes which use the components
// of a Geometry. Geometry classes support the concept of applying a
// GeometryComponentFilter filter to a geometry. The filter is applied to every
// component of a geometry, as well as to the geometry itself. (For instance, in
// a Polygon, all the LinearRing components for the shell and holes are visited,
// as well as the polygon itself. In order to process only atomic components,
// the Filter method code must explicitly handle only LineStrings, LinearRings
// and Points.
//
// A GeometryComponentFilter filter can either record information about the
// Geometry or change the Geometry in some way.
//
// GeometryComponentFilter is an example of the Gang-of-Four Visitor pattern.
type Geom_GeometryComponentFilter interface {
	// Filter performs an operation with or on a geometry component.
	Filter(geom *Geom_Geometry)

	// Marker method for type identification.
	IsGeom_GeometryComponentFilter()
}
