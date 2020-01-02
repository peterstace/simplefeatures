package geom

import (
	"bytes"
	"database/sql/driver"
	"io"
	"unsafe"
)

// MultiLineString is a multicurve whose elements are LineStrings.
//
// Its assertions are:
//
// 1. It must be made of up zero or more valid LineStrings.
type MultiLineString struct {
	lines []LineString
}

// NewMultiLineString creates a MultiLineString from its constintuent
// LineStrings.
func NewMultiLineString(lines []LineString, opts ...ConstructorOption) MultiLineString {
	return MultiLineString{lines}
}

// NewMultiLineStringC creates a MultiLineString from its coordinates. The
// first dimension of the coordinates slice indicates the LineString, and the
// second dimension indicates the Coordinate within a LineString.
func NewMultiLineStringC(coords [][]Coordinates, opts ...ConstructorOption) (MultiLineString, error) {
	var lines []LineString
	for _, c := range coords {
		if len(c) == 0 {
			continue
		}
		line, err := NewLineStringC(c, opts...)
		if err != nil {
			return MultiLineString{}, err
		}
		lines = append(lines, line)
	}
	return MultiLineString{lines}, nil
}

// NewMultiLineStringXY creates a MultiLineString from its XYs. The
// first dimension of the XYs slice indicates the LineString, and the
// second dimension indicates the XY within a LineString.
func NewMultiLineStringXY(pts [][]XY, opts ...ConstructorOption) (MultiLineString, error) {
	return NewMultiLineStringC(twoDimXYToCoords(pts))
}

// AsGeometry converts this MultiLineString into a Geometry.
func (m MultiLineString) AsGeometry() Geometry {
	return Geometry{multiLineStringTag, unsafe.Pointer(&m)}
}

// NumLineStrings gives the number of LineString elements in the
// MultiLineString.
func (m MultiLineString) NumLineStrings() int {
	return len(m.lines)
}

// LineStringN gives the nth (zero indexed) LineString element.
func (m MultiLineString) LineStringN(n int) LineString {
	return m.lines[n]
}

func (m MultiLineString) AsText() string {
	return string(m.AppendWKT(nil))
}

func (m MultiLineString) AppendWKT(dst []byte) []byte {
	dst = append(dst, []byte("MULTILINESTRING")...)
	if len(m.lines) == 0 {
		return append(dst, []byte(" EMPTY")...)
	}
	dst = append(dst, '(')
	for i, line := range m.lines {
		dst = line.appendWKTBody(dst)
		if i != len(m.lines)-1 {
			dst = append(dst, ',')
		}
	}
	return append(dst, ')')
}

// IsSimple returns true iff the following conditions hold:
//
// 1. Each element (a LineString) is simple.
//
// 2. The intersection between any two elements occurs at points that are on
// the boundaries of both elements.
func (m MultiLineString) IsSimple() bool {
	for _, ls := range m.lines {
		if !ls.IsSimple() {
			return false
		}
	}
	for i := 0; i < len(m.lines); i++ {
		for j := i + 1; j < len(m.lines); j++ {
			inter := mustIntersection(m.lines[i], m.lines[j])
			bound := mustIntersection(m.lines[i].Boundary(), m.lines[j].Boundary())
			eq, err := inter.Equals(mustIntersection(inter, bound))
			if err != nil {
				panic(err) // Equals is implemented for all of the required types here.
			}
			if !eq {
				return false
			}
		}
	}
	return true
}

func (m MultiLineString) Intersection(g GeometryX) (GeometryX, error) {
	return intersection(m, g)
}

func (m MultiLineString) Intersects(g GeometryX) bool {
	return hasIntersection(m, g)
}

func (m MultiLineString) IsEmpty() bool {
	return len(m.lines) == 0
}

func (m MultiLineString) Dimension() int {
	return 1
}

func (m MultiLineString) Equals(other GeometryX) (bool, error) {
	return equals(m, other)
}

