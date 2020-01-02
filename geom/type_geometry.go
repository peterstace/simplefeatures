package geom

import (
	"encoding/json"
)

// GeometryX is the most general type of geometry supported, and exposes common
// behaviour. All geometry types implement this interface.
type GeometryX interface {
	// Intersection returns a geometric object that represents the point set
	// intersection of this geometry with another geometry.
	//
	// It is not implemented for all possible pairs of geometries, and returns
	// an error in those cases.
	Intersection(GeometryX) (GeometryX, error)

	// Intersects returns true if the intersection of this gemoetry with the
	// specified other geometry is not empty, or false if it is empty.
	Intersects(GeometryX) bool

	// IsEmpty returns true if this object an empty geometry.
	IsEmpty() bool

	// Dimension returns the dimension of the geometry. This is 0 for empty
	// geometries, 0 for points, 1 for curves, and 2 for surfaces. For mixed
	// geometries, it is the maximum dimension over the collection.
	Dimension() int

	// Envelope returns the axis aligned bounding box that most tightly
	// surrounds the geometry. Envelopes are not defined for empty geometries,
	// in which case the returned flag will be false.
	Envelope() (Envelope, bool)

	// Equals checks if this geometry is equal to another geometry. Two
	// geometries are equal if they contain exactly the same points.
	//
	// It is not implemented for all possible pairs of geometries, and returns
	// an error in those cases.
	Equals(GeometryX) (bool, error)

	// Boundary returns the GeometryX representing the limit of this geometry.
	Boundary() GeometryX

	// Convex hull returns a GeometryX that represents the smallest convex set
	// that contains this geometry.
	ConvexHull() GeometryX

	// convexHullPointset returns the list of points that must be considered
	// when finding the convex hull.
	convexHullPointSet() []XY

	// TransformXY transforms this GeometryX into another geometry according the
	// mapping provided by the XY function. Some classes of mappings (such as
	// affine transformations) will preserve the validity this GeometryX in the
	// transformed GeometryX, in which case no error will be returned. Other
	// types of transformations may result in a validation error if their
	// mapping results in an invalid GeometryX.
	TransformXY(func(XY) XY, ...ConstructorOption) (GeometryX, error)

	// EqualsExact checks if this geometry is equal to another geometry from a
	// structural pointwise equality perspective. Geometries that are
	// structurally equal are defined by exactly same control points in the
	// same order. Note that even if two geometries are spatially equal (i.e.
	// represent the same point set), they may not be defined by exactly the
	// same way. Ordering differences and numeric tolerances can be accounted
	// for using options.
	EqualsExact(GeometryX, ...EqualsExactOption) bool

	// IsValid returns if the current geometry is valid. It is useful to use when
	// validation is disabled at constructing, for example, json.Unmarshal
	IsValid() bool

	json.Marshaler
}
