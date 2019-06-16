package simplefeatures

// MultiPoint is a 0-dimensional geometric collection of points. The points are
// not connected or ordered.
type MultiPoint struct {
	pts []Point
}

func NewMultiPoint(pts []Point) MultiPoint {
	return MultiPoint{pts}
}

func NewMultiPointFromCoords(coords []OptionalCoordinates) (MultiPoint, error) {
	var pts []Point
	for _, c := range coords {
		if c.Empty {
			continue
		}
		pt, err := NewPointFromCoords(c.Value)
		if err != nil {
			return MultiPoint{}, err
		}
		pts = append(pts, pt)
	}
	return MultiPoint{pts}, nil
}

func (m MultiPoint) AsText() []byte {
	return m.AppendWKT(nil)
}

func (m MultiPoint) AppendWKT(dst []byte) []byte {
	dst = append(dst, []byte("MULTIPOINT")...)
	if len(m.pts) == 0 {
		return append(dst, []byte(" EMPTY")...)
	}
	dst = append(dst, '(')
	for i, pt := range m.pts {
		dst = pt.appendWKTBody(dst)
		if i != len(m.pts)-1 {
			dst = append(dst, ',')
		}
	}
	return append(dst, ')')
}

func (m MultiPoint) IsSimple() bool {
	panic("not implemented")
}

func (m MultiPoint) Intersection(g Geometry) Geometry {
	return intersection(m, g)
}

func (m MultiPoint) IsEmpty() bool {
	return len(m.pts) == 0
}

func (m MultiPoint) Dimension() int {
	return 0
}

func (m MultiPoint) Equals(other Geometry) bool {
	return equals(m, other)
}
