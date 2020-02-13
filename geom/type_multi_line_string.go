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
			// Ignore any intersections if the lines are *exactly* the same
			// (ignoring order). This is to match PostGIS and libgeos
			// behaviour. The OGC spec is ambiguous around this case, so it's
			// just easier to follow other implementations for better
			// interoperability.
			if m.lines[i].EqualsExact(m.lines[j].AsGeometry(), IgnoreOrder) {
				continue
			}

			inter := mustIntersection(
				m.lines[i].AsGeometry(),
				m.lines[j].AsGeometry(),
			)
			if inter.IsEmpty() {
				continue
			}
			bound := mustIntersection(
				m.lines[i].Boundary().AsGeometry(),
				m.lines[j].Boundary().AsGeometry(),
			)
			if !inter.EqualsExact(mustIntersection(inter, bound), IgnoreOrder) {
				return false
			}
		}
	}
	return true
}

func (m MultiLineString) Intersection(g Geometry) (Geometry, error) {
	return intersection(m.AsGeometry(), g)
}

func (m MultiLineString) Intersects(g Geometry) bool {
	return hasIntersection(m.AsGeometry(), g)
}

func (m MultiLineString) IsEmpty() bool {
	return len(m.lines) == 0
}

func (m MultiLineString) Equals(other Geometry) (bool, error) {
	return equals(m.AsGeometry(), other)
}

func (m MultiLineString) Envelope() (Envelope, bool) {
	if len(m.lines) == 0 {
		return Envelope{}, false
	}
	env := mustEnv(m.lines[0].Envelope())
	for _, line := range m.lines[1:] {
		e := mustEnv(line.Envelope())
		env = env.ExpandToIncludeEnvelope(e)
	}
	return env, true
}
func (m MultiLineString) Boundary() MultiPoint {
	counts := make(map[XY]int)
	var uniqueEndpoints []Point
	for _, ls := range m.lines {
		if ls.IsClosed() {
			continue
		}
		for _, pt := range [2]Point{
			ls.StartPoint(),
			ls.EndPoint(),
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

func (m MultiLineString) ConvexHull() Geometry {
	return convexHull(m.AsGeometry())
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
func (m MultiLineString) TransformXY(fn func(XY) XY, opts ...ConstructorOption) (Geometry, error) {
	coords := m.Coordinates()
	transform2dCoords(coords, fn)
	mls, err := NewMultiLineStringC(coords, opts...)
	return mls.AsGeometry(), err
}

// EqualsExact checks if this MultiLineString is exactly equal to another MultiLineString.
func (m MultiLineString) EqualsExact(other Geometry, opts ...EqualsExactOption) bool {
	return other.IsMultiLineString() &&
		multiLineStringExactEqual(m, other.AsMultiLineString(), opts)
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

// Reverse in the case of MultiLineString outputs each component line string in their
// original order, each individually reversed.
func (m MultiLineString) Reverse() MultiLineString {
	linestrings := make([]LineString, len(m.lines))
	// Form the reversed slice.
	for i := 0; i < len(m.lines); i++ {
		linestrings[i] = m.lines[i].Reverse()
	}
	return NewMultiLineString(linestrings)
}
