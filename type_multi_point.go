package simplefeatures

import (
	"database/sql/driver"
)

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
	return MultiPoint{pts}
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

// NumPoints gives the number of element points making up the MultiPoint.
func (m MultiPoint) NumPoints() int {
	return len(m.pts)
}

// PointN gives the nth (zero indexed) Point.
func (m MultiPoint) PointN(n int) Point {
	return m.pts[n]
}

func (m MultiPoint) AsText() string {
	return string(m.AppendWKT(nil))
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

// IsSimple returns true iff no two of its points are equal.
func (m MultiPoint) IsSimple() bool {
	seen := make(map[xyHash]bool)
	for _, p := range m.pts {
		h := p.coords.XY.hash()
		if seen[h] {
			return false
		}
		seen[h] = true
	}
	return true
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

func (m MultiPoint) Value() (driver.Value, error) {
	return m.AsText(), nil
}
