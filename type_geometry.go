package simplefeatures

import "io"

// Geometry is the most general type of geometry supported, and exposes common
// behaviour. All geometry types implement this interface.
type Geometry interface {
	// AsText returns the WKT representation of the geometry.
	AsText() string

	// AppendWKT appends the WKT representation of the geometry to dst and
	// returns the resultant slice.
	AppendWKT(dst []byte) []byte

	// AsBinary writes the WKB (Well Known Binary) representation of the
	// geometry to the writer.
	AsBinary(w io.Writer) error

	// Intersection returns a geometric object that represents the point set
	// intersection of this geometry with another geometry.
	Intersection(Geometry) Geometry

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

	// Equals checks if this geometry is equal to another geometrie. Two
	// geometries are equal if they contain exactly the same points.
	Equals(Geometry) bool

	// Boundary returns the Geometry representing the limit of this geometry.
	Boundary() Geometry

	// Convex hull returns a Geometry that represents the smallest convex set
	// that contains this geometry.
	ConvexHull() Geometry

	// convexHullPointset returns the list of points that must be considered
	// when finding the convex hull.
	convexHullPointSet() []XY
}

// HeterogenousGeometry are geometries that contain a single element, or
// elements all of the same type. Specifically, all geometries are heterogenous
// except for GeometryCollection.
type HeterogenousGeometry interface {
	Geometry

	// IsSimple returns true iff the geometry doesn't contain any anomalous
	// geometry points such as self intersection or self tangency. The precise
	// condition will differ for each type of geometry.
	IsSimple() bool
}
