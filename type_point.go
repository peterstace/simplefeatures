package gatig

// Point is a 0-dimensional geometry, and represents a single location in a
// coordinate space.
type Point struct {
	x, y  float64
	empty bool
}

// NewPoint creates a new point.
func NewPoint(x, y float64) Point {
	return Point{x, y, false}
}

// NewEmptyPoint creates an empty point.
func NewEmptyPoint() Point {
	return Point{empty: true}
}
