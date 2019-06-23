package simplefeatures

// MultiPoint is a 0-dimensional geometric collection of points. The points are
// not connected or ordered.
type MultiPoint struct {
	pts []Point
}

func NewMultiPoint(pts []Point) MultiPoint {
	// Deduplicate
	ptSet := make(map[xyString]Point)
	for _, p := range pts {
		ptSet[xykey(p.coords.XY)] = p
	}
	ptSlice := make([]Point, 0, len(ptSet))
	for _, p := range ptSet {
		ptSlice = append(ptSlice, p)
	}
	return MultiPoint{ptSlice}
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
	return NewMultiPoint(pts), nil
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

// This could just be "return true" because we de-duplicate points.
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

func (m MultiPoint) Envelope() (Envelope, bool) {
	if len(m.pts) == 0 {
		return Envelope{}, false
	}
	env := NewEnvelope(m.pts[0].coords.XY)
	for _, pt := range m.pts[1:] {
		env = env.Extend(pt.coords.XY)
	}
	return env, true
}
