package simplefeatures

// MultiPoint is a 0-dimensional geometric collection of points. The points are
// not connected or ordered.
//
// Its assertions are:
//
// 1. It must be made up of 0 or more valid Points.
type MultiPoint struct {
	pts []Point
}

func NewMultiPoint(pts []Point) MultiPoint {
	// Deduplicate
	dedupe := make([]Point, 0, len(pts))
	seen := make(map[xyHash]bool)
	for _, pt := range pts {
		h := pt.coords.hash()
		if !seen[h] {
			seen[h] = true
			dedupe = append(dedupe, pt)
		}
	}
	return MultiPoint{dedupe}
}

func NewMultiPointFromCoords(coords []OptionalCoordinates) (MultiPoint, error) {
	var pts []Point
	for _, c := range coords {
		if c.Empty {
			continue
		}
		pt := NewPointFromCoords(c.Value)
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

func (m MultiPoint) Boundary() Geometry {
	// This is a little bit more complicated than it really has to be (it just
	// has to always return an empty set). However, this is the behavour of
	// Postgis.
	if m.IsEmpty() {
		return m
	}
	return NewGeometryCollection(nil)
}