func (m MultiLineString) Envelope() (Envelope, bool) {
	if len(m.lines) == 0 {
		return Envelope{}, false
	}
	env := mustEnvelope(m.lines[0])
	for _, line := range m.lines[1:] {
		e := mustEnvelope(line)
		env = env.ExpandToIncludeEnvelope(e)
	}
	return env, true
}
func (m MultiLineString) Boundary() GeometryX {
	if m.IsEmpty() {
		// Postgis behaviour (but any other empty set would be ok).
		return NewMultiLineString(nil)
	}

	counts := make(map[XY]int)
	var uniqueEndpoints []Point
	for _, ls := range m.lines {
		if ls.IsClosed() {
			continue
		}
		for _, pt := range [2]Point{
			NewPointC(ls.lines[0].a),
			NewPointC(ls.lines[len(ls.lines)-1].b),
		} {
			_, seen := counts[pt.coords.XY]
			if !seen {
				uniqueEndpoints = append(uniqueEndpoints, pt)
			}
			counts[pt.coords.XY]++
		}
	}

	var bound []Point
	for _, pt := range uniqueEndpoints {
		if counts[pt.coords.XY]%2 == 1 {
			bound = append(bound, pt)
		}
	}
	return NewMultiPoint(bound)
}

func (m MultiLineString) Value() (driver.Value, error) {
	var buf bytes.Buffer
	err := m.AsBinary(&buf)
	return buf.Bytes(), err
}

func (m MultiLineString) AsBinary(w io.Writer) error {
	marsh := newWKBMarshaller(w)
	marsh.writeByteOrder()
	marsh.writeGeomType(wkbGeomTypeMultiLineString)
	n := m.NumLineStrings()
	marsh.writeCount(n)
	for i := 0; i < n; i++ {
		ls := m.LineStringN(i)
		marsh.setErr(ls.AsBinary(w))
	}
	return marsh.err
}

func (m MultiLineString) ConvexHull() GeometryX {
	return convexHull(m)
}

func (m MultiLineString) convexHullPointSet() []XY {
	var points []XY
	n := m.NumLineStrings()
	for i := 0; i < n; i++ {
		line := m.LineStringN(i)
		m := line.NumPoints()
		for j := 0; j < m; j++ {
			points = append(points, line.PointN(j).XY())
		}
	}
	return points
}

func (m MultiLineString) MarshalJSON() ([]byte, error) {
	return marshalGeoJSON("MultiLineString", m.Coordinates())
}

// Coordinates returns the coordinates of each constintuent LineString in the
// MultiLineString.
func (m MultiLineString) Coordinates() [][]Coordinates {
	numLines := m.NumLineStrings()
	coords := make([][]Coordinates, numLines)
	for i := 0; i < numLines; i++ {
		numPts := m.LineStringN(i).NumPoints()
		coords[i] = make([]Coordinates, numPts)
		for j := 0; j < numPts; j++ {
			coords[i][j] = m.LineStringN(i).PointN(j).Coordinates()
		}
	}
	return coords
}

// TransformXY transforms this MultiLineString into another MultiLineString according to fn.
func (m MultiLineString) TransformXY(fn func(XY) XY, opts ...ConstructorOption) (GeometryX, error) {
	coords := m.Coordinates()
	transform2dCoords(coords, fn)
	return NewMultiLineStringC(coords, opts...)
}

// EqualsExact checks if this MultiLineString is exactly equal to another MultiLineString.
func (m MultiLineString) EqualsExact(other GeometryX, opts ...EqualsExactOption) bool {
	o, ok := other.(MultiLineString)
	return ok && multiLineStringExactEqual(m, o, opts)
}

// IsValid checks if this MultiLineString is valid
func (m MultiLineString) IsValid() bool {
	_, err := NewMultiLineStringC(m.Coordinates())
	return err == nil
}

// Length gives the sum of the lengths of the constituent members of the multi
// line string.
func (m MultiLineString) Length() float64 {
	var sum float64
	for _, ln := range m.lines {
		sum += ln.Length()
	}
	return sum
}
