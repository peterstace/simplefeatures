package geom

// Coordinates represents a point location.
type Coordinates struct {
	// XY represents the XY position of the point location.
	XY

	// Z represents the height of the location. Its value is zero
	// for non-3D coordinate types.
	Z float64

	// M represents a user defined measure associated with the
	// location. Its value is zero for non-measure coordinate
	// types.
	M float64

	// Type indicates the coordinates type, and therefore whether
	// or not Z and M are populated.
	Type CoordinatesType
}
