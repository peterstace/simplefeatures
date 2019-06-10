package simplefeatures

// MultiPolygon is a multi surface whose elements are polygons.
type MultiPolygon struct {
	polys []Polygon
}

func NewMultiPolygon(polys []Polygon) (MultiPolygon, error) {
	// TODO: The interiors of 2 polygons must not intersect.
	// TODO: The boundaries of 2 polygons may touch only at a finite number of points.
	return MultiPolygon{polys}, nil
}

func NewMultiPolygonFromCoords(coords [][][]Coordinates) (MultiPolygon, error) {
	polys := make([]Polygon, len(coords))
	for i, c := range coords {
		poly, err := NewPolygonFromCoords(c)
		if err != nil {
			return MultiPolygon{}, err
		}
		polys[i] = poly
	}
	return NewMultiPolygon(polys)
}

func (m MultiPolygon) AsText() []byte {
	return m.AppendWKT(nil)
}

func (m MultiPolygon) AppendWKT(dst []byte) []byte {
	dst = append(dst, []byte("MULTIPOLYGON")...)
	if len(m.polys) == 0 {
		return append(dst, []byte(" EMPTY")...)
	}
	dst = append(dst, '(')
	for i, poly := range m.polys {
		dst = poly.appendWKTBody(dst)
		if i != len(m.polys)-1 {
			dst = append(dst, ',')
		}
	}
	return append(dst, ')')
}

func (m MultiPolygon) IsSimple() bool {
	panic("not implemented")
}

func (m MultiPolygon) Intersection(Geometry) Geometry {
	panic("not implemented")
}
