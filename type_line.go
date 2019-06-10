package simplefeatures

// Line is a LineString with exactly two points.
type Line struct {
	ls LineString
}

// NewLine creates a line segment given the coordinates of its two endpoints.
func NewLine(c1, c2 Coordinates) (Line, error) {
	ls, err := NewLineString([]Coordinates{c1, c2})
	return Line{ls}, err
}

var _ Geometry = Line{}

func (n Line) AsText() []byte {
	return n.ls.AsText()
}

func (n Line) AppendWKT(dst []byte) []byte {
	return n.ls.AppendWKT(dst)
}

func (n Line) IsSimple() bool {
	return n.ls.IsSimple()
}
