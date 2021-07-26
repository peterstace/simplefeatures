package geom

// Coordinates represents a point location. Coordinates values may be
// constructed manually using the type definition directly. Alternatively, one
// of the New(XYZM)Coordinates constructor functions can be used.
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

// NewXYCoordinates constructs a new set of coordinates of type XY.
func NewXYCoordinates(x, y float64) Coordinates {
	return Coordinates{
		Type: DimXY,
		XY:   XY{x, y},
	}
}

// NewXYZCoordinates constructs a new set of coordinates of type XYZ.
func NewXYZCoordinates(x, y, z float64) Coordinates {
	return Coordinates{
		Type: DimXYZ,
		XY:   XY{x, y},
		Z:    z,
	}
}

// NewXYMCoordinates constructs a new set of coordinates of type XYM.
func NewXYMCoordinates(x, y, m float64) Coordinates {
	return Coordinates{
		Type: DimXYM,
		XY:   XY{x, y},
		M:    m,
	}
}

// NewXYZMCoordinates constructs a new set of coordinates of type XYZM.
func NewXYZMCoordinates(x, y, z, m float64) Coordinates {
	return Coordinates{
		Type: DimXYZM,
		XY:   XY{x, y},
		Z:    z,
		M:    m,
	}
}
