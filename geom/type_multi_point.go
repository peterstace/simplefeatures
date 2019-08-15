package geom

import (
	"database/sql/driver"
	"io"
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

// NewMultiPointOC creates a new MultiPoint consisting of a Point for each
// non-empty OptionalCoordinate.
func NewMultiPointOC(coords []OptionalCoordinates) MultiPoint {
	var pts []Point
	for _, c := range coords {
		if c.Empty {
			continue
		}
		pt := NewPointC(c.Value)
		pts = append(pts, pt)
	}
	return NewMultiPoint(pts)
}

// NewMultiPointC creates a new MultiPoint consisting of a point for each coordinate.
func NewMultiPointC(coords []Coordinates) MultiPoint {
	var pts []Point
	for _, c := range coords {
		pt := NewPointC(c)
		pts = append(pts, pt)
	}
	return NewMultiPoint(pts)
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

func (m MultiPoint) AsBinary(w io.Writer) error {
	marsh := newWKBMarshaller(w)
	marsh.writeByteOrder()
	marsh.writeGeomType(wkbGeomTypeMultiPoint)
	n := m.NumPoints()
	marsh.writeCount(n)
	for i := 0; i < n; i++ {
		pt := m.PointN(i)
		marsh.setErr(pt.AsBinary(w))
	}
	return marsh.err
}

// ConvexHull finds the convex hull of the set of points. This may either be
// the empty set, a single point, a line, or a polygon.
func (m MultiPoint) ConvexHull() Geometry {
	return convexHull(m)
}

func (m MultiPoint) convexHullPointSet() []XY {
	n := m.NumPoints()
	points := make([]XY, n)
	for i := 0; i < n; i++ {
		points[i] = m.PointN(i).XY()
	}
	return points
}

func (m MultiPoint) MarshalJSON() ([]byte, error) {
	return marshalGeoJSON("MultiPoint", m.Coordinates())
}

// Coordinates returns the coordinates of the points represented by the
// MultiPoint.
func (m MultiPoint) Coordinates() []Coordinates {
	coords := make([]Coordinates, len(m.pts))
	for i := range coords {
		coords[i] = m.pts[i].Coordinates()
	}
	return coords
}
