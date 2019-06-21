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
	var polys []Polygon
	for _, c := range coords {
		if len(c) == 0 {
			continue
		}
		poly, err := NewPolygonFromCoords(c)
		if err != nil {
			return MultiPolygon{}, err
		}
		polys = append(polys, poly)
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

func (m MultiPolygon) Intersection(g Geometry) Geometry {
	return intersection(m, g)
}

func (m MultiPolygon) IsEmpty() bool {
	for _, p := range m.polys {
		if !p.IsEmpty() {
			return false
		}
	}
	return true
}

func (m MultiPolygon) Dimension() int {
	if m.IsEmpty() {
		return 0
	}
	return 2
}

func (m MultiPolygon) Equals(other Geometry) bool {
	return equals(m, other)
}

func (m MultiPolygon) Envelope() (Envelope, bool) {
	if len(m.polys) == 0 {
		return Envelope{}, false
	}
	env := mustEnvelope(m.polys[0])
	for _, poly := range m.polys[1:] {
		env = env.Union(mustEnvelope(poly))
	}
	return env, true
}
