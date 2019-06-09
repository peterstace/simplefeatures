package simplefeatures

// MultiPoint is a 0-dimensional geometric collection of points. The points are
// not connected or ordered.
type MultiPoint struct {
	pts []Point
}

func NewMultiPoint(pts []Point) (MultiPoint, error) {
	// TODO: error checking
	return MultiPoint{pts}, nil
}

func NewMultiPointFromCoords(coords []OptionalCoordinates) (MultiPoint, error) {
	pts := make([]Point, len(coords))
	for i, c := range coords {
		pt, err := NewPointFromOptionalCoords(c)
		if err != nil {
			return MultiPoint{}, err
		}
		pts[i] = pt
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
