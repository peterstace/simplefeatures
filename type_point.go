package simplefeatures

// Point is a 0-dimensional geometry, and represents a single location in a
// coordinate space.
//
// There aren't any assertions about what constitutes a valid point, other than
// that it must be able to be represented by an XY.
type Point struct {
	coords Coordinates
}

// NewPoint creates a new point from an XY.
func NewPoint(xy XY) Point {
	return NewPointXY(xy.X, xy.Y)
}

// NewPointXY creates a new point from an X and a Y.
func NewPointXY(x, y Scalar) Point {
	return NewPointFromCoords(Coordinates{XY{x, y}})
}

// NewPointFromCoords creates a new point gives its coordinates.
func NewPointFromCoords(c Coordinates) Point {
	return Point{coords: c}
}

func (p Point) AsText() []byte {
	return p.AppendWKT(nil)
}

func (p Point) AppendWKT(dst []byte) []byte {
	dst = append(dst, []byte("POINT")...)
	return p.appendWKTBody(dst)
}

func (p Point) appendWKTBody(dst []byte) []byte {
	dst = append(dst, '(')
	dst = p.coords.X.appendAsFloat(dst)
	dst = append(dst, ' ')
	dst = p.coords.Y.appendAsFloat(dst)
	return append(dst, ')')
}

func (p Point) IsSimple() bool {
	panic("not implemented")
}

func (p Point) Intersection(g Geometry) Geometry {
	return intersection(p, g)
}

func (p Point) IsEmpty() bool {
	return false
}

func (p Point) Dimension() int {
	return 0
}

func (p Point) Equals(other Geometry) bool {
	return equals(p, other)
}

func (p Point) Envelope() (Envelope, bool) {
	return NewEnvelope(p.coords.XY), true
}

func (p Point) Boundary() Geometry {
	return NewGeometryCollection(nil)
}
