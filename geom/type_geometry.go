package geom

// GeometryX is the most general type of geometry supported, and exposes common
// behaviour. All geometry types implement this interface.
type GeometryX interface {
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
}
