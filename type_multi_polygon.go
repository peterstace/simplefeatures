package simplefeatures

import "errors"

// MultiPolygon is a multi surface whose elements are polygons.
//
// Its assertions are:
//
// 1. It must be made up of zero or more valid Polygons.
//
// 2. The interiors of any two polygons must not intersect. TODO: this is not
// yet implemented.
//
// 3. The boundaries of any two polygons may touch only at a finite number of
// points. TODO: this is not yet implemented.
type MultiPolygon struct {
	polys []Polygon
}

func NewMultiPolygon(polys []Polygon) (MultiPolygon, error) {
	// TODO: implement assertions

	for i := 0; i < len(polys); i++ {
		for j := i + 1; j < len(polys); j++ {
			bound1 := polys[i].Boundary()
			bound2 := polys[j].Boundary()
			inter := bound1.Intersection(bound2)
			if inter.Dimension() > 0 {
				return MultiPolygon{}, errors.New("the boundaries of the polygon elements of multipolygons must only intersect at points")
			}
		}
	}

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
	return len(m.polys) == 0
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

func (m MultiPolygon) Boundary() Geometry {
	if m.IsEmpty() {
		return m
	}
	var bounds []LineString
	for _, poly := range m.polys {
		bounds = append(bounds, poly.outer.ls)
		for _, inner := range poly.holes {
			bounds = append(bounds, inner.ls)
		}
	}
	return NewMultiLineString(bounds)
}
