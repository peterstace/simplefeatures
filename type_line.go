package simplefeatures

// Line is a LineString with exactly two points.
type Line struct {
	ls LineString
}

// NewLine creates a line given the two points that define it.
func NewLine(p1, p2 Point) (Line, error) {
	ls, err := NewLineString([]Point{p1, p2})
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
