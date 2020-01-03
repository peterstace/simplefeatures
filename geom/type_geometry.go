package geom

// GeometryX is the most general type of geometry supported, and exposes common
// behaviour. All geometry types implement this interface.
type GeometryX interface {
	// Equals checks if this geometry is equal to another geometry. Two
	// geometries are equal if they contain exactly the same points.
	//
	// It is not implemented for all possible pairs of geometries, and returns
	// an error in those cases.
	Equals(GeometryX) (bool, error)
}
