package simplefeatures

type Geometry interface {
	// AsText returns the WKT representation of the geometry.
	AsText() []byte

	// AppendWKT appends the WKT representation of the geometry to dst and
	// returns the resultant slice.
	AppendWKT(dst []byte) []byte

	// IsSimple returns true iff the geometry doesn't contain any anomalous
	// geometry points such as self intersection or self tangency. The precise
	// condition will differ for each type of geometry.
	IsSimple() bool

	// Intersection returns a geometric object that represents the point set
	// intersection of this geometry with another geometry.
	Intersection(Geometry) Geometry

	// IsEmpty returns true if this object an empty geometry.
	IsEmpty() bool

	// Dimension returns the dimension of the geometry. This is 0 for empty
	// geometries, 0 for points, 1 for curves, and 2 for surfaces. For mixed
	// geometries, it is the maximum dimension over the collection.
	Dimension() int

	// Equals checks if this geometry is equal to another geometrie. Two
	// geometries are equal if they contain exactly the same points.
	Equals(Geometry) bool

	// FiniteNumberOfPoints checks if the geometry represents a finite number
	// of points. If it does, then the returned int is the number of distinct
	// points represented.
	FiniteNumberOfPoints() (int, bool)
}
