package simplefeatures

// Point is a 0-dimensional geometry, and represents a single location in a
// coordinate space.
type Point struct {
	x, y  float64
	empty bool
}

// NewPoint creates a new point.
func NewPoint(x, y float64) (Point, error) {
	// TODO: Inf and NaN not allowed.
	return Point{x, y, false}, nil
}

// NewEmptyPoint creates an empty point.
func NewEmptyPoint() Point {
	return Point{empty: true}
}

// NewPointFromCoords creates a new point gives its coordinates.
func NewPointFromCoords(c Coordinates) (Point, error) {
	return NewPoint(c.X, c.Y)
}

// NewPointFromOptionalCoords creates a new point given its coordinates (which
// may be empty).
func NewPointFromOptionalCoords(c OptionalCoordinates) (Point, error) {
	if c.Empty {
		return NewEmptyPoint(), nil
	}
	return NewPoint(c.Value.X, c.Value.Y)
}
