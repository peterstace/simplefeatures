package jts

// Geom_GeometryFilter is an interface for classes which use the elements of a
// Geometry. GeometryCollection classes support the concept of applying a
// GeometryFilter to the Geometry. The filter is applied to every element
// Geometry. A GeometryFilter can either record information about the Geometry
// or change the Geometry in some way.
//
// GeometryFilter is an example of the Gang-of-Four Visitor pattern.
type Geom_GeometryFilter interface {
	// Filter performs an operation with or on geom.
	Filter(geom *Geom_Geometry)

	// Marker method for type identification.
	IsGeom_GeometryFilter()
}
